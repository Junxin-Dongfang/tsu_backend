-- =============================================================================
-- 000006_create_skills_base.down.sql
-- 技能基础系统回滚：删除技能、动作、效果、判定系统的所有表、视图、函数和类型
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 1. 删除视图
-- --------------------------------------------------------------------------------

DROP VIEW IF EXISTS buff_effects_detail;
DROP VIEW IF EXISTS action_effects_detail;
DROP VIEW IF EXISTS skill_action_relations;

-- --------------------------------------------------------------------------------
-- 2. 删除函数
-- --------------------------------------------------------------------------------

DROP FUNCTION IF EXISTS calculate_hit_rate(UUID, UUID, UUID);

-- --------------------------------------------------------------------------------
-- 3. 删除表（按依赖关系顺序删除）
-- --------------------------------------------------------------------------------

-- 删除用户技能表
DROP TABLE IF EXISTS hero_skills;

-- 删除关联表
DROP TABLE IF EXISTS buff_effects;
DROP TABLE IF EXISTS skill_unlock_actions;
DROP TABLE IF EXISTS action_effects;

-- 删除主要实体表
DROP TABLE IF EXISTS action_flags;
DROP TABLE IF EXISTS buffs;
DROP TABLE IF EXISTS actions;
DROP TABLE IF EXISTS skill_level_configs;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS effects;

-- 删除元数据表
DROP TABLE IF EXISTS action_type_definitions;
DROP TABLE IF EXISTS range_config_rules;
DROP TABLE IF EXISTS formula_variables;
DROP TABLE IF EXISTS effect_type_definitions;
DROP TABLE IF EXISTS damage_types;
DROP TABLE IF EXISTS action_categories;
DROP TABLE IF EXISTS skill_categories;

-- --------------------------------------------------------------------------------
-- 4. 删除枚举类型
-- --------------------------------------------------------------------------------

DROP TYPE IF EXISTS action_type_enum;
DROP TYPE IF EXISTS modifier_type_enum;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE '技能基础系统回滚完成';
    RAISE NOTICE '============================================';
    RAISE NOTICE '已删除:';
    RAISE NOTICE '  ✓ 所有技能系统相关表';
    RAISE NOTICE '  ✓ 所有动作系统相关表';
    RAISE NOTICE '  ✓ 所有效果系统相关表';
    RAISE NOTICE '  ✓ 所有视图和函数';
    RAISE NOTICE '  ✓ 所有枚举类型';
    RAISE NOTICE '============================================';
END $$;
