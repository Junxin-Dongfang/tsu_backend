-- 回滚物品职业限制设计修复

-- 1. 重新添加 required_class_id 字段
ALTER TABLE game_config.items ADD COLUMN IF NOT EXISTS required_class_id UUID;

-- 2. 从关联表迁移回单个职业ID（只取第一个关联的职业）
UPDATE game_config.items i
SET required_class_id = (
    SELECT class_id 
    FROM game_config.item_class_relations 
    WHERE item_id = i.id 
    LIMIT 1
)
WHERE EXISTS (
    SELECT 1 
    FROM game_config.item_class_relations 
    WHERE item_id = i.id
);

-- 3. 重新添加外键约束
ALTER TABLE game_config.items 
ADD CONSTRAINT items_required_class_id_fkey 
FOREIGN KEY (required_class_id) REFERENCES game_config.classes(id);

-- 4. 删除关联表
DROP TABLE IF EXISTS game_config.item_class_relations;

