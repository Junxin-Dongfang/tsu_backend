-- =============================================================================
-- Drop Heroes System
-- 删除英雄系统相关表和函数（按相反顺序）
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除触发器
-- --------------------------------------------------------------------------------
DROP TRIGGER IF EXISTS initialize_hero_attributes_trigger ON heroes;

-- --------------------------------------------------------------------------------
-- 删除函数
-- --------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS trigger_initialize_hero_attributes();
DROP FUNCTION IF EXISTS get_hero_details(UUID);
DROP FUNCTION IF EXISTS calculate_hero_power(UUID);
DROP FUNCTION IF EXISTS initialize_hero_attributes(UUID, UUID);
DROP FUNCTION IF EXISTS calculate_hero_final_attribute();

-- --------------------------------------------------------------------------------
-- 删除表（按依赖关系逆序）
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS hero_experience_logs;
DROP TABLE IF EXISTS hero_skills;
DROP TABLE IF EXISTS hero_equipment;
DROP TABLE IF EXISTS hero_attributes;
DROP TABLE IF EXISTS heroes;

-- --------------------------------------------------------------------------------
-- 删除枚举类型
-- --------------------------------------------------------------------------------
DROP TYPE IF EXISTS equipment_slot_enum;
DROP TYPE IF EXISTS hero_status_enum;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Heroes System 删除完成';
    RAISE NOTICE '============================================';
END $$;