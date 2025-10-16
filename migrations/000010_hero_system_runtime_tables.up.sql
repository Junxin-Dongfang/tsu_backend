-- ============================================
-- 英雄系统运行时表（game_runtime schema）
-- ============================================

-- 1. 英雄职业历史表
CREATE TABLE game_runtime.hero_class_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES game_config.classes(id),
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acquisition_type VARCHAR(20) NOT NULL DEFAULT 'initial',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_hero_class_history_hero ON game_runtime.hero_class_history(hero_id);
CREATE INDEX idx_hero_class_history_current ON game_runtime.hero_class_history(hero_id, is_current) 
WHERE is_current = TRUE;
CREATE INDEX idx_hero_class_history_type ON game_runtime.hero_class_history(hero_id, acquisition_type);

CREATE TRIGGER update_hero_class_history_updated_at 
    BEFORE UPDATE ON game_runtime.hero_class_history 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_runtime.hero_class_history IS '英雄职业历史表';
COMMENT ON COLUMN game_runtime.hero_class_history.acquisition_type 
IS '获得方式：initial(初始)、advancement(进阶-保留技能池)、transfer(转职-放弃旧技能池)';

-- 2. 属性操作历史表（栈式回退）
CREATE TABLE game_runtime.hero_attribute_operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    attribute_code VARCHAR(32) NOT NULL,
    points_added INTEGER NOT NULL,
    xp_spent INTEGER NOT NULL,
    value_before INTEGER NOT NULL,
    value_after INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rollback_deadline TIMESTAMPTZ NOT NULL,
    rolled_back_at TIMESTAMPTZ
);

CREATE INDEX idx_hero_attr_ops_hero ON game_runtime.hero_attribute_operations(hero_id);
CREATE INDEX idx_hero_attr_ops_rollback ON game_runtime.hero_attribute_operations(hero_id, attribute_code, rolled_back_at)
WHERE rolled_back_at IS NULL;
CREATE INDEX idx_hero_attr_ops_cleanup ON game_runtime.hero_attribute_operations(rolled_back_at, rollback_deadline)
WHERE rolled_back_at IS NOT NULL;

COMMENT ON TABLE game_runtime.hero_attribute_operations IS '属性操作历史表（支持栈式回退）';
COMMENT ON COLUMN game_runtime.hero_attribute_operations.rollback_deadline IS '回退截止时间（创建时间+1小时）';
COMMENT ON COLUMN game_runtime.hero_attribute_operations.rolled_back_at IS '实际回退时间（NULL表示未回退）';

-- 3. 技能操作历史表（栈式回退）
CREATE TABLE game_runtime.hero_skill_operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    hero_skill_id UUID NOT NULL REFERENCES game_runtime.hero_skills(id) ON DELETE CASCADE,
    levels_added INTEGER NOT NULL,
    xp_spent INTEGER NOT NULL,
    gold_spent INTEGER NOT NULL DEFAULT 0,
    materials_spent JSONB DEFAULT '[]',
    level_before INTEGER NOT NULL,
    level_after INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rollback_deadline TIMESTAMPTZ NOT NULL,
    rolled_back_at TIMESTAMPTZ
);

CREATE INDEX idx_hero_skill_ops_skill ON game_runtime.hero_skill_operations(hero_skill_id);
CREATE INDEX idx_hero_skill_ops_rollback ON game_runtime.hero_skill_operations(hero_skill_id, rolled_back_at)
WHERE rolled_back_at IS NULL;
CREATE INDEX idx_hero_skill_ops_cleanup ON game_runtime.hero_skill_operations(rolled_back_at, rollback_deadline)
WHERE rolled_back_at IS NOT NULL;

COMMENT ON TABLE game_runtime.hero_skill_operations IS '技能操作历史表（支持栈式回退）';
COMMENT ON COLUMN game_runtime.hero_skill_operations.rollback_deadline IS '回退截止时间（创建时间+1小时）';
COMMENT ON COLUMN game_runtime.hero_skill_operations.rolled_back_at IS '实际回退时间（NULL表示未回退）';

-- 4. 修改 hero_skills 表
ALTER TABLE game_runtime.hero_skills
ADD COLUMN IF NOT EXISTS first_learned_at TIMESTAMPTZ DEFAULT NOW();

COMMENT ON COLUMN game_runtime.hero_skills.first_learned_at IS '首次学习时间';
COMMENT ON COLUMN game_runtime.hero_skills.learned_method IS '技能来源：class_unlock(职业解锁)、equipment(装备提供)、quest(任务奖励)、manual(主动学习)等';

-- 5. 创建英雄属性计算视图
CREATE OR REPLACE VIEW game_runtime.hero_computed_attributes AS
SELECT 
  h.id as hero_id,
  hat.id as attribute_type_id,
  hat.attribute_code,
  hat.attribute_name,
  -- 基础加点值
  COALESCE((h.allocated_attributes->hat.attribute_code->>'value')::INTEGER, 1) as base_value,
  -- 职业加成（仅当前职业）
  COALESCE(cab.base_bonus_value, 0) + 
    CASE WHEN cab.bonus_per_level THEN COALESCE(cab.per_level_bonus_value * h.current_level, 0) 
    ELSE 0 END as class_bonus,
  -- 最终值（未来可加入技能、装备加成）
  COALESCE((h.allocated_attributes->hat.attribute_code->>'value')::INTEGER, 1) +
  COALESCE(cab.base_bonus_value, 0) + 
    CASE WHEN cab.bonus_per_level THEN COALESCE(cab.per_level_bonus_value * h.current_level, 0) 
    ELSE 0 END as final_value
FROM game_runtime.heroes h
CROSS JOIN game_config.hero_attribute_type hat
LEFT JOIN game_runtime.hero_class_history hch ON hch.hero_id = h.id AND hch.is_current = TRUE
LEFT JOIN game_config.class_attribute_bonuses cab ON cab.class_id = hch.class_id AND cab.attribute_id = hat.id
WHERE h.deleted_at IS NULL AND hat.is_active = TRUE AND hat.category = 'basic';

COMMENT ON VIEW game_runtime.hero_computed_attributes IS '英雄属性计算视图（仅包含当前职业加成）';

