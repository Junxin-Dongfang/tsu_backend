-- =============================================================================
-- 英雄设计系统
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 属性类型定义表
-- -----------------------------------------------------------------------------

-- 属性分类枚举
CREATE TYPE attribute_category_enum AS ENUM ('base', 'derived', 'resistance');

--属性数据类型枚举
CREATE TYPE attribute_data_type_enum AS ENUM ('integer', 'percentage');

CREATE TABLE IF NOT EXISTS hero_attribute_type
(
    id          UUID         PRIMARY KEY DEFAULT (UUID()) COMMENT '属性类型唯一标识',
    attribute_code        VARCHAR(32)  NOT NULL UNIQUE COMMENT '属性类型代码',
    attribute_name        VARCHAR(64) NOT NULL UNIQUE COMMENT '属性类型名称',

    category    attribute_category_enum NOT NULL COMMENT '属性分类',
    data_type   attribute_data_type_enum NOT NULL COMMENT '属性数据类型',

    min_value   decimal(10, 2)      COMMENT '属性最小值',
    max_value   decimal(10, 2)      COMMENT '属性最大值',
    default_value decimal(10, 2)    COMMENT '属性默认值',

    calculation_formula TEXT COMMENT '计算公式, 用于除了base属性的计算',
    dependency_attributes TEXT COMMENT '依赖的属性代码列表, 用于计算公式中',

    icon        VARCHAR(256) COMMENT '属性类型图标URL',
    color      VARCHAR(16)  COMMENT '属性类型颜色值',
    unit       VARCHAR(16)  COMMENT '属性单位',
    display_order SMALLINT          DEFAULT 0 COMMENT '显示顺序',
    is_active   BOOLEAN      DEFAULT TRUE COMMENT '是否启用',
    is_visible  BOOLEAN      DEFAULT TRUE COMMENT '是否在UI中显示',

    description TEXT         COMMENT '属性类型描述',
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP
) COMMENT '英雄属性类型表';

--属性类型索引
CREATE INDEX idx_hero_attribute_type_category ON hero_attribute_type(category);
CREATE INDEX idx_hero_attribute_type_is_active ON hero_attribute_type(is_active);
CREATE INDEX idx_hero_attribute_type_is_visible ON hero_attribute_type(is_visible);
CREATE INDEX idx_hero_attribute_type_display_order ON hero_attribute_type(display_order);
CREATE UNIQUE INDEX idx_hero_attribute_type_code_name ON hero_attribute_type(attribute_code) WHERE deleted_at IS NULL;

-- 触发器: 自动更新更新时间
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- -----------------------------------------------------------------------------
-- 标签定义表
-- -----------------------------------------------------------------------------
--标签类型枚举
CREATE TYPE tag_type_enum AS ENUM ('class', 'skill', 'equipment');

CREATE TABLE IF NOT EXISTS tags (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 标签信息
    tag_code      VARCHAR(32) NOT NULL UNIQUE COMMENT '标签代码',
    tag_name      VARCHAR(64) NOT NULL COMMENT '标签名称',
    description   TEXT COMMENT '标签描述',

    tag_type      tag_type_enum NOT NULL COMMENT '标签类型',
    color         VARCHAR(16) COMMENT '标签颜色值',
    icon          VARCHAR(256) COMMENT '标签图标URL',
    is_active     BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    display_order SMALLINT DEFAULT 0 COMMENT '显示顺序',

    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at    TIMESTAMP WITH TIME ZONE COMMENT '删除时间'
) COMMENT '标签表';

-- 标签索引
CREATE INDEX idx_tags_tag_type ON tags(tag_type);
CREATE INDEX idx_tags_is_active ON tags(is_active);
CREATE INDEX idx_tags_display_order ON tags(display_order);
CREATE UNIQUE INDEX idx_tags_tag_type_code ON tags(tag_code) WHERE deleted_at IS NULL;

-- 触发器: 自动更新更新时间
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;