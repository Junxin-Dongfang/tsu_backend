-- =============================================================================
-- Create Attribute System
-- 属性系统：英雄属性类型定义
-- 依赖：000001_create_core_infrastructure
-- =============================================================================

-- 属性分类枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'attribute_category_enum') THEN
        CREATE TYPE attribute_category_enum AS ENUM (
            'basic',      -- 基础属性
            'derived',    -- 派生属性
            'resistance'  -- 抗性属性
        );
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 英雄属性类型表
-- 定义游戏中所有可用的属性类型，如力量、敏捷、智力等
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_attribute_type (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 属性基础信息
    attribute_code          VARCHAR(32) NOT NULL UNIQUE,     -- 属性代码 (如 STRENGTH)
    attribute_name          VARCHAR(64) NOT NULL,            -- 属性名称 (如 "力量")
    category                attribute_category_enum NOT NULL DEFAULT 'basic', -- 属性分类
    data_type              data_type_enum NOT NULL DEFAULT 'integer', -- 数据类型

    -- 数值范围限制
    min_value              DECIMAL(10,2),                    -- 最小值
    max_value              DECIMAL(10,2),                    -- 最大值
    default_value          DECIMAL(10,2),                    -- 默认值

    -- 高级配置
    calculation_formula    TEXT,                             -- 计算公式 (如 "base_value * level_modifier")
    attribute_dependencies  JSONB,                             -- 依赖的其他属性 (JSON 格式)

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

-- 获取属性类型选项的函数（用于下拉选择）
CREATE OR REPLACE FUNCTION get_attribute_type_options()
RETURNS TABLE (
    id UUID,
    attribute_code VARCHAR,
    attribute_name VARCHAR,
    category attribute_category_enum,
    data_type data_type_enum
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
    RAISE NOTICE '包含: 属性类型定义、基础数据';
    RAISE NOTICE '已插入 % 个属性类型', (SELECT COUNT(*) FROM hero_attribute_type WHERE deleted_at IS NULL);
    RAISE NOTICE '============================================';
END $$;