-- =============================================================================
-- Rollback Equipment System
-- 回滚装备系统的所有变更
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除运行时表 (game_runtime schema)
-- --------------------------------------------------------------------------------

DROP TABLE IF EXISTS game_runtime.item_operation_logs CASCADE;
DROP TABLE IF EXISTS game_runtime.item_drop_records CASCADE;
DROP TABLE IF EXISTS game_runtime.hero_equipment_slots CASCADE;
DROP TABLE IF EXISTS game_runtime.player_items CASCADE;

-- --------------------------------------------------------------------------------
-- 删除配置表 (game_config schema)
-- --------------------------------------------------------------------------------

DROP TABLE IF EXISTS game_config.socket_type_configs CASCADE;
DROP TABLE IF EXISTS game_config.gem_effect_configs CASCADE;
DROP TABLE IF EXISTS game_config.equipment_set_configs CASCADE;
DROP TABLE IF EXISTS game_config.item_drop_configs CASCADE;
DROP TABLE IF EXISTS game_config.equipment_slot_configs CASCADE;
DROP TABLE IF EXISTS game_config.items CASCADE;

-- --------------------------------------------------------------------------------
-- 删除枚举类型
-- --------------------------------------------------------------------------------

DROP TYPE IF EXISTS drop_source_enum CASCADE;
DROP TYPE IF EXISTS uniqueness_type_enum CASCADE;
DROP TYPE IF EXISTS gem_color_enum CASCADE;
DROP TYPE IF EXISTS socket_size_enum CASCADE;
DROP TYPE IF EXISTS item_location_enum CASCADE;
DROP TYPE IF EXISTS material_type_enum CASCADE;
DROP TYPE IF EXISTS item_quality_enum CASCADE;
DROP TYPE IF EXISTS item_type_enum CASCADE;

