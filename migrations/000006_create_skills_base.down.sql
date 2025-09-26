-- =============================================================================
-- Drop Skills Base System
-- 删除技能基础系统相关表和函数（按相反顺序）
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除函数
-- --------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS get_skill_details(INTEGER, INTEGER);
DROP FUNCTION IF EXISTS check_skill_learning_requirements(INTEGER, UUID);

-- --------------------------------------------------------------------------------
-- 删除表（按依赖关系逆序）
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS skill_effect_templates;
DROP TABLE IF EXISTS skill_learning_requirements;
DROP TABLE IF EXISTS skill_level_configs;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS skill_categories;
DROP TABLE IF EXISTS skill_config_versions;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Base System 删除完成';
    RAISE NOTICE '============================================';
END $$;