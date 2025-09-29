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
-- 删除视图
-- --------------------------------------------------------------------------------
DROP VIEW IF EXISTS skill_action_relations;
DROP VIEW IF EXISTS active_skill_config_summary;

-- --------------------------------------------------------------------------------
-- 删除函数
-- --------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS get_available_actions(UUID, action_type_enum);
DROP FUNCTION IF EXISTS calculate_skill_damage(INTEGER, UUID, UUID, INTEGER);

-- --------------------------------------------------------------------------------
-- 删除表（按依赖关系逆序）
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS range_config_rules;
DROP TABLE IF EXISTS damage_types;
DROP TABLE IF EXISTS action_flags;
DROP TABLE IF EXISTS buffs;
DROP TABLE IF EXISTS skill_unlock_actions;
DROP TABLE IF EXISTS actions;
DROP TABLE IF EXISTS action_categories;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Advanced System 删除完成';
    RAISE NOTICE '============================================';
END $$;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Base System 删除完成';
    RAISE NOTICE '============================================';
END $$;