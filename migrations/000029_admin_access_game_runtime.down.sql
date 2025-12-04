-- Revoke grants from tsu_admin_user on game_runtime schema
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA game_runtime FROM tsu_admin_user;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA game_runtime FROM tsu_admin_user;
REVOKE USAGE ON SCHEMA game_runtime FROM tsu_admin_user;

-- Remove default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA game_runtime
  REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM tsu_admin_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA game_runtime
  REVOKE USAGE, SELECT ON SEQUENCES FROM tsu_admin_user;
