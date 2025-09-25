-- 回滚000002_init_create_users_business_extensions.up.sql，从下往上执行
DROP TABLE IF EXISTS financial_transactions CASCADE;
DROP TYPE IF EXISTS payment_method_enum CASCADE;
DROP TYPE IF EXISTS transaction_status_enum CASCADE;
DROP TYPE IF EXISTS transaction_type_enum CASCADE;
DROP FUNCTION IF EXISTS is_user_premium CASCADE;
DROP TABLE IF EXISTS user_finances CASCADE;

