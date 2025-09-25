-- Rollback Classes System
-- =============================================================================

-- 删除表（按依赖关系顺序，从子表到父表）
DROP TABLE IF EXISTS class_advanced_requirements CASCADE;
DROP TABLE IF EXISTS class_attribute_bonuses CASCADE;
DROP TABLE IF EXISTS class_tag_relations CASCADE;
DROP TABLE IF EXISTS classes CASCADE;

-- 删除枚举类型
DROP TYPE IF EXISTS class_tier_enum;
