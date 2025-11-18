-- 000024_add_world_drop_items.down.sql
-- 回滚世界掉落物品子表

-- 回填 world_drop_configs.item_id（若存在对应条目则保留原值）
UPDATE game_config.world_drop_configs AS cfg
SET item_id = sub.item_id
FROM (
    SELECT world_drop_config_id, item_id
    FROM game_config.world_drop_items
    WHERE deleted_at IS NULL
    ORDER BY created_at ASC
) AS sub
WHERE cfg.id = sub.world_drop_config_id;

DROP TRIGGER IF EXISTS update_world_drop_items_updated_at ON game_config.world_drop_items;
DROP TABLE IF EXISTS game_config.world_drop_items;
