-- 000028_complete_warehouse_system.up.sql
-- 增强仓库/背包生产级保障：唯一约束、堆叠限制、容量配置、入库审计、索引优化

-- 1) 仓库物品唯一约束与堆叠上限
ALTER TABLE game_runtime.team_warehouse_items
    ADD CONSTRAINT uq_team_warehouse_items_warehouse_item UNIQUE (warehouse_id, item_id);

ALTER TABLE game_runtime.team_warehouse_items
    ADD CONSTRAINT check_team_warehouse_items_quantity_max CHECK (quantity <= 999);

-- 2) 分配历史查询优化：接收者索引
CREATE INDEX IF NOT EXISTS idx_loot_history_recipient ON game_runtime.team_loot_distribution_history(recipient_hero_id);

-- 3) 入库审计日志表
CREATE TABLE IF NOT EXISTS game_runtime.team_warehouse_loot_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES game_runtime.team_warehouses(id) ON DELETE CASCADE,
    source_dungeon_id UUID REFERENCES game_config.dungeons(id),
    gold_amount BIGINT NOT NULL DEFAULT 0 CHECK (gold_amount >= 0),
    items JSONB NOT NULL DEFAULT '[]', -- [{item_id, item_type, quantity}]
    result TEXT NOT NULL,             -- success / failed
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE game_runtime.team_warehouse_loot_log IS '团队仓库战利品入库审计日志';
COMMENT ON COLUMN game_runtime.team_warehouse_loot_log.items IS '入库物品明细 JSON 数组';
COMMENT ON COLUMN game_runtime.team_warehouse_loot_log.result IS '入库结果：success/failed';

CREATE INDEX IF NOT EXISTS idx_warehouse_loot_log_wh_created ON game_runtime.team_warehouse_loot_log(warehouse_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_warehouse_loot_log_dungeon ON game_runtime.team_warehouse_loot_log(source_dungeon_id);

-- 4) 背包/仓库容量配置表
CREATE TABLE IF NOT EXISTS game_config.inventory_capacities (
    location item_location_enum PRIMARY KEY,
    max_slots INT NOT NULL CHECK (max_slots > 0),
    max_stack INT NOT NULL CHECK (max_stack > 0)
);

COMMENT ON TABLE game_config.inventory_capacities IS '背包/仓库/储物位置容量与堆叠上限配置';

-- 默认配置（可按运营需要调整）
INSERT INTO game_config.inventory_capacities (location, max_slots, max_stack)
VALUES
    ('backpack', 120, 999),
    ('warehouse', 300, 999),
    ('storage', 200, 999)
ON CONFLICT (location) DO NOTHING;

-- 5) 英雄钱包
CREATE TABLE IF NOT EXISTS game_runtime.hero_wallets (
    hero_id UUID PRIMARY KEY REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    gold_amount BIGINT NOT NULL DEFAULT 0 CHECK (gold_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE game_runtime.hero_wallets IS '英雄钱包表，存储英雄金币余额';

-- 6) 背包归属强化：为 player_items 增加 hero_id 与位置约束 + 部分索引
ALTER TABLE game_runtime.player_items
    ADD COLUMN IF NOT EXISTS hero_id UUID;

-- 当在背包/已装备位置时要求 hero_id 非空；NOT VALID 以兼容存量，后续可 VALIDATE
ALTER TABLE game_runtime.player_items
    ADD CONSTRAINT check_player_items_hero_location
    CHECK (
        (item_location IN ('backpack','equipped') AND hero_id IS NOT NULL)
        OR (item_location NOT IN ('backpack','equipped'))
    ) NOT VALID;

-- 部分索引提升查询性能（大表场景）
CREATE INDEX IF NOT EXISTS idx_player_items_hero_location
    ON game_runtime.player_items(hero_id, item_location)
    WHERE item_location IN ('backpack','equipped') AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_player_items_owner_location
    ON game_runtime.player_items(owner_id, item_location)
    WHERE item_location NOT IN ('backpack','equipped') AND deleted_at IS NULL;
