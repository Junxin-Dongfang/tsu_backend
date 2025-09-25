-- Rollback Hero System
-- =============================================================================

-- 删除表
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS hero_attribute_type CASCADE;

-- 删除枚举类型
DROP TYPE IF EXISTS tag_type_enum;
DROP TYPE IF EXISTS attribute_data_type_enum;
DROP TYPE IF EXISTS attribute_category_enum;
