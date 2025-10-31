-- =============================================================================
-- Create Equipment System
-- 装备系统：装备配置、装备实例、装备槽位、装备掉落等
-- 依赖：000004_create_game_config_schema_tables, 000005_create_game_runtime_schema_tables
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 枚举类型定义
-- --------------------------------------------------------------------------------

-- 物品类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'item_type_enum') THEN
        CREATE TYPE item_type_enum AS ENUM (
            'equipment',        -- 装备
            'consumable',       -- 消耗品
            'gem',              -- 宝石
            'repair_material',  -- 修复材料
            'enhancement_material', -- 强化材料
            'quest_item',       -- 任务物品
            'material',         -- 材料
            'other'             -- 其他
        );
    END IF;
END $$;

-- 物品品质枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'item_quality_enum') THEN
        CREATE TYPE item_quality_enum AS ENUM (
            'poor',       -- 劣质
            'normal',     -- 普通
            'fine',       -- 精致
            'excellent',  -- 优良
            'superb',     -- 极佳
            'master',     -- 大师
            'epic',       -- 史诗
            'legendary',  -- 传奇
            'mythic'      -- 神话
        );
    END IF;
END $$;

-- 材料类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'material_type_enum') THEN
        CREATE TYPE material_type_enum AS ENUM (
            'metal',      -- 金属
            'leather',    -- 皮革
            'fabric',     -- 布料
            'wood',       -- 木材
            'gem',        -- 宝石
            'bone',       -- 骨头
            'stone',      -- 石头
            'composite'   -- 复合材料
        );
    END IF;
END $$;

-- 物品位置枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'item_location_enum') THEN
        CREATE TYPE item_location_enum AS ENUM (
            'equipped',         -- 已装备
            'backpack',         -- 背包
            'warehouse',        -- 仓库
            'storage',          -- 储藏室
            'guild_warehouse',  -- 公会仓库
            'market',           -- 市场(上架中)
            'mail'              -- 邮件
        );
    END IF;
END $$;

-- 宝石孔位大小枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'socket_size_enum') THEN
        CREATE TYPE socket_size_enum AS ENUM (
            'small',  -- 小孔
            'large'   -- 大孔
        );
    END IF;
END $$;

-- 宝石颜色枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gem_color_enum') THEN
        CREATE TYPE gem_color_enum AS ENUM (
            'red',      -- 红色
            'yellow',   -- 黄色
            'blue',     -- 蓝色
            'green',    -- 绿色
            'purple',   -- 紫色
            'orange',   -- 橙色
            'white',    -- 白色
            'black'     -- 黑色
        );
    END IF;
END $$;

-- 装备唯一性枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uniqueness_type_enum') THEN
        CREATE TYPE uniqueness_type_enum AS ENUM (
            'none',      -- 无唯一性
            'account',   -- 账户唯一
            'character', -- 角色唯一
            'team',      -- 队伍唯一
            'guild'      -- 公会唯一
        );
    END IF;
END $$;

-- 掉落来源枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'drop_source_enum') THEN
        CREATE TYPE drop_source_enum AS ENUM (
            'dungeon',   -- 副本
            'quest',     -- 任务
            'activity',  -- 活动
            'shop',      -- 商店
            'craft',     -- 制作
            'reward',    -- 奖励
            'other'      -- 其他
        );
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 配置表 (game_config schema)
-- --------------------------------------------------------------------------------

-- 装备/物品配置表
CREATE TABLE IF NOT EXISTS game_config.items (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 基本信息
    item_code         VARCHAR(64) NOT NULL UNIQUE,         -- 物品代码
    item_name         VARCHAR(128) NOT NULL,               -- 物品名称
    item_type         item_type_enum NOT NULL,             -- 物品类型
    item_quality      item_quality_enum NOT NULL,          -- 物品品质
    item_level        SMALLINT NOT NULL DEFAULT 1,         -- 物品等级
    description       TEXT,                                 -- 物品描述
    icon_url          VARCHAR(500),                        -- 图标URL

    -- 装备专用字段
    equip_slot        slot_type_enum,                      -- 可装备位置
    required_class_id UUID REFERENCES game_config.classes(id), -- 职业要求
    required_level    SMALLINT,                             -- 等级要求
    material_type     material_type_enum,                  -- 材料类型
    max_durability    INTEGER,                              -- 最大耐久度
    uniqueness_type   uniqueness_type_enum DEFAULT 'none', -- 唯一性类型

    -- 装备效果 (JSON格式)
    out_of_combat_effects JSONB,                           -- 局外效果(直接加属性)
    in_combat_effects     JSONB,                           -- 局内效果(战斗时触发)
    use_effects           JSONB,                           -- 使用效果(消耗品)
    provided_skills       UUID[],                          -- 提供的技能ID数组

    -- 孔位配置
    socket_type       VARCHAR(10),                         -- 孔位类型(A-I)
    socket_count      SMALLINT DEFAULT 0,                  -- 孔位数量

    -- 强化配置
    enhancement_material_id UUID REFERENCES game_config.items(id), -- 强化材料ID
    enhancement_cost_gold   INTEGER,                       -- 强化金币消耗

    -- 宝石专用字段
    gem_color         gem_color_enum,                      -- 宝石颜色
    gem_size          socket_size_enum,                    -- 宝石大小

    -- 修复材料专用字段
    repair_durability_amount INTEGER,                      -- 修复的耐久度数量
    repair_applicable_quality item_quality_enum[],         -- 适用的装备品质数组
    repair_material_type      material_type_enum,          -- 修复的材料类型

    -- 堆叠和价值
    max_stack_size    INTEGER DEFAULT 1,                   -- 最大堆叠数量
    base_value        INTEGER DEFAULT 0,                   -- 基础价值(金币)
    is_tradable       BOOLEAN DEFAULT TRUE,                -- 是否可交易
    is_droppable      BOOLEAN DEFAULT TRUE,                -- 是否可丢弃

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 装备配置表索引
CREATE INDEX IF NOT EXISTS idx_items_item_code ON game_config.items(item_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_items_item_type ON game_config.items(item_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_items_item_quality ON game_config.items(item_quality) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_items_item_level ON game_config.items(item_level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_items_equip_slot ON game_config.items(equip_slot) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_items_is_active ON game_config.items(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 装备配置表触发器
CREATE TRIGGER update_items_updated_at
    BEFORE UPDATE ON game_config.items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 装备槽位配置表
CREATE TABLE IF NOT EXISTS game_config.equipment_slot_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    class_id          UUID NOT NULL REFERENCES game_config.classes(id) ON DELETE CASCADE, -- 职业ID
    slot_type         slot_type_enum NOT NULL,             -- 槽位类型
    default_count     SMALLINT NOT NULL DEFAULT 1,         -- 默认槽位数量
    max_count         SMALLINT NOT NULL DEFAULT 1,         -- 最大槽位数量
    unlock_level      SMALLINT DEFAULT 1,                  -- 解锁等级

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个职业的每种槽位类型只能有一条配置
    UNIQUE(class_id, slot_type)
);

-- 装备槽位配置表索引
CREATE INDEX IF NOT EXISTS idx_equipment_slot_configs_class_id ON game_config.equipment_slot_configs(class_id);
CREATE INDEX IF NOT EXISTS idx_equipment_slot_configs_slot_type ON game_config.equipment_slot_configs(slot_type);

-- 装备槽位配置表触发器
CREATE TRIGGER update_equipment_slot_configs_updated_at
    BEFORE UPDATE ON game_config.equipment_slot_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 装备掉落配置表
CREATE TABLE IF NOT EXISTS game_config.item_drop_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 掉落配置
    item_id           UUID NOT NULL REFERENCES game_config.items(id) ON DELETE CASCADE, -- 物品ID
    drop_source       drop_source_enum NOT NULL,           -- 掉落来源
    source_id         VARCHAR(64),                         -- 来源ID(副本ID/任务ID等)
    min_level         SMALLINT DEFAULT 1,                  -- 最低等级
    max_level         SMALLINT DEFAULT 100,                -- 最高等级
    drop_rate         DECIMAL(5,4) NOT NULL,               -- 掉落概率(0.0001-1.0000)
    drop_cooldown     INTEGER DEFAULT 0,                   -- 掉落冷却时间(秒)

    -- 品质权重配置 (JSON格式)
    quality_weights   JSONB,                                -- 品质权重 {"poor":50,"normal":30,"fine":15,"excellent":5}

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 装备掉落配置表索引
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_item_id ON game_config.item_drop_configs(item_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_drop_source ON game_config.item_drop_configs(drop_source) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_source_id ON game_config.item_drop_configs(source_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_level_range ON game_config.item_drop_configs(min_level, max_level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_item_drop_configs_is_active ON game_config.item_drop_configs(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 装备掉落配置表触发器
CREATE TRIGGER update_item_drop_configs_updated_at
    BEFORE UPDATE ON game_config.item_drop_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 装备套装配置表
CREATE TABLE IF NOT EXISTS game_config.equipment_set_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 套装信息
    set_code          VARCHAR(64) NOT NULL UNIQUE,         -- 套装代码
    set_name          VARCHAR(128) NOT NULL,               -- 套装名称
    description       TEXT,                                 -- 套装描述
    set_tag_id        UUID REFERENCES game_config.tags(id), -- 套装标签ID(用于匹配装备)

    -- 套装效果配置 (JSON数组)
    -- 格式: [{"piece_count":2,"effects":[...]},{"piece_count":4,"effects":[...]}]
    set_effects       JSONB NOT NULL,                      -- 套装效果配置

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 装备套装配置表索引
CREATE INDEX IF NOT EXISTS idx_equipment_set_configs_set_code ON game_config.equipment_set_configs(set_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_equipment_set_configs_set_tag_id ON game_config.equipment_set_configs(set_tag_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_equipment_set_configs_is_active ON game_config.equipment_set_configs(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 装备套装配置表触发器
CREATE TRIGGER update_equipment_set_configs_updated_at
    BEFORE UPDATE ON game_config.equipment_set_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 宝石效果配置表
CREATE TABLE IF NOT EXISTS game_config.gem_effect_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 宝石组合配置 (JSON数组)
    -- 单个宝石: [{"gem_id":"uuid","count":1}]
    -- 组合宝石: [{"color":"red","size":"large","count":1},{"color":"yellow","size":"small","count":2}]
    gem_combination   JSONB NOT NULL,                      -- 宝石组合

    -- 效果配置 (JSON格式)
    effects           JSONB NOT NULL,                      -- 效果配置

    -- 描述
    description       TEXT,                                 -- 效果描述

    -- 状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 宝石效果配置表索引
CREATE INDEX IF NOT EXISTS idx_gem_effect_configs_is_active ON game_config.gem_effect_configs(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- 宝石效果配置表触发器
CREATE TRIGGER update_gem_effect_configs_updated_at
    BEFORE UPDATE ON game_config.gem_effect_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 孔位类型配置表
CREATE TABLE IF NOT EXISTS game_config.socket_type_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 孔位类型信息
    socket_type_code  VARCHAR(10) NOT NULL UNIQUE,         -- 孔位类型代码(A-I)
    socket_type_name  VARCHAR(64) NOT NULL,                -- 孔位类型名称
    description       TEXT,                                 -- 描述

    -- 孔位配置 (JSON格式)
    -- 格式: {"small":2,"large":1} 表示2个小孔+1个大孔
    socket_layout     JSONB NOT NULL,                      -- 孔位布局

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 孔位类型配置表索引
CREATE INDEX IF NOT EXISTS idx_socket_type_configs_socket_type_code ON game_config.socket_type_configs(socket_type_code) WHERE deleted_at IS NULL;

-- 孔位类型配置表触发器
CREATE TRIGGER update_socket_type_configs_updated_at
    BEFORE UPDATE ON game_config.socket_type_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 运行时表 (game_runtime schema)
-- --------------------------------------------------------------------------------

-- 玩家装备实例表
CREATE TABLE IF NOT EXISTS game_runtime.player_items (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    item_id           UUID NOT NULL REFERENCES game_config.items(id) ON DELETE RESTRICT, -- 装备配置ID
    owner_id          UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,         -- 所有者ID

    -- 来源信息
    source_type       drop_source_enum NOT NULL,           -- 来源类型
    source_id         VARCHAR(64),                         -- 来源ID

    -- 位置信息
    item_location     item_location_enum NOT NULL DEFAULT 'backpack', -- 物品位置
    location_index    INTEGER,                              -- 位置索引(背包格子编号)

    -- 装备实例属性
    current_durability INTEGER,                             -- 当前耐久度
    max_durability_override INTEGER,                        -- 最大耐久度覆盖(如果与配置不同)
    enhancement_level SMALLINT DEFAULT 0,                   -- 强化等级
    used_count        INTEGER DEFAULT 0,                    -- 已使用次数

    -- 镶嵌信息 (JSON数组)
    -- 格式: [{"socket_index":0,"gem_item_id":"uuid"},{"socket_index":1,"gem_item_id":"uuid"}]
    socketed_gems     JSONB,                                -- 镶嵌的宝石

    -- 绑定信息
    is_bound          BOOLEAN DEFAULT FALSE,                -- 是否绑定
    bound_account_id  UUID,                                  -- 绑定的账户ID

    -- 堆叠信息(消耗品)
    stack_count       INTEGER DEFAULT 1,                    -- 堆叠数量

    -- 价值信息
    current_value     INTEGER,                              -- 当前价值
    market_min_price  INTEGER,                              -- 市场最低价

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ                          -- 软删除
);

-- 玩家装备实例表索引
CREATE INDEX IF NOT EXISTS idx_player_items_item_id ON game_runtime.player_items(item_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_player_items_owner_id ON game_runtime.player_items(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_player_items_location ON game_runtime.player_items(item_location) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_player_items_owner_location ON game_runtime.player_items(owner_id, item_location) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_player_items_source ON game_runtime.player_items(source_type, source_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_player_items_created_at ON game_runtime.player_items(created_at) WHERE deleted_at IS NULL;

-- 玩家装备实例表触发器
CREATE TRIGGER update_player_items_updated_at
    BEFORE UPDATE ON game_runtime.player_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 英雄装备槽位表
CREATE TABLE IF NOT EXISTS game_runtime.hero_equipment_slots (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    hero_id           UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE, -- 英雄ID
    slot_type         slot_type_enum NOT NULL,             -- 槽位类型
    slot_index        SMALLINT NOT NULL DEFAULT 0,         -- 槽位索引(同类型槽位的序号,从0开始)

    -- 装备信息
    equipped_item_id  UUID REFERENCES game_runtime.player_items(id) ON DELETE SET NULL, -- 已装备物品ID

    -- 槽位状态
    is_unlocked       BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否已解锁
    unlock_level      SMALLINT DEFAULT 1,                  -- 解锁等级
    added_by_item_id  UUID REFERENCES game_runtime.player_items(id) ON DELETE CASCADE, -- 由哪个装备增加的槽位

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个英雄的每种槽位类型的每个索引只能有一个槽位
    UNIQUE(hero_id, slot_type, slot_index)
);

-- 英雄装备槽位表索引
CREATE INDEX IF NOT EXISTS idx_hero_equipment_slots_hero_id ON game_runtime.hero_equipment_slots(hero_id);
CREATE INDEX IF NOT EXISTS idx_hero_equipment_slots_slot_type ON game_runtime.hero_equipment_slots(slot_type);
CREATE INDEX IF NOT EXISTS idx_hero_equipment_slots_equipped_item_id ON game_runtime.hero_equipment_slots(equipped_item_id) WHERE equipped_item_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_hero_equipment_slots_is_unlocked ON game_runtime.hero_equipment_slots(is_unlocked) WHERE is_unlocked = TRUE;

-- 英雄装备槽位表触发器
CREATE TRIGGER update_hero_equipment_slots_updated_at
    BEFORE UPDATE ON game_runtime.hero_equipment_slots
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 装备掉落记录表
CREATE TABLE IF NOT EXISTS game_runtime.item_drop_records (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 掉落信息
    item_instance_id  UUID NOT NULL REFERENCES game_runtime.player_items(id) ON DELETE CASCADE, -- 装备实例ID
    item_config_id    UUID NOT NULL REFERENCES game_config.items(id) ON DELETE RESTRICT,        -- 装备配置ID

    -- 来源信息
    drop_source       drop_source_enum NOT NULL,           -- 掉落来源
    source_id         VARCHAR(64),                         -- 来源ID(副本ID/任务ID等)

    -- 获得者信息
    receiver_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE, -- 获得者ID
    team_id           UUID,                                 -- 队伍ID(如果是组队)

    -- 掉落时的状态
    player_level      SMALLINT,                            -- 玩家等级
    item_quality      item_quality_enum,                   -- 掉落的品质

    -- 时间戳
    dropped_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 装备掉落记录表索引
CREATE INDEX IF NOT EXISTS idx_item_drop_records_item_instance_id ON game_runtime.item_drop_records(item_instance_id);
CREATE INDEX IF NOT EXISTS idx_item_drop_records_item_config_id ON game_runtime.item_drop_records(item_config_id);
CREATE INDEX IF NOT EXISTS idx_item_drop_records_receiver_id ON game_runtime.item_drop_records(receiver_id);
CREATE INDEX IF NOT EXISTS idx_item_drop_records_drop_source ON game_runtime.item_drop_records(drop_source);
CREATE INDEX IF NOT EXISTS idx_item_drop_records_dropped_at ON game_runtime.item_drop_records(dropped_at);

-- 装备操作日志表
CREATE TABLE IF NOT EXISTS game_runtime.item_operation_logs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 操作信息
    item_instance_id  UUID NOT NULL,                       -- 装备实例ID(不设置外键,因为装备可能被删除)
    operation_type    VARCHAR(32) NOT NULL,                -- 操作类型(equip/unequip/move/discard/enhance/repair/socket)
    operator_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE, -- 操作者ID

    -- 操作前后状态 (JSON格式)
    state_before      JSONB,                                -- 操作前状态
    state_after       JSONB,                                -- 操作后状态

    -- 操作结果
    is_success        BOOLEAN NOT NULL,                    -- 是否成功
    error_message     TEXT,                                 -- 错误信息(如果失败)

    -- 时间戳
    operated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 装备操作日志表索引
CREATE INDEX IF NOT EXISTS idx_item_operation_logs_item_instance_id ON game_runtime.item_operation_logs(item_instance_id);
CREATE INDEX IF NOT EXISTS idx_item_operation_logs_operation_type ON game_runtime.item_operation_logs(operation_type);
CREATE INDEX IF NOT EXISTS idx_item_operation_logs_operator_id ON game_runtime.item_operation_logs(operator_id);
CREATE INDEX IF NOT EXISTS idx_item_operation_logs_operated_at ON game_runtime.item_operation_logs(operated_at);
CREATE INDEX IF NOT EXISTS idx_item_operation_logs_is_success ON game_runtime.item_operation_logs(is_success) WHERE is_success = FALSE;

-- --------------------------------------------------------------------------------
-- 注释
-- --------------------------------------------------------------------------------

COMMENT ON TABLE game_config.items IS '装备/物品配置表 - 存储所有物品的模板配置';
COMMENT ON TABLE game_config.equipment_slot_configs IS '装备槽位配置表 - 定义每个职业的装备槽位';
COMMENT ON TABLE game_config.item_drop_configs IS '装备掉落配置表 - 定义装备掉落规则';
COMMENT ON TABLE game_config.equipment_set_configs IS '装备套装配置表 - 定义装备套装效果';
COMMENT ON TABLE game_config.gem_effect_configs IS '宝石效果配置表 - 定义宝石和宝石组合效果';
COMMENT ON TABLE game_config.socket_type_configs IS '孔位类型配置表 - 定义孔位类型和布局';

COMMENT ON TABLE game_runtime.player_items IS '玩家装备实例表 - 存储玩家拥有的所有物品实例';
COMMENT ON TABLE game_runtime.hero_equipment_slots IS '英雄装备槽位表 - 存储每个英雄的装备槽位';
COMMENT ON TABLE game_runtime.item_drop_records IS '装备掉落记录表 - 记录所有装备掉落历史';
COMMENT ON TABLE game_runtime.item_operation_logs IS '装备操作日志表 - 记录所有装备操作日志';

COMMENT ON COLUMN game_config.items.out_of_combat_effects IS '局外效果 - 直接影响英雄属性的效果,格式: [{"Data_type":"Status","Data_ID":"MAX_HP","Bouns_type":"bonus","Bouns_Number":"5"}]';
COMMENT ON COLUMN game_config.items.in_combat_effects IS '局内效果 - 战斗时触发的效果,格式: [{"Effect":"APPLY_BUFF","params":{...}}]';
COMMENT ON COLUMN game_config.items.socket_type IS '孔位类型 - A-I,对应socket_type_configs表的socket_type_code';
COMMENT ON COLUMN game_config.equipment_set_configs.set_effects IS '套装效果 - 格式: [{"piece_count":2,"effects":[...]},{"piece_count":4,"effects":[...]}]';
COMMENT ON COLUMN game_config.gem_effect_configs.gem_combination IS '宝石组合 - 格式: [{"gem_id":"uuid","count":1}] 或 [{"color":"red","size":"large","count":1}]';
COMMENT ON COLUMN game_runtime.player_items.socketed_gems IS '镶嵌的宝石 - 格式: [{"socket_index":0,"gem_item_id":"uuid"}]';
COMMENT ON COLUMN game_runtime.hero_equipment_slots.added_by_item_id IS '由哪个装备增加的槽位 - 如果是装备效果增加的槽位,记录装备ID,卸下装备时删除槽位';

