-- =============================================================================
-- Create Game Config Schema Tables
-- 游戏配置数据：所有策划/运营通过后台管理的静态配置数据
-- 依赖：000002_create_core_infrastructure
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 枚举类型定义（game_config 专用）
-- --------------------------------------------------------------------------------

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

-- 修正类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'modifier_type_enum') THEN
        CREATE TYPE modifier_type_enum AS ENUM (
            'advantage',      -- 优势
            'disadvantage',   -- 劣势
            'bonus_flat',     -- 固定加值
            'bonus_percent',  -- 百分比加成
            'penalty_flat',   -- 固定减值
            'penalty_percent' -- 百分比减值
        );
    END IF;
END $$;

-- 动作类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'action_type_enum') THEN
        CREATE TYPE action_type_enum AS ENUM (
            'main',       -- 主要动作
            'minor',      -- 次要动作
            'reaction'    -- 反应动作
        );
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 标签表 (游戏配置数据 - game_config schema)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.tags (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 标签信息
    tag_code      VARCHAR(32) NOT NULL UNIQUE,              -- 标签代码
    tag_name      VARCHAR(64) NOT NULL,                     -- 标签名称
    color         VARCHAR(16),                              -- 标签颜色
    icon          VARCHAR(256),                             -- 标签图标
    description   TEXT,                                     -- 标签描述
    category      tag_type_enum NOT NULL,                  -- 标签类别

    -- 状态
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,           -- 是否启用
    display_order INTEGER NOT NULL DEFAULT 0,              -- 显示顺序

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ                              -- 软删除
);

-- 标签表索引
CREATE INDEX IF NOT EXISTS idx_tags_is_active ON game_config.tags(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tags_display_order ON game_config.tags(display_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_code_unique ON game_config.tags(tag_code) WHERE deleted_at IS NULL;

-- 标签表触发器
CREATE TRIGGER update_tags_updated_at
    BEFORE UPDATE ON game_config.tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 标签关联信息表 (游戏配置数据 - game_config schema)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.tags_relations (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 关联信息
    tag_id        UUID NOT NULL REFERENCES game_config.tags(id) ON DELETE CASCADE, -- 标签 ID（显式指定 schema）
    entity_type   VARCHAR(64) NOT NULL,                                -- 关联实体类型 (如 'hero', 'item')
    entity_id     UUID NOT NULL,                                       -- 关联实体 ID

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ                                        -- 软删除
);

-- 标签关联信息表索引
CREATE INDEX IF NOT EXISTS idx_tags_relations_tag_id ON game_config.tags_relations(tag_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tags_relations_entity ON game_config.tags_relations(entity_type, entity_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_relations_unique ON game_config.tags_relations(tag_id, entity_type, entity_id) WHERE deleted_at IS NULL;

-- 标签关联信息表触发器
CREATE TRIGGER update_tags_relations_updated_at
    BEFORE UPDATE ON game_config.tags_relations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 英雄属性类型表
-- 定义游戏中所有可用的属性类型，如力量、敏捷、智力等
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.hero_attribute_type (
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
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_category ON game_config.hero_attribute_type(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_data_type ON game_config.hero_attribute_type(data_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_is_active ON game_config.hero_attribute_type(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_is_visible ON game_config.hero_attribute_type(is_visible) WHERE is_visible = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_attribute_type_display_order ON game_config.hero_attribute_type(display_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_hero_attribute_type_code_unique ON game_config.hero_attribute_type(attribute_code) WHERE deleted_at IS NULL;

-- 英雄属性类型表触发器
CREATE TRIGGER update_hero_attribute_type_updated_at
    BEFORE UPDATE ON game_config.hero_attribute_type
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

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
    FROM game_config.hero_attribute_type hat
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
    FROM game_config.hero_attribute_type
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
-- 职业表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.classes (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    -- 职业基本信息
    class_code    VARCHAR(32) NOT NULL UNIQUE,                         -- 职业代码
    class_name    VARCHAR(64) NOT NULL,                                -- 职业名称
    description   TEXT,                                                -- 职业描述
    lore_text     TEXT,                                     -- 职业背景故事

    -- 职业特色
    specialty     TEXT,                             -- 职业特长描述
    playstyle    TEXT,                             -- 职业玩法风格描述

    -- 职业等级和阶级
    tier          class_tier_enum NOT NULL,               -- 职业阶级
    promotion_count SMALLINT DEFAULT 0 CHECK (promotion_count >= 0), -- 转职次数加成

    -- 显示配置
    icon          VARCHAR(256),                            -- 职业图标URL
    color         VARCHAR(16),                             -- 职业代表颜色值

    -- 状态控制
    is_active     BOOLEAN DEFAULT TRUE,                    -- 是否启用
    is_visible    BOOLEAN DEFAULT TRUE,                    -- 是否在UI中显示
    display_order SMALLINT DEFAULT 0,                      -- 显示顺序

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 创建时间
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 更新时间
    deleted_at    TIMESTAMPTZ                              -- 删除时间 (软删除)
);

-- 职业表索引
CREATE INDEX IF NOT EXISTS idx_classes_tier ON game_config.classes(tier) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_classes_is_active ON game_config.classes(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_classes_is_visible ON game_config.classes(is_visible) WHERE is_visible = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_classes_display_order ON game_config.classes(display_order) WHERE deleted_at IS NULL;

-- 职业表触发器
CREATE TRIGGER update_classes_updated_at
    BEFORE UPDATE ON game_config.classes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业属性加成表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.class_attribute_bonuses (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    class_id      UUID NOT NULL REFERENCES game_config.classes(id) ON DELETE CASCADE,                           -- 职业ID
    attribute_id  UUID NOT NULL REFERENCES game_config.hero_attribute_type(id) ON DELETE CASCADE,                           -- 属性ID (引用 hero_attribute_type)

    base_bonus_value DECIMAL(10,2) NOT NULL DEFAULT 0,     -- 基础加成值

    bonus_per_level BOOLEAN NOT NULL DEFAULT FALSE,        -- 是否随等级加成
    per_level_bonus_value DECIMAL(10,2) NOT NULL DEFAULT 0, -- 每级加成值

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 创建时间
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 更新时间

    -- 唯一约束：每个职业的每个属性只能有一个加成配置
    UNIQUE(class_id, attribute_id)
);

-- 职业属性加成索引
CREATE INDEX IF NOT EXISTS idx_class_attribute_bonuses_class_id ON game_config.class_attribute_bonuses(class_id);
CREATE INDEX IF NOT EXISTS idx_class_attribute_bonuses_attribute_id ON game_config.class_attribute_bonuses(attribute_id);

-- 职业属性加成表触发器
CREATE TRIGGER update_class_attribute_bonuses_updated_at
    BEFORE UPDATE ON game_config.class_attribute_bonuses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业进阶要求表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.class_advanced_requirements (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    from_class_id UUID NOT NULL REFERENCES game_config.classes(id) ON DELETE CASCADE,                           -- 当前职业ID
    to_class_id   UUID NOT NULL REFERENCES game_config.classes(id) ON DELETE CASCADE,                           -- 目标职业ID

    -- 基础要求
    required_level INT NOT NULL CHECK (required_level > 0), -- 所需等级
    required_honor INT NOT NULL DEFAULT 0 CHECK (required_honor >= 0), -- 所需荣誉值
    required_job_change_count INT NOT NULL DEFAULT 1 CHECK (required_job_change_count >= 0), -- 所需转职次数

    -- 复杂要求 (JSON 格式)
    required_attributes JSONB,                              -- 所需属性要求, 格式: {"attribute_code": required_value, ...}
    required_skills JSONB,                                  -- 所需技能要求, 格式: {"skill_id": level, ...}
    required_items JSONB,                                   -- 所需物品, 格式: {"item_id": quantity, ...}

    -- 状态控制
    is_active     BOOLEAN DEFAULT TRUE,                     -- 是否启用
    display_order SMALLINT DEFAULT 0,                       -- 显示顺序

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),       -- 创建时间
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),       -- 更新时间
    deleted_at    TIMESTAMPTZ                              -- 删除时间 (软删除)
);

-- 职业进阶要求索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_class_advanced_requirements_from_to ON game_config.class_advanced_requirements(from_class_id, to_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_to_class_id ON game_config.class_advanced_requirements(to_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_display_order ON game_config.class_advanced_requirements(display_order) WHERE deleted_at IS NULL;

-- 职业进阶要求表触发器
CREATE TRIGGER update_class_advanced_requirements_updated_at
    BEFORE UPDATE ON game_config.class_advanced_requirements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业相关函数
-- --------------------------------------------------------------------------------

-- 检查职业进阶是否满足属性要求的函数
CREATE OR REPLACE FUNCTION check_class_advancement_requirements(
    p_from_class_id_param UUID,
    p_to_class_id_param UUID,
    p_hero_level_param INT,
    p_hero_attributes_param JSONB
) RETURNS TABLE (
    is_eligible BOOLEAN,
    missing_requirements TEXT[]
) AS $$
DECLARE
    v_req_record RECORD;
    v_missing_reqs TEXT[] := '{}';
BEGIN
    -- 获取进阶要求
    SELECT * INTO v_req_record
    FROM game_config.class_advanced_requirements
    WHERE from_class_id = p_from_class_id_param
      AND to_class_id = p_to_class_id_param
      AND is_active = TRUE
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RETURN QUERY SELECT FALSE, ARRAY['进阶路径不存在或未激活'];
        RETURN;
    END IF;

    -- 检查等级要求
    IF p_hero_level_param < v_req_record.required_level THEN
        v_missing_reqs := array_append(v_missing_reqs, format('需要等级 %s (当前 %s)', v_req_record.required_level, p_hero_level_param));
    END IF;

    -- 检查属性要求
    -- 没有属性要求则跳过
    IF v_req_record.required_attributes IS NOT NULL THEN
        DECLARE
            v_req_key TEXT;
            v_req_value NUMERIC;
            v_player_value NUMERIC;
        BEGIN
            FOR v_req_key, v_req_value IN SELECT key, value::NUMERIC FROM jsonb_each_text(v_req_record.required_attributes)
            LOOP
                -- 使用 ->> 操作符获取文本值，然后转换为 NUMERIC
                v_player_value := (p_hero_attributes_param ->> v_req_key)::NUMERIC;

                -- 检查玩家是否拥有该属性且满足要求
                IF v_player_value IS NULL OR v_player_value < v_req_value THEN
                    v_missing_reqs := array_append(v_missing_reqs, format('需要属性 %s 至少 %s (当前 %s)', v_req_key, v_req_value, COALESCE(v_player_value::TEXT, '无')));
                END IF;
            END LOOP;
        END;
    END IF;
    -- TODO: 检查其他要求 (属性、技能等)
    -- 这里可以根据需要扩展更复杂的检查逻辑

    RETURN QUERY SELECT array_length(v_missing_reqs, 1) IS NULL, v_missing_reqs;
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 技能类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.skill_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    category_code VARCHAR(32) NOT NULL,
    category_name VARCHAR(64) NOT NULL,
    description TEXT,

    -- 显示配置
    icon VARCHAR(256),
    color VARCHAR(16),
    display_order INTEGER DEFAULT 0,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_skill_categories_code_unique
    ON game_config.skill_categories(category_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_categories_active
    ON game_config.skill_categories(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE TRIGGER update_skill_categories_updated_at
    BEFORE UPDATE ON game_config.skill_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.skill_categories IS '技能类别表';

-- --------------------------------------------------------------------------------
-- 动作类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.action_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    category_code VARCHAR(32) NOT NULL,
    category_name VARCHAR(64) NOT NULL,
    description TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_action_categories_code_unique
    ON game_config.action_categories(category_code) WHERE deleted_at IS NULL;

CREATE TRIGGER update_action_categories_updated_at
    BEFORE UPDATE ON game_config.action_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.action_categories IS '动作类别表';

-- --------------------------------------------------------------------------------
-- 伤害类型定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.damage_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),  -- 'physical', 'magical', 'elemental', 'special'

    -- 抗性配置
    resistance_attribute_code VARCHAR(50),
    damage_reduction_attribute_code VARCHAR(50),  -- 新增：固定值减免属性
    resistance_cap INTEGER DEFAULT 75 CHECK (resistance_cap >= 0 AND resistance_cap <= 100),

    -- 视觉效果
    color VARCHAR(16),
    icon VARCHAR(256),

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_damage_types_code_unique
    ON game_config.damage_types(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_damage_types_category
    ON game_config.damage_types(category) WHERE deleted_at IS NULL;

CREATE TRIGGER update_damage_types_updated_at
    BEFORE UPDATE ON game_config.damage_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.damage_types IS '伤害类型定义表';
COMMENT ON COLUMN game_config.damage_types.resistance_cap IS '该伤害类型的抗性上限(%)';
COMMENT ON COLUMN game_config.damage_types.damage_reduction_attribute_code IS '固定值减免属性代码（如FIRE_DmgReduce）';

-- --------------------------------------------------------------------------------
-- 元效果类型定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.effect_type_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    effect_type_code VARCHAR(50) NOT NULL,
    effect_type_name VARCHAR(100) NOT NULL,

    description TEXT,
    parameter_list TEXT[],        -- 参数列表
    parameter_descriptions TEXT,  -- 参数说明
    parameter_definitions JSONB,  -- 参数详细定义（类型、范围等）

    failure_handling VARCHAR(50), -- 失败处理方式：skip_remaining, continue
    json_template JSONB,          -- JSON配置模板

    example TEXT,
    notes TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_effect_type_definitions_code_unique
    ON game_config.effect_type_definitions(effect_type_code) WHERE deleted_at IS NULL;

CREATE TRIGGER update_effect_type_definitions_updated_at
    BEFORE UPDATE ON game_config.effect_type_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.effect_type_definitions IS '元效果类型定义表 - 定义所有可用的元效果类型及其参数规范';
COMMENT ON COLUMN game_config.effect_type_definitions.effect_type_code IS '元效果代码（如HIT_CHECK、DMG_CALCULATION）';
COMMENT ON COLUMN game_config.effect_type_definitions.parameter_list IS '参数名称数组';
COMMENT ON COLUMN game_config.effect_type_definitions.failure_handling IS '失败处理：skip_remaining跳过后续效果, continue继续执行';

-- --------------------------------------------------------------------------------
-- 公式变量定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.formula_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    variable_code VARCHAR(50) NOT NULL,
    variable_name VARCHAR(100) NOT NULL,
    variable_type VARCHAR(50) NOT NULL,  -- 'attribute', 'target', 'skill_data', 'equipment_data'
    scope VARCHAR(50) NOT NULL,          -- 'character', 'action', 'skill', 'equipment', 'global'
    data_type VARCHAR(20) NOT NULL,      -- 'integer', 'decimal', 'string', 'object', 'boolean'

    description TEXT,
    example TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_formula_variables_code_unique
    ON game_config.formula_variables(variable_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_formula_variables_scope
    ON game_config.formula_variables(scope) WHERE deleted_at IS NULL;

CREATE TRIGGER update_formula_variables_updated_at
    BEFORE UPDATE ON game_config.formula_variables
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.formula_variables IS '公式变量定义表 - 定义所有可在配置公式中使用的变量';
COMMENT ON COLUMN game_config.formula_variables.variable_type IS '变量类型：基础属性、目标选择、技能数据、装备数据';
COMMENT ON COLUMN game_config.formula_variables.scope IS '变量作用域：角色、动作、技能、装备、全局';

-- --------------------------------------------------------------------------------
-- 射程配置规则表（元数据）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.range_config_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    parameter_type VARCHAR(50) NOT NULL,    -- 'range', 'positions', 'depth'
    parameter_format VARCHAR(100) NOT NULL, -- 格式说明（如'N', '0~N', '最近N位'）

    description TEXT,
    example VARCHAR(200),
    notes TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE game_config.range_config_rules IS '射程配置规则表 - 文档用途，说明射程参数的格式规范';

-- --------------------------------------------------------------------------------
-- 动作类型定义表（元数据）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.action_type_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    action_type VARCHAR(20) NOT NULL,
    description TEXT,
    per_turn_limit INTEGER,  -- 每回合限制次数
    usage_timing TEXT,       -- 使用时机说明
    example TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_action_type_definitions_type_unique
    ON game_config.action_type_definitions(action_type);

COMMENT ON TABLE game_config.action_type_definitions IS '动作类型定义表 - 说明main/minor/reaction的规则';

-- --------------------------------------------------------------------------------
-- 效果定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 效果基本信息
    effect_code VARCHAR(50) NOT NULL,
    effect_name VARCHAR(100) NOT NULL,
    effect_type VARCHAR(50) NOT NULL,  -- 关联 effect_type_definitions

    -- 效果参数（标准化JSON）
    parameters JSONB NOT NULL DEFAULT '{}',

    -- 计算公式（可选）
    calculation_formula TEXT,

    -- 触发条件
    trigger_condition JSONB,
    trigger_chance DECIMAL(5,2) DEFAULT 100.00 CHECK (trigger_chance >= 0 AND trigger_chance <= 100),

    -- 目标过滤
    target_filter JSONB,

    -- 特征标签（冗余字段，方便快速查询）
    feature_tags TEXT[],

    -- 视觉/音效配置
    visual_config JSONB,
    sound_config JSONB,

    -- 描述模板
    description TEXT,
    tooltip_template TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_effects_code_unique
    ON game_config.effects(effect_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_effects_type
    ON game_config.effects(effect_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_effects_feature_tags
    ON game_config.effects USING GIN(feature_tags);

CREATE TRIGGER update_effects_updated_at
    BEFORE UPDATE ON game_config.effects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.effects IS '效果定义表 - 所有游戏效果的原子单位';
COMMENT ON COLUMN game_config.effects.effect_type IS '效果类型，关联effect_type_definitions表';
COMMENT ON COLUMN game_config.effects.parameters IS 'JSON格式的效果参数，不同effect_type有不同的参数结构';

-- --------------------------------------------------------------------------------
-- 技能定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 技能基本信息
    skill_code VARCHAR(50) NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_type skill_type_enum NOT NULL DEFAULT 'weapon',
    category_id UUID REFERENCES game_config.skill_categories(id),

    -- 等级配置
    max_level INTEGER DEFAULT 10 CHECK (max_level >= 1 AND max_level <= 100),

    -- 特征和效果
    feature_tags TEXT[],
    passive_effects JSONB,  -- 新增：被动效果配置

    -- 学习要求
    required_level INTEGER DEFAULT 1 CHECK (required_level >= 1),
    required_class_codes VARCHAR(32)[],
    prerequisite_skill_codes VARCHAR(50)[],

    -- 显示配置
    description TEXT,
    detailed_description TEXT,
    icon VARCHAR(256),

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_code_unique
    ON game_config.skills(skill_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_type
    ON game_config.skills(skill_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_category
    ON game_config.skills(category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_feature_tags
    ON game_config.skills USING GIN(feature_tags);

CREATE TRIGGER update_skills_updated_at
    BEFORE UPDATE ON game_config.skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.skills IS '技能定义表';
COMMENT ON COLUMN game_config.skills.passive_effects IS '被动效果配置（JSONB格式）';

-- --------------------------------------------------------------------------------
-- 技能等级配置表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.skill_level_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id UUID NOT NULL REFERENCES game_config.skills(id) ON DELETE CASCADE,

    level_number INTEGER NOT NULL CHECK (level_number >= 1),

    -- 对动作/效果的加成
    damage_multiplier DECIMAL(5,2) DEFAULT 1.0,
    healing_multiplier DECIMAL(5,2) DEFAULT 1.0,
    duration_modifier INTEGER DEFAULT 0,
    range_modifier INTEGER DEFAULT 0,
    cooldown_modifier INTEGER DEFAULT 0,
    mana_cost_modifier INTEGER DEFAULT 0,

    -- 其他加成配置
    effect_modifiers JSONB,

    -- 升级消耗
    upgrade_cost_xp INTEGER DEFAULT 0,
    upgrade_cost_gold INTEGER DEFAULT 0,
    upgrade_materials JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(skill_id, level_number)
);

CREATE INDEX IF NOT EXISTS idx_skill_level_configs_skill
    ON game_config.skill_level_configs(skill_id) WHERE deleted_at IS NULL;

CREATE TRIGGER update_skill_level_configs_updated_at
    BEFORE UPDATE ON game_config.skill_level_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.skill_level_configs IS '技能等级配置表 - 每个技能每个等级的具体数值';

-- --------------------------------------------------------------------------------
-- 动作定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- 动作基本信息
    action_code VARCHAR(50) NOT NULL,
    action_name VARCHAR(100) NOT NULL,
    action_category_id UUID REFERENCES game_config.action_categories(id),
    action_type action_type_enum NOT NULL DEFAULT 'main',

    -- 关联技能（可选）
    related_skill_id UUID REFERENCES game_config.skills(id),

    -- 动作参数
    feature_tags TEXT[],
    range_config JSONB NOT NULL DEFAULT '{}',  -- 射程配置（range, positions, depth）
    target_config JSONB DEFAULT '{"type": "single", "ally": false}',
    area_config JSONB DEFAULT '{"type": "single"}',

    -- 消耗和限制
    action_point_cost INTEGER DEFAULT 1 CHECK (action_point_cost >= 0),
    mana_cost INTEGER DEFAULT 0 CHECK (mana_cost >= 0),
    mana_cost_formula TEXT,  -- 新增：支持公式（如"50+2*skill_level"）
    cooldown_turns INTEGER DEFAULT 0 CHECK (cooldown_turns >= 0),
    uses_per_battle INTEGER,

    -- 命中率配置
    hit_rate_config JSONB DEFAULT '{
        "base_hit_rate": 75,
        "use_attacker_accuracy": true,
        "accuracy_attribute": "ACCURACY",
        "accuracy_multiplier": 1.0,
        "use_target_defense": false,
        "min_hit_rate": 5,
        "max_hit_rate": 95
    }'::jsonb,

    -- 效果配置（兼容Excel导入）
    legacy_effect_config JSONB,  -- 新增：存储Excel原始配置

    -- 使用条件
    requirements JSONB,
    start_flags TEXT[],  -- 开始标记数组

    -- 动画和视觉效果
    animation_config JSONB,
    visual_effects JSONB,
    sound_effects JSONB,

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_actions_code_unique
    ON game_config.actions(action_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_category
    ON game_config.actions(action_category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_skill
    ON game_config.actions(related_skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_type
    ON game_config.actions(action_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_feature_tags
    ON game_config.actions USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_actions_start_flags
    ON game_config.actions USING GIN(start_flags);

CREATE TRIGGER update_actions_updated_at
    BEFORE UPDATE ON game_config.actions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.actions IS '动作定义表';
COMMENT ON COLUMN game_config.actions.hit_rate_config IS '命中率计算配置';
COMMENT ON COLUMN game_config.actions.legacy_effect_config IS 'Excel原始效果配置（用于兼容导入）';
COMMENT ON COLUMN game_config.actions.mana_cost_formula IS 'MP消耗公式（如"50+2*skill_level"），优先于mana_cost';
COMMENT ON COLUMN game_config.actions.start_flags IS '开始标记数组（如STARTING_ATTACK_ACTION）';

-- --------------------------------------------------------------------------------
-- Action与Effect关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.action_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    action_id UUID NOT NULL REFERENCES game_config.actions(id) ON DELETE CASCADE,
    effect_id UUID NOT NULL REFERENCES game_config.effects(id) ON DELETE CASCADE,

    execution_order INTEGER DEFAULT 0,

    -- 参数覆盖
    parameter_overrides JSONB,

    -- 条件执行
    is_conditional BOOLEAN DEFAULT FALSE,
    condition_config JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(action_id, effect_id, execution_order)
);

CREATE INDEX IF NOT EXISTS idx_action_effects_action
    ON game_config.action_effects(action_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_effects_effect
    ON game_config.action_effects(effect_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_effects_order
    ON game_config.action_effects(action_id, execution_order);

COMMENT ON TABLE game_config.action_effects IS 'Action与Effect关联表';

-- --------------------------------------------------------------------------------
-- 技能解锁动作关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.skill_unlock_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id UUID NOT NULL REFERENCES game_config.skills(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES game_config.actions(id) ON DELETE CASCADE,

    unlock_level INTEGER NOT NULL DEFAULT 1 CHECK (unlock_level >= 1),
    is_default BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(skill_id, action_id)
);

CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_skill
    ON game_config.skill_unlock_actions(skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_action
    ON game_config.skill_unlock_actions(action_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_level
    ON game_config.skill_unlock_actions(unlock_level);

COMMENT ON TABLE game_config.skill_unlock_actions IS '技能解锁动作关联表';

-- --------------------------------------------------------------------------------
-- Buff定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.buffs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    -- Buff基本信息
    buff_code VARCHAR(50) NOT NULL,
    buff_name VARCHAR(100) NOT NULL,
    buff_type VARCHAR(50) NOT NULL DEFAULT 'buff',  -- 'buff', 'debuff', 'neutral'

    -- Buff分类
    category VARCHAR(50),
    feature_tags TEXT[],

    -- 持续时间配置
    default_duration INTEGER DEFAULT 1,
    max_duration INTEGER DEFAULT 10,
    min_duration INTEGER DEFAULT 1,

    -- 叠加规则
    stack_rule VARCHAR(50) DEFAULT 'no_stack',  -- 'no_stack', 'stackable', 'replace', 'refresh'
    max_stacks INTEGER DEFAULT 1,

    -- 效果描述和参数（新增）
    effect_description TEXT,     -- 效果描述（给玩家看的）
    parameter_list TEXT[],       -- 参数列表（如["N"]）
    parameter_definitions JSONB, -- 参数定义（如{"N": {"type":"percentage", "min":0, "max":100}}）

    -- 触发事件
    trigger_events TEXT[],  -- 'turn_start', 'turn_end', 'take_damage', etc.

    -- 优劣势支持
    provides_advantage BOOLEAN DEFAULT FALSE,
    provides_disadvantage BOOLEAN DEFAULT FALSE,
    advantage_applies_to TEXT[],  -- ['hit_rate', 'critical_rate', 'saving_throw']

    -- 视觉效果
    visual_effects JSONB,
    icon VARCHAR(256),
    color VARCHAR(16),

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_buffs_code_unique
    ON game_config.buffs(buff_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_type
    ON game_config.buffs(buff_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_category
    ON game_config.buffs(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_feature_tags
    ON game_config.buffs USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_buffs_trigger_events
    ON game_config.buffs USING GIN(trigger_events);

CREATE TRIGGER update_buffs_updated_at
    BEFORE UPDATE ON game_config.buffs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.buffs IS 'Buff定义表';
COMMENT ON COLUMN game_config.buffs.effect_description IS 'Buff效果的文字描述（面向玩家）';
COMMENT ON COLUMN game_config.buffs.parameter_list IS 'Buff的参数名称列表';
COMMENT ON COLUMN game_config.buffs.parameter_definitions IS 'Buff参数的详细定义（类型、范围等）';

-- --------------------------------------------------------------------------------
-- Buff与Effect关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.buff_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    buff_id UUID NOT NULL REFERENCES game_config.buffs(id) ON DELETE CASCADE,
    effect_id UUID NOT NULL REFERENCES game_config.effects(id) ON DELETE CASCADE,

    trigger_timing VARCHAR(50) NOT NULL,  -- 'on_apply', 'turn_start', 'turn_end', 'on_remove'
    execution_order INTEGER DEFAULT 0,

    parameter_overrides JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(buff_id, effect_id, trigger_timing, execution_order)
);

CREATE INDEX IF NOT EXISTS idx_buff_effects_buff
    ON game_config.buff_effects(buff_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buff_effects_effect
    ON game_config.buff_effects(effect_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buff_effects_timing
    ON game_config.buff_effects(trigger_timing);

COMMENT ON TABLE game_config.buff_effects IS 'Buff与Effect关联表';

-- --------------------------------------------------------------------------------
-- 动作Flag定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS game_config.action_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    flag_code VARCHAR(50) NOT NULL,
    flag_name VARCHAR(100) NOT NULL,
    category VARCHAR(50),  -- 'action_chain', 'check_status', 'modifier'

    -- 持续时间配置
    duration_type VARCHAR(20) DEFAULT 'action',  -- 'action', 'turn', 'battle', 'permanent'
    default_duration VARCHAR(20) DEFAULT '1',

    -- 自动移除条件
    auto_remove_condition VARCHAR(100),  -- 'act_end', 'next_act_start', 'turn_start', 'turn_end'
    remove_on_events TEXT[],

    -- Flag属性
    is_visible BOOLEAN DEFAULT FALSE,
    is_stackable BOOLEAN DEFAULT FALSE,
    max_stacks INTEGER DEFAULT 1,

    -- 优劣势支持
    provides_advantage BOOLEAN DEFAULT FALSE,
    provides_disadvantage BOOLEAN DEFAULT FALSE,
    advantage_applies_to TEXT[],

    -- 效果
    flag_effects JSONB,
    modifier_effects JSONB,

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_action_flags_code_unique
    ON game_config.action_flags(flag_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_flags_category
    ON game_config.action_flags(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_flags_duration_type
    ON game_config.action_flags(duration_type) WHERE deleted_at IS NULL;

CREATE TRIGGER update_action_flags_updated_at
    BEFORE UPDATE ON game_config.action_flags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.action_flags IS '动作Flag定义表';
COMMENT ON COLUMN game_config.action_flags.auto_remove_condition IS '自动移除条件（act_end/next_act_start/turn_start/turn_end等）';

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Game Config Schema Tables 创建完成';
    RAISE NOTICE '包含: 标签、属性类型、职业、技能、动作、Buff等配置表';
    RAISE NOTICE '注意: 系统必需数据请通过 seeds/system/ 导入';
    RAISE NOTICE '============================================';
END $$;
