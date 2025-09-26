-- =============================================================================
-- Create Attribute System
-- 属性系统：英雄属性类型定义
-- 依赖：000001_create_core_infrastructure
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 英雄属性类型表 (已在项目中实现)
-- 定义游戏中所有可用的属性类型，如力量、敏捷、智力等
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_attribute_type (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 属性基础信息
    attribute_code          VARCHAR(32) NOT NULL UNIQUE,     -- 属性代码 (如 STRENGTH)
    attribute_name          VARCHAR(64) NOT NULL,            -- 属性名称 (如 "力量")
    category                attribute_category_enum NOT NULL DEFAULT 'basic', -- 属性分类
    data_type              attribute_data_type_enum NOT NULL DEFAULT 'integer', -- 数据类型

    -- 数值范围限制
    min_value              DECIMAL(10,2),                    -- 最小值
    max_value              DECIMAL(10,2),                    -- 最大值
    default_value          DECIMAL(10,2),                    -- 默认值

    -- 高级配置
    calculation_formula    TEXT,                             -- 计算公式 (如 "base_value * level_modifier")
    dependency_attributes  TEXT,                             -- 依赖的其他属性 (JSON 格式)

    -- 显示配置
    icon                   VARCHAR(256),                     -- 图标 URL
    color                  VARCHAR(16),                      -- 颜色代码 (如 #FF5733)
    unit                   VARCHAR(16),                      -- 单位 (如 "点", "%")
    display_order          INTEGER NOT NULL DEFAULT 0,      -- 显示顺序

    -- 状态控制
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,   -- 是否启用
    is_visible             BOOLEAN NOT NULL DEFAULT TRUE,   -- 是否在UI中显示
    description            TEXT,                             -- 属性描述

    -- 时间戳
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ                        -- 软删除
);

-- 英雄属性类型表索引
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_category ON hero_attribute_type(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_data_type ON hero_attribute_type(data_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_is_active ON hero_attribute_type(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_is_visible ON hero_attribute_type(is_visible) WHERE is_visible = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_display_order ON hero_attribute_type(display_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_hero_attribute_type_code_unique ON hero_attribute_type(attribute_code) WHERE deleted_at IS NULL;

-- 英雄属性类型表触发器
CREATE TRIGGER update_hero_attribute_type_updated_at
    BEFORE UPDATE ON hero_attribute_type
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 插入基础属性类型数据
-- --------------------------------------------------------------------------------

INSERT INTO hero_attribute_type (
    attribute_code, attribute_name, category, data_type,
    min_value, max_value, default_value,
    icon, color, unit, display_order,
    description
) VALUES
    -- 基础属性
    ('STRENGTH', '力量', 'basic', 'integer', 1, 100, 10,
     '/icons/strength.png', '#E74C3C', '点', 1,
     '决定物理攻击力和负重能力'),

    ('AGILITY', '敏捷', 'basic', 'integer', 1, 100, 10,
     '/icons/agility.png', '#2ECC71', '点', 2,
     '决定攻击速度、移动速度和闪避能力'),

    ('INTELLIGENCE', '智力', 'basic', 'integer', 1, 100, 10,
     '/icons/intelligence.png', '#3498DB', '点', 3,
     '决定魔法攻击力和魔法值'),

    ('CONSTITUTION', '体质', 'basic', 'integer', 1, 100, 10,
     '/icons/constitution.png', '#E67E22', '点', 4,
     '决定生命值和生命恢复速度'),

    -- 战斗属性
    ('ATTACK_POWER', '攻击力', 'combat', 'integer', 0, 9999, 100,
     '/icons/attack.png', '#C0392B', '点', 10,
     '物理攻击造成的基础伤害'),

    ('MAGIC_POWER', '法术强度', 'combat', 'integer', 0, 9999, 100,
     '/icons/magic.png', '#8E44AD', '点', 11,
     '魔法攻击造成的基础伤害'),

    ('DEFENSE', '防御力', 'combat', 'integer', 0, 9999, 50,
     '/icons/defense.png', '#7F8C8D', '点', 12,
     '减少受到的物理伤害'),

    ('MAGIC_RESISTANCE', '魔法抗性', 'combat', 'integer', 0, 9999, 50,
     '/icons/magic_resist.png', '#9B59B6', '点', 13,
     '减少受到的魔法伤害'),

    -- 百分比属性
    ('CRITICAL_CHANCE', '暴击率', 'combat', 'percentage', 0, 100, 5,
     '/icons/critical.png', '#F39C12', '%', 20,
     '造成暴击的概率'),

    ('CRITICAL_DAMAGE', '暴击伤害', 'combat', 'percentage', 100, 300, 150,
     '/icons/critical_dmg.png', '#D68910', '%', 21,
     '暴击时造成的额外伤害'),

    ('DODGE_CHANCE', '闪避率', 'combat', 'percentage', 0, 95, 5,
     '/icons/dodge.png', '#28B463', '%', 22,
     '闪避攻击的概率'),

    -- 特殊属性
    ('MOVEMENT_SPEED', '移动速度', 'special', 'integer', 1, 10, 3,
     '/icons/speed.png', '#1ABC9C', '格', 30,
     '每回合可移动的格数'),

    ('ACTION_POINTS', '行动点', 'special', 'integer', 1, 10, 2,
     '/icons/action.png', '#F1C40F', '点', 31,
     '每回合可执行的行动次数'),

    -- 抗性属性
    ('FIRE_RESISTANCE', '火焰抗性', 'resistance', 'percentage', -100, 100, 0,
     '/icons/fire_resist.png', '#E74C3C', '%', 40,
     '对火焰伤害的抗性'),

    ('ICE_RESISTANCE', '冰霜抗性', 'resistance', 'percentage', -100, 100, 0,
     '/icons/ice_resist.png', '#5DADE2', '%', 41,
     '对冰霜伤害的抗性'),

    ('LIGHTNING_RESISTANCE', '雷电抗性', 'resistance', 'percentage', -100, 100, 0,
     '/icons/lightning_resist.png', '#F7DC6F', '%', 42,
     '对雷电伤害的抗性')

ON CONFLICT (attribute_code) DO NOTHING;

-- --------------------------------------------------------------------------------
-- 属性标签表 (用于属性分类和筛选)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS attribute_tags (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 标签信息
    tag_code      VARCHAR(32) NOT NULL UNIQUE,              -- 标签代码
    tag_name      VARCHAR(64) NOT NULL,                     -- 标签名称
    color         VARCHAR(16),                              -- 标签颜色
    icon          VARCHAR(256),                             -- 标签图标
    description   TEXT,                                     -- 标签描述

    -- 状态
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,           -- 是否启用
    display_order INTEGER NOT NULL DEFAULT 0,              -- 显示顺序

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ                              -- 软删除
);

-- 属性标签表索引
CREATE INDEX IF NOT EXISTS idx_attribute_tags_is_active ON attribute_tags(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attribute_tags_display_order ON attribute_tags(display_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_attribute_tags_code_unique ON attribute_tags(tag_code) WHERE deleted_at IS NULL;

-- 属性标签表触发器
CREATE TRIGGER update_attribute_tags_updated_at
    BEFORE UPDATE ON attribute_tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 属性类型与标签关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_attribute_type_tags (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attribute_type_id     UUID NOT NULL REFERENCES hero_attribute_type(id) ON DELETE CASCADE,
    tag_id               UUID NOT NULL REFERENCES attribute_tags(id) ON DELETE CASCADE,

    -- 时间戳
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束
    UNIQUE(attribute_type_id, tag_id)
);

-- 属性类型标签关联表索引
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_tags_attribute ON hero_attribute_type_tags(attribute_type_id);
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_tags_tag ON hero_attribute_type_tags(tag_id);

-- --------------------------------------------------------------------------------
-- 插入基础标签数据
-- --------------------------------------------------------------------------------

INSERT INTO attribute_tags (tag_code, tag_name, color, icon, description, display_order) VALUES
    ('COMBAT', '战斗', '#E74C3C', '/icons/combat.png', '与战斗相关的属性', 1),
    ('MAGIC', '魔法', '#8E44AD', '/icons/magic.png', '与魔法相关的属性', 2),
    ('PHYSICAL', '物理', '#34495E', '/icons/physical.png', '与物理能力相关的属性', 3),
    ('DEFENSIVE', '防御', '#7F8C8D', '/icons/defense.png', '与防御相关的属性', 4),
    ('MOBILITY', '机动', '#1ABC9C', '/icons/mobility.png', '与移动和行动相关的属性', 5),
    ('ELEMENTAL', '元素', '#F39C12', '/icons/elemental.png', '与元素伤害相关的属性', 6)
ON CONFLICT (tag_code) DO NOTHING;

-- --------------------------------------------------------------------------------
-- 属性相关函数
-- --------------------------------------------------------------------------------

-- 获取属性类型选项的函数（用于下拉选择）
CREATE OR REPLACE FUNCTION get_attribute_type_options()
RETURNS TABLE (
    id UUID,
    attribute_code VARCHAR,
    attribute_name VARCHAR,
    category attribute_category_enum,
    data_type attribute_data_type_enum
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        hat.id,
        hat.attribute_code,
        hat.attribute_name,
        hat.category,
        hat.data_type
    FROM hero_attribute_type hat
    WHERE hat.is_active = TRUE
      AND hat.deleted_at IS NULL
    ORDER BY hat.display_order ASC, hat.attribute_name ASC;
END;
$$ LANGUAGE plpgsql;

-- 验证属性值是否在有效范围内的函数
CREATE OR REPLACE FUNCTION validate_attribute_value(
    attribute_code_param VARCHAR,
    value_param DECIMAL
) RETURNS BOOLEAN AS $$
DECLARE
    min_val DECIMAL;
    max_val DECIMAL;
    found BOOLEAN := FALSE;
BEGIN
    SELECT min_value, max_value, TRUE
    INTO min_val, max_val, found
    FROM hero_attribute_type
    WHERE attribute_code = attribute_code_param
      AND is_active = TRUE
      AND deleted_at IS NULL;

    -- 如果属性类型不存在，返回 FALSE
    IF NOT found THEN
        RETURN FALSE;
    END IF;

    -- 检查值是否在范围内
    RETURN (min_val IS NULL OR value_param >= min_val)
       AND (max_val IS NULL OR value_param <= max_val);
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Attribute System 创建完成';
    RAISE NOTICE '包含: 属性类型定义、属性标签、基础数据';
    RAISE NOTICE '已插入 % 个属性类型', (SELECT COUNT(*) FROM hero_attribute_type WHERE deleted_at IS NULL);
    RAISE NOTICE '已插入 % 个属性标签', (SELECT COUNT(*) FROM attribute_tags WHERE deleted_at IS NULL);
    RAISE NOTICE '============================================';
END $$;