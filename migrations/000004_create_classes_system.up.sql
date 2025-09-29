-- =============================================================================
-- Create Classes System
-- 职业系统：职业定义、属性加成、进阶路径等
-- 依赖：000001_create_core_infrastructure, 000003_create_attribute_system
-- =============================================================================

-- 为tag_type_enum标签类型枚举添加class
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumtypid = 'tag_type_enum'::regtype AND enumlabel = 'class') THEN
        ALTER TYPE tag_type_enum ADD VALUE 'class';
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 职业表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS classes (
    -- 职业基本信息
    class_code    VARCHAR(32) PRIMARY KEY NOT NULL UNIQUE,             -- 职业代码
    class_name    VARCHAR(64) NOT NULL,                    -- 职业名称
    description   TEXT,                                     -- 职业描述
    lore_text     TEXT,                                     -- 职业背景故事
    
    -- 职业特色
    specialty     TEXT,                             -- 职业特长描述
    playstyle    TEXT,                             -- 职业玩法风格描述

    -- 职业等级和阶级
    tier          class_tier_enum NOT NULL,               -- 职业阶级 (1-5)
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
CREATE INDEX IF NOT EXISTS idx_classes_tier ON classes(tier) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_classes_is_active ON classes(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_classes_is_visible ON classes(is_visible) WHERE is_visible = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_classes_display_order ON classes(display_order) WHERE deleted_at IS NULL;

-- 职业表触发器
CREATE TRIGGER update_classes_updated_at
    BEFORE UPDATE ON classes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业属性加成表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS class_attribute_bonuses (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    class_id      UUID NOT NULL,                           -- 职业ID
    attribute_id  UUID NOT NULL,                           -- 属性ID (引用 hero_attribute_type)

    base_bonus_value DECIMAL(10,2) NOT NULL DEFAULT 0,     -- 基础加成值

    bonus_per_level BOOLEAN NOT NULL DEFAULT FALSE,        -- 是否随等级加成
    per_level_bonus_value DECIMAL(10,2) NOT NULL DEFAULT 0, -- 每级加成值

    -- 时间戳
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 创建时间
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 更新时间

    -- 唯一约束：每个职业的每个属性只能有一个加成配置
    UNIQUE(class_id, attribute_id),

    -- 外键约束
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (attribute_id) REFERENCES hero_attribute_type(id) ON DELETE CASCADE
);

-- 职业属性加成索引
CREATE INDEX IF NOT EXISTS idx_class_attribute_bonuses_class_id ON class_attribute_bonuses(class_id);
CREATE INDEX IF NOT EXISTS idx_class_attribute_bonuses_attribute_id ON class_attribute_bonuses(attribute_id);

-- 职业属性加成表触发器
CREATE TRIGGER update_class_attribute_bonuses_updated_at
    BEFORE UPDATE ON class_attribute_bonuses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业进阶要求表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS class_advanced_requirements (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),

    from_class_id UUID NOT NULL,                           -- 当前职业ID
    to_class_id   UUID NOT NULL,                           -- 目标职业ID

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
    deleted_at    TIMESTAMPTZ,                              -- 删除时间 (软删除)

    -- 外键约束
    FOREIGN KEY (from_class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (to_class_id) REFERENCES classes(id) ON DELETE CASCADE
);

-- 职业进阶要求索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_class_advanced_requirements_from_to ON class_advanced_requirements(from_class_id, to_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_to_class_id ON class_advanced_requirements(to_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_display_order ON class_advanced_requirements(display_order) WHERE deleted_at IS NULL;

-- 职业进阶要求表触发器
CREATE TRIGGER update_class_advanced_requirements_updated_at
    BEFORE UPDATE ON class_advanced_requirements
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
    FROM class_advanced_requirements
    WHERE from_class_id = p_from_class_id_param
      AND to_class_id = p_to_class_id_param
      AND is_active = TRUE
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RETURN QUERY SELECT FALSE, ARRAY['进阶路径不存在或未激活'];
        RETURN;
    END IF;

    -- 检查等级要求
    IF hero_level_param < v_req_record.required_level THEN
        v_missing_reqs := array_append(v_missing_reqs, format('需要等级 %s (当前 %s)', v_req_record.required_level, hero_level_param));
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
                v_player_value := (hero_attributes_param ->> v_req_key)::NUMERIC;

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
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Classes System 创建完成';
    RAISE NOTICE '包含: 职业定义、标签系统、属性加成、进阶路径';
    RAISE NOTICE '已插入 % 个职业', (SELECT COUNT(*) FROM classes WHERE deleted_at IS NULL);
    RAISE NOTICE '已插入 % 个标签', (SELECT COUNT(*) FROM tags WHERE deleted_at IS NULL);
    RAISE NOTICE '============================================';
END $$;