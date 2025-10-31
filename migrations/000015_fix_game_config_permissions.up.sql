-- 修复game_config schema的权限问题
-- 确保各个服务用户对game_config schema有正确的权限

-- 1. 授予tsu_admin_user完整权限（Admin Server需要读写game_config）
GRANT USAGE ON SCHEMA game_config TO tsu_admin_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA game_config TO tsu_admin_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA game_config TO tsu_admin_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT ALL PRIVILEGES ON TABLES TO tsu_admin_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT ALL PRIVILEGES ON SEQUENCES TO tsu_admin_user;

-- 2. 授予tsu_game_user只读权限（Game Server只需要读取game_config）
GRANT USAGE ON SCHEMA game_config TO tsu_game_user;
GRANT SELECT ON ALL TABLES IN SCHEMA game_config TO tsu_game_user;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA game_config TO tsu_game_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT SELECT ON TABLES TO tsu_game_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT USAGE ON SEQUENCES TO tsu_game_user;

-- 3. 授予tsu_user完整权限（用于迁移和管理）
GRANT USAGE ON SCHEMA game_config TO tsu_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA game_config TO tsu_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA game_config TO tsu_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT ALL PRIVILEGES ON TABLES TO tsu_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT ALL PRIVILEGES ON SEQUENCES TO tsu_user;

