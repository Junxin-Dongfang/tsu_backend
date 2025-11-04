-- 000020_add_missing_attribute_types.up.sql
-- 添加怪物系统需要的缺失属性类型

-- 添加先攻属性类型
INSERT INTO game_config.hero_attribute_type (
    attribute_code,
    attribute_name,
    category,
    data_type,
    calculation_formula,
    description,
    display_order,
    is_active,
    is_visible
) VALUES (
    'INITIATIVE',
    '先攻',
    'derived',
    'integer',
    'AGI*2+WIS',
    '决定战斗中的行动顺序，敏捷和感知的综合体现',
    20,
    TRUE,
    TRUE
) ON CONFLICT (attribute_code) DO NOTHING;

-- 添加体质豁免属性类型
INSERT INTO game_config.hero_attribute_type (
    attribute_code,
    attribute_name,
    category,
    data_type,
    calculation_formula,
    description,
    display_order,
    is_active,
    is_visible
) VALUES (
    'BODY_RESIST',
    '体质豁免',
    'resistance',
    'integer',
    'VIT*2+WLP',
    '抵抗毒素、疾病等身体伤害的能力',
    30,
    TRUE,
    TRUE
) ON CONFLICT (attribute_code) DO NOTHING;

-- 添加魔法豁免属性类型
INSERT INTO game_config.hero_attribute_type (
    attribute_code,
    attribute_name,
    category,
    data_type,
    calculation_formula,
    description,
    display_order,
    is_active,
    is_visible
) VALUES (
    'MAGIC_RESIST',
    '魔法豁免',
    'resistance',
    'integer',
    'WLP*2+WIS',
    '抵抗魔法效果的能力',
    31,
    TRUE,
    TRUE
) ON CONFLICT (attribute_code) DO NOTHING;

-- 添加精神豁免属性类型
INSERT INTO game_config.hero_attribute_type (
    attribute_code,
    attribute_name,
    category,
    data_type,
    calculation_formula,
    description,
    display_order,
    is_active,
    is_visible
) VALUES (
    'MENTAL_RESIST',
    '精神豁免',
    'resistance',
    'integer',
    'WIS*2+WLP',
    '抵抗精神控制、幻觉等效果的能力',
    32,
    TRUE,
    TRUE
) ON CONFLICT (attribute_code) DO NOTHING;

-- 添加环境豁免属性类型
INSERT INTO game_config.hero_attribute_type (
    attribute_code,
    attribute_name,
    category,
    data_type,
    calculation_formula,
    description,
    display_order,
    is_active,
    is_visible
) VALUES (
    'ENVIRONMENT_RESIST',
    '环境豁免',
    'resistance',
    'integer',
    'VIT+WLP',
    '抵抗极端环境（高温、严寒、高压等）的能力',
    33,
    TRUE,
    TRUE
) ON CONFLICT (attribute_code) DO NOTHING;

