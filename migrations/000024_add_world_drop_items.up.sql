-- 000024_add_world_drop_items.up.sql
-- 为世界掉落配置引入物品子表，支持一个配置关联多条物品记录。

CREATE TABLE IF NOT EXISTS game_config.world_drop_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    world_drop_config_id UUID NOT NULL REFERENCES game_config.world_drop_configs(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES game_config.items(id) ON DELETE CASCADE,
    drop_rate DECIMAL(10,5),
    drop_weight INTEGER,
    min_quantity INTEGER NOT NULL DEFAULT 1,
    max_quantity INTEGER NOT NULL DEFAULT 1,
    min_level INTEGER,
    max_level INTEGER,
    guaranteed_drop BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 触发器：更新 updated_at
CREATE TRIGGER update_world_drop_items_updated_at
    BEFORE UPDATE ON game_config.world_drop_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 唯一索引：同一配置下不得重复同一物品（忽略已删除）
CREATE UNIQUE INDEX IF NOT EXISTS idx_world_drop_items_unique_item_per_config
    ON game_config.world_drop_items(world_drop_config_id, item_id)
    WHERE deleted_at IS NULL;

-- 查询辅助索引
CREATE INDEX IF NOT EXISTS idx_world_drop_items_config_id
    ON game_config.world_drop_items(world_drop_config_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_world_drop_items_item_id
    ON game_config.world_drop_items(item_id)
    WHERE deleted_at IS NULL;

-- 约束：必须至少提供 drop_rate 或 drop_weight，并且取值合法
ALTER TABLE game_config.world_drop_items
    ADD CONSTRAINT chk_world_drop_item_probability
    CHECK ( (drop_rate IS NOT NULL AND drop_rate > 0 AND drop_rate <= 1)
            OR (drop_weight IS NOT NULL AND drop_weight > 0) );

ALTER TABLE game_config.world_drop_items
    ADD CONSTRAINT chk_world_drop_item_quantity
    CHECK (min_quantity > 0 AND max_quantity > 0 AND min_quantity <= max_quantity);

ALTER TABLE game_config.world_drop_items
    ADD CONSTRAINT chk_world_drop_item_level
    CHECK ( (min_level IS NULL OR max_level IS NULL) OR (min_level <= max_level) );

-- 迁移现有 world_drop_configs 数据到新表，保留旧列用于兼容
INSERT INTO game_config.world_drop_items (
    id, world_drop_config_id, item_id, drop_rate, min_quantity, max_quantity,
    guaranteed_drop, metadata, created_at, updated_at
)
SELECT
    uuid_generate_v7(),
    id,
    item_id,
    base_drop_rate,
    1,
    1,
    FALSE,
    '{}'::jsonb,
    NOW(),
    NOW()
FROM game_config.world_drop_configs
WHERE deleted_at IS NULL
ON CONFLICT DO NOTHING;

COMMENT ON TABLE game_config.world_drop_items IS '世界掉落配置的物品子表';
COMMENT ON COLUMN game_config.world_drop_items.drop_rate IS '概率模式下的掉落概率(0-1]';
COMMENT ON COLUMN game_config.world_drop_items.drop_weight IS '权重模式下的权重值(>0)';
COMMENT ON COLUMN game_config.world_drop_items.guaranteed_drop IS '是否为保底掉落';
