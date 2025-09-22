CREATE EXTENSION IF NOT EXISTS pgcrypto; -- 提供 gen_random_uuid()

SET TIME ZONE 'UTC';

-- 通用更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW(); 
   RETURN NEW;
END;
$$ language 'plpgsql';

--创建性别枚举类型
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_enum') THEN
        CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other', 'prefer_not_to_say');
    END IF;
END;
$$;

-- =============================================================================
-- 用户核心表：整合认证信息与业务数据
-- 注意：这个表与 Kratos identities 表保持松耦合，可以独立存在
-- =============================================================================

CREATE TABLE IF NOT EXISTS users (
    -- 主键：与 Kratos identity ID 对应
    id                 UUID PRIMARY KEY,
    
    -- 业务核心字段
    is_premium         BOOLEAN NOT NULL DEFAULT FALSE, --是否为高级用户
    diamond_count      INTEGER NOT NULL DEFAULT 0 CHECK (diamond_count >= 0),--钻石数量

    --用户信息
    username          VARCHAR(50) UNIQUE NOT NULL,
    email             VARCHAR(255) UNIQUE NOT NULL,
    phone_number      VARCHAR(20) UNIQUE,
    
    -- 用户状态管理
    is_banned          BOOLEAN NOT NULL DEFAULT FALSE,
    ban_until          TIMESTAMPTZ,
    ban_reason         TEXT,
    
    -- 个人资料
    avatar_url         VARCHAR(500),-- 头像URL
    bio                TEXT,-- 个人简介
    display_name       VARCHAR(100),-- 显示名称
    birth_date         DATE,-- 出生日期
    gender             gender_enum,-- 性别
    timezone           VARCHAR(50) DEFAULT 'UTC',-- 时区
    language           VARCHAR(10) DEFAULT 'zh-CN',-- 语言偏好
    
    -- 业务统计
    total_spent        DECIMAL(12,2) NOT NULL DEFAULT 0.00,-- 总消费金额
    referral_code      VARCHAR(20) UNIQUE,-- 推荐码
    referred_by        UUID REFERENCES users(id),-- 被推荐人
    referral_count     INTEGER NOT NULL DEFAULT 0 CHECK (referral_count >= 0),-- 推荐人数

    --登录追踪
    last_login_at      TIMESTAMPTZ,-- 上次登录时间
    last_login_ip      VARCHAR(45),-- 上次登录IP
    login_count        INTEGER NOT NULL DEFAULT 0 CHECK (login_count >= 0),-- 登录次数

    -- 时间戳
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ -- 软删除
);

-- 用户表索引
CREATE UNIQUE INDEX idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_is_premium ON users(is_premium) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_is_banned ON users(is_banned) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_referred_by ON users(referred_by) WHERE deleted_at IS NULL;

-- 用户表触发器
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- 设备类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'device_type_enum') THEN
        CREATE TYPE device_type_enum AS ENUM ('mobile', 'desktop', 'tablet', 'other');
    END IF;
END;
$$;

-- 登录方式枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'login_method_enum') THEN
        CREATE TYPE login_method_enum AS ENUM ('password', 'oauth', 'sms', 'magic_link', 'biometric');
    END IF;
END;

-- 登录结果枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'login_status_enum') THEN
        CREATE TYPE login_status_enum AS ENUM ('success', 'failed', 'blocked');
    END IF;
END;

-- =============================================================================
-- 用户登录历史表
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_login_history (
    id             BIGSERIAL PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 登录信息
    login_time     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    logout_time    TIMESTAMPTZ,
    session_duration INTEGER, -- 会话时长（秒）
    
    -- 客户端信息
    ip_address     VARCHAR(45) NOT NULL,
    user_agent     TEXT,
    device_type    device_type_enum,
    browser_name   VARCHAR(100),
    browser_version VARCHAR(50),
    os_name        VARCHAR(100),
    os_version     VARCHAR(50),
    
    -- 地理位置
    country        VARCHAR(100),
    region         VARCHAR(100),
    city           VARCHAR(100),
    
    -- 登录方式
    login_method   VARCHAR(50) NOT NULL DEFAULT 'password', -- password, oauth, sms, etc.
    oauth_provider VARCHAR(50), -- google, facebook, github, etc.
    
    -- 安全信息
    is_suspicious  BOOLEAN NOT NULL DEFAULT FALSE,
    risk_score     INTEGER CHECK (risk_score >= 0 AND risk_score <= 100),
    
    -- 状态
    status         login_status_enum NOT NULL DEFAULT 'success' CHECK (status IN ('success', 'failed', 'blocked'))
);

-- 登录历史索引
CREATE INDEX idx_login_history_user_id ON user_login_history(user_id);
CREATE INDEX idx_login_history_login_time ON user_login_history(login_time);
CREATE INDEX idx_login_history_ip_address ON user_login_history(ip_address);
CREATE INDEX idx_login_history_status ON user_login_history(status);
CREATE INDEX idx_login_history_suspicious ON user_login_history(is_suspicious) WHERE is_suspicious = TRUE;

-- =============================================================================
-- 用户设置表
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_settings (
    id               BIGSERIAL PRIMARY KEY,
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 通知设置
    email_notifications BOOLEAN NOT NULL DEFAULT TRUE,-- 是否启用邮件通知
    sms_notifications   BOOLEAN NOT NULL DEFAULT FALSE,-- 是否启用短信通知
    push_notifications  BOOLEAN NOT NULL DEFAULT TRUE,-- 是否启用推送通知
    marketing_emails    BOOLEAN NOT NULL DEFAULT FALSE,-- 是否接收营销邮件

    -- 隐私设置
    profile_visibility  VARCHAR(20) NOT NULL DEFAULT 'public' CHECK (profile_visibility IN ('public', 'friends', 'private')),
    show_online_status  BOOLEAN NOT NULL DEFAULT TRUE,-- 是否显示在线状态
    allow_friend_requests BOOLEAN NOT NULL DEFAULT TRUE,-- 是否允许好友请求
    
    -- 安全设置
    two_factor_enabled  BOOLEAN NOT NULL DEFAULT FALSE,-- 是否启用双因素认证
    login_alerts        BOOLEAN NOT NULL DEFAULT TRUE,-- 是否启用登录提醒

    -- 其他设置
    theme              VARCHAR(20) NOT NULL DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),-- 主题
    
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id)
);

CREATE TRIGGER update_user_settings_updated_at 
    BEFORE UPDATE ON user_settings 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();