-- 000017_add_equipment_slots.down.sql
-- 回滚装备槽位配置表

-- 1. 删除装备槽位表
DROP TABLE IF EXISTS game_config.equipment_slots;

-- 2. 删除槽位类型枚举
DROP TYPE IF EXISTS game_config.slot_type_enum;

