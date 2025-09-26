-- =============================================================================
-- Create Classes System
-- 职业系统：职业定义、属性加成、进阶路径等
-- 依赖：000001_create_core_infrastructure, 000003_create_attribute_system
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 标签表 (通用标签系统)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS tags (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 标签信息
    tag_code      VARCHAR(32) NOT NULL UNIQUE,              -- 标签代码
    tag_name      VARCHAR(64) NOT NULL,                     -- 标签名称
    tag_type      VARCHAR(32) NOT NULL DEFAULT 'general',   -- 标签类型 (class, skill, item 等)
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

-- 标签表索引
CREATE INDEX IF NOT EXISTS idx_tags_tag_type ON tags(tag_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tags_is_active ON tags(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tags_display_order ON tags(display_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_code_unique ON tags(tag_code) WHERE deleted_at IS NULL;

-- 标签表触发器
CREATE TRIGGER update_tags_updated_at
    BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS classes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 职业基本信息
    class_code    VARCHAR(32) NOT NULL UNIQUE,             -- 职业代码
    class_name    VARCHAR(64) NOT NULL,                    -- 职业名称
    description   TEXT,                                     -- 职业描述
    lore_text     TEXT,                                     -- 职业背景故事

    -- 职业等级和阶级
    tier          class_tier_enum NOT NULL,               -- 职业阶级 (1-5)
    promotion_count SMALLINT DEFAULT 0 CHECK (promotion_count >= 0), -- 转职次数奖励

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
CREATE UNIQUE INDEX IF NOT EXISTS idx_classes_class_code ON classes(class_code) WHERE deleted_at IS NULL;

-- 职业表触发器
CREATE TRIGGER update_classes_updated_at
    BEFORE UPDATE ON classes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业标签关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS class_tag_relations (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_id   UUID NOT NULL,                              -- 职业ID
    tag_id     UUID NOT NULL,                              -- 标签ID

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),         -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),         -- 更新时间

    -- 唯一约束
    UNIQUE(class_id, tag_id),

    -- 外键约束
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- 职业标签关联索引
CREATE INDEX IF NOT EXISTS idx_class_tag_relations_class_id ON class_tag_relations(class_id);
CREATE INDEX IF NOT EXISTS idx_class_tag_relations_tag_id ON class_tag_relations(tag_id);

-- 职业标签关联表触发器
CREATE TRIGGER update_class_tag_relations_updated_at
    BEFORE UPDATE ON class_tag_relations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业属性加成表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS class_attribute_bonuses (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_id      UUID NOT NULL,                           -- 职业ID
    attribute_id  UUID NOT NULL,                           -- 属性ID (引用 hero_attribute_type)

    -- 加成值配置
    base_bonus_value DECIMAL(10,2) NOT NULL DEFAULT 0,     -- 基础加成值
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
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

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
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_from_class_id ON class_advanced_requirements(from_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_to_class_id ON class_advanced_requirements(to_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advanced_requirements_display_order ON class_advanced_requirements(display_order) WHERE deleted_at IS NULL;

-- 职业进阶要求表触发器
CREATE TRIGGER update_class_advanced_requirements_updated_at
    BEFORE UPDATE ON class_advanced_requirements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 职业进阶路径表 (进阶路径的视图化表示)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS class_advancement_paths (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    path_name     VARCHAR(64) NOT NULL,                     -- 进阶路径名称
    path_code     VARCHAR(32) NOT NULL UNIQUE,              -- 进阶路径代码
    description   TEXT,                                      -- 路径描述

    -- 路径配置
    start_class_id UUID NOT NULL,                           -- 起始职业ID
    path_classes   JSONB NOT NULL,                          -- 路径中的职业列表, 格式: [class_id1, class_id2, ...]
    min_level      INT DEFAULT 1 CHECK (min_level > 0),     -- 最低等级要求

    -- 状态控制
    is_active      BOOLEAN DEFAULT TRUE,                    -- 是否启用
    display_order  INTEGER DEFAULT 0,                       -- 显示顺序

    -- 时间戳
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 创建时间
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- 更新时间
    deleted_at     TIMESTAMPTZ,                             -- 删除时间 (软删除)

    -- 外键约束
    FOREIGN KEY (start_class_id) REFERENCES classes(id) ON DELETE CASCADE
);

-- 职业进阶路径索引
CREATE INDEX IF NOT EXISTS idx_class_advancement_paths_start_class ON class_advancement_paths(start_class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_advancement_paths_is_active ON class_advancement_paths(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_class_advancement_paths_code ON class_advancement_paths(path_code) WHERE deleted_at IS NULL;

-- 职业进阶路径表触发器
CREATE TRIGGER update_class_advancement_paths_updated_at
    BEFORE UPDATE ON class_advancement_paths
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 插入基础标签数据
-- --------------------------------------------------------------------------------

INSERT INTO tags (tag_code, tag_name, tag_type, color, icon, description, display_order) VALUES
    -- 职业标签
    ('MELEE', '近战', 'class', '#E74C3C', '/icons/melee.png', '近距离作战的职业', 1),
    ('RANGED', '远程', 'class', '#3498DB', '/icons/ranged.png', '远距离作战的职业', 2),
    ('MAGIC', '法术', 'class', '#9B59B6', '/icons/magic.png', '使用魔法的职业', 3),
    ('SUPPORT', '辅助', 'class', '#2ECC71', '/icons/support.png', '提供支援的职业', 4),
    ('TANK', '坦克', 'class', '#95A5A6', '/icons/tank.png', '承受伤害的职业', 5),
    ('DPS', '输出', 'class', '#E67E22', '/icons/dps.png', '高伤害输出的职业', 6),

    -- 进阶标签
    ('BASIC', '基础', 'class', '#7F8C8D', '/icons/basic.png', '基础职业', 10),
    ('ADVANCED', '进阶', 'class', '#F39C12', '/icons/advanced.png', '进阶职业', 11),
    ('ELITE', '精英', 'class', '#8E44AD', '/icons/elite.png', '精英职业', 12),
    ('LEGENDARY', '传奇', 'class', '#C0392B', '/icons/legendary.png', '传奇职业', 13),
    ('MYTHIC', '神话', 'class', '#D4AF37', '/icons/mythic.png', '神话职业', 14)
ON CONFLICT (tag_code) DO NOTHING;

-- --------------------------------------------------------------------------------
-- 插入示例职业数据
-- --------------------------------------------------------------------------------

INSERT INTO classes (class_code, class_name, description, lore_text, tier, icon, color, display_order) VALUES
    ('WARRIOR', '战士', '近战物理职业，具有高血量和防御力', '战场上的勇士，以坚韧不拔著称', '1', '/icons/warrior.png', '#C0392B', 1),
    ('ARCHER', '弓箭手', '远程物理职业，具有高敏捷和精准度', '精通弓术的猎手，百步穿杨', '1', '/icons/archer.png', '#27AE60', 2),
    ('MAGE', '法师', '远程魔法职业，具有强大的法术能力', '掌握神秘力量的学者', '1', '/icons/mage.png', '#3498DB', 3),
    ('ROGUE', '盗贼', '敏捷型职业，擅长潜行和暴击', '行走在阴影中的刺客', '1', '/icons/rogue.png', '#7F8C8D', 4),

    -- 进阶职业示例
    ('PALADIN', '圣骑士', '战士进阶职业，具有神圣力量', '正义与光明的化身', '2', '/icons/paladin.png', '#F1C40F', 10),
    ('BERSERKER', '狂战士', '战士进阶职业，狂怒状态下战力倍增', '战场上的疯狂战士', '2', '/icons/berserker.png', '#E74C3C', 11)
ON CONFLICT (class_code) DO NOTHING;

-- --------------------------------------------------------------------------------
-- 职业相关函数
-- --------------------------------------------------------------------------------

-- 获取职业进阶路径的函数
CREATE OR REPLACE FUNCTION get_class_advancement_paths(class_id_param UUID)
RETURNS TABLE (
    path_id UUID,
    path_name VARCHAR,
    target_class_id UUID,
    target_class_name VARCHAR,
    required_level INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        car.id,
        c_to.class_name,
        car.to_class_id,
        c_to.class_name,
        car.required_level
    FROM class_advanced_requirements car
    JOIN classes c_to ON car.to_class_id = c_to.id
    WHERE car.from_class_id = class_id_param
      AND car.is_active = TRUE
      AND car.deleted_at IS NULL
      AND c_to.deleted_at IS NULL
    ORDER BY car.display_order ASC, car.required_level ASC;
END;
$$ LANGUAGE plpgsql;

-- 检查职业进阶是否满足条件的函数
CREATE OR REPLACE FUNCTION check_class_advancement_requirements(
    from_class_id_param UUID,
    to_class_id_param UUID,
    hero_level_param INT,
    hero_attributes_param JSONB
) RETURNS TABLE (
    is_eligible BOOLEAN,
    missing_requirements TEXT[]
) AS $$
DECLARE
    req_record RECORD;
    missing_reqs TEXT[] := '{}';
    is_valid BOOLEAN := FALSE;
BEGIN
    -- 获取进阶要求
    SELECT * INTO req_record
    FROM class_advanced_requirements
    WHERE from_class_id = from_class_id_param
      AND to_class_id = to_class_id_param
      AND is_active = TRUE
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RETURN QUERY SELECT FALSE, ARRAY['进阶路径不存在或未激活'];
        RETURN;
    END IF;

    -- 检查等级要求
    IF hero_level_param < req_record.required_level THEN
        missing_reqs := array_append(missing_reqs, format('需要等级 %s (当前 %s)', req_record.required_level, hero_level_param));
    END IF;

    -- TODO: 检查其他要求 (属性、技能等)
    -- 这里可以根据需要扩展更复杂的检查逻辑

    is_valid := array_length(missing_reqs, 1) IS NULL;

    RETURN QUERY SELECT is_valid, missing_reqs;
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