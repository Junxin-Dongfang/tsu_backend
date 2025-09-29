-- =============================================================================
-- 000006_create_skills_base.up.sql
-- 技能基础系统：技能、动作、效果、判定系统（完整版）
-- 依赖：000001_create_core_infrastructure, 000003_create_attribute_system, 000004_create_classes_system
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 1. 枚举类型定义
-- --------------------------------------------------------------------------------

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
-- 2. 技能类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_categories (
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
    ON skill_categories(category_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_categories_active 
    ON skill_categories(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE TRIGGER update_skill_categories_updated_at
    BEFORE UPDATE ON skill_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE skill_categories IS '技能类别表';

-- --------------------------------------------------------------------------------
-- 3. 动作类别表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_categories (
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
    ON action_categories(category_code) WHERE deleted_at IS NULL;

CREATE TRIGGER update_action_categories_updated_at
    BEFORE UPDATE ON action_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE action_categories IS '动作类别表';

-- --------------------------------------------------------------------------------
-- 4. 伤害类型定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS damage_types (
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
    ON damage_types(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_damage_types_category 
    ON damage_types(category) WHERE deleted_at IS NULL;

CREATE TRIGGER update_damage_types_updated_at
    BEFORE UPDATE ON damage_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE damage_types IS '伤害类型定义表';
COMMENT ON COLUMN damage_types.resistance_cap IS '该伤害类型的抗性上限(%)';
COMMENT ON COLUMN damage_types.damage_reduction_attribute_code IS '固定值减免属性代码（如FIRE_DmgReduce）';

-- --------------------------------------------------------------------------------
-- 5. 元效果类型定义表（新增）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS effect_type_definitions (
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
    ON effect_type_definitions(effect_type_code) WHERE deleted_at IS NULL;

CREATE TRIGGER update_effect_type_definitions_updated_at
    BEFORE UPDATE ON effect_type_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE effect_type_definitions IS '元效果类型定义表 - 定义所有可用的元效果类型及其参数规范';
COMMENT ON COLUMN effect_type_definitions.effect_type_code IS '元效果代码（如HIT_CHECK、DMG_CALCULATION）';
COMMENT ON COLUMN effect_type_definitions.parameter_list IS '参数名称数组';
COMMENT ON COLUMN effect_type_definitions.failure_handling IS '失败处理：skip_remaining跳过后续效果, continue继续执行';

-- --------------------------------------------------------------------------------
-- 6. 公式变量定义表（新增）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS formula_variables (
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
    ON formula_variables(variable_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_formula_variables_scope 
    ON formula_variables(scope) WHERE deleted_at IS NULL;

CREATE TRIGGER update_formula_variables_updated_at
    BEFORE UPDATE ON formula_variables
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE formula_variables IS '公式变量定义表 - 定义所有可在配置公式中使用的变量';
COMMENT ON COLUMN formula_variables.variable_type IS '变量类型：基础属性、目标选择、技能数据、装备数据';
COMMENT ON COLUMN formula_variables.scope IS '变量作用域：角色、动作、技能、装备、全局';

-- --------------------------------------------------------------------------------
-- 7. 射程配置规则表（新增 - 元数据）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS range_config_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    parameter_type VARCHAR(50) NOT NULL,    -- 'range', 'positions', 'depth'
    parameter_format VARCHAR(100) NOT NULL, -- 格式说明（如'N', '0~N', '最近N位'）
    
    description TEXT,
    example VARCHAR(200),
    notes TEXT,
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE range_config_rules IS '射程配置规则表 - 文档用途，说明射程参数的格式规范';

-- --------------------------------------------------------------------------------
-- 8. 动作类型定义表（新增 - 元数据）
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_type_definitions (
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
    ON action_type_definitions(action_type);

COMMENT ON TABLE action_type_definitions IS '动作类型定义表 - 说明main/minor/reaction的规则';

-- --------------------------------------------------------------------------------
-- 9. 效果定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS effects (
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
    ON effects(effect_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_effects_type 
    ON effects(effect_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_effects_feature_tags 
    ON effects USING GIN(feature_tags);

CREATE TRIGGER update_effects_updated_at
    BEFORE UPDATE ON effects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE effects IS '效果定义表 - 所有游戏效果的原子单位';
COMMENT ON COLUMN effects.effect_type IS '效果类型，关联effect_type_definitions表';
COMMENT ON COLUMN effects.parameters IS 'JSON格式的效果参数，不同effect_type有不同的参数结构';

-- --------------------------------------------------------------------------------
-- 10. 技能定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- 技能基本信息
    skill_code VARCHAR(50) NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_type skill_type_enum NOT NULL DEFAULT 'weapon',
    category_id UUID REFERENCES skill_categories(id),
    
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
    ON skills(skill_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_type 
    ON skills(skill_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_category 
    ON skills(category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skills_feature_tags 
    ON skills USING GIN(feature_tags);

CREATE TRIGGER update_skills_updated_at
    BEFORE UPDATE ON skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE skills IS '技能定义表';
COMMENT ON COLUMN skills.passive_effects IS '被动效果配置（JSONB格式）';

-- --------------------------------------------------------------------------------
-- 11. 技能等级配置表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_level_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    
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
    ON skill_level_configs(skill_id) WHERE deleted_at IS NULL;

CREATE TRIGGER update_skill_level_configs_updated_at
    BEFORE UPDATE ON skill_level_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE skill_level_configs IS '技能等级配置表 - 每个技能每个等级的具体数值';

-- --------------------------------------------------------------------------------
-- 12. 动作定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- 动作基本信息
    action_code VARCHAR(50) NOT NULL,
    action_name VARCHAR(100) NOT NULL,
    action_category_id UUID REFERENCES action_categories(id),
    action_type action_type_enum NOT NULL DEFAULT 'main',
    
    -- 关联技能（可选）
    related_skill_id UUID REFERENCES skills(id),
    
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
    ON actions(action_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_category 
    ON actions(action_category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_skill 
    ON actions(related_skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_type 
    ON actions(action_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actions_feature_tags 
    ON actions USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_actions_start_flags 
    ON actions USING GIN(start_flags);

CREATE TRIGGER update_actions_updated_at
    BEFORE UPDATE ON actions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE actions IS '动作定义表';
COMMENT ON COLUMN actions.hit_rate_config IS '命中率计算配置';
COMMENT ON COLUMN actions.legacy_effect_config IS 'Excel原始效果配置（用于兼容导入）';
COMMENT ON COLUMN actions.mana_cost_formula IS 'MP消耗公式（如"50+2*skill_level"），优先于mana_cost';
COMMENT ON COLUMN actions.start_flags IS '开始标记数组（如STARTING_ATTACK_ACTION）';

-- --------------------------------------------------------------------------------
-- 13. Action与Effect关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    effect_id UUID NOT NULL REFERENCES effects(id) ON DELETE CASCADE,
    
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
    ON action_effects(action_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_effects_effect 
    ON action_effects(effect_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_effects_order 
    ON action_effects(action_id, execution_order);

COMMENT ON TABLE action_effects IS 'Action与Effect关联表';

-- --------------------------------------------------------------------------------
-- 14. 技能解锁动作关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS skill_unlock_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    
    unlock_level INTEGER NOT NULL DEFAULT 1 CHECK (unlock_level >= 1),
    is_default BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(skill_id, action_id)
);

CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_skill 
    ON skill_unlock_actions(skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_action 
    ON skill_unlock_actions(action_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_skill_unlock_actions_level 
    ON skill_unlock_actions(unlock_level);

COMMENT ON TABLE skill_unlock_actions IS '技能解锁动作关联表';

-- --------------------------------------------------------------------------------
-- 15. Buff定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS buffs (
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
    ON buffs(buff_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_type 
    ON buffs(buff_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_category 
    ON buffs(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buffs_feature_tags 
    ON buffs USING GIN(feature_tags);
CREATE INDEX IF NOT EXISTS idx_buffs_trigger_events 
    ON buffs USING GIN(trigger_events);

CREATE TRIGGER update_buffs_updated_at
    BEFORE UPDATE ON buffs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE buffs IS 'Buff定义表';
COMMENT ON COLUMN buffs.effect_description IS 'Buff效果的文字描述（面向玩家）';
COMMENT ON COLUMN buffs.parameter_list IS 'Buff的参数名称列表';
COMMENT ON COLUMN buffs.parameter_definitions IS 'Buff参数的详细定义（类型、范围等）';

-- --------------------------------------------------------------------------------
-- 16. Buff与Effect关联表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS buff_effects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    buff_id UUID NOT NULL REFERENCES buffs(id) ON DELETE CASCADE,
    effect_id UUID NOT NULL REFERENCES effects(id) ON DELETE CASCADE,
    
    trigger_timing VARCHAR(50) NOT NULL,  -- 'on_apply', 'turn_start', 'turn_end', 'on_remove'
    execution_order INTEGER DEFAULT 0,
    
    parameter_overrides JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(buff_id, effect_id, trigger_timing, execution_order)
);

CREATE INDEX IF NOT EXISTS idx_buff_effects_buff 
    ON buff_effects(buff_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buff_effects_effect 
    ON buff_effects(effect_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_buff_effects_timing 
    ON buff_effects(trigger_timing);

COMMENT ON TABLE buff_effects IS 'Buff与Effect关联表';

-- --------------------------------------------------------------------------------
-- 17. 动作Flag定义表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS action_flags (
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
    ON action_flags(flag_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_flags_category 
    ON action_flags(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_action_flags_duration_type 
    ON action_flags(duration_type) WHERE deleted_at IS NULL;

CREATE TRIGGER update_action_flags_updated_at
    BEFORE UPDATE ON action_flags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE action_flags IS '动作Flag定义表';
COMMENT ON COLUMN action_flags.auto_remove_condition IS '自动移除条件（act_end/next_act_start/turn_start/turn_end等）';

-- --------------------------------------------------------------------------------
-- 18. Effect与Tag的关联通过已有的 tags_relations 表实现
-- --------------------------------------------------------------------------------
-- 无需创建单独的 effect_tags 表
-- 使用方式：entity_type='effect', entity_id=effect.id

-- --------------------------------------------------------------------------------
-- 19. 英雄技能表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    hero_id UUID NOT NULL,
    skill_id UUID NOT NULL REFERENCES skills(id),
    skill_code VARCHAR(32) NOT NULL,
    
    -- 技能等级信息
    skill_level INTEGER NOT NULL DEFAULT 1 CHECK (skill_level >= 1),
    skill_experience INTEGER NOT NULL DEFAULT 0 CHECK (skill_experience >= 0),
    max_level INTEGER NOT NULL DEFAULT 10 CHECK (max_level >= 1),
    
    -- 技能状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_equipped BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- 获得信息
    learned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    learned_method VARCHAR(50) DEFAULT 'class_unlock',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(hero_id, skill_id),
    UNIQUE(hero_id, skill_code)
);

CREATE INDEX IF NOT EXISTS idx_hero_skills_hero ON hero_skills(hero_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_skills_skill ON hero_skills(skill_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_hero_skills_equipped ON hero_skills(is_equipped) WHERE is_equipped = TRUE;

CREATE TRIGGER update_hero_skills_updated_at
    BEFORE UPDATE ON hero_skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE hero_skills IS '英雄技能表';

-- --------------------------------------------------------------------------------
-- 20. 核心函数
-- --------------------------------------------------------------------------------

-- 计算命中率
CREATE OR REPLACE FUNCTION calculate_hit_rate(
    p_action_id UUID,
    p_attacker_id UUID,
    p_target_id UUID
) RETURNS DECIMAL AS $$
DECLARE
    v_config JSONB;
    v_base_hit_rate DECIMAL;
    v_attacker_accuracy DECIMAL := 0;
    v_target_defense DECIMAL := 0;
    v_final_hit_rate DECIMAL;
    v_min_rate DECIMAL;
    v_max_rate DECIMAL;
BEGIN
    -- 获取动作的命中率配置
    SELECT hit_rate_config INTO v_config
    FROM actions WHERE id = p_action_id AND deleted_at IS NULL;
    
    IF v_config IS NULL THEN
        RETURN 75.0;  -- 默认命中率
    END IF;
    
    v_base_hit_rate := (v_config->>'base_hit_rate')::DECIMAL;
    v_min_rate := (v_config->>'min_hit_rate')::DECIMAL;
    v_max_rate := (v_config->>'max_hit_rate')::DECIMAL;
    
    -- 获取攻击者命中属性
    IF (v_config->>'use_attacker_accuracy')::BOOLEAN THEN
        SELECT ha.final_value INTO v_attacker_accuracy
        FROM hero_attributes ha
        JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
        WHERE ha.hero_id = p_attacker_id 
          AND hat.attribute_code = v_config->>'accuracy_attribute'
          AND ha.deleted_at IS NULL;
        
        v_attacker_accuracy := COALESCE(v_attacker_accuracy, 0) * 
                               COALESCE((v_config->>'accuracy_multiplier')::DECIMAL, 1.0);
    END IF;
    
    -- 获取目标防御属性（如果配置需要）
    IF (v_config->>'use_target_defense')::BOOLEAN THEN
        SELECT ha.final_value INTO v_target_defense
        FROM hero_attributes ha
        JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
        WHERE ha.hero_id = p_target_id 
          AND hat.attribute_code = v_config->>'defense_attribute'
          AND ha.deleted_at IS NULL;
        
        v_target_defense := COALESCE(v_target_defense, 0) * 
                            COALESCE((v_config->>'defense_multiplier')::DECIMAL, 1.0);
    END IF;
    
    -- 计算最终命中率
    v_final_hit_rate := v_base_hit_rate + v_attacker_accuracy - v_target_defense;
    
    -- 限制在最小/最大范围内
    v_final_hit_rate := GREATEST(v_min_rate, LEAST(v_max_rate, v_final_hit_rate));
    
    RETURN v_final_hit_rate;
END;
$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_hit_rate IS '计算动作的命中率（基于配置和角色属性）';

-- --------------------------------------------------------------------------------
-- 21. 视图定义
-- --------------------------------------------------------------------------------

-- 技能与动作关联视图
CREATE OR REPLACE VIEW skill_action_relations AS
SELECT
    s.skill_code,
    s.skill_name,
    a.action_code,
    a.action_name,
    sua.unlock_level,
    sua.is_default
FROM skills s
JOIN skill_unlock_actions sua ON s.id = sua.skill_id
JOIN actions a ON sua.action_id = a.id
WHERE s.is_active = TRUE 
  AND a.is_active = TRUE 
  AND s.deleted_at IS NULL 
  AND a.deleted_at IS NULL
  AND sua.deleted_at IS NULL;

COMMENT ON VIEW skill_action_relations IS '技能与动作关联视图';

-- 动作及其效果视图
CREATE OR REPLACE VIEW action_effects_detail AS
SELECT
    a.action_code,
    a.action_name,
    e.effect_code,
    e.effect_name,
    e.effect_type,
    ae.execution_order,
    ae.parameter_overrides
FROM actions a
JOIN action_effects ae ON a.id = ae.action_id
JOIN effects e ON ae.effect_id = e.id
WHERE a.is_active = TRUE 
  AND e.is_active = TRUE
  AND a.deleted_at IS NULL 
  AND e.deleted_at IS NULL
  AND ae.deleted_at IS NULL
ORDER BY a.action_code, ae.execution_order;

COMMENT ON VIEW action_effects_detail IS '动作及其效果详情视图';

-- Buff及其效果视图
CREATE OR REPLACE VIEW buff_effects_detail AS
SELECT
    b.buff_code,
    b.buff_name,
    e.effect_code,
    e.effect_name,
    be.trigger_timing,
    be.execution_order
FROM buffs b
JOIN buff_effects be ON b.id = be.buff_id
JOIN effects e ON be.effect_id = e.id
WHERE b.is_active = TRUE 
  AND e.is_active = TRUE
  AND b.deleted_at IS NULL 
  AND e.deleted_at IS NULL
  AND be.deleted_at IS NULL
ORDER BY b.buff_code, be.trigger_timing, be.execution_order;

COMMENT ON VIEW buff_effects_detail IS 'Buff及其效果详情视图';

-- --------------------------------------------------------------------------------
-- 22. 插入基础数据
-- --------------------------------------------------------------------------------

-- 插入技能类别
INSERT INTO skill_categories (category_code, category_name, description, icon, color, display_order)
VALUES 
    ('COMBAT', '战斗技能', '用于战斗的主动和被动技能', '/icons/combat.png', '#E74C3C', 1),
    ('SUPPORT', '辅助技能', '提供支援和治疗的技能', '/icons/support.png', '#2ECC71', 2),
    ('PASSIVE', '被动技能', '自动生效的被动能力', '/icons/passive.png', '#3498DB', 3),
    ('ULTIMATE', '终极技能', '强大的终极能力', '/icons/ultimate.png', '#9B59B6', 4)
ON CONFLICT (category_code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入动作类别
INSERT INTO action_categories (category_code, category_name, description)
VALUES 
    ('BASIC_ATTACK', '基础攻击', '基础的物理和魔法攻击动作'),
    ('BASIC_SAVE', '豁免检定', '需要目标进行豁免的动作'),
    ('DEFENSIVE', '防御动作', '防御、格挡、闪避等防御性动作'),
    ('MOVEMENT', '移动动作', '移动、传送、位置变换等动作'),
    ('SPECIAL', '特殊动作', '特殊效果动作'),
    ('SUMMON', '召唤', '召唤生物或单位'),
    ('COMMAND', '指挥', '指挥友军或召唤物'),
    ('GUARD', '守护', '守护盟友'),
    ('GROUND_EFFECT', '地面效果', '创造地面效果'),
    ('ITEM_USAGE', '物品使用', '使用物品'),
    ('BASIC_REACTION', '反应动作', '反应性动作')
ON CONFLICT (category_code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入伤害类型
INSERT INTO damage_types (code, name, category, resistance_attribute_code, damage_reduction_attribute_code, resistance_cap, color)
VALUES 
    ('PHYSICAL', '物理伤害', 'physical', 'PHYSICAL_RESIST', NULL, 75, '#8B4513'),
    ('MAGICAL', '魔法伤害', 'magical', 'MAGIC_RESISTANCE', NULL, 75, '#4B0082'),
    ('FIRE', '火焰伤害', 'elemental', 'FIRE_RESIST', 'FIRE_DmgReduce', 75, '#FF4500'),
    ('ICE', '冰霜伤害', 'elemental', 'ICE_RESIST', NULL, 75, '#00BFFF'),
    ('LIGHTNING', '雷电伤害', 'elemental', 'LIGHTNING_RESISTANCE', NULL, 75, '#FFD700'),
    ('POISON', '毒素伤害', 'elemental', 'POISON_RESISTANCE', NULL, 75, '#32CD32'),
    ('PURE', '真实伤害', 'special', NULL, NULL, 0, '#FFFFFF'),
    ('ALL', '所有伤害', 'all', NULL, NULL, 75, '#FFFFFF'),
    ('NON_PHY', '非物理伤害', 'elemental', NULL, NULL, 75, '#9370DB'),
    ('SLASH', '挥砍伤害', 'physical', 'PHYSICAL_RESIST', NULL, 75, '#CD853F'),
    ('PIERCE', '穿刺伤害', 'physical', 'PHYSICAL_RESIST', NULL, 75, '#DAA520'),
    ('BLUDGE', '钝击伤害', 'physical', 'PHYSICAL_RESIST', NULL, 75, '#A0522D')
ON CONFLICT (code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入元效果类型定义
INSERT INTO effect_type_definitions (
    effect_type_code, 
    effect_type_name, 
    description, 
    parameter_list,
    parameter_descriptions,
    failure_handling,
    json_template
)
VALUES 
    ('HIT_CHECK', '命中检定', 
     '进行攻击命中判定，计算命中成功率并赋予相应FLAG',
     ARRAY['MATCH_TARGET', 'MATCH_FLAG', 'Key_Attribute_1', 'Key_Attribute_2', 'AccSkill_Type'],
     'MATCH_TARGET:匹配FLAG的目标; MATCH_FLAG:需要匹配的FLAG; Key_Attribute_1:关键属性1; Key_Attribute_2:关键属性2; AccSkill_Type:精准技能类型',
     'skip_remaining',
     '{"effect":"HIT_CHECK","params":{"MATCH_TARGET":"0","MATCH_FLAG":"0","Key_Attribute_1":"STR","Key_Attribute_2":"AGI","AccSkill_Type":"SWORD_PROFICIENCY"},"order":0}'::jsonb),
    
    ('CRIT_CHECK', '暴击检定',
     '进行暴击检定，成功时赋予Crit_Confirmed标记并设置暴击倍率',
     ARRAY['MATCH_TARGET', 'MATCH_FLAG', 'Act_Crit_%', 'Crit_Multiplier'],
     'Act_Crit_%:基础暴击率; Crit_Multiplier:暴击倍率修正',
     'continue',
     '{"effect":"CRIT_CHECK","params":{"MATCH_TARGET":"Target","MATCH_FLAG":"Hit_SUCCESS","Act_Crit_%":"2","Crit_Multiplier":"0"},"order":1}'::jsonb),
    
    ('DMG_CALCULATION', '伤害计算',
     '计算最终伤害值，考虑抗性、暴击等因素',
     ARRAY['MATCH_TARGET', 'MATCH_FLAG', 'Base_DMG', 'Base_DMG_type', 'DMG_Attribute_1', 'DMG_Attribute_2', 'Action_Mult', 'Current_Crit_Mult'],
     'Base_DMG:基础伤害; Base_DMG_type:伤害类型; DMG_Attribute_1:伤害属性1; DMG_Attribute_2:伤害属性2; Action_Mult:动作倍率; Current_Crit_Mult:当前暴击倍率',
     'continue',
     '{"effect":"DMG_CALCULATION","params":{"MATCH_TARGET":"Target","MATCH_FLAG":"Hit_SUCCESS","Base_DMG":"Weapon_DMG","Base_DMG_type":"PHYSICAL","DMG_Attribute_1":"STR","DMG_Attribute_2":"AGI","Action_Mult":"1","Current_Crit_Mult":"get_Crit_Multiplier"},"order":2}'::jsonb),
    
    ('APPLY_BUFF', '施加影响',
     '将指定BUFF施加到目标身上',
     ARRAY['MATCH_TARGET', 'MATCH_FLAG', 'Target', 'Buff_ID', 'Duration', 'Caster_Level'],
     'Target:施加目标; Buff_ID:BUFF的ID; Duration:持续时间; Caster_Level:施放者等级',
     'continue',
     '{"effect":"APPLY_BUFF","params":{"MATCH_TARGET":"self","MATCH_FLAG":"Deal_DMG_SUCCESS","Target":"Target","Buff_ID":"Weapon_Buff","Duration":"3","Caster_Level":"Self_Level"},"order":3}'::jsonb),
    
    ('SAVE_CHECK', '豁免检定',
     '进行豁免检定，失败时赋予SAVE_CHECK_FAIL标记',
     ARRAY['MATCH_TARGET', 'MATCH_FLAG', 'save_type'],
     'save_type:豁免类型(body/mind/magic/environment)',
     'continue',
     '{"effect":"SAVE_CHECK","params":{"MATCH_TARGET":"0","MATCH_FLAG":"0","save_type":"magic"},"order":0}'::jsonb),
    
    ('MOVEMENT_TO_DIRECTION', '移动',
     '向指定方向移动指定距离',
     ARRAY['MATCH_TARGET', 'MATCH_FLAG', 'Movement_Direction', 'Movement_length', 'Block', 'Terrain', 'Break'],
     'Movement_Direction:移动方向; Movement_length:移动距离; Block:是否会被阻挡; Terrain:是否触发地形效果; Break:是否可被打断',
     'continue',
     '{"effect":"MOVEMENT_TO_DIRECTION","params":{"MATCH_TARGET":"0","MATCH_FLAG":"0","Movement_Direction":"Right","Movement_length":"4","Block":"1","Terrain":"1","Break":"1"},"order":1}'::jsonb),
    
    ('REACTION_TRIGGER', '反应触发',
     '监听目标并在满足条件时触发后续效果',
     ARRAY['Listen_Target', 'Listen_Time', 'Listen_Condition', 'Trigger_Cost_Type', 'Trigger_Cost_Number'],
     'Listen_Target:监听目标; Listen_Time:监听时间; Listen_Condition:触发条件; Trigger_Cost_Type:触发消耗类型; Trigger_Cost_Number:消耗数量',
     'skip_remaining',
     '{"effect":"REACTION_TRIGGER","params":{"Listen_Target":"Target","Listen_Time":"3","Listen_Condition":"Target_get_FLAG{STARTING_FLAW_ACTION}","Trigger_Cost_Type":"reaction","Trigger_Cost_Number":"1"},"order":0}'::jsonb)
ON CONFLICT (effect_type_code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入公式变量定义
INSERT INTO formula_variables (variable_code, variable_name, variable_type, scope, data_type, description, example)
VALUES 
    ('STR', '力量', 'attribute', 'character', 'integer', '角色的力量属性值', 'STR*2+10'),
    ('AGI', '敏捷', 'attribute', 'character', 'integer', '角色的敏捷属性值', 'AGI*1.5'),
    ('INT', '智力', 'attribute', 'character', 'integer', '角色的智力属性值', 'INT*2+skill_level*5'),
    ('VIT', '体质', 'attribute', 'character', 'integer', '角色的体质属性值', 'VIT+skill_level'),
    ('WIP', '意志', 'attribute', 'character', 'integer', '角色的意志属性值', 'WIP*2'),
    ('Target', '动作目标', 'target', 'action', 'object', '继承动作选择的目标', ''),
    ('Self', '自身', 'target', 'action', 'object', '动作使用者自身', ''),
    ('skill_level', '技能等级', 'skill_data', 'skill', 'integer', '当前技能的等级', '10+2*skill_level'),
    ('Weapon_DMG', '武器伤害', 'equipment_data', 'equipment', 'integer', '当前使用武器的基础伤害', 'Weapon_DMG+STR'),
    ('Weapon_Buff', '武器效果', 'equipment_data', 'equipment', 'string', '当前使用武器的附加效果ID', '')
ON CONFLICT (variable_code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入射程配置规则
INSERT INTO range_config_rules (parameter_type, parameter_format, description, example, notes)
VALUES 
    ('range', 'N', '固定射程N格', 'range:6', '表示射程为6格'),
    ('range', '0~N', '射程0到N格', 'range:0~6', '表示射程0-6格可选'),
    ('positions', '最近N位', '选择最近的N个位置', 'positions:最近1位', '选择最近的1个单位'),
    ('positions', '最远N位', '选择最远的N个位置', 'positions:最远2位', '选择最远的2个单位'),
    ('positions', '自身', '目标为自身', 'positions:自身', ''),
    ('depth', '单体', '只影响单个目标', 'depth:单体', ''),
    ('depth', '全体', '影响范围内所有单位', 'depth:全体', '')
ON CONFLICT DO NOTHING;

-- 插入动作类型定义
INSERT INTO action_type_definitions (action_type, description, per_turn_limit, usage_timing, example)
VALUES 
    ('main', '主要动作，每回合1次', 1, '回合中', '普通攻击、施放法术'),
    ('minor', '次要动作，每回合1次', 1, '回合中', '移动、使用物品'),
    ('reaction', '反应动作，每回合多次', NULL, '其他单位行动时', '借机攻击、反击')
ON CONFLICT (action_type) DO NOTHING;

-- 插入Action Flag
INSERT INTO action_flags (
    flag_code, 
    flag_name, 
    category, 
    duration_type,
    auto_remove_condition,
    is_visible,
    description
)
VALUES 
    ('CRIT_CONFIRMED', '确认暴击', 'action_chain', 'action', 'act_end', FALSE, '记录本动作是否成功暴击'),
    ('HIT_SUCCESS', '命中检定成功', 'check_status', 'action', 'next_act_start', FALSE, '标记本动作成功命中'),
    ('HIT_FAIL', '命中检定失败', 'check_status', 'action', 'next_act_start', FALSE, '标记本动作未命中'),
    ('DEAL_DMG_SUCCESS', '成功造成伤害', 'check_status', 'action', 'next_act_start', FALSE, '标记本动作成功造成伤害'),
    ('DEAL_DMG_FAIL', '未成功造成伤害', 'check_status', 'action', 'next_act_start', FALSE, '标记本动作未成功造成伤害'),
    ('DEAL_BUFF_SUCCESS', '成功施加影响', 'check_status', 'action', 'next_act_start', FALSE, '标记本动作成功施加影响'),
    ('DEAL_BUFF_FAIL', '未成功施加影响', 'check_status', 'action', 'next_act_start', FALSE, '标记本动作未成功施加影响'),
    ('STARTING_ATTACK_ACTION', '即将使用攻击动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用一个带有攻击特征的动作'),
    ('STARTING_MAIN_ACTION', '即将使用主要动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用一个主要动作'),
    ('STARTING_MINOR_ACTION', '即将使用次要动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用一个次要动作'),
    ('STARTING_FLAW_ACTION', '即将使用破绽动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用一个会露出破绽的动作'),
    ('STARTING_MAGIC_ACTION', '即将使用魔法动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用一个魔法动作'),
    ('STARTING_MOVEMENT_ACTION', '即将使用移动动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用一个移动动作'),
    ('STARTING_USAGE_ACTION', '即将使用物品动作', 'check_status', 'action', 'this_act_end', FALSE, '说明即将使用物品'),
    ('SAVE_CHECK_SUCCESS', '豁免检定成功', 'check_status', 'action', 'next_act_start', FALSE, '标记豁免检定成功'),
    ('SAVE_CHECK_FAIL', '豁免检定失败', 'check_status', 'action', 'next_act_start', FALSE, '标记豁免检定失败')
ON CONFLICT (flag_code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入Buff
INSERT INTO buffs (
    buff_code,
    buff_name,
    buff_type,
    category,
    feature_tags,
    default_duration,
    effect_description,
    stack_rule,
    parameter_list,
    parameter_definitions,
    description
)
VALUES 
    ('UNIT_TARGET', '召唤物目标', 'neutral', 'magic', 
     ARRAY['summon'], 3,
     '亡灵召唤物优先攻击带有此特征的单位',
     'no_stack', NULL, NULL,
     '引导亡灵优先攻击此单位'),
    
    ('ICED_LAND', '冻结地面', 'debuff', 'environment',
     ARRAY['ground', 'negative', 'ice', 'magic'], 3,
     '触发地面效果的单位立刻停止移动；如果其等级比施法者低，还会倒地',
     'no_stack', NULL, NULL,
     '地面效果，阻碍移动'),
    
    ('BLEED', '流血', 'debuff', 'body',
     ARRAY['negative', 'bleed', 'continuous_damage'], 3,
     'HP每回合减少N%*角色等级，无伤害类型；受到任何恢复生命值效果时立刻结束',
     'stackable', ARRAY['N'],
     '{"N":{"type":"percentage","description":"伤害百分比","min":0,"max":100}}'::jsonb,
     '使角色持续受到流血伤害'),
    
    ('BURN', '灼烧', 'debuff', 'body',
     ARRAY['negative', 'fire', 'continuous_damage'], 3,
     'HP每回合减少N%*角色等级，火焰伤害类型。任何寒冰伤害都会立刻结束灼烧',
     'stackable', ARRAY['N'],
     '{"N":{"type":"percentage","description":"伤害百分比","min":0,"max":100}}'::jsonb,
     '使角色持续受到火焰伤害'),
    
    ('FROZEN', '冰冻', 'debuff', 'body',
     ARRAY['negative', 'ice', 'continuous_damage'], 3,
     'HP每回合减少N%*角色等级，寒冰伤害类型。任何火焰伤害都会立刻结束冰冻',
     'stackable', ARRAY['N'],
     '{"N":{"type":"percentage","description":"伤害百分比","min":0,"max":100}}'::jsonb,
     '使角色持续受到寒冰伤害，并减弱身体抵抗'),
    
    ('SHOCK', '震慑', 'debuff', 'body',
     ARRAY['negative', 'disability'], 1,
     '跳过角色的回合开始恢复动作阶段',
     'no_stack', NULL, NULL,
     '使角色1轮无法行动'),
    
    ('MOVE_FORBIDDEN', '禁足', 'debuff', 'body',
     ARRAY['negative', 'disability', 'movement'], 1,
     '禁止使用任何带有移动特征的动作或效果。被动移动转化为与距离等额的钝击伤害',
     'no_stack', NULL, NULL,
     '将角色禁锢在原地'),
    
    ('SLOWED', '缓慢', 'debuff', 'body',
     ARRAY['negative', 'disability'], 3,
     '任何移动动作的距离减半。抵消加速状态',
     'no_stack', NULL, NULL,
     '减慢角色的移动'),
    
    ('QUICKEN', '加速', 'buff', 'body',
     ARRAY['positive', 'movement'], 3,
     '任何移动动作的移动距离翻倍。抵消缓慢状态',
     'no_stack', NULL, NULL,
     '加快角色的移动')
ON CONFLICT (buff_code) WHERE deleted_at IS NULL DO NOTHING;

-- 插入属性类型（如果不存在）
INSERT INTO hero_attribute_type (attribute_code, attribute_name, category, data_type, default_value, min_value, max_value, unit, description)
VALUES 
    ('BASE_INITIATIVE', '基础先攻', 'basic', 'integer', 10, 0, 999, '点', '影响行动顺序'),
    ('ACCURACY', '精准', 'derived', 'integer', 0, 0, 9999, '点', '命中率计算基础值'),
    ('DODGE', '闪避', 'derived', 'integer', 0, 0, 9999, '点', '闪避率计算基础值'),
    ('CRITICAL_RATE', '暴击率', 'basic', 'percentage', 5, 0, 100, '%', '暴击概率'),
    ('CRITICAL_DAMAGE', '暴击伤害', 'basic', 'percentage', 150, 100, 500, '%', '暴击时的伤害倍率'),
    ('PHYSICAL_RESIST', '物理抗性', 'resistance', 'percentage', 0, -9999, 90, '%', '对物理伤害的抗性/脆弱性'),
    ('FIRE_RESIST', '火焰抗性', 'resistance', 'percentage', 0, -9999, 90, '%', '对火焰伤害的抗性/脆弱性'),
    ('ICE_RESIST', '寒冰抗性', 'resistance', 'percentage', 0, -9999, 90, '%', '对寒冰伤害的抗性/脆弱性')
ON CONFLICT (attribute_code) WHERE deleted_at IS NULL DO NOTHING;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE '技能基础系统创建完成（完整版）';
    RAISE NOTICE '============================================';
    RAISE NOTICE '核心特性:';
    RAISE NOTICE '  ✓ 优劣势判定系统';
    RAISE NOTICE '  ✓ 伤害浮动 50%%-150%%（应用层实现）';
    RAISE NOTICE '  ✓ 抗性上限控制';
    RAISE NOTICE '  ✓ 元效果类型定义系统';
    RAISE NOTICE '  ✓ 公式变量管理';
    RAISE NOTICE '  ✓ Effect独立管理（可复用）';
    RAISE NOTICE '  ✓ Excel配置兼容';
    RAISE NOTICE '============================================';
    RAISE NOTICE '数据统计:';
    RAISE NOTICE '  • % 个技能类别', (SELECT COUNT(*) FROM skill_categories WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个动作类别', (SELECT COUNT(*) FROM action_categories WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个伤害类型', (SELECT COUNT(*) FROM damage_types WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个元效果类型', (SELECT COUNT(*) FROM effect_type_definitions WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个公式变量', (SELECT COUNT(*) FROM formula_variables WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个Action Flag', (SELECT COUNT(*) FROM action_flags WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个Buff', (SELECT COUNT(*) FROM buffs WHERE deleted_at IS NULL);
    RAISE NOTICE '  • % 个射程配置规则', (SELECT COUNT(*) FROM range_config_rules);
    RAISE NOTICE '============================================';
END $;