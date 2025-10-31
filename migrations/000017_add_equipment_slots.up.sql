-- 000017_add_equipment_slots.up.sql
-- 创建装备槽位配置表

-- 1. 创建槽位类型枚举
CREATE TYPE game_config.slot_type_enum AS ENUM (
    'weapon',    -- 武器槽位（主手、副手、双手）
    'armor',     -- 护甲槽位（头、胸、腿、手、脚等）
    'accessory', -- 饰品槽位（项链、戒指、徽章等）
    'special'    -- 特殊槽位（坐骑、宠物等扩展槽位）
);

-- 2. 创建装备槽位表
CREATE TABLE game_config.equipment_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_code VARCHAR(50) UNIQUE NOT NULL,      -- 槽位代码，如 "mainhand", "head"
    slot_name VARCHAR(100) NOT NULL,            -- 槽位名称，如 "主手", "头部"
    slot_type game_config.slot_type_enum NOT NULL, -- 槽位类型
    display_order INT NOT NULL DEFAULT 0,       -- UI显示顺序
    icon VARCHAR(255),                          -- 槽位图标路径
    description TEXT,                           -- 槽位描述
    is_active BOOLEAN NOT NULL DEFAULT true,    -- 是否启用
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP                        -- 软删除时间戳
);

-- 3. 创建索引
CREATE INDEX idx_equipment_slots_slot_type ON game_config.equipment_slots(slot_type);
CREATE INDEX idx_equipment_slots_is_active ON game_config.equipment_slots(is_active);
CREATE INDEX idx_equipment_slots_deleted_at ON game_config.equipment_slots(deleted_at);
CREATE INDEX idx_equipment_slots_display_order ON game_config.equipment_slots(display_order);

-- 4. 添加注释
COMMENT ON TABLE game_config.equipment_slots IS '装备槽位配置表';
COMMENT ON COLUMN game_config.equipment_slots.slot_code IS '槽位代码（唯一标识）';
COMMENT ON COLUMN game_config.equipment_slots.slot_name IS '槽位显示名称';
COMMENT ON COLUMN game_config.equipment_slots.slot_type IS '槽位类型分类';
COMMENT ON COLUMN game_config.equipment_slots.display_order IS 'UI显示顺序（数值越小越靠前）';
COMMENT ON COLUMN game_config.equipment_slots.icon IS '槽位图标路径';
COMMENT ON COLUMN game_config.equipment_slots.description IS '槽位描述';
COMMENT ON COLUMN game_config.equipment_slots.is_active IS '是否启用';
COMMENT ON COLUMN game_config.equipment_slots.deleted_at IS '软删除时间戳';

-- 5. 插入初始槽位数据（标准装备槽位）

-- 武器槽位
INSERT INTO game_config.equipment_slots (slot_code, slot_name, slot_type, display_order, description) VALUES
('mainhand', '主手', 'weapon', 10, '主手武器槽位'),
('offhand', '副手', 'weapon', 20, '副手武器或盾牌槽位'),
('twohand', '双手', 'weapon', 15, '双手武器槽位（占用主手和副手）');

-- 护甲槽位
INSERT INTO game_config.equipment_slots (slot_code, slot_name, slot_type, display_order, description) VALUES
('head', '头部', 'armor', 30, '头盔槽位'),
('chest', '胸部', 'armor', 40, '胸甲槽位'),
('legs', '腿部', 'armor', 50, '腿甲槽位'),
('hands', '手部', 'armor', 60, '手套槽位'),
('feet', '脚部', 'armor', 70, '靴子槽位'),
('shoulders', '肩部', 'armor', 35, '肩甲槽位'),
('waist', '腰部', 'armor', 45, '腰带槽位');

-- 饰品槽位
INSERT INTO game_config.equipment_slots (slot_code, slot_name, slot_type, display_order, description) VALUES
('neck', '项链', 'accessory', 80, '项链槽位'),
('ring1', '戒指1', 'accessory', 90, '第一个戒指槽位'),
('ring2', '戒指2', 'accessory', 100, '第二个戒指槽位'),
('trinket1', '饰品1', 'accessory', 110, '第一个饰品槽位'),
('trinket2', '饰品2', 'accessory', 120, '第二个饰品槽位'),
('back', '披风', 'accessory', 75, '披风槽位');

-- 特殊槽位（可选，用于扩展）
INSERT INTO game_config.equipment_slots (slot_code, slot_name, slot_type, display_order, description, is_active) VALUES
('mount', '坐骑', 'special', 200, '坐骑槽位', false),
('pet', '宠物', 'special', 210, '宠物槽位', false),
('costume', '时装', 'special', 220, '时装槽位', false);

