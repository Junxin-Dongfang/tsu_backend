-- =============================================================================
-- Create Skills Advanced System
-- 技能高级系统：动作、效果、Buff、战斗逻辑等
-- 依赖：000006_create_skills_base
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 新增枚举类型
-- --------------------------------------------------------------------------------

-- 修正类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'modifier_type_enum') THEN
        CREATE TYPE modifier_type_enum AS ENUM (
            'advantage',      -- 优势
            'disadvantage',   -- 劣势
            'bonus_flat',     -- 固定加值
            'bonus_percent',  -- 百分比加成
            'penalty_flat',   -- 固定减值
            'penalty_percent' -- 百分比减值
        );
    END IF;
END $$;

- --------------------------------------------------------------------------------
-- 5. 效果定义表（核心新增）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- 效果基本信息
    effect_code VARCHAR(50) NOT NULL,
    effect_name VARCHAR(100) NOT NULL,
    effect_type VARCHAR(50) NOT NULL,  -- 'damage', 'heal', 'apply_buff', 'modify_attribute'等
    
    -- 效果参数（标准化JSON）
    parameters JSONB NOT NULL DEFAULT '{}',
    
    -- 计算公式（可选）
    calculation_formula TEXT,
    
    -- 触发条件
    trigger_condition JSONB,
    trigger_chance DECIMAL(5,2) DEFAULT 100.00 CHECK (trigger_chance >= 0 AND trigger_chance <= 100),
    
    -- 目标过滤
    target_filter JSONB,
    
    -- 特征标签（冗余字段，方便快速查询）
    feature_tags TEXT[],
    
    -- 视觉/音效配置
    visual_config JSONB,
    sound_config JSONB,
    
    -- 描述模板
    description TEXT,
    tooltip_template TEXT,
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_effects_code_unique 
    ON effects(effect_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_effects_type 
    ON effects(effect_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_effects_feature_tags 
    ON effects USING GIN(feature_tags);

CREATE TRIGGER update_effects_updated_at
    BEFORE UPDATE ON effects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE effects IS '效果定义表 - 所有游戏效果的原子单位';
COMMENT ON COLUMN effects.effect_type IS '效果类型: damage, heal, apply_buff, remove_buff, modify_attribute, summon, teleport等';
COMMENT ON COLUMN effects.parameters IS 'JSON格式的效果参数，不同effect_type有不同的参数结构';

-- --------------------------------------------------------------------------------
-- 动作类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    category_code VARCHAR(32) NOT NULL,
    category_name VARCHAR(64) NOT NULL,
    description TEXT,
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 动作类别索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_action_categories_code_unique 
    ON action_categories(category_code) WHERE deleted_at IS NULL;

-- 更新时间戳触发器
CREATE TRIGGER update_action_categories_updated_at
    BEFORE UPDATE ON action_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 8. 动作定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- 动作基本信息
    action_code VARCHAR(50) NOT NULL,
    action_name VARCHAR(100) NOT NULL,
    action_category_id UUID REFERENCES action_categories(id),
    action_type action_type_enum NOT NULL DEFAULT 'main',
    
    -- 关联技能（可选）
    related_skill_id UUID REFERENCES skills(id),
    
    -- 动作参数
    feature_tags TEXT[],
    range_config JSONB NOT NULL DEFAULT '{"range": 1}',
    target_config JSONB DEFAULT '{"type": "single", "ally": false}',
    area_config JSONB DEFAULT '{"type": "single"}',
    
    -- 消耗和限制
    action_point_cost INTEGER DEFAULT 1 CHECK (action_point_cost >= 0),
    mana_cost INTEGER DEFAULT 0 CHECK (mana_cost >= 0),
    cooldown_turns INTEGER DEFAULT 0 CHECK (cooldown_turns >= 0),
    uses_per_battle INTEGER,
    
    -- 命中率配置（新增）
    hit_rate_config JSONB DEFAULT '{
        "base_hit_rate": 75,
        "use_attacker_accuracy": true,
        "accuracy_attribute": "ACCURACY",
        "accuracy_multiplier": 1.0,
        "use_target_defense": false,
        "min_hit_rate": 5,
        "max_hit_rate": 95
    }'::jsonb,
    
    -- 使用条件
    requirements JSONB,
    start_flags TEXT[],
    
    -- 动画和视觉效果
    animation_config JSONB,
    visual_effects JSONB,
    sound_effects JSONB,
    
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_actions_code_unique 
    ON actions(action_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_category 
    ON actions(action_category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_skill 
    ON actions(related_skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_type 
    ON actions(action_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_feature_tags 
    ON actions USING GIN(feature_tags);

CREATE TRIGGER update_actions_updated_at
    BEFORE UPDATE ON actions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 9. Action与Effect关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    effect_id UUID NOT NULL REFERENCES effects(id) ON DELETE CASCADE,
    
    execution_order INTEGER DEFAULT 0,
    
    -- 参数覆盖
    parameter_overrides JSONB,
    
    -- 条件执行
    is_conditional BOOLEAN DEFAULT FALSE,
    condition_config JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(action_id, effect_id)
);

-- Action与Effect关联索引
CREATE INDEX IF NOT EXISTS idx_action_effects_action 
    ON action_effects(action_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_effects_effect 
    ON action_effects(effect_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_effects_order 
    ON action_effects(action_id, execution_order);

-- Action与Effect关联触发器
CREATE TRIGGER update_action_effects_updated_at
    BEFORE UPDATE ON action_effects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 10. 技能解锁动作关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_unlock_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    
    unlock_level INTEGER NOT NULL DEFAULT 1 CHECK (unlock_level >= 1),
    is_default BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(skill_id, action_id)
);

-- 技能解锁动作关联索引
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_skill 
    ON skill_unlock_actions(skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_action 
    ON skill_unlock_actions(action_id) WHERE deleted_at IS NULL;

-- 技能解锁动作关联触发器
CREATE TRIGGER update_skill_unlock_actions_updated_at
    BEFORE UPDATE ON skill_unlock_actions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 11. Buff定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS buffs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- Buff基本信息
    buff_code VARCHAR(50) NOT NULL,
    buff_name VARCHAR(100) NOT NULL,
    buff_type VARCHAR(50) NOT NULL DEFAULT 'buff',
    
    -- Buff分类
    category VARCHAR(50),
    feature_tags TEXT[],
    
    -- 持续时间配置
    default_duration INTEGER DEFAULT 1,
    max_duration INTEGER DEFAULT 10,
    min_duration INTEGER DEFAULT 1,
    
    -- 叠加规则
    stack_rule VARCHAR(50) DEFAULT 'no_stack',
    max_stacks INTEGER DEFAULT 1,
    
    -- 触发事件
    trigger_events TEXT[],
    
    -- 优劣势支持（新增）
    provides_advantage BOOLEAN DEFAULT FALSE,
    provides_disadvantage BOOLEAN DEFAULT FALSE,
    advantage_applies_to TEXT[],
    
    -- 视觉效果
    visual_effects JSONB,
    icon VARCHAR(256),
    color VARCHAR(16),
    
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_buffs_code_unique 
    ON buffs(buff_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_type 
    ON buffs(buff_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_category 
    ON buffs(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_feature_tags 
    ON buffs USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_buffs_trigger_events 
    ON buffs USING GIN(trigger_events);

CREATE TRIGGER update_buffs_updated_at
    BEFORE UPDATE ON buffs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 12. Buff与Effect关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS buff_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    buff_id UUID NOT NULL REFERENCES buffs(id) ON DELETE CASCADE,
    effect_id UUID NOT NULL REFERENCES effects(id) ON DELETE CASCADE,
    
    trigger_timing VARCHAR(50) NOT NULL,  -- 'on_apply', 'turn_start', 'turn_end', 'on_remove'
    execution_order INTEGER DEFAULT 0,
    
    parameter_overrides JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(buff_id, effect_id, trigger_timing)
);

-- Buff与Effect关联索引
CREATE INDEX IF NOT EXISTS idx_buff_effects_buff 
    ON buff_effects(buff_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buff_effects_effect 
    ON buff_effects(effect_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buff_effects_timing 
    ON buff_effects(trigger_timing);


-- --------------------------------------------------------------------------------
-- 动作Flag定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_flags (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    flag_code VARCHAR(50) NOT NULL,
    flag_name VARCHAR(100) NOT NULL,
    category VARCHAR(50), -- 'action_chain', 'status_check', 'modifier'

    -- 持续时间配置
    duration_type VARCHAR(20) DEFAULT 'action', -- 'action', 'turn', 'battle', 'permanent'
    default_duration VARCHAR(20) DEFAULT '1', -- 默认持续时间，可以是数字或特殊值如 "-"

    -- 自动移除条件
    auto_remove_condition VARCHAR(100), -- 'action_end', 'turn_start', 'turn_end'
    remove_on_events TEXT[], -- 触发移除的事件

    -- Flag属性
    is_visible BOOLEAN DEFAULT FALSE, -- 是否在UI中显示
    is_stackable BOOLEAN DEFAULT FALSE, -- 是否可叠加
    max_stacks INTEGER DEFAULT 1,

    -- 效果
    flag_effects JSONB, -- Flag产生的效果
    modifier_effects JSONB, -- 修饰符效果

    provides_advantage BOOLEAN DEFAULT FALSE,-- 是否提供优势
    provides_disadvantage BOOLEAN DEFAULT FALSE,-- 是否提供劣势
    advantage_applies_to TEXT[];--优劣势作用范围: hit_rate, critical_rate, saving_throw等

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, flag_code)
);

-- 动作Flag定义索引
CREATE INDEX IF NOT EXISTS idx_action_flags_version ON action_flags(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_action_flags_category ON action_flags(category);
CREATE INDEX IF NOT EXISTS idx_action_flags_duration_type ON action_flags(duration_type);

-- --------------------------------------------------------------------------------
-- 伤害类型定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS damage_types (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50), -- 'physical', 'magical', 'elemental', 'pure'

    -- 抗性关联
    resistance_attribute VARCHAR(50), -- 对应的抗性属性代码
    damage_reduction_formula TEXT, -- 伤害减免计算公式

    resistance_cap INTEGER DEFAULT 75 CHECK (resistance_cap >= 0 AND resistance_cap <= 100), -- 最大抗性上限

    -- 视觉效果
    color VARCHAR(7), -- 十六进制颜色
    icon VARCHAR(256), -- 伤害类型图标
    particle_effect VARCHAR(256), -- 粒子效果

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, code)
);

-- 伤害类型索引
CREATE INDEX IF NOT EXISTS idx_damage_types_version ON damage_types(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_damage_types_category ON damage_types(category);

-- --------------------------------------------------------------------------------
-- 射程配置规则表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS range_config_rules (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    parameter_type VARCHAR(50) NOT NULL, -- 'range', 'positions', 'area', 'line'
    parameter_format VARCHAR(100) NOT NULL, -- 'N', '0~N', 'M~N', '最近N个', '直线N格'

    -- 规则说明
    description TEXT,
    example VARCHAR(200), -- 示例
    validation_pattern VARCHAR(200), -- 验证正则表达式
    notes TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 射程配置规则索引
CREATE INDEX IF NOT EXISTS idx_range_config_rules_type ON range_config_rules(parameter_type);

-- --------------------------------------------------------------------------------
-- 插入基础数据
-- --------------------------------------------------------------------------------

-- 插入动作类别示例数据
INSERT INTO action_categories (version_id, category_code, category_name, description, default_action_type) VALUES
    (1, 'BASIC_ATTACK', '基础攻击', '基础的物理和魔法攻击动作', 'main'),
    (1, 'DEFENSIVE', '防御动作', '防御、格挡、闪避等防御性动作', 'reaction'),
    (1, 'MOVEMENT', '移动动作', '移动、传送、位置变换等动作', 'minor')
ON CONFLICT (version_id, category_code) DO NOTHING;

-- 插入基础伤害类型
INSERT INTO damage_types (version_id, code, name, category, resistance_attribute, color) VALUES
    (1, 'PHYSICAL', '物理伤害', 'physical', 'DEFENSE', '#8B4513'),
    (1, 'MAGICAL', '魔法伤害', 'magical', 'MAGIC_RESISTANCE', '#4B0082'),
    (1, 'FIRE', '火焰伤害', 'elemental', 'FIRE_RESISTANCE', '#FF4500'),
    (1, 'ICE', '冰霜伤害', 'elemental', 'ICE_RESISTANCE', '#00BFFF'),
    (1, 'LIGHTNING', '雷电伤害', 'elemental', 'LIGHTNING_RESISTANCE', '#FFD700')
ON CONFLICT (version_id, code) DO NOTHING;

-- 插入基础Action Flag
INSERT INTO action_flags (version_id, flag_code, flag_name, category, duration_type, description) VALUES
    (1, 'ATTACKING', '攻击状态', 'action_chain', 'action', '正在执行攻击动作'),
    (1, 'DEFENDING', '防御状态', 'action_chain', 'turn', '处于防御状态，减少受到的伤害'),
    (1, 'MOVED', '已移动', 'status_check', 'turn', '本回合已经移动过')
ON CONFLICT (version_id, flag_code) DO NOTHING;

-- --------------------------------------------------------------------------------
-- 高级技能相关函数
-- --------------------------------------------------------------------------------

-- 计算技能伤害的函数
CREATE OR REPLACE FUNCTION calculate_skill_damage(
    action_id_param INTEGER,
    caster_id UUID,
    target_id UUID,
    skill_level_param INTEGER DEFAULT 1
) RETURNS TABLE (
    base_damage INTEGER,
    final_damage INTEGER,
    damage_type VARCHAR,
    is_critical BOOLEAN
) AS $$
DECLARE
    action_record RECORD;
    caster_attack DECIMAL;
    target_defense DECIMAL;
    damage_multiplier DECIMAL := 1.0;
    crit_chance DECIMAL;
    crit_damage DECIMAL;
    is_crit BOOLEAN := FALSE;
    base_dmg INTEGER;
    final_dmg INTEGER;
BEGIN
    -- 获取动作信息
    SELECT * INTO action_record FROM actions WHERE id = action_id_param;

    -- 获取施法者攻击力
    SELECT final_value INTO caster_attack
    FROM hero_attributes ha
    JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
    WHERE ha.hero_id = caster_id AND hat.attribute_code = 'ATTACK_POWER';
    caster_attack := COALESCE(caster_attack, 100);

    -- 获取目标防御力
    SELECT final_value INTO target_defense
    FROM hero_attributes ha
    JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
    WHERE ha.hero_id = target_id AND hat.attribute_code = 'DEFENSE';
    target_defense := COALESCE(target_defense, 50);

    -- 获取技能等级倍数
    SELECT damage_multiplier INTO damage_multiplier
    FROM skill_level_configs slc
    JOIN skill_unlock_actions sua ON slc.skill_id = sua.skill_id
    WHERE sua.action_id = action_id_param AND slc.level_number = skill_level_param;
    damage_multiplier := COALESCE(damage_multiplier, 1.0);

    -- 计算基础伤害
    base_dmg := (caster_attack * damage_multiplier)::INTEGER;

    -- 检查暴击
    SELECT final_value INTO crit_chance
    FROM hero_attributes ha
    JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
    WHERE ha.hero_id = caster_id AND hat.attribute_code = 'CRITICAL_CHANCE';
    crit_chance := COALESCE(crit_chance, 5) / 100.0;

    IF random() < crit_chance THEN
        is_crit := TRUE;
        SELECT final_value INTO crit_damage
        FROM hero_attributes ha
        JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
        WHERE ha.hero_id = caster_id AND hat.attribute_code = 'CRITICAL_DAMAGE';
        crit_damage := COALESCE(crit_damage, 150) / 100.0;
        base_dmg := (base_dmg * crit_damage)::INTEGER;
    END IF;

    -- 应用防御减免
    final_dmg := GREATEST(1, base_dmg - target_defense::INTEGER);

    RETURN QUERY SELECT base_dmg, final_dmg, 'PHYSICAL'::VARCHAR, is_crit;
END;
$$ LANGUAGE plpgsql;

-- 获取可用动作列表的函数
CREATE OR REPLACE FUNCTION get_available_actions(
    hero_id_param UUID,
    action_type_filter action_type_enum DEFAULT NULL
) RETURNS TABLE (
    action_id INTEGER,
    action_name VARCHAR,
    action_type action_type_enum,
    mana_cost INTEGER,
    ap_cost INTEGER,
    cooldown_turns INTEGER,
    skill_name VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.action_name,
        a.action_type,
        a.mana_cost,
        a.action_point_cost,
        a.cooldown_turns,
        s.skill_name
    FROM actions a
    JOIN skill_unlock_actions sua ON a.id = sua.action_id
    JOIN skills s ON sua.skill_id = s.id
    JOIN hero_skills hs ON s.skill_code = hs.skill_code
    WHERE hs.hero_id = hero_id_param
      AND hs.is_active = TRUE
      AND hs.skill_level >= sua.unlock_level
      AND a.is_active = TRUE
      AND (action_type_filter IS NULL OR a.action_type = action_type_filter)
    ORDER BY a.action_type, a.action_name;
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 视图定义（便于查询）
-- --------------------------------------------------------------------------------

-- 活跃技能配置概览视图
CREATE OR REPLACE VIEW active_skill_config_summary AS
SELECT
    cv.version_number,
    cv.description,
    cv.created_at as version_created_at,
    (SELECT COUNT(*) FROM skills s WHERE s.version_id = cv.id AND s.is_active = true) as skills_count,
    (SELECT COUNT(*) FROM actions a WHERE a.version_id = cv.id AND a.is_active = true) as actions_count,
    (SELECT COUNT(*) FROM buffs b WHERE b.version_id = cv.id AND b.is_active = true) as buffs_count,
    (SELECT COUNT(*) FROM damage_types dt WHERE dt.version_id = cv.id AND dt.is_active = true) as damage_types_count
FROM skill_config_versions cv
WHERE cv.is_active = true;

-- 技能与解锁动作关联视图
CREATE OR REPLACE VIEW skill_action_relations AS
SELECT
    s.skill_code,
    s.skill_name,
    a.action_code,
    a.action_name,
    sua.unlock_level,
    s.version_id
FROM skills s
JOIN skill_unlock_actions sua ON s.id = sua.skill_id
JOIN actions a ON sua.action_id = a.id
WHERE s.is_active = true AND a.is_active = true;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Advanced System 创建完成';
    RAISE NOTICE '包含: 动作系统、Buff系统、战斗逻辑、效果配置';
    RAISE NOTICE '已插入 % 个动作类别', (SELECT COUNT(*) FROM action_categories WHERE is_active = TRUE);
    RAISE NOTICE '已插入 % 个伤害类型', (SELECT COUNT(*) FROM damage_types WHERE is_active = TRUE);
    RAISE NOTICE '已插入 % 个动作标识', (SELECT COUNT(*) FROM action_flags WHERE is_active = TRUE);
    RAISE NOTICE '============================================';
END $$;

-- =============================================================================
-- Create Skills Base System
-- 技能基础系统：技能定义、类别、等级配置等
-- 依赖：000001_create_core_infrastructure, 000004_create_classes_system
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 版本管理表 (技能配置版本管理)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_config_versions (
    id SERIAL PRIMARY KEY,
    version_number VARCHAR(20) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(100)
);

-- 确保只有一个活跃版本（唯一索引）
CREATE UNIQUE INDEX IF NOT EXISTS idx_skill_single_active_version ON skill_config_versions(is_active) WHERE is_active = TRUE;

-- --------------------------------------------------------------------------------
-- 技能类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    category_code VARCHAR(32) NOT NULL,-- 类别代码
    category_name VARCHAR(64) NOT NULL,-- 类别名称
    description TEXT,-- 类别描述
    
    -- 显示配置
    icon VARCHAR(256),-- 类别图标
    color VARCHAR(16),-- 类别颜色
    display_order INTEGER DEFAULT 0,-- 显示顺序
    
    is_active BOOLEAN DEFAULT TRUE,-- 是否启用
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 技能类别索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_skill_categories_code_unique 
    ON skill_categories(category_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_categories_active 
    ON skill_categories(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 更新时间戳触发器
CREATE TRIGGER update_skill_categories_updated_at
    BEFORE UPDATE ON skill_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 技能定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- 技能基本信息
    skill_code VARCHAR(50) NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_type skill_type_enum NOT NULL DEFAULT 'weapon',
    category_id UUID REFERENCES skill_categories(id),
    
    -- 等级配置
    max_level INTEGER DEFAULT 10 CHECK (max_level >= 1 AND max_level <= 100),
    
    -- 特征
    feature_tags TEXT[],
    
    -- 学习要求
    required_level INTEGER DEFAULT 1 CHECK (required_level >= 1),
    required_class_codes VARCHAR(32)[],
    prerequisite_skill_codes VARCHAR(50)[],
    
    -- 显示配置
    description TEXT,
    detailed_description TEXT,
    icon VARCHAR(256),
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 技能定义索引

CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_code_unique 
    ON skills(skill_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_type 
    ON skills(skill_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_category 
    ON skills(category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_feature_tags 
    ON skills USING GIN(feature_tags);

-- 更新时间戳触发器
CREATE TRIGGER update_skills_updated_at
    BEFORE UPDATE ON skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- --------------------------------------------------------------------------------
-- 7. 技能等级配置表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_level_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    
    level_number INTEGER NOT NULL CHECK (level_number >= 1),
    
    -- 对动作/效果的加成
    damage_multiplier DECIMAL(5,2) DEFAULT 1.0,
    healing_multiplier DECIMAL(5,2) DEFAULT 1.0,
    duration_modifier INTEGER DEFAULT 0,
    range_modifier INTEGER DEFAULT 0,
    cooldown_modifier INTEGER DEFAULT 0,
    mana_cost_modifier INTEGER DEFAULT 0,
    
    -- 其他加成配置
    effect_modifiers JSONB,
    
    -- 升级消耗
    upgrade_cost_xp INTEGER DEFAULT 0,
    upgrade_cost_gold INTEGER DEFAULT 0,
    upgrade_materials JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(skill_id, level_number)
);

CREATE INDEX IF NOT EXISTS idx_skill_level_configs_skill 
    ON skill_level_configs(skill_id) WHERE deleted_at IS NULL;

CREATE TRIGGER update_skill_level_configs_updated_at
    BEFORE UPDATE ON skill_level_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- --------------------------------------------------------------------------------
-- 技能学习条件表 (复杂的学习条件配置)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_learning_requirements (
    id SERIAL PRIMARY KEY,
    skill_id INTEGER REFERENCES skills(id) ON DELETE CASCADE,

    -- 条件类型
    requirement_type VARCHAR(50) NOT NULL, -- 'level', 'attribute', 'skill', 'item', 'quest'
    requirement_key VARCHAR(100) NOT NULL, -- 具体的条件键 (如属性代码、技能ID等)
    requirement_value INTEGER NOT NULL, -- 要求的数值

    -- 条件描述
    description TEXT,
    error_message TEXT, -- 不满足条件时的错误信息

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 技能学习条件索引
CREATE INDEX IF NOT EXISTS idx_skill_learning_requirements_skill ON skill_learning_requirements(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_learning_requirements_type ON skill_learning_requirements(requirement_type);

-- --------------------------------------------------------------------------------
-- 技能效果模板表 (可重用的技能效果)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_effect_templates (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    template_code VARCHAR(50) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    effect_type VARCHAR(50) NOT NULL, -- 'damage', 'heal', 'buff', 'debuff', 'summon'

    -- 效果参数模板
    parameter_template JSONB NOT NULL, -- 参数模板
    calculation_formula TEXT, -- 计算公式

    -- 动画和视觉效果
    visual_effects JSONB, -- 视觉效果配置
    sound_effects JSONB, -- 音效配置

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, template_code)
);

-- 技能效果模板索引
CREATE INDEX IF NOT EXISTS idx_skill_effect_templates_version ON skill_effect_templates(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_skill_effect_templates_type ON skill_effect_templates(effect_type);

-- --------------------------------------------------------------------------------
-- 插入基础数据
-- --------------------------------------------------------------------------------

-- 插入初始版本
INSERT INTO skill_config_versions (version_number, description, is_active, created_by)
VALUES ('1.0.0', '技能系统初始版本', TRUE, 'system')
ON CONFLICT (version_number) DO NOTHING;

-- 插入技能类别
INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'COMBAT', '战斗技能', '用于战斗的主动和被动技能', '/icons/combat.png', '#E74C3C', 1
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'SUPPORT', '辅助技能', '提供支援和治疗的技能', '/icons/support.png', '#2ECC71', 2
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'PASSIVE', '被动技能', '自动生效的被动能力', '/icons/passive.png', '#3498DB', 3
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'ULTIMATE', '终极技能', '强大的终极能力', '/icons/ultimate.png', '#9B59B6', 4
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

-- --------------------------------------------------------------------------------
-- 技能相关函数
-- --------------------------------------------------------------------------------

-- 检查技能学习条件的函数
CREATE OR REPLACE FUNCTION check_skill_learning_requirements(
    skill_id_param INTEGER,
    hero_id_param UUID
) RETURNS TABLE (
    is_eligible BOOLEAN,
    missing_requirements TEXT[]
) AS $$
DECLARE
    req_record RECORD;
    missing_reqs TEXT[] := '{}';
    is_valid BOOLEAN := TRUE;
    hero_level INTEGER;
    hero_attr_value DECIMAL;
    hero_skill_level INTEGER;
BEGIN
    -- 获取英雄基本信息
    SELECT level INTO hero_level FROM heroes WHERE id = hero_id_param;

    -- 检查所有学习要求
    FOR req_record IN
        SELECT requirement_type, requirement_key, requirement_value, error_message
        FROM skill_learning_requirements
        WHERE skill_id = skill_id_param AND is_active = TRUE
    LOOP
        CASE req_record.requirement_type
            WHEN 'level' THEN
                IF hero_level < req_record.requirement_value THEN
                    missing_reqs := array_append(missing_reqs,
                        COALESCE(req_record.error_message, format('需要等级 %s', req_record.requirement_value)));
                    is_valid := FALSE;
                END IF;

            WHEN 'attribute' THEN
                -- 检查属性要求
                SELECT ha.final_value INTO hero_attr_value
                FROM hero_attributes ha
                JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
                WHERE ha.hero_id = hero_id_param AND hat.attribute_code = req_record.requirement_key;

                IF COALESCE(hero_attr_value, 0) < req_record.requirement_value THEN
                    missing_reqs := array_append(missing_reqs,
                        COALESCE(req_record.error_message, format('需要 %s 达到 %s', req_record.requirement_key, req_record.requirement_value)));
                    is_valid := FALSE;
                END IF;

            WHEN 'skill' THEN
                -- 检查前置技能要求
                SELECT skill_level INTO hero_skill_level
                FROM hero_skills hs
                JOIN skills s ON hs.skill_code = s.skill_code
                WHERE hs.hero_id = hero_id_param AND s.id = req_record.requirement_key::INTEGER;

                IF COALESCE(hero_skill_level, 0) < req_record.requirement_value THEN
                    missing_reqs := array_append(missing_reqs,
                        COALESCE(req_record.error_message, '缺少前置技能'));
                    is_valid := FALSE;
                END IF;

            -- 可以继续添加其他条件类型
        END CASE;
    END LOOP;

    RETURN QUERY SELECT is_valid, missing_reqs;
END;
$$ LANGUAGE plpgsql;

-- 获取技能详细信息的函数
CREATE OR REPLACE FUNCTION get_skill_details(skill_id_param INTEGER, skill_level_param INTEGER DEFAULT 1)
RETURNS TABLE (
    skill_id INTEGER,
    skill_name VARCHAR,
    skill_type skill_type_enum,
    category_name VARCHAR,
    max_level INTEGER,
    current_damage_multiplier DECIMAL,
    current_mana_cost INTEGER,
    current_cooldown INTEGER,
    description TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.id,
        s.skill_name,
        s.skill_type,
        sc.category_name,
        s.max_level,
        COALESCE(slc.damage_multiplier, 1.0),
        s.base_mana_cost + COALESCE(slc.mana_cost_modifier, 0),
        s.base_cooldown + COALESCE(slc.cooldown_modifier, 0),
        s.description
    FROM skills s
    JOIN skill_categories sc ON s.category_id = sc.id
    LEFT JOIN skill_level_configs slc ON s.id = slc.skill_id AND slc.level_number = skill_level_param
    WHERE s.id = skill_id_param AND s.is_active = TRUE;
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Base System 创建完成';
    RAISE NOTICE '包含: 技能定义、类别、等级配置、学习条件';
    RAISE NOTICE '已插入 % 个技能类别', (SELECT COUNT(*) FROM skill_categories WHERE is_active = TRUE);
    RAISE NOTICE '============================================';
END $$;