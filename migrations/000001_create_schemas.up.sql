-- =============================================================================
-- Create Database Schemas
-- 创建数据库 Schema：按照模块职责划分数据归属
-- =============================================================================

-- --------------------------------------------------------------------------------
-- Schema 创建
-- --------------------------------------------------------------------------------

-- Auth Schema: 认证授权相关数据（Auth Module 拥有）
CREATE SCHEMA IF NOT EXISTS auth;
COMMENT ON SCHEMA auth IS '认证授权相关数据：用户信息，登录历史，会话，权限等';

-- Game Config Schema: 游戏配置数据（Admin Module 拥有）
CREATE SCHEMA IF NOT EXISTS game_config;
COMMENT ON SCHEMA game_config IS '游戏配置相关数据：职业，技能，物品，怪物等';

-- Game Runtime Schema: 游戏运行时数据（Game Module 拥有）
CREATE SCHEMA IF NOT EXISTS game_runtime;
COMMENT ON SCHEMA game_runtime IS '游戏运行时相关数据：玩家角色，背包，属性数据等';

-- Admin Schema: 后台管理数据（Admin Module 拥有）
CREATE SCHEMA IF NOT EXISTS admin;
COMMENT ON SCHEMA admin IS '后台管理数据：操作日志，系统配置，审计记录等';

-- --------------------------------------------------------------------------------
-- 数据库用户创建
-- --------------------------------------------------------------------------------

-- 创建各模块专用数据库用户
-- 注意：密码应该从环境变量或密钥管理系统中获取，这里仅为示例

DO $$
BEGIN
    -- Auth 模块用户
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tsu_auth_user') THEN
        CREATE ROLE tsu_auth_user WITH LOGIN PASSWORD 'tsu_auth_password';
        RAISE NOTICE '创建用户: tsu_auth_user';
    END IF;

    -- Game 模块用户
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tsu_game_user') THEN
        CREATE ROLE tsu_game_user WITH LOGIN PASSWORD 'tsu_game_password';
        RAISE NOTICE '创建用户: tsu_game_user';
    END IF;

    -- Admin 模块用户
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tsu_admin_user') THEN
        CREATE ROLE tsu_admin_user WITH LOGIN PASSWORD 'tsu_admin_password';
        RAISE NOTICE '创建用户: tsu_admin_user';
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- Schema 基础权限配置
-- --------------------------------------------------------------------------------

-- 1. Auth Schema 权限
GRANT USAGE ON SCHEMA auth TO tsu_auth_user;
GRANT ALL PRIVILEGES ON SCHEMA auth TO tsu_auth_user;

-- 其他模块对 auth 只读权限
GRANT USAGE ON SCHEMA auth TO tsu_game_user;
GRANT USAGE ON SCHEMA auth TO tsu_admin_user;

-- 2. Game Config Schema 权限
GRANT USAGE ON SCHEMA game_config TO tsu_admin_user;
GRANT ALL PRIVILEGES ON SCHEMA game_config TO tsu_admin_user;

-- Game 模块对 game_config 只读权限
GRANT USAGE ON SCHEMA game_config TO tsu_game_user;

-- 3. Game Runtime Schema 权限
GRANT USAGE ON SCHEMA game_runtime TO tsu_game_user;
GRANT ALL PRIVILEGES ON SCHEMA game_runtime TO tsu_game_user;

-- 4. Admin Schema 权限
GRANT USAGE ON SCHEMA admin TO tsu_admin_user;
GRANT ALL PRIVILEGES ON SCHEMA admin TO tsu_admin_user;

-- 5. Public Schema 权限（枚举类型、扩展）
GRANT USAGE ON SCHEMA public TO tsu_auth_user, tsu_game_user, tsu_admin_user;

-- --------------------------------------------------------------------------------
-- 默认权限配置（对未来创建的表自动授权）
-- --------------------------------------------------------------------------------

-- Auth Schema: 未来创建的表，tsu_auth_user 自动拥有所有权限
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO tsu_auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT USAGE, SELECT ON SEQUENCES TO tsu_auth_user;

-- 其他模块对 auth 未来的表只读
ALTER DEFAULT PRIVILEGES IN SCHEMA auth GRANT SELECT ON TABLES TO tsu_game_user, tsu_admin_user;

-- Game Config Schema: tsu_admin_user 拥有所有权限
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO tsu_admin_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT USAGE, SELECT ON SEQUENCES TO tsu_admin_user;

-- Game 模块对 game_config 只读
ALTER DEFAULT PRIVILEGES IN SCHEMA game_config GRANT SELECT ON TABLES TO tsu_game_user;

-- Game Runtime Schema: tsu_game_user 拥有所有权限
ALTER DEFAULT PRIVILEGES IN SCHEMA game_runtime GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO tsu_game_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA game_runtime GRANT USAGE, SELECT ON SEQUENCES TO tsu_game_user;

-- Admin Schema: tsu_admin_user 拥有所有权限
ALTER DEFAULT PRIVILEGES IN SCHEMA admin GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO tsu_admin_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA admin GRANT USAGE, SELECT ON SEQUENCES TO tsu_admin_user;

-- --------------------------------------------------------------------------------
-- 迁移执行用户权限
-- --------------------------------------------------------------------------------

-- 授予当前执行迁移的用户（应该是 tsu_user 或超级用户）所有 schema 的完整权限
DO $$
DECLARE
    migration_user TEXT;
BEGIN
    SELECT current_user INTO migration_user;

    -- 授予所有 schema 的完整权限（用于执行迁移）
    EXECUTE format('GRANT ALL PRIVILEGES ON SCHEMA auth TO %I', migration_user);
    EXECUTE format('GRANT ALL PRIVILEGES ON SCHEMA game_config TO %I', migration_user);
    EXECUTE format('GRANT ALL PRIVILEGES ON SCHEMA game_runtime TO %I', migration_user);
    EXECUTE format('GRANT ALL PRIVILEGES ON SCHEMA admin TO %I', migration_user);
    EXECUTE format('GRANT ALL PRIVILEGES ON SCHEMA public TO %I', migration_user);

    RAISE NOTICE '迁移执行用户 % 已获得所有权限', migration_user;
END $$;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Database Schemas 创建完成';
    RAISE NOTICE '包含:';
    RAISE NOTICE '  - auth: 认证授权数据 (Owner: tsu_auth_user)';
    RAISE NOTICE '  - game_config: 游戏配置数据 (Owner: tsu_admin_user)';
    RAISE NOTICE '  - game_runtime: 游戏运行时数据 (Owner: tsu_game_user)';
    RAISE NOTICE '  - admin: 后台管理数据 (Owner: tsu_admin_user)';
    RAISE NOTICE '';
    RAISE NOTICE '数据库用户:';
    RAISE NOTICE '  - tsu_auth_user: Auth 模块专用';
    RAISE NOTICE '  - tsu_game_user: Game 模块专用';
    RAISE NOTICE '  - tsu_admin_user: Admin 模块专用';
    RAISE NOTICE '============================================';
END $$;
