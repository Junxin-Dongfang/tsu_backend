-- =============================================================================
-- Rollback Item Drop System Redesign
-- 回滚装备掉落系统重新设计
-- =============================================================================

-- 删除运行时表
DROP TABLE IF EXISTS game_runtime.world_drop_stats CASCADE;

-- 删除配置表
DROP TABLE IF EXISTS game_config.world_drop_configs CASCADE;
DROP TABLE IF EXISTS game_config.drop_pool_items CASCADE;
DROP TABLE IF EXISTS game_config.drop_pools CASCADE;

-- 恢复旧的掉落配置表
CREATE TABLE IF NOT EXISTS game_config.item_drop_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 掉落配置
    item_id           UUID NOT NULL REFERENCES game_config.items(id) ON DELETE CASCADE,
    drop_source       drop_source_enum NOT NULL,
    source_id         VARCHAR(64),
    min_level         SMALLINT DEFAULT 1,
    max_level         SMALLINT DEFAULT 100,
    drop_rate         DECIMAL(5,4) NOT NULL,
    drop_cooldown     INTEGER DEFAULT 0,

    -- 品质权重配置 (JSON格式)
    quality_weights   JSONB,

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

-- 恢复索引
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_item_id ON game_config.item_drop_configs(item_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_drop_source ON game_config.item_drop_configs(drop_source) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_source_id ON game_config.item_drop_configs(source_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_level_range ON game_config.item_drop_configs(min_level, max_level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_is_active ON game_config.item_drop_configs(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 恢复触发器
CREATE TRIGGER update_item_drop_configs_updated_at
    BEFORE UPDATE ON game_config.item_drop_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

