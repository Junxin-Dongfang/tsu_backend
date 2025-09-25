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

    --用户信息
    username          VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名，唯一且不可更改',
    nickname          VARCHAR(50) COMMENT '昵称，可更改',
    email             VARCHAR(255) UNIQUE NOT NULL COMMENT '邮箱',
    phone_number      VARCHAR(20) UNIQUE COMMENT '手机号码',

    -- 用户状态管理
    is_banned          BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否被封禁',
    ban_until          TIMESTAMPTZ COMMENT '封禁截止时间',
    ban_reason         TEXT COMMENT '封禁原因',

    -- 个人资料
    avatar_url         VARCHAR(500) COMMENT '头像URL',
    bio                TEXT COMMENT '个人简介',
    birth_date         DATE COMMENT '出生日期',
    gender             gender_enum COMMENT '性别',
    timezone           VARCHAR(50) DEFAULT 'UTC' COMMENT '时区',
    language           VARCHAR(10) DEFAULT 'zh-CN' COMMENT '语言偏好',

    --登录追踪
    last_login_at      TIMESTAMPTZ COMMENT '上次登录时间',
    last_login_ip      VARCHAR(45) COMMENT '上次登录IP',
    login_count        INTEGER NOT NULL DEFAULT 0 CHECK (login_count >= 0) COMMENT '登录次数',

    -- 时间戳
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ -- 软删除
);

-- 用户表索引
CREATE UNIQUE INDEX idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL AND deleted_at IS NULL;
--为被封禁的用户创建索引，优化查询
CREATE INDEX idx_users_is_banned_true ON users(is_banned) WHERE is_banned = TRUE AND deleted_at IS NULL;
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
$$;

-- 登录结果枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'login_status_enum') THEN
        CREATE TYPE login_status_enum AS ENUM ('success', 'failed', 'blocked');
    END IF;
END;
$$;

-- =============================================================================
-- 用户登录历史表
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_login_history (
    id             BIGSERIAL PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id) COMMENT '用户ID' ON DELETE CASCADE,
    
    -- 登录信息
    login_time     TIMESTAMPTZ NOT NULL COMMENT '登录时间' DEFAULT NOW(),
    logout_time    TIMESTAMPTZ COMMENT '登出时间',
    session_duration INTEGER COMMENT '会话时长（秒）',

    -- 客户端信息
    ip_address     VARCHAR(45) NOT NULL COMMENT 'IP地址',
    user_agent     TEXT COMMENT '用户代理',
    device_type    device_type_enum COMMENT '设备类型',
    browser_name   VARCHAR(100) COMMENT '浏览器名称',
    browser_version VARCHAR(50) COMMENT '浏览器版本',
    os_name        VARCHAR(100) COMMENT '操作系统名称',
    os_version     VARCHAR(50) COMMENT '操作系统版本',

    -- 地理位置
    country        VARCHAR(100) COMMENT '国家',
    region         VARCHAR(100) COMMENT '地区',
    city           VARCHAR(100) COMMENT '城市',

    -- 登录方式
    login_method   login_method_enum NOT NULL DEFAULT 'password' COMMENT '登录方式',
    oauth_provider VARCHAR(50) COMMENT 'OAuth提供商',
    
    -- 安全信息
    is_suspicious  BOOLEAN NOT NULL COMMENT '是否可疑' DEFAULT FALSE,
    risk_score     INTEGER CHECK (risk_score >= 0 AND risk_score <= 100) COMMENT '风险评分',

    -- 状态
    status         login_status_enum NOT NULL COMMENT '登录状态' DEFAULT 'success'
);

-- 登录历史索引
CREATE INDEX idx_login_history_user_id_login_time ON user_login_history(user_id, login_time DESC);
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
    
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id)
);

CREATE TRIGGER update_user_settings_updated_at 
    BEFORE UPDATE ON user_settings 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();