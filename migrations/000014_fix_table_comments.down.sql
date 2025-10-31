-- =============================================================================
-- Rollback Table Comments Fix
-- 回滚表注释修正
-- =============================================================================

-- 恢复原来的注释
COMMENT ON TABLE game_runtime.player_items IS '玩家装备实例表 - 存储玩家拥有的所有物品实例';
COMMENT ON TABLE game_runtime.item_drop_records IS '装备掉落记录表 - 记录所有装备掉落历史';
COMMENT ON TABLE game_runtime.item_operation_logs IS '装备操作日志表 - 记录所有装备操作日志';
COMMENT ON TABLE game_config.items IS '装备/物品配置表 - 存储所有物品的模板配置';

