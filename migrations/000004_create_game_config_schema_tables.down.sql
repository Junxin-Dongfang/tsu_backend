-- =============================================================================
-- Drop Game Config Schema Tables
-- 删除游戏配置数据表（按相反顺序）
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除关联表（先删除有外键的表）
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS game_config.buff_effects;
DROP TABLE IF EXISTS game_config.action_effects;
DROP TABLE IF EXISTS game_config.skill_unlock_actions;
DROP TABLE IF EXISTS game_config.class_skill_pools;
DROP TABLE IF EXISTS game_config.class_attribute_bonuses;
DROP TABLE IF EXISTS game_config.class_advanced_requirements;
DROP TABLE IF EXISTS game_config.tags_relations;

-- --------------------------------------------------------------------------------
-- 删除主表（再删除被引用的表）
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS game_config.action_flags;
DROP TABLE IF EXISTS game_config.buffs;
DROP TABLE IF EXISTS game_config.actions;
DROP TABLE IF EXISTS game_config.effects;
DROP TABLE IF EXISTS game_config.skills;
DROP TABLE IF EXISTS game_config.classes;
DROP TABLE IF EXISTS game_config.hero_attribute_type;
DROP TABLE IF EXISTS game_config.tags;

-- --------------------------------------------------------------------------------
-- 删除元数据表
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS game_config.action_type_definitions;
DROP TABLE IF EXISTS game_config.range_config_rules;
DROP TABLE IF EXISTS game_config.formula_variables;
DROP TABLE IF EXISTS game_config.effect_type_definitions;
DROP TABLE IF EXISTS game_config.damage_types;
DROP TABLE IF EXISTS game_config.action_categories;
DROP TABLE IF EXISTS game_config.skill_categories;

-- --------------------------------------------------------------------------------
-- 删除函数
-- --------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS check_class_advancement_requirements(UUID, UUID, INT, JSONB);
DROP FUNCTION IF EXISTS validate_attribute_value(VARCHAR, DECIMAL);
DROP FUNCTION IF EXISTS get_attribute_type_options();

-- --------------------------------------------------------------------------------
-- 删除枚举类型（如果没有被其他对象使用）
-- --------------------------------------------------------------------------------
DROP TYPE IF EXISTS action_type_enum;
DROP TYPE IF EXISTS modifier_type_enum;
DROP TYPE IF EXISTS attribute_category_enum;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Game Config Schema Tables 删除完成';
    RAISE NOTICE '============================================';
END $$;
