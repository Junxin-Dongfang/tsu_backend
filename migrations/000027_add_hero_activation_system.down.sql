-- 英雄激活与上下文管理系统迁移回滚

-- ============================================================
-- 第 1 步：删除触发器和函数
-- ============================================================
DROP TRIGGER IF EXISTS check_current_hero_activated
    ON game_runtime.current_hero_contexts;

DROP FUNCTION IF EXISTS validate_current_hero_is_activated();

DROP TRIGGER IF EXISTS update_current_hero_contexts_updated_at
    ON game_runtime.current_hero_contexts;

-- ============================================================
-- 第 2 步：删除 current_hero_contexts 表
-- ============================================================
DROP TABLE IF EXISTS game_runtime.current_hero_contexts;

-- ============================================================
-- 第 3 步：删除 heroes 表的 is_activated 字段和索引
-- ============================================================
DROP INDEX IF EXISTS idx_heroes_user_activated;

ALTER TABLE game_runtime.heroes
DROP COLUMN IF EXISTS is_activated;
