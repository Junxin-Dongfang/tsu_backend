-- =============================================================================
-- Fix Table Comments
-- 修正表注释,使描述更准确
-- 依赖：000012_create_equipment_system
-- =============================================================================

-- 修正 player_items 表注释
COMMENT ON TABLE game_runtime.player_items IS '玩家物品实例表 - 存储玩家拥有的所有物品实例(装备、消耗品、材料等)';

-- 修正相关字段注释
COMMENT ON COLUMN game_runtime.player_items.item_id IS '物品配置ID - 关联到 game_config.items 表';
COMMENT ON COLUMN game_runtime.player_items.owner_id IS '所有者ID - 拥有此物品的玩家ID';
COMMENT ON COLUMN game_runtime.player_items.current_durability IS '当前耐久度 - 仅装备类物品有效';
COMMENT ON COLUMN game_runtime.player_items.enhancement_level IS '强化等级 - 仅装备类物品有效';
COMMENT ON COLUMN game_runtime.player_items.socketed_gems IS '镶嵌的宝石 - 仅装备类物品有效,格式: [{"socket_index":0,"gem_item_id":"uuid"}]';
COMMENT ON COLUMN game_runtime.player_items.stack_count IS '堆叠数量 - 仅消耗品和材料有效';

-- 修正 item_drop_records 表注释
COMMENT ON TABLE game_runtime.item_drop_records IS '物品掉落记录表 - 记录所有物品掉落历史(装备、消耗品、材料等)';
COMMENT ON COLUMN game_runtime.item_drop_records.item_instance_id IS '物品实例ID - 关联到 game_runtime.player_items 表';
COMMENT ON COLUMN game_runtime.item_drop_records.item_config_id IS '物品配置ID - 关联到 game_config.items 表';

-- 修正 item_operation_logs 表注释
COMMENT ON TABLE game_runtime.item_operation_logs IS '物品操作日志表 - 记录所有物品操作日志(装备、消耗品、材料等)';
COMMENT ON COLUMN game_runtime.item_operation_logs.item_instance_id IS '物品实例ID - 不设置外键,因为物品可能被删除';
COMMENT ON COLUMN game_runtime.item_operation_logs.operation_type IS '操作类型 - equip(穿戴)/unequip(卸下)/move(移动)/discard(丢弃)/enhance(强化)/repair(修复)/socket(镶嵌)/use(使用)';

-- 修正 items 表注释
COMMENT ON TABLE game_config.items IS '物品配置表 - 存储所有物品的模板配置(装备、消耗品、宝石、材料等)';
COMMENT ON COLUMN game_config.items.item_type IS '物品类型 - equipment(装备)/consumable(消耗品)/gem(宝石)/repair_material(修复材料)/enhancement_material(强化材料)/material(材料)/other(其他)';

