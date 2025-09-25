-- =============================================================================
-- Create Views for Class Management API
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 职业统计视图 - 用于获取职业英雄数量统计
-- 注意：heroes表尚未创建，暂时提供默认值
-- -----------------------------------------------------------------------------
CREATE OR REPLACE VIEW class_hero_stats AS
SELECT
    c.id as class_id,
    0 as total_heroes,
    0 as active_heroes,
    0.0 as average_level,
    0 as max_level
FROM classes c
WHERE c.deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- 职业详细信息视图 - 包含关联的属性加成和标签信息
-- -----------------------------------------------------------------------------
CREATE OR REPLACE VIEW class_details AS
SELECT
    c.*,
    -- 属性加成统计
    COALESCE(ab.bonus_count, 0) as attribute_bonus_count,
    -- 标签统计
    COALESCE(ct.tag_count, 0) as tag_count,
    -- 进阶要求统计
    COALESCE(ar_from.advancement_paths, 0) as available_advancement_paths,
    COALESCE(ar_to.advancement_sources, 0) as advancement_source_count
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
    INNER JOIN tags t ON t.id = ctr.tag_id AND t.deleted_at IS NULL
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
WHERE c.deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- 职业标签详细视图 - 用于获取职业的标签信息
-- -----------------------------------------------------------------------------
CREATE OR REPLACE VIEW class_tags_view AS
SELECT
    ctr.class_id,
    t.id as tag_id,
    t.tag_code,
    t.tag_name,
    t.description as tag_description,
    t.color as tag_color,
    t.icon as tag_icon,
    t.display_order as tag_display_order,
    ctr.created_at as relation_created_at
FROM class_tag_relations ctr
INNER JOIN tags t ON t.id = ctr.tag_id
WHERE t.deleted_at IS NULL
  AND t.is_active = true;

-- -----------------------------------------------------------------------------
-- 进阶路径详细视图 - 用于进阶系统查询
-- -----------------------------------------------------------------------------
CREATE OR REPLACE VIEW class_advancement_paths AS
SELECT
    car.id as requirement_id,
    car.from_class_id,
    fc.class_name as from_class_name,
    fc.class_code as from_class_code,
    car.to_class_id,
    tc.class_name as to_class_name,
    tc.class_code as to_class_code,
    car.required_level,
    car.required_honor,
    car.required_job_change_count,
    car.required_attributes,
    car.required_skills,
    car.is_active,
    car.display_order,
    car.created_at,
    car.updated_at
FROM class_advanced_requirements car
INNER JOIN classes fc ON fc.id = car.from_class_id AND fc.deleted_at IS NULL
INNER JOIN classes tc ON tc.id = car.to_class_id AND tc.deleted_at IS NULL
WHERE car.deleted_at IS NULL;