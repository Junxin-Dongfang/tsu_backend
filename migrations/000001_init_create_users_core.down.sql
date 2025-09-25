-- Rollback core user tables
DROP TABLE IF EXISTS user_settings CASCADE;
DROP TABLE IF EXISTS user_login_history CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Drop enums
DROP TYPE IF EXISTS gender_enum;
DROP TYPE IF EXISTS device_type_enum;
DROP TYPE IF EXISTS login_method_enum;
DROP TYPE IF EXISTS login_status_enum;