-- =============================================================================
-- Redesign Item Drop System
-- 重新设计装备掉落系统,支持多层次掉落机制
-- 依赖：000012_create_equipment_system
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除旧的掉落配置表
-- --------------------------------------------------------------------------------

DROP TABLE IF EXISTS game_config.item_drop_configs CASCADE;

-- --------------------------------------------------------------------------------
-- 新的掉落系统表
-- --------------------------------------------------------------------------------

-- 掉落池配置表
CREATE TABLE IF NOT EXISTS game_config.drop_pools (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 掉落池信息
    pool_code         VARCHAR(64) NOT NULL UNIQUE,         -- 掉落池代码
    pool_name         VARCHAR(128) NOT NULL,               -- 掉落池名称
    pool_type         VARCHAR(32) NOT NULL,                -- 掉落池类型(monster/world/event/boss)
    description       TEXT,                                 -- 描述

    -- 掉落规则
    min_drops         SMALLINT DEFAULT 0,                  -- 最少掉落数量
    max_drops         SMALLINT DEFAULT 1,                  -- 最多掉落数量
    guaranteed_drops  SMALLINT DEFAULT 0,                  -- 保底掉落数量

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 掉落池配置表索引
CREATE INDEX IF NOT EXISTS idx_drop_pools_pool_code ON game_config.drop_pools(pool_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_drop_pools_pool_type ON game_config.drop_pools(pool_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_drop_pools_is_active ON game_config.drop_pools(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 掉落池配置表触发器
CREATE TRIGGER update_drop_pools_updated_at
    BEFORE UPDATE ON game_config.drop_pools
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 掉落池物品配置表
CREATE TABLE IF NOT EXISTS game_config.drop_pool_items (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    drop_pool_id      UUID NOT NULL REFERENCES game_config.drop_pools(id) ON DELETE CASCADE, -- 掉落池ID
    item_id           UUID NOT NULL REFERENCES game_config.items(id) ON DELETE CASCADE,      -- 物品ID

    -- 掉落概率和权重
    drop_weight       INTEGER NOT NULL DEFAULT 1,          -- 掉落权重(用于计算概率)
    drop_rate         DECIMAL(5,4),                        -- 固定掉落概率(0.0001-1.0000,如果设置则忽略权重)

    -- 品质随机配置 (JSON格式)
    -- 格式: {"poor":50,"normal":30,"fine":15,"excellent":5}
    quality_weights   JSONB,                                -- 品质权重(如果为空则使用物品配置的品质)

    -- 数量配置
    min_quantity      INTEGER DEFAULT 1,                   -- 最少数量(消耗品)
    max_quantity      INTEGER DEFAULT 1,                   -- 最多数量(消耗品)

    -- 等级限制
    min_level         SMALLINT,                             -- 最低等级要求
    max_level         SMALLINT,                             -- 最高等级要求

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 掉落池物品配置表索引
CREATE INDEX IF NOT EXISTS idx_drop_pool_items_drop_pool_id ON game_config.drop_pool_items(drop_pool_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_drop_pool_items_item_id ON game_config.drop_pool_items(item_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_drop_pool_items_level_range ON game_config.drop_pool_items(min_level, max_level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_drop_pool_items_is_active ON game_config.drop_pool_items(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 掉落池物品配置表触发器
CREATE TRIGGER update_drop_pool_items_updated_at
    BEFORE UPDATE ON game_config.drop_pool_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 世界掉落配置表
CREATE TABLE IF NOT EXISTS game_config.world_drop_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    item_id           UUID NOT NULL REFERENCES game_config.items(id) ON DELETE CASCADE, -- 物品ID

    -- 全局限制
    total_drop_limit  INTEGER,                             -- 全局掉落总数限制(NULL表示无限制)
    daily_drop_limit  INTEGER,                             -- 每日掉落数量限制
    hourly_drop_limit INTEGER,                             -- 每小时掉落数量限制

    -- 掉落间隔
    min_drop_interval INTEGER,                             -- 最小掉落间隔(秒)
    max_drop_interval INTEGER,                             -- 最大掉落间隔(秒)

    -- 触发条件 (JSON格式)
    -- 格式: {"type":"level_range","min_level":10,"max_level":20}
    -- 或: {"type":"dungeon_type","dungeon_types":["elite","boss"]}
    -- 或: {"type":"event","event_id":"xxx"}
    trigger_conditions JSONB,                              -- 触发条件

    -- 掉落概率
    base_drop_rate    DECIMAL(5,4) NOT NULL,              -- 基础掉落概率
    
    -- 掉落概率修正因子 (JSON格式)
    -- 格式: {"dungeon_level_bonus":0.01,"team_size_bonus":0.05}
    drop_rate_modifiers JSONB,                             -- 掉落概率修正因子

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 世界掉落配置表索引
CREATE INDEX IF NOT EXISTS idx_world_drop_configs_item_id ON game_config.world_drop_configs(item_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_world_drop_configs_is_active ON game_config.world_drop_configs(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 世界掉落配置表触发器
CREATE TRIGGER update_world_drop_configs_updated_at
    BEFORE UPDATE ON game_config.world_drop_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 世界掉落统计表 (运行时表)
CREATE TABLE IF NOT EXISTS game_runtime.world_drop_stats (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    world_drop_config_id UUID NOT NULL REFERENCES game_config.world_drop_configs(id) ON DELETE CASCADE, -- 世界掉落配置ID
    item_id           UUID NOT NULL REFERENCES game_config.items(id) ON DELETE RESTRICT,                -- 物品ID

    -- 统计信息
    total_dropped     INTEGER NOT NULL DEFAULT 0,          -- 总掉落数量
    last_drop_at      TIMESTAMPTZ,                         -- 上次掉落时间
    daily_dropped     INTEGER NOT NULL DEFAULT 0,          -- 今日掉落数量
    daily_reset_at    TIMESTAMPTZ,                         -- 每日重置时间
    hourly_dropped    INTEGER NOT NULL DEFAULT 0,          -- 本小时掉落数量
    hourly_reset_at   TIMESTAMPTZ,                         -- 每小时重置时间

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束
    UNIQUE(world_drop_config_id)
);

-- 世界掉落统计表索引
CREATE INDEX IF NOT EXISTS idx_world_drop_stats_item_id ON game_runtime.world_drop_stats(item_id);
CREATE INDEX IF NOT EXISTS idx_world_drop_stats_last_drop_at ON game_runtime.world_drop_stats(last_drop_at);

-- 世界掉落统计表触发器
CREATE TRIGGER update_world_drop_stats_updated_at
    BEFORE UPDATE ON game_runtime.world_drop_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 注释
-- --------------------------------------------------------------------------------

COMMENT ON TABLE game_config.drop_pools IS '掉落池配置表 - 定义各种掉落池(怪物掉落池、世界掉落池、事件掉落池等)';
COMMENT ON TABLE game_config.drop_pool_items IS '掉落池物品配置表 - 定义掉落池中的物品及其掉落概率';
COMMENT ON TABLE game_config.world_drop_configs IS '世界掉落配置表 - 定义全局掉落限制和触发机制';
COMMENT ON TABLE game_runtime.world_drop_stats IS '世界掉落统计表 - 记录世界掉落的统计信息,用于限制掉落数量和间隔';

COMMENT ON COLUMN game_config.drop_pools.pool_type IS '掉落池类型 - monster(怪物掉落池)/world(世界掉落池)/event(事件掉落池)/boss(Boss掉落池)';
COMMENT ON COLUMN game_config.drop_pools.guaranteed_drops IS '保底掉落数量 - 至少掉落多少个物品';
COMMENT ON COLUMN game_config.drop_pool_items.drop_weight IS '掉落权重 - 用于计算相对概率,权重越高掉落概率越大';
COMMENT ON COLUMN game_config.drop_pool_items.drop_rate IS '固定掉落概率 - 如果设置则忽略权重,直接使用此概率';
COMMENT ON COLUMN game_config.world_drop_configs.trigger_conditions IS '触发条件 - JSON格式,定义何时可以掉落此物品';
COMMENT ON COLUMN game_config.world_drop_configs.drop_rate_modifiers IS '掉落概率修正因子 - JSON格式,根据不同条件修正掉落概率';
COMMENT ON COLUMN game_runtime.world_drop_stats.total_dropped IS '总掉落数量 - 用于限制全局掉落总数';
COMMENT ON COLUMN game_runtime.world_drop_stats.last_drop_at IS '上次掉落时间 - 用于限制掉落间隔';

