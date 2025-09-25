-- =============================================================================
-- Create Classes System
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 职业表
-- --------------------------------------------------------------------------------

--职业阶级枚举，1-基础，2-进阶，3-精英，4-传奇，5-神话
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'class_tier_enum') THEN
        CREATE TYPE class_tier_enum AS ENUM ('1', '2', '3', '4', '5');
    END IF;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE IF NOT EXISTS classes(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_code    VARCHAR(32) NOT NULL UNIQUE, -- 职业代码
    class_name    VARCHAR(64) NOT NULL, -- 职业名称
    description   TEXT, -- 职业描述
    lore_text    TEXT, -- 职业背景故事

    tier         class_tier_enum NOT NULL, -- 职业阶级

    promotion_count SMALLINT DEFAULT 0, -- 转职次数奖励,默认为0

    icon         VARCHAR(256), -- 职业图标URL
    color        VARCHAR(16), -- 职业代表颜色值

    is_active     BOOLEAN DEFAULT TRUE, -- 是否启用
    is_visible    BOOLEAN DEFAULT TRUE, -- 是否在UI中显示
    display_order SMALLINT DEFAULT 0, -- 显示顺序

    created_at    TIMESTAMPTZ DEFAULT NOW(), -- 创建时间
    updated_at    TIMESTAMPTZ DEFAULT NOW(), -- 更新时间
    deleted_at    TIMESTAMPTZ -- 删除时间
);

-- 职业索引
CREATE INDEX idx_classes_tier ON classes(tier);
CREATE INDEX idx_classes_is_active ON classes(is_active);
CREATE INDEX idx_classes_is_visible ON classes(is_visible);
CREATE INDEX idx_classes_display_order ON classes(display_order);
CREATE UNIQUE INDEX idx_classes_class_code ON classes(class_code) WHERE deleted_at IS NULL;

-- 职业表触发器
CREATE TRIGGER update_classes_updated_at 
    BEFORE UPDATE ON classes 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业tag关联表
-- --------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS class_tag_relations(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_id   UUID NOT NULL, -- 职业ID
    tag_id     UUID NOT NULL, -- 标签ID

    created_at TIMESTAMPTZ DEFAULT NOW(), -- 创建时间
    updated_at TIMESTAMPTZ DEFAULT NOW(), -- 更新时间

    UNIQUE(class_id, tag_id),

    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- 职业标签关联索引
CREATE INDEX idx_class_tag_relations_class_id ON class_tag_relations(class_id);
CREATE INDEX idx_class_tag_relations_tag_id ON class_tag_relations(tag_id);

-- 职业标签关联表触发器
CREATE TRIGGER update_class_tag_relations_updated_at 
    BEFORE UPDATE ON class_tag_relations 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- -----------------------------------------------------------------------------
-- 职业属性加成表
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS class_attribute_bonuses (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_id      UUID NOT NULL, -- 职业ID
    attribute_id UUID NOT NULL, -- 属性ID

    base_bonus_value DECIMAL(10,2) NOT NULL, -- 基础加成值
    per_level_bonus_value DECIMAL(10,2) NOT NULL, -- 每级加成值

    created_at    TIMESTAMPTZ DEFAULT NOW(), -- 创建时间
    updated_at    TIMESTAMPTZ DEFAULT NOW(), -- 更新时间

    UNIQUE(class_id, attribute_id),

    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (attribute_id) REFERENCES hero_attribute_type(id) ON DELETE CASCADE
);

-- 职业属性加成索引
CREATE INDEX idx_class_attribute_bonuses_class_id ON class_attribute_bonuses(class_id);
CREATE INDEX idx_class_attribute_bonuses_attribute_id ON class_attribute_bonuses(attribute_id);

-- 职业属性加成表触发器
CREATE TRIGGER update_class_attribute_bonuses_updated_at 
    BEFORE UPDATE ON class_attribute_bonuses 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- -----------------------------------------------------------------------------
-- 职业进阶要求表
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS class_advanced_requirements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    from_class_id UUID NOT NULL, -- 当前职业ID
    to_class_id   UUID NOT NULL, -- 目标职业ID

    required_level INT NOT NULL, -- 所需等级
    required_honor INT NOT NULL DEFAULT 0, -- 所需荣誉值
    required_job_change_count INT NOT NULL DEFAULT 1, -- 所需转职次数

    required_attributes JSONB, -- 所需属性要求, 格式: {"attribute_code": required_value, ...}
    required_skills JSONB, -- 所需技能要求, 格式: {"skill_id": level, ...}

    is_active     BOOLEAN DEFAULT TRUE, -- 是否启用
    display_order SMALLINT DEFAULT 0, -- 显示顺序

    created_at    TIMESTAMPTZ DEFAULT NOW(), -- 创建时间
    updated_at    TIMESTAMPTZ DEFAULT NOW(), -- 更新时间
    deleted_at    TIMESTAMPTZ, -- 删除时间

    FOREIGN KEY (from_class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (to_class_id) REFERENCES classes(id) ON DELETE CASCADE
);

-- 职业进阶要求索引
CREATE UNIQUE INDEX idx_class_advanced_requirements_from_to ON class_advanced_requirements(from_class_id, to_class_id);
CREATE INDEX idx_class_advanced_requirements_from_class_id ON class_advanced_requirements(from_class_id);
CREATE INDEX idx_class_advanced_requirements_to_class_id ON class_advanced_requirements(to_class_id);
CREATE INDEX idx_class_advanced_requirements_display_order ON class_advanced_requirements(display_order);

-- 职业进阶要求表触发器
CREATE TRIGGER update_class_advanced_requirements_updated_at 
    BEFORE UPDATE ON class_advanced_requirements 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
