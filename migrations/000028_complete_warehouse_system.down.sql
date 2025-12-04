-- 000028_complete_warehouse_system.down.sql

-- 4) 删除容量配置表
DROP TABLE IF EXISTS game_config.inventory_capacities;

-- 3) 删除入库审计日志表
DROP TABLE IF EXISTS game_runtime.team_warehouse_loot_log;

-- 2) 移除接收者索引
DROP INDEX IF EXISTS idx_loot_history_recipient;

-- 1) 移除仓库物品约束
ALTER TABLE game_runtime.team_warehouse_items
    DROP CONSTRAINT IF EXISTS check_team_warehouse_items_quantity_max;

ALTER TABLE game_runtime.team_warehouse_items
    DROP CONSTRAINT IF EXISTS uq_team_warehouse_items_warehouse_item;

-- 5) 删除英雄钱包
DROP TABLE IF EXISTS game_runtime.hero_wallets;

-- 6) 回滚 player_items 扩展
DROP INDEX IF EXISTS idx_player_items_hero_location;
DROP INDEX IF EXISTS idx_player_items_owner_location;
ALTER TABLE game_runtime.player_items DROP CONSTRAINT IF EXISTS check_player_items_hero_location;
ALTER TABLE game_runtime.player_items DROP COLUMN IF EXISTS hero_id;
