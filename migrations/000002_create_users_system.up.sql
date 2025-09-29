-- =============================================================================
-- Create Users System
-- 用户系统：用户核心信息、财务、登录历史等
-- 依赖：000001_create_core_infrastructure
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 用户相关枚举类型
-- --------------------------------------------------------------------------------

-- 性别枚举类型
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_enum') THEN
        CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other', 'prefer_not_to_say');
    END IF;
END $$;

-- 设备类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'device_type_enum') THEN
        CREATE TYPE device_type_enum AS ENUM ('mobile', 'desktop', 'tablet', 'other');
    END IF;
END $$;

-- 登录方式枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'login_method_enum') THEN
        CREATE TYPE login_method_enum AS ENUM ('password', 'oauth', 'sms', 'magic_link', 'biometric');
    END IF;
END $$;

-- 登录结果枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'login_status_enum') THEN
        CREATE TYPE login_status_enum AS ENUM ('success', 'failed', 'blocked');
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 用户核心表
-- 注意：这个表与 Kratos identities 表保持松耦合，可以独立存在
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS users (
    -- 主键：与 Kratos identity ID 对应
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 用户信息（用户名，唯一且不可更改）
    username          VARCHAR(50) UNIQUE NOT NULL,
    nickname          VARCHAR(50), -- 昵称，可更改
    email             VARCHAR(255) UNIQUE NOT NULL, -- 邮箱
    phone_number      VARCHAR(20) UNIQUE, -- 手机号码

    -- 用户状态管理（是否被封禁）
    is_banned          BOOLEAN NOT NULL DEFAULT FALSE,
    ban_until          TIMESTAMPTZ, -- 封禁截止时间
    ban_reason         TEXT, -- 封禁原因

    -- 个人资料（头像URL）
    avatar_url         VARCHAR(500),
    bio                TEXT, -- 个人简介
    birth_date         DATE, -- 出生日期
    gender             gender_enum, -- 性别
    timezone           VARCHAR(50) DEFAULT 'UTC', -- 时区
    language           VARCHAR(10) DEFAULT 'zh-CN', -- 语言偏好

    -- 登录追踪（上次登录时间）
    last_login_at      TIMESTAMPTZ,
    last_login_ip      INET, -- 上次登录IP
    login_count        INTEGER NOT NULL DEFAULT 0 CHECK (login_count >= 0), -- 登录次数

    -- 时间戳
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ -- 软删除
);

-- 用户表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_unique ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users(email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone_number) WHERE phone_number IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_banned_true ON users(is_banned) WHERE is_banned = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
-- 用户表触发器
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- --------------------------------------------------------------------------------
-- 用户登录历史表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_login_history (
    id             BIGSERIAL PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- 用户ID

    -- 登录信息（登录时间）
    login_time     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    logout_time    TIMESTAMPTZ, -- 登出时间
    session_duration INTEGER, -- 会话时长（秒）

    -- 客户端信息（IP地址）
    ip_address     VARCHAR(45) NOT NULL,
    user_agent     TEXT, -- 用户代理
    device_type    device_type_enum, -- 设备类型
    browser_name   VARCHAR(100), -- 浏览器名称
    browser_version VARCHAR(50), -- 浏览器版本
    os_name        VARCHAR(100), -- 操作系统名称
    os_version     VARCHAR(50), -- 操作系统版本

    -- 地理位置（国家）
    country        VARCHAR(100),
    region         VARCHAR(100), -- 地区
    city           VARCHAR(100), -- 城市

    -- 登录方式
    login_method   login_method_enum NOT NULL DEFAULT 'password',
    oauth_provider VARCHAR(50), -- OAuth提供商

    -- 安全信息（是否可疑）
    is_suspicious  BOOLEAN NOT NULL DEFAULT FALSE,
    risk_score     INTEGER CHECK (risk_score >= 0 AND risk_score <= 100), -- 风险评分

    -- 状态（登录状态）
    status         login_status_enum NOT NULL DEFAULT 'success'
);

-- 登录历史索引
CREATE INDEX IF NOT EXISTS idx_login_history_user_id_login_time ON user_login_history(user_id, login_time DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_login_time ON user_login_history(login_time);
CREATE INDEX IF NOT EXISTS idx_login_history_ip_address ON user_login_history(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_history_status ON user_login_history(status);
CREATE INDEX IF NOT EXISTS idx_login_history_suspicious ON user_login_history(is_suspicious) WHERE is_suspicious = TRUE;

-- --------------------------------------------------------------------------------
-- 用户相关函数
-- --------------------------------------------------------------------------------

-- 判断用户是否为高级用户的函数
CREATE OR REPLACE FUNCTION is_user_premium(user_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    premium_status BOOLEAN;
BEGIN
    SELECT
        CASE
            WHEN premium_start IS NOT NULL AND premium_expiry IS NOT NULL AND NOW() BETWEEN premium_start AND premium_expiry THEN TRUE
            ELSE FALSE
        END
    INTO premium_status
    FROM user_finances
    WHERE user_id = user_uuid AND deleted_at IS NULL;

    -- 没有找到用户财务记录等情况，返回 FALSE
    RETURN COALESCE(premium_status, FALSE);
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 设置时区
-- --------------------------------------------------------------------------------
SET TIME ZONE 'UTC';

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Users System 创建完成';
    RAISE NOTICE '包含: 用户核心表、财务表、交易记录、登录历史';
    RAISE NOTICE '============================================';
END $$;