-- =============================================================================
-- Create Game Runtime Schema Tables
-- 游戏运行时数据：玩家游戏过程中产生的动态数据
-- 依赖：000003_create_users_system, 000004_create_game_config_schema_tables
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 枚举类型定义（game_runtime 专用）
-- --------------------------------------------------------------------------------

-- 英雄状态枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'hero_status_enum') THEN
        CREATE TYPE hero_status_enum AS ENUM (
            'active',     -- 激活状态
            'inactive'   -- 非激活状态
        );
    END IF;
END $$;

-- 装备槽位类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'slot_type_enum') THEN
        CREATE TYPE slot_type_enum AS ENUM (
            'head',           -- 头部
            'eyes',           -- 眼部
            'ears',           -- 耳部
            'neck',           -- 项链
            'cloak',          -- 披风
            'chest',          -- 躯干
            'belt',           -- 腰带
            'shoulder',       -- 肩膀
            'wrist',          -- 手腕
            'gloves',         -- 手套
            'legs',           -- 腿部
            'feet',           -- 脚部
            'ring',           -- 戒指
            'badge',          -- 勋章
            'coat',           -- 外套
            'pocket',         -- 口袋
            'summon_mount',   -- 召唤物/坐骑
            'mainhand',       -- 主手武器
            'offhand',        -- 副手武器
            'twohand',        -- 双手武器
            'special'         -- 特殊
        );
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 英雄表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_runtime.heroes (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    user_id           UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,-- 所属用户（显式指定 schema）

    -- 职业信息
    class_id          UUID NOT NULL REFERENCES game_config.classes(id) ON DELETE RESTRICT,-- 所属职业（显式指定 schema）
    promotion_count   SMALLINT DEFAULT 0 CHECK (promotion_count >= 0), -- 可用转职次数

    -- 英雄基本信息
    hero_name         VARCHAR(64) NOT NULL,                -- 英雄名称
    description       TEXT,                                -- 英雄描述

    -- 英雄等级和经验
    current_level      SMALLINT NOT NULL DEFAULT 1 CHECK (current_level >= 1),-- 英雄等级

    experience_total   BIGINT NOT NULL DEFAULT 0 CHECK (experience_total >= 0),-- 总经验值
    experience_available BIGINT NOT NULL DEFAULT 0 CHECK (experience_available >= 0),-- 当前经验值
    experience_spent BIGINT NOT NULL DEFAULT 0 CHECK (experience_spent >= 0),-- 已使用经验值

    -- 英雄状态
    status           hero_status_enum NOT NULL DEFAULT 'active', -- 英雄状态
    last_login_at     TIMESTAMPTZ,                         -- 上次登录时间

    -- 英雄外观
    avatar_url        VARCHAR(500),                        -- 头像URL

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_battle_at    TIMESTAMPTZ,                         -- 上次战斗时间
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 英雄表索引
CREATE INDEX IF NOT EXISTS idx_heroes_user_id ON game_runtime.heroes(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_class_id ON game_runtime.heroes(class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_current_level ON game_runtime.heroes(current_level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_status ON game_runtime.heroes(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_created_at ON game_runtime.heroes(created_at) WHERE deleted_at IS NULL;

-- 英雄表触发器
CREATE TRIGGER update_heroes_updated_at
    BEFORE UPDATE ON game_runtime.heroes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 英雄技能表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_runtime.hero_skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES game_config.skills(id),
    skill_code VARCHAR(32) NOT NULL,

    -- 技能等级信息
    skill_level INTEGER NOT NULL DEFAULT 1 CHECK (skill_level >= 1),
    skill_experience INTEGER NOT NULL DEFAULT 0 CHECK (skill_experience >= 0),
    max_level INTEGER NOT NULL DEFAULT 10 CHECK (max_level >= 1),

    -- 技能状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_equipped BOOLEAN NOT NULL DEFAULT FALSE,

    -- 获得信息
    learned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    learned_method VARCHAR(50) DEFAULT 'class_unlock',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(hero_id, skill_id),
    UNIQUE(hero_id, skill_code)
);

CREATE INDEX IF NOT EXISTS idx_hero_skills_hero ON game_runtime.hero_skills(hero_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_skills_skill ON game_runtime.hero_skills(skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_skills_equipped ON game_runtime.hero_skills(is_equipped) WHERE is_equipped = TRUE;

CREATE TRIGGER update_hero_skills_updated_at
    BEFORE UPDATE ON game_runtime.hero_skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_runtime.hero_skills IS '英雄技能表';

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Game Runtime Schema Tables 创建完成';
    RAISE NOTICE '包含: 英雄实例、英雄技能等运行时数据';
    RAISE NOTICE '============================================';
END $$;
