-- =============================================================================
-- Create Skills Advanced System
-- 技能高级系统：动作、效果、Buff、战斗逻辑等
-- 依赖：000006_create_skills_base
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 动作类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_categories (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    category_code VARCHAR(50) NOT NULL,
    category_name VARCHAR(100) NOT NULL,
    description TEXT,

    -- 动作类别属性
    default_action_type action_type_enum DEFAULT 'main',
    default_range INTEGER DEFAULT 1,
    default_area INTEGER DEFAULT 1,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, category_code)
);

-- 动作类别索引
CREATE INDEX IF NOT EXISTS idx_action_categories_version ON action_categories(version_id, is_active);

-- --------------------------------------------------------------------------------
-- 动作定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS actions (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    -- 动作基本信息
    action_code VARCHAR(50) NOT NULL,
    action_name VARCHAR(100) NOT NULL,
    action_category_id INTEGER REFERENCES action_categories(id),
    action_type action_type_enum NOT NULL DEFAULT 'main',

    -- 关联技能
    related_skill_id INTEGER REFERENCES skills(id),
    skill_level_required INTEGER DEFAULT 1,

    -- 动作参数
    feature_tags TEXT[], -- 动作特征数组
    range_config JSONB NOT NULL DEFAULT '{"range": 1}', -- 射程配置
    target_config JSONB DEFAULT '{"type": "single", "ally": false}', -- 目标配置
    area_config JSONB DEFAULT '{"type": "single"}', -- 范围配置

    -- 消耗和限制
    action_point_cost INTEGER DEFAULT 1 CHECK (action_point_cost >= 0),
    mana_cost INTEGER DEFAULT 0 CHECK (mana_cost >= 0),
    cooldown_turns INTEGER DEFAULT 0 CHECK (cooldown_turns >= 0),
    uses_per_battle INTEGER, -- 每场战斗使用次数限制

    -- 效果配置
    effect_config JSONB NOT NULL DEFAULT '{}', -- 完整效果配置JSON
    start_flags TEXT[], -- 动作开始时赋予的flag数组
    requirements JSONB, -- 使用条件

    -- 动画和视觉效果
    animation_config JSONB, -- 动画配置
    visual_effects JSONB, -- 视觉效果
    sound_effects JSONB, -- 音效配置

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, action_code)
);

-- 动作定义索引
CREATE INDEX IF NOT EXISTS idx_actions_version_active ON actions(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_actions_code ON actions(action_code, version_id);
CREATE INDEX IF NOT EXISTS idx_actions_category ON actions(action_category_id);
CREATE INDEX IF NOT EXISTS idx_actions_skill ON actions(related_skill_id);
CREATE INDEX IF NOT EXISTS idx_actions_type ON actions(action_type);

-- 特征标签索引（支持数组查询）
CREATE INDEX IF NOT EXISTS idx_actions_feature_tags ON actions USING GIN(feature_tags);

-- --------------------------------------------------------------------------------
-- 技能解锁动作关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_unlock_actions (
    id SERIAL PRIMARY KEY,
    skill_id INTEGER REFERENCES skills(id) ON DELETE CASCADE,
    action_id INTEGER REFERENCES actions(id) ON DELETE CASCADE,

    unlock_level INTEGER NOT NULL DEFAULT 1 CHECK (unlock_level >= 1),
    is_default BOOLEAN DEFAULT FALSE, -- 是否为技能默认动作

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(skill_id, action_id)
);

-- 技能解锁动作关联索引
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_skill_id ON skill_unlock_actions(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_action_id ON skill_unlock_actions(action_id);
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_level ON skill_unlock_actions(unlock_level);

-- --------------------------------------------------------------------------------
-- Buff定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS buffs (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    -- Buff基本信息
    buff_code VARCHAR(50) NOT NULL,
    buff_name VARCHAR(100) NOT NULL,
    buff_type VARCHAR(50) NOT NULL DEFAULT 'buff', -- 'buff', 'debuff', 'neutral'

    -- Buff分类
    category VARCHAR(50), -- 'magic', 'physical', 'environmental', 'curse'
    feature_tags TEXT[], -- buff特征数组

    -- 持续时间配置
    default_duration INTEGER DEFAULT 1, -- 默认持续时间（回合数）
    max_duration INTEGER DEFAULT 10, -- 最大持续时间
    min_duration INTEGER DEFAULT 1, -- 最小持续时间

    -- 叠加规则
    stack_rule VARCHAR(50) DEFAULT 'no_stack', -- 'no_stack', 'stackable', 'replace', 'refresh'
    max_stacks INTEGER DEFAULT 1, -- 最大叠加层数

    -- 效果配置
    effect_description TEXT, -- 效果描述
    parameters TEXT[], -- 参数名称数组
    parameter_definitions JSONB, -- 参数定义JSON
    effect_formula TEXT, -- 效果计算公式

    -- 触发条件
    trigger_events TEXT[], -- 触发事件数组 ('turn_start', 'turn_end', 'take_damage', etc.)
    application_conditions JSONB, -- 应用条件

    -- 视觉效果
    visual_effects JSONB, -- 视觉效果配置
    icon VARCHAR(256), -- Buff图标
    color VARCHAR(16), -- Buff颜色

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, buff_code)
);

-- Buff定义索引
CREATE INDEX IF NOT EXISTS idx_buffs_version_active ON buffs(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_buffs_code ON buffs(buff_code, version_id);
CREATE INDEX IF NOT EXISTS idx_buffs_type ON buffs(buff_type);
CREATE INDEX IF NOT EXISTS idx_buffs_category ON buffs(category);
CREATE INDEX IF NOT EXISTS idx_buffs_stack_rule ON buffs(stack_rule);

-- 特征标签和触发事件索引
CREATE INDEX IF NOT EXISTS idx_buffs_feature_tags ON buffs USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_buffs_trigger_events ON buffs USING GIN(trigger_events);

-- --------------------------------------------------------------------------------
-- 动作Flag定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_flags (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    flag_code VARCHAR(50) NOT NULL,
    flag_name VARCHAR(100) NOT NULL,
    category VARCHAR(50), -- 'action_chain', 'status_check', 'modifier'

    -- 持续时间配置
    duration_type VARCHAR(20) DEFAULT 'action', -- 'action', 'turn', 'battle', 'permanent'
    default_duration VARCHAR(20) DEFAULT '1', -- 默认持续时间，可以是数字或特殊值如 "-"

    -- 自动移除条件
    auto_remove_condition VARCHAR(100), -- 'action_end', 'turn_start', 'turn_end'
    remove_on_events TEXT[], -- 触发移除的事件

    -- Flag属性
    is_visible BOOLEAN DEFAULT FALSE, -- 是否在UI中显示
    is_stackable BOOLEAN DEFAULT FALSE, -- 是否可叠加
    max_stacks INTEGER DEFAULT 1,

    -- 效果
    flag_effects JSONB, -- Flag产生的效果
    modifier_effects JSONB, -- 修饰符效果

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, flag_code)
);

-- 动作Flag定义索引
CREATE INDEX IF NOT EXISTS idx_action_flags_version ON action_flags(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_action_flags_category ON action_flags(category);
CREATE INDEX IF NOT EXISTS idx_action_flags_duration_type ON action_flags(duration_type);

-- --------------------------------------------------------------------------------
-- 伤害类型定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS damage_types (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50), -- 'physical', 'magical', 'elemental', 'pure'

    -- 抗性关联
    resistance_attribute VARCHAR(50), -- 对应的抗性属性代码
    damage_reduction_formula TEXT, -- 伤害减免计算公式

    -- 视觉效果
    color VARCHAR(7), -- 十六进制颜色
    icon VARCHAR(256), -- 伤害类型图标
    particle_effect VARCHAR(256), -- 粒子效果

    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(version_id, code)
);

-- 伤害类型索引
CREATE INDEX IF NOT EXISTS idx_damage_types_version ON damage_types(version_id, is_active);
CREATE INDEX IF NOT EXISTS idx_damage_types_category ON damage_types(category);

-- --------------------------------------------------------------------------------
-- 射程配置规则表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS range_config_rules (
    id SERIAL PRIMARY KEY,
    version_id INTEGER REFERENCES skill_config_versions(id) ON DELETE CASCADE,

    parameter_type VARCHAR(50) NOT NULL, -- 'range', 'positions', 'area', 'line'
    parameter_format VARCHAR(100) NOT NULL, -- 'N', '0~N', 'M~N', '最近N个', '直线N格'

    -- 规则说明
    description TEXT,
    example VARCHAR(200), -- 示例
    validation_pattern VARCHAR(200), -- 验证正则表达式
    notes TEXT,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 射程配置规则索引
CREATE INDEX IF NOT EXISTS idx_range_config_rules_type ON range_config_rules(parameter_type);

-- --------------------------------------------------------------------------------
-- 插入基础数据
-- --------------------------------------------------------------------------------

-- 插入动作类别示例数据
INSERT INTO action_categories (version_id, category_code, category_name, description, default_action_type) VALUES
    (1, 'BASIC_ATTACK', '基础攻击', '基础的物理和魔法攻击动作', 'main'),
    (1, 'DEFENSIVE', '防御动作', '防御、格挡、闪避等防御性动作', 'reaction'),
    (1, 'MOVEMENT', '移动动作', '移动、传送、位置变换等动作', 'minor')
ON CONFLICT (version_id, category_code) DO NOTHING;

-- 插入基础伤害类型
INSERT INTO damage_types (version_id, code, name, category, resistance_attribute, color) VALUES
    (1, 'PHYSICAL', '物理伤害', 'physical', 'DEFENSE', '#8B4513'),
    (1, 'MAGICAL', '魔法伤害', 'magical', 'MAGIC_RESISTANCE', '#4B0082'),
    (1, 'FIRE', '火焰伤害', 'elemental', 'FIRE_RESISTANCE', '#FF4500'),
    (1, 'ICE', '冰霜伤害', 'elemental', 'ICE_RESISTANCE', '#00BFFF'),
    (1, 'LIGHTNING', '雷电伤害', 'elemental', 'LIGHTNING_RESISTANCE', '#FFD700')
ON CONFLICT (version_id, code) DO NOTHING;

-- 插入基础Action Flag
INSERT INTO action_flags (version_id, flag_code, flag_name, category, duration_type, description) VALUES
    (1, 'ATTACKING', '攻击状态', 'action_chain', 'action', '正在执行攻击动作'),
    (1, 'DEFENDING', '防御状态', 'action_chain', 'turn', '处于防御状态，减少受到的伤害'),
    (1, 'MOVED', '已移动', 'status_check', 'turn', '本回合已经移动过')
ON CONFLICT (version_id, flag_code) DO NOTHING;

-- --------------------------------------------------------------------------------
-- 高级技能相关函数
-- --------------------------------------------------------------------------------

-- 计算技能伤害的函数
CREATE OR REPLACE FUNCTION calculate_skill_damage(
    action_id_param INTEGER,
    caster_id UUID,
    target_id UUID,
    skill_level_param INTEGER DEFAULT 1
) RETURNS TABLE (
    base_damage INTEGER,
    final_damage INTEGER,
    damage_type VARCHAR,
    is_critical BOOLEAN
) AS $$
DECLARE
    action_record RECORD;
    caster_attack DECIMAL;
    target_defense DECIMAL;
    damage_multiplier DECIMAL := 1.0;
    crit_chance DECIMAL;
    crit_damage DECIMAL;
    is_crit BOOLEAN := FALSE;
    base_dmg INTEGER;
    final_dmg INTEGER;
BEGIN
    -- 获取动作信息
    SELECT * INTO action_record FROM actions WHERE id = action_id_param;

    -- 获取施法者攻击力
    SELECT final_value INTO caster_attack
    FROM hero_attributes ha
    JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
    WHERE ha.hero_id = caster_id AND hat.attribute_code = 'ATTACK_POWER';
    caster_attack := COALESCE(caster_attack, 100);

    -- 获取目标防御力
    SELECT final_value INTO target_defense
    FROM hero_attributes ha
    JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
    WHERE ha.hero_id = target_id AND hat.attribute_code = 'DEFENSE';
    target_defense := COALESCE(target_defense, 50);

    -- 获取技能等级倍数
    SELECT damage_multiplier INTO damage_multiplier
    FROM skill_level_configs slc
    JOIN skill_unlock_actions sua ON slc.skill_id = sua.skill_id
    WHERE sua.action_id = action_id_param AND slc.level_number = skill_level_param;
    damage_multiplier := COALESCE(damage_multiplier, 1.0);

    -- 计算基础伤害
    base_dmg := (caster_attack * damage_multiplier)::INTEGER;

    -- 检查暴击
    SELECT final_value INTO crit_chance
    FROM hero_attributes ha
    JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
    WHERE ha.hero_id = caster_id AND hat.attribute_code = 'CRITICAL_CHANCE';
    crit_chance := COALESCE(crit_chance, 5) / 100.0;

    IF random() < crit_chance THEN
        is_crit := TRUE;
        SELECT final_value INTO crit_damage
        FROM hero_attributes ha
        JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
        WHERE ha.hero_id = caster_id AND hat.attribute_code = 'CRITICAL_DAMAGE';
        crit_damage := COALESCE(crit_damage, 150) / 100.0;
        base_dmg := (base_dmg * crit_damage)::INTEGER;
    END IF;

    -- 应用防御减免
    final_dmg := GREATEST(1, base_dmg - target_defense::INTEGER);

    RETURN QUERY SELECT base_dmg, final_dmg, 'PHYSICAL'::VARCHAR, is_crit;
END;
$$ LANGUAGE plpgsql;

-- 获取可用动作列表的函数
CREATE OR REPLACE FUNCTION get_available_actions(
    hero_id_param UUID,
    action_type_filter action_type_enum DEFAULT NULL
) RETURNS TABLE (
    action_id INTEGER,
    action_name VARCHAR,
    action_type action_type_enum,
    mana_cost INTEGER,
    ap_cost INTEGER,
    cooldown_turns INTEGER,
    skill_name VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.action_name,
        a.action_type,
        a.mana_cost,
        a.action_point_cost,
        a.cooldown_turns,
        s.skill_name
    FROM actions a
    JOIN skill_unlock_actions sua ON a.id = sua.action_id
    JOIN skills s ON sua.skill_id = s.id
    JOIN hero_skills hs ON s.skill_code = hs.skill_code
    WHERE hs.hero_id = hero_id_param
      AND hs.is_active = TRUE
      AND hs.skill_level >= sua.unlock_level
      AND a.is_active = TRUE
      AND (action_type_filter IS NULL OR a.action_type = action_type_filter)
    ORDER BY a.action_type, a.action_name;
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 视图定义（便于查询）
-- --------------------------------------------------------------------------------

-- 活跃技能配置概览视图
CREATE OR REPLACE VIEW active_skill_config_summary AS
SELECT
    cv.version_number,
    cv.description,
    cv.created_at as version_created_at,
    (SELECT COUNT(*) FROM skills s WHERE s.version_id = cv.id AND s.is_active = true) as skills_count,
    (SELECT COUNT(*) FROM actions a WHERE a.version_id = cv.id AND a.is_active = true) as actions_count,
    (SELECT COUNT(*) FROM buffs b WHERE b.version_id = cv.id AND b.is_active = true) as buffs_count,
    (SELECT COUNT(*) FROM damage_types dt WHERE dt.version_id = cv.id AND dt.is_active = true) as damage_types_count
FROM skill_config_versions cv
WHERE cv.is_active = true;

-- 技能与解锁动作关联视图
CREATE OR REPLACE VIEW skill_action_relations AS
SELECT
    s.skill_code,
    s.skill_name,
    a.action_code,
    a.action_name,
    sua.unlock_level,
    s.version_id
FROM skills s
JOIN skill_unlock_actions sua ON s.id = sua.skill_id
JOIN actions a ON sua.action_id = a.id
WHERE s.is_active = true AND a.is_active = true;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Skills Advanced System 创建完成';
    RAISE NOTICE '包含: 动作系统、Buff系统、战斗逻辑、效果配置';
    RAISE NOTICE '已插入 % 个动作类别', (SELECT COUNT(*) FROM action_categories WHERE is_active = TRUE);
    RAISE NOTICE '已插入 % 个伤害类型', (SELECT COUNT(*) FROM damage_types WHERE is_active = TRUE);
    RAISE NOTICE '已插入 % 个动作标识', (SELECT COUNT(*) FROM action_flags WHERE is_active = TRUE);
    RAISE NOTICE '============================================';
END $$;