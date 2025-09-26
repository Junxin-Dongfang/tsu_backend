-- =============================================================================
-- Drop Users System
-- 删除用户系统相关表和函数（按相反顺序）
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除函数
-- --------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS get_user_diamond_balance(UUID);
DROP FUNCTION IF EXISTS is_user_premium(UUID);

-- --------------------------------------------------------------------------------
-- 删除表（按依赖关系逆序）
-- --------------------------------------------------------------------------------
DROP TABLE IF EXISTS user_login_history;
DROP TABLE IF EXISTS financial_transactions;
DROP TABLE IF EXISTS user_finances;
DROP TABLE IF EXISTS users;

-- --------------------------------------------------------------------------------
-- 删除枚举类型（如果没有被其他对象使用）
-- --------------------------------------------------------------------------------
DROP TYPE IF EXISTS payment_method_enum;
DROP TYPE IF EXISTS transaction_status_enum;
DROP TYPE IF EXISTS transaction_type_enum;
DROP TYPE IF EXISTS login_status_enum;
DROP TYPE IF EXISTS login_method_enum;
DROP TYPE IF EXISTS device_type_enum;
DROP TYPE IF EXISTS gender_enum;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Users System 删除完成';
    RAISE NOTICE '============================================';
END $$;