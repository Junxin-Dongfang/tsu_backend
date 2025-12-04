-- Grant tsu_admin_user permissions to read/write game_runtime objects
GRANT USAGE ON SCHEMA game_runtime TO tsu_admin_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA game_runtime TO tsu_admin_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA game_runtime TO tsu_admin_user;

-- Ensure future tables/sequences in game_runtime also grant to tsu_admin_user
ALTER DEFAULT PRIVILEGES IN SCHEMA game_runtime
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO tsu_admin_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA game_runtime
  GRANT USAGE, SELECT ON SEQUENCES TO tsu_admin_user;
