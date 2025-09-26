-- =============================================================================
-- Drop Skills Advanced System
-- 删除技能高级系统相关表和函数（按相反顺序）
-- =============================================================================

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