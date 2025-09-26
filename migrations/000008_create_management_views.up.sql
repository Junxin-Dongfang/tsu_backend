-- =============================================================================
-- Create Management Views
-- 管理视图：为管理 API 提供便捷的查询视图
-- 依赖：000002_create_users_system, 000003_create_attribute_system, 000004_create_classes_system, 000005_create_heroes_system, 000006_create_skills_base, 000007_create_skills_advanced
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 用户统计视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW user_statistics AS
SELECT
    u.id as user_id,
    u.username,
    u.email,
    u.created_at as user_created_at,
    u.last_login_at,
    u.login_count,

    -- 英雄统计
    COALESCE(h_stats.hero_count, 0) as hero_count,
    COALESCE(h_stats.active_hero_count, 0) as active_hero_count,
    COALESCE(h_stats.max_hero_level, 0) as max_hero_level,
    COALESCE(h_stats.total_battles, 0) as total_battles,

    -- 财务信息
    COALESCE(uf.current_diamonds, 0) as current_diamonds,
    COALESCE(uf.total_spent_amount, 0) as total_spent_amount,
    is_user_premium(u.id) as is_premium_user
FROM users u
LEFT JOIN user_finances uf ON u.id = uf.user_id AND uf.deleted_at IS NULL
LEFT JOIN (
    SELECT
        user_id,
        COUNT(*) as hero_count,
        COUNT(*) FILTER (WHERE status = 'active') as active_hero_count,
        MAX(level) as max_hero_level,
        SUM(total_battles) as total_battles
    FROM heroes
    WHERE deleted_at IS NULL
    GROUP BY user_id
) h_stats ON u.id = h_stats.user_id
WHERE u.deleted_at IS NULL;

-- --------------------------------------------------------------------------------
-- 职业统计视图 (更新版，基于实际的英雄表)
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW class_hero_stats AS
SELECT
    c.id as class_id,
    c.class_code,
    c.class_name,

    -- 英雄统计
    COALESCE(h_stats.total_heroes, 0) as total_heroes,
    COALESCE(h_stats.active_heroes, 0) as active_heroes,
    COALESCE(h_stats.average_level, 0.0) as average_level,
    COALESCE(h_stats.max_level, 0) as max_level,
    COALESCE(h_stats.total_battles, 0) as total_battles,
    COALESCE(h_stats.total_victories, 0) as total_victories,

    -- 职业配置统计
    COALESCE(ab_stats.attribute_bonus_count, 0) as attribute_bonus_count,
    COALESCE(t_stats.tag_count, 0) as tag_count
FROM classes c
LEFT JOIN (
    SELECT
        class_id,
        COUNT(*) as total_heroes,
        COUNT(*) FILTER (WHERE status = 'active') as active_heroes,
        AVG(level) as average_level,
        MAX(level) as max_level,
        SUM(total_battles) as total_battles,
        SUM(victories) as total_victories
    FROM heroes
    WHERE deleted_at IS NULL
    GROUP BY class_id
) h_stats ON c.id = h_stats.class_id
LEFT JOIN (
    SELECT
        class_id,
        COUNT(*) as attribute_bonus_count
    FROM class_attribute_bonuses
    GROUP BY class_id
) ab_stats ON c.id = ab_stats.class_id
LEFT JOIN (
    SELECT
        ctr.class_id,
        COUNT(*) as tag_count
    FROM class_tag_relations ctr
    INNER JOIN tags t ON t.id = ctr.tag_id AND t.deleted_at IS NULL AND t.is_active = TRUE
    GROUP BY ctr.class_id
) t_stats ON c.id = t_stats.class_id
WHERE c.deleted_at IS NULL;

-- --------------------------------------------------------------------------------
-- 职业详细信息视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW class_details AS
SELECT
    c.*,
    -- 属性加成统计
    COALESCE(ab.bonus_count, 0) as attribute_bonus_count,
    -- 标签统计
    COALESCE(ct.tag_count, 0) as tag_count,
    -- 进阶要求统计
    COALESCE(ar_from.advancement_paths, 0) as available_advancement_paths,
    COALESCE(ar_to.advancement_sources, 0) as advancement_source_count,
    -- 英雄统计
    COALESCE(hs.total_heroes, 0) as total_heroes,
    COALESCE(hs.active_heroes, 0) as active_heroes
FROM classes c
-- 属性加成统计子查询
LEFT JOIN (
    SELECT
        class_id,
        COUNT(*) as bonus_count
    FROM class_attribute_bonuses
    GROUP BY class_id
) ab ON ab.class_id = c.id
-- 标签统计子查询
LEFT JOIN (
    SELECT
        ctr.class_id,
        COUNT(*) as tag_count
    FROM class_tag_relations ctr
    INNER JOIN tags t ON t.id = ctr.tag_id AND t.deleted_at IS NULL AND t.is_active = TRUE
    GROUP BY ctr.class_id
) ct ON ct.class_id = c.id
-- 进阶路径统计（作为源职业）
LEFT JOIN (
    SELECT
        from_class_id,
        COUNT(*) as advancement_paths
    FROM class_advanced_requirements
    WHERE deleted_at IS NULL AND is_active = true
    GROUP BY from_class_id
) ar_from ON ar_from.from_class_id = c.id
-- 进阶来源统计（作为目标职业）
LEFT JOIN (
    SELECT
        to_class_id,
        COUNT(*) as advancement_sources
    FROM class_advanced_requirements
    WHERE deleted_at IS NULL AND is_active = true
    GROUP BY to_class_id
) ar_to ON ar_to.to_class_id = c.id
-- 英雄统计
LEFT JOIN (
    SELECT
        class_id,
        COUNT(*) as total_heroes,
        COUNT(*) FILTER (WHERE status = 'active') as active_heroes
    FROM heroes
    WHERE deleted_at IS NULL
    GROUP BY class_id
) hs ON c.id = hs.class_id
WHERE c.deleted_at IS NULL;

-- --------------------------------------------------------------------------------
-- 职业标签详细视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW class_tags_view AS
SELECT
    ctr.class_id,
    t.id as tag_id,
    t.tag_code,
    t.tag_name,
    t.tag_type,
    t.description as tag_description,
    t.color as tag_color,
    t.icon as tag_icon,
    t.display_order as tag_display_order,
    t.is_active as tag_is_active,
    ctr.created_at as relation_created_at
FROM class_tag_relations ctr
INNER JOIN tags t ON t.id = ctr.tag_id
WHERE t.deleted_at IS NULL
ORDER BY t.display_order ASC, t.tag_name ASC;

-- --------------------------------------------------------------------------------
-- 进阶路径详细视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW class_advancement_paths_view AS
SELECT
    car.id as requirement_id,
    car.from_class_id,
    fc.class_name as from_class_name,
    fc.class_code as from_class_code,
    fc.tier as from_class_tier,
    car.to_class_id,
    tc.class_name as to_class_name,
    tc.class_code as to_class_code,
    tc.tier as to_class_tier,
    car.required_level,
    car.required_honor,
    car.required_job_change_count,
    car.required_attributes,
    car.required_skills,
    car.required_items,
    car.is_active,
    car.display_order,
    car.created_at,
    car.updated_at
FROM class_advanced_requirements car
INNER JOIN classes fc ON fc.id = car.from_class_id AND fc.deleted_at IS NULL
INNER JOIN classes tc ON tc.id = car.to_class_id AND tc.deleted_at IS NULL
WHERE car.deleted_at IS NULL
ORDER BY car.display_order ASC, car.required_level ASC;

-- --------------------------------------------------------------------------------
-- 英雄详细视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW hero_details AS
SELECT
    h.id as hero_id,
    h.user_id,
    u.username,
    h.hero_name,
    h.level,
    h.experience,
    h.status,
    h.health_points,
    h.max_health_points,
    h.mana_points,
    h.max_mana_points,

    -- 职业信息
    c.id as class_id,
    c.class_name,
    c.class_code,
    c.tier as class_tier,
    c.icon as class_icon,
    c.color as class_color,

    -- 战斗统计
    h.total_battles,
    h.victories,
    h.defeats,
    CASE
        WHEN h.total_battles > 0 THEN ROUND((h.victories::DECIMAL / h.total_battles * 100), 2)
        ELSE 0
    END as win_rate,
    h.total_damage_dealt,
    h.total_damage_taken,

    -- 战力计算（可能较慢，仅在需要时使用）
    calculate_hero_power(h.id) as total_power,

    -- 时间信息
    h.created_at,
    h.last_battle_at
FROM heroes h
INNER JOIN users u ON h.user_id = u.id AND u.deleted_at IS NULL
INNER JOIN classes c ON h.class_id = c.id AND c.deleted_at IS NULL
WHERE h.deleted_at IS NULL;

-- --------------------------------------------------------------------------------
-- 英雄属性详细视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW hero_attribute_details AS
SELECT
    ha.hero_id,
    h.hero_name,
    hat.id as attribute_type_id,
    hat.attribute_code,
    hat.attribute_name,
    hat.category as attribute_category,
    hat.data_type as attribute_data_type,
    hat.unit as attribute_unit,

    -- 属性值详情
    ha.base_value,
    ha.bonus_value,
    ha.equipment_bonus,
    ha.temporary_bonus,
    ha.final_value,

    -- 显示配置
    hat.icon as attribute_icon,
    hat.color as attribute_color,
    hat.display_order,
    hat.is_visible

FROM hero_attributes ha
INNER JOIN heroes h ON ha.hero_id = h.id AND h.deleted_at IS NULL
INNER JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id AND hat.deleted_at IS NULL
WHERE hat.is_active = TRUE
ORDER BY hat.display_order ASC, hat.attribute_name ASC;

-- --------------------------------------------------------------------------------
-- 英雄技能详细视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW hero_skill_details AS
SELECT
    hs.hero_id,
    h.hero_name,
    s.id as skill_id,
    s.skill_code,
    s.skill_name,
    s.skill_type,
    sc.category_name as skill_category,
    hs.skill_level,
    hs.skill_experience,
    hs.max_level,
    hs.is_active as skill_is_active,
    hs.is_equipped,
    hs.learned_at,
    hs.learned_method,

    -- 技能配置
    s.max_level as skill_max_level,
    s.base_cooldown,
    s.base_mana_cost,
    s.description as skill_description,
    s.icon as skill_icon

FROM hero_skills hs
INNER JOIN heroes h ON hs.hero_id = h.id AND h.deleted_at IS NULL
INNER JOIN skills s ON hs.skill_code = s.skill_code AND s.is_active = TRUE
LEFT JOIN skill_categories sc ON s.category_id = sc.id AND sc.is_active = TRUE
ORDER BY s.skill_type ASC, hs.skill_level DESC, s.skill_name ASC;

-- --------------------------------------------------------------------------------
-- 属性类型管理视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW attribute_type_management AS
SELECT
    hat.id,
    hat.attribute_code,
    hat.attribute_name,
    hat.category,
    hat.data_type,
    hat.min_value,
    hat.max_value,
    hat.default_value,
    hat.unit,
    hat.icon,
    hat.color,
    hat.display_order,
    hat.is_active,
    hat.is_visible,
    hat.description,

    -- 使用统计
    COALESCE(hero_usage_stats.hero_count, 0) as heroes_using_count,
    COALESCE(class_usage_stats.class_bonus_count, 0) as class_bonus_count,

    -- 标签信息
    COALESCE(tag_stats.tag_count, 0) as tag_count,

    hat.created_at,
    hat.updated_at
FROM hero_attribute_type hat
LEFT JOIN (
    SELECT
        attribute_type_id,
        COUNT(DISTINCT hero_id) as hero_count
    FROM hero_attributes
    GROUP BY attribute_type_id
) hero_usage_stats ON hat.id = hero_usage_stats.attribute_type_id
LEFT JOIN (
    SELECT
        attribute_id,
        COUNT(*) as class_bonus_count
    FROM class_attribute_bonuses
    GROUP BY attribute_id
) class_usage_stats ON hat.id = class_usage_stats.attribute_id
LEFT JOIN (
    SELECT
        attribute_type_id,
        COUNT(*) as tag_count
    FROM hero_attribute_type_tags
    GROUP BY attribute_type_id
) tag_stats ON hat.id = tag_stats.attribute_type_id
WHERE hat.deleted_at IS NULL
ORDER BY hat.display_order ASC, hat.attribute_name ASC;

-- --------------------------------------------------------------------------------
-- 技能系统统计视图
-- --------------------------------------------------------------------------------

CREATE OR REPLACE VIEW skill_system_statistics AS
SELECT
    'skills' as category,
    'active' as status,
    COUNT(*) as count
FROM skills
WHERE is_active = TRUE

UNION ALL

SELECT
    'skill_categories' as category,
    'active' as status,
    COUNT(*) as count
FROM skill_categories
WHERE is_active = TRUE

UNION ALL

SELECT
    'actions' as category,
    'active' as status,
    COUNT(*) as count
FROM actions
WHERE is_active = TRUE

UNION ALL

SELECT
    'buffs' as category,
    'active' as status,
    COUNT(*) as count
FROM buffs
WHERE is_active = TRUE

UNION ALL

SELECT
    'damage_types' as category,
    'active' as status,
    COUNT(*) as count
FROM damage_types
WHERE is_active = TRUE;

-- --------------------------------------------------------------------------------
-- 创建视图索引（对于经常查询的视图字段）
-- --------------------------------------------------------------------------------

-- 为视图中经常用作筛选条件的字段创建索引
-- 注意：PostgreSQL 不支持直接在视图上创建索引，但可以在基础表上创建适当的索引来优化视图查询

-- 英雄相关索引优化
CREATE INDEX IF NOT EXISTS idx_heroes_user_class_level ON heroes(user_id, class_id, level DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_status_level ON heroes(status, level DESC) WHERE deleted_at IS NULL;

-- 英雄属性索引优化
CREATE INDEX IF NOT EXISTS idx_hero_attributes_final_value ON hero_attributes(final_value DESC);

-- 技能相关索引优化
CREATE INDEX IF NOT EXISTS idx_hero_skills_hero_equipped ON hero_skills(hero_id, is_equipped) WHERE is_equipped = TRUE;

-- 职业相关索引优化
CREATE INDEX IF NOT EXISTS idx_class_attribute_bonuses_class_attr ON class_attribute_bonuses(class_id, attribute_id);

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Management Views 创建完成';
    RAISE NOTICE '包含: 用户统计、职业管理、英雄详情、属性管理、技能统计等视图';
    RAISE NOTICE '创建了 % 个管理视图', 8;
    RAISE NOTICE '============================================';
END $$;