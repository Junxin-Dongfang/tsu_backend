-- =============================================================================
-- Create Skills Base System
-- 技能基础系统：技能定义、类别、等级配置等
-- 依赖：000001_create_core_infrastructure, 000004_create_classes_system
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 版本管理表 (技能配置版本管理)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_config_versions (
    id SERIAL PRIMARY KEY,
    version_number VARCHAR(20) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(100)
);

-- 确保只有一个活跃版本（唯一索引）
CREATE UNIQUE INDEX IF NOT EXISTS idx_skill_single_active_version ON skill_config_versions(is_active) WHERE is_active = TRUE;

-- --------------------------------------------------------------------------------
-- 技能类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_categories (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    category_code VARCHAR(50) NOT NULL,
    category_name VARCHAR(100) NOT NULL,
    description TEXT,

    -- 显示配置
    icon VARCHAR(256),
    color VARCHAR(16),
    display_order INTEGER DEFAULT 0,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, category_code)
);

-- 技能类别索引
CREATE INDEX IF NOT EXISTS idx_skill_categories_version_active ON skill_categories(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_skill_categories_display_order ON skill_categories(display_order);

-- --------------------------------------------------------------------------------
-- 技能定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skills (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    -- 技能基本信息
    skill_code VARCHAR(50) NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_type skill_type_enum NOT NULL DEFAULT 'active',
    category_id INTEGER REFERENCES skill_categories(id),

    -- 等级配置
    max_level INTEGER DEFAULT 10 CHECK (max_level >= 1 AND max_level <= 100),
    base_cooldown INTEGER DEFAULT 0 CHECK (base_cooldown >= 0), -- 基础冷却时间(秒)
    base_mana_cost INTEGER DEFAULT 0 CHECK (base_mana_cost >= 0), -- 基础魔法消耗

    -- 特征和效果
    feature_tags TEXT[], -- 特征标签数组
    passive_effects JSONB, -- 被动效果JSON配置
    active_effects JSONB, -- 主动效果JSON配置

    -- 学习要求
    required_level INTEGER DEFAULT 1 CHECK (required_level >= 1),
    required_class_ids UUID[], -- 可学习的职业ID数组
    prerequisite_skills INTEGER[], -- 前置技能ID数组

    -- 显示配置
    description TEXT,
    detailed_description TEXT, -- 详细描述
    icon VARCHAR(256),
    animation VARCHAR(256), -- 动画效果

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, skill_code)
);

-- 技能定义索引
CREATE INDEX IF NOT EXISTS idx_skills_version_active ON skills(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_skills_code ON skills(skill_code, version_id);
CREATE INDEX IF NOT EXISTS idx_skills_type ON skills(skill_type);
CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category_id);
CREATE INDEX IF NOT EXISTS idx_skills_max_level ON skills(max_level);

-- 特征标签索引（支持数组查询）
CREATE INDEX IF NOT EXISTS idx_skills_feature_tags ON skills USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_skills_required_class_ids ON skills USING GIN(required_class_ids);

-- --------------------------------------------------------------------------------
-- 技能等级配置表 (每个技能每个等级的具体数值)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_level_configs (
    id SERIAL PRIMARY KEY,
    skill_id INTEGER REFERENCES skills(id) ON DELETE CASCADE,

    level_number INTEGER NOT NULL CHECK (level_number >= 1),

    -- 等级特定数值
    damage_multiplier DECIMAL(5,2) DEFAULT 1.0, -- 伤害倍数
    healing_multiplier DECIMAL(5,2) DEFAULT 1.0, -- 治疗倍数
    duration_seconds INTEGER DEFAULT 0, -- 持续时间
    range_modifier INTEGER DEFAULT 0, -- 射程修正
    area_modifier INTEGER DEFAULT 0, -- 范围修正

    -- 消耗修正
    mana_cost_modifier INTEGER DEFAULT 0,
    cooldown_modifier INTEGER DEFAULT 0,

    -- 等级特定效果
    level_effects JSONB, -- 该等级的特殊效果

    -- 升级消耗
    upgrade_cost_xp INTEGER DEFAULT 0,
    upgrade_cost_gold INTEGER DEFAULT 0,
    upgrade_materials JSONB, -- 升级所需材料

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(skill_id, level_number)
);

-- 技能等级配置索引
CREATE INDEX IF NOT EXISTS idx_skill_level_configs_skill_id ON skill_level_configs(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_level_configs_level ON skill_level_configs(level_number);

-- --------------------------------------------------------------------------------
-- 技能学习条件表 (复杂的学习条件配置)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_learning_requirements (
    id SERIAL PRIMARY KEY,
    skill_id INTEGER REFERENCES skills(id) ON DELETE CASCADE,

    -- 条件类型
    requirement_type VARCHAR(50) NOT NULL, -- 'level', 'attribute', 'skill', 'item', 'quest'
    requirement_key VARCHAR(100) NOT NULL, -- 具体的条件键 (如属性代码、技能ID等)
    requirement_value INTEGER NOT NULL, -- 要求的数值

    -- 条件描述
    description TEXT,
    error_message TEXT, -- 不满足条件时的错误信息

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 技能学习条件索引
CREATE INDEX IF NOT EXISTS idx_skill_learning_requirements_skill ON skill_learning_requirements(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_learning_requirements_type ON skill_learning_requirements(requirement_type);

-- --------------------------------------------------------------------------------
-- 技能效果模板表 (可重用的技能效果)
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_effect_templates (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    template_code VARCHAR(50) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    effect_type VARCHAR(50) NOT NULL, -- 'damage', 'heal', 'buff', 'debuff', 'summon'

    -- 效果参数模板
    parameter_template JSONB NOT NULL, -- 参数模板
    calculation_formula TEXT, -- 计算公式

    -- 动画和视觉效果
    visual_effects JSONB, -- 视觉效果配置
    sound_effects JSONB, -- 音效配置

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, template_code)
);

-- 技能效果模板索引
CREATE INDEX IF NOT EXISTS idx_skill_effect_templates_version ON skill_effect_templates(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_skill_effect_templates_type ON skill_effect_templates(effect_type);

-- --------------------------------------------------------------------------------
-- 插入基础数据
-- --------------------------------------------------------------------------------

-- 插入初始版本
INSERT INTO skill_config_versions (version_number, description, is_active, created_by)
VALUES ('1.0.0', '技能系统初始版本', TRUE, 'system')
ON CONFLICT (version_number) DO NOTHING;

-- 插入技能类别
INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'COMBAT', '战斗技能', '用于战斗的主动和被动技能', '/icons/combat.png', '#E74C3C', 1
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'SUPPORT', '辅助技能', '提供支援和治疗的技能', '/icons/support.png', '#2ECC71', 2
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'PASSIVE', '被动技能', '自动生效的被动能力', '/icons/passive.png', '#3498DB', 3
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

INSERT INTO skill_categories (version_id, category_code, category_name, description, icon, color, display_order)
SELECT 1, 'ULTIMATE', '终极技能', '强大的终极能力', '/icons/ultimate.png', '#9B59B6', 4
WHERE EXISTS (SELECT 1 FROM skill_config_versions WHERE id = 1);

-- --------------------------------------------------------------------------------
-- 技能相关函数
-- --------------------------------------------------------------------------------

-- 检查技能学习条件的函数
CREATE OR REPLACE FUNCTION check_skill_learning_requirements(
    skill_id_param INTEGER,
    hero_id_param UUID
) RETURNS TABLE (
    is_eligible BOOLEAN,
    missing_requirements TEXT[]
) AS $$
DECLARE
    req_record RECORD;
    missing_reqs TEXT[] := '{}';
    is_valid BOOLEAN := TRUE;
    hero_level INTEGER;
    hero_attr_value DECIMAL;
    hero_skill_level INTEGER;
BEGIN
    -- 获取英雄基本信息
    SELECT level INTO hero_level FROM heroes WHERE id = hero_id_param;

    -- 检查所有学习要求
    FOR req_record IN
        SELECT requirement_type, requirement_key, requirement_value, error_message
        FROM skill_learning_requirements
        WHERE skill_id = skill_id_param AND is_active = TRUE
    LOOP
        CASE req_record.requirement_type
            WHEN 'level' THEN
                IF hero_level < req_record.requirement_value THEN
                    missing_reqs := array_append(missing_reqs,
                        COALESCE(req_record.error_message, format('需要等级 %s', req_record.requirement_value)));
                    is_valid := FALSE;
                END IF;

            WHEN 'attribute' THEN
                -- 检查属性要求
                SELECT ha.final_value INTO hero_attr_value
                FROM hero_attributes ha
                JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
                WHERE ha.hero_id = hero_id_param AND hat.attribute_code = req_record.requirement_key;

                IF COALESCE(hero_attr_value, 0) < req_record.requirement_value THEN
                    missing_reqs := array_append(missing_reqs,
                        COALESCE(req_record.error_message, format('需要 %s 达到 %s', req_record.requirement_key, req_record.requirement_value)));
                    is_valid := FALSE;
                END IF;

            WHEN 'skill' THEN
                -- 检查前置技能要求
                SELECT skill_level INTO hero_skill_level
                FROM hero_skills hs
                JOIN skills s ON hs.skill_code = s.skill_code
                WHERE hs.hero_id = hero_id_param AND s.id = req_record.requirement_key::INTEGER;

                IF COALESCE(hero_skill_level, 0) < req_record.requirement_value THEN
                    missing_reqs := array_append(missing_reqs,
                        COALESCE(req_record.error_message, '缺少前置技能'));
                    is_valid := FALSE;
                END IF;

            -- 可以继续添加其他条件类型
        END CASE;
    END LOOP;

    RETURN QUERY SELECT is_valid, missing_reqs;
END;
$$ LANGUAGE plpgsql;

-- 获取技能详细信息的函数
CREATE OR REPLACE FUNCTION get_skill_details(skill_id_param INTEGER, skill_level_param INTEGER DEFAULT 1)
RETURNS TABLE (
    skill_id INTEGER,
    skill_name VARCHAR,
    skill_type skill_type_enum,
    category_name VARCHAR,
    max_level INTEGER,
    current_damage_multiplier DECIMAL,
    current_mana_cost INTEGER,
    current_cooldown INTEGER,
    description TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.id,
        s.skill_name,
        s.skill_type,
        sc.category_name,
        s.max_level,
        COALESCE(slc.damage_multiplier, 1.0),
        s.base_mana_cost + COALESCE(slc.mana_cost_modifier, 0),
        s.base_cooldown + COALESCE(slc.cooldown_modifier, 0),
        s.description
    FROM skills s
    JOIN skill_categories sc ON s.category_id = sc.id
    LEFT JOIN skill_level_configs slc ON s.id = slc.skill_id AND slc.level_number = skill_level_param
    WHERE s.id = skill_id_param AND s.is_active = TRUE;
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Base System 创建完成';
    RAISE NOTICE '包含: 技能定义、类别、等级配置、学习条件';
    RAISE NOTICE '已插入 % 个技能类别', (SELECT COUNT(*) FROM skill_categories WHERE is_active = TRUE);
    RAISE NOTICE '============================================';
END $$;