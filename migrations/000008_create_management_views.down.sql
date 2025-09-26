-- =============================================================================
-- Drop Management Views
-- 删除管理视图（按相反顺序）
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除视图
-- --------------------------------------------------------------------------------
DROP VIEW IF EXISTS skill_system_statistics;
DROP VIEW IF EXISTS attribute_type_management;
DROP VIEW IF EXISTS hero_skill_details;
DROP VIEW IF EXISTS hero_attribute_details;
DROP VIEW IF EXISTS hero_details;
DROP VIEW IF EXISTS class_advancement_paths_view;
DROP VIEW IF EXISTS class_tags_view;
DROP VIEW IF EXISTS class_details;
DROP VIEW IF EXISTS class_hero_stats;
DROP VIEW IF EXISTS user_statistics;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Management Views 删除完成';
    RAISE NOTICE '============================================';
END $$;