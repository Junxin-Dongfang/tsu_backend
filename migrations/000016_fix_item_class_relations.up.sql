-- 修复物品职业限制设计
-- 使用关联表实现多对多关系，支持通用装备、单职业装备、多职业装备

-- 1. 创建物品-职业关联表
CREATE TABLE IF NOT EXISTS game_config.item_class_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL,
    class_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT item_class_relations_item_id_fkey FOREIGN KEY (item_id) REFERENCES game_config.items(id) ON DELETE CASCADE,
    CONSTRAINT item_class_relations_class_id_fkey FOREIGN KEY (class_id) REFERENCES game_config.classes(id) ON DELETE CASCADE,
    CONSTRAINT item_class_relations_unique UNIQUE (item_id, class_id)
);

-- 2. 添加索引以优化查询性能
CREATE INDEX IF NOT EXISTS idx_item_class_relations_item_id ON game_config.item_class_relations(item_id);
CREATE INDEX IF NOT EXISTS idx_item_class_relations_class_id ON game_config.item_class_relations(class_id);

-- 3. 添加表注释
COMMENT ON TABLE game_config.item_class_relations IS '物品-职业关联表：定义哪些职业可以使用哪些物品。没有关联记录表示通用装备（所有职业都能使用）';
COMMENT ON COLUMN game_config.item_class_relations.id IS '主键';
COMMENT ON COLUMN game_config.item_class_relations.item_id IS '物品ID';
COMMENT ON COLUMN game_config.item_class_relations.class_id IS '职业ID';
COMMENT ON COLUMN game_config.item_class_relations.created_at IS '创建时间';

-- 4. 迁移现有数据：将 items.required_class_id 迁移到关联表
INSERT INTO game_config.item_class_relations (item_id, class_id)
SELECT id, required_class_id 
FROM game_config.items 
WHERE required_class_id IS NOT NULL
ON CONFLICT (item_id, class_id) DO NOTHING;

-- 5. 删除旧字段
ALTER TABLE game_config.items DROP COLUMN IF EXISTS required_class_id;

-- 6. 授予权限
GRANT ALL PRIVILEGES ON game_config.item_class_relations TO tsu_admin_user;
GRANT SELECT ON game_config.item_class_relations TO tsu_game_user;
GRANT ALL PRIVILEGES ON game_config.item_class_relations TO postgres;

