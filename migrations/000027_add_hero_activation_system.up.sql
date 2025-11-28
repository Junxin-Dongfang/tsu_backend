-- 英雄激活与上下文管理系统迁移
-- 添加英雄激活状态和当前操作英雄跟踪功能

-- ============================================================
-- 第 1 步：添加 is_activated 字段到 heroes 表
-- ============================================================
ALTER TABLE game_runtime.heroes
ADD COLUMN is_activated BOOLEAN NOT NULL DEFAULT FALSE;

-- 创建索引：优化已激活英雄查询性能
CREATE INDEX idx_heroes_user_activated
ON game_runtime.heroes(user_id, is_activated)
WHERE is_activated = TRUE AND deleted_at IS NULL;

-- ============================================================
-- 第 2 步：创建 current_hero_contexts 表
-- ============================================================
CREATE TABLE IF NOT EXISTS game_runtime.current_hero_contexts (
    user_id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE RESTRICT,

    switched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引：用于通过 hero_id 查询（如删除英雄时）
CREATE INDEX idx_current_hero_contexts_hero_id
ON game_runtime.current_hero_contexts(hero_id);

-- ============================================================
-- 第 3 步：创建触发器维护 updated_at
-- ============================================================
-- 注意：这里假设 update_updated_at_column() 函数已存在
-- 如果不存在，取消注释下面的定义
/*
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
*/

CREATE TRIGGER update_current_hero_contexts_updated_at
    BEFORE UPDATE ON game_runtime.current_hero_contexts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 第 4 步：创建验证触发器（确保当前英雄已激活）
-- ============================================================
CREATE OR REPLACE FUNCTION validate_current_hero_is_activated()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM game_runtime.heroes
        WHERE id = NEW.hero_id
          AND user_id = NEW.user_id
          AND is_activated = TRUE
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'hero_id must reference an activated hero belonging to user_id';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_current_hero_activated
    BEFORE INSERT OR UPDATE ON game_runtime.current_hero_contexts
    FOR EACH ROW EXECUTE FUNCTION validate_current_hero_is_activated();

-- ============================================================
-- 第 5 步：数据迁移（向后兼容）
-- ============================================================
-- 将现有英雄设为已激活（向后兼容：假设现有英雄都应该可用）
UPDATE game_runtime.heroes
SET is_activated = TRUE
WHERE deleted_at IS NULL;

-- 为每个用户设置当前操作英雄（选择第一个已激活的英雄）
INSERT INTO game_runtime.current_hero_contexts (user_id, hero_id, switched_at)
SELECT DISTINCT ON (h.user_id)
    h.user_id,
    h.id,
    NOW()
FROM game_runtime.heroes h
WHERE h.is_activated = TRUE
    AND h.deleted_at IS NULL
ORDER BY h.user_id, h.created_at ASC
ON CONFLICT (user_id) DO NOTHING;
