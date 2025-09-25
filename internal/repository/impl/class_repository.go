package impl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity"
	"tsu-self/internal/repository/interfaces"
	"tsu-self/internal/repository/query"
)

type classRepository struct {
	db *sql.DB
}

// NewClassRepository 创建职业仓储实例
func NewClassRepository(db *sql.DB) interfaces.ClassRepository {
	return &classRepository{db: db}
}

// Create 创建职业
func (r *classRepository) Create(ctx context.Context, class *entity.Class) error {
	return class.Insert(ctx, r.db, boil.Infer())
}

// GetByID 根据ID获取职业
func (r *classRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Class, error) {
	return entity.FindClass(ctx, r.db, id.String())
}

// GetByCode 根据代码获取职业
func (r *classRepository) GetByCode(ctx context.Context, code string) (*entity.Class, error) {
	return entity.Classes(
		qm.Where("class_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)
}

// Update 更新职业
func (r *classRepository) Update(ctx context.Context, class *entity.Class) error {
	_, err := class.Update(ctx, r.db, boil.Infer())
	return err
}

// Delete 删除职业（软删除）
func (r *classRepository) Delete(ctx context.Context, id uuid.UUID) error {
	class, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	class.DeletedAt = null.TimeFrom(class.UpdatedAt.Time)
	_, err = class.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at"))
	return err
}

// List 获取职业列表
func (r *classRepository) List(ctx context.Context, params *query.ClassListParams) ([]*entity.Class, int64, error) {
	// 构建查询条件
	whereClauses := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	if params.Tier != nil {
		whereClauses = append(whereClauses, qm.Where("tier = ?", fmt.Sprintf("%d", *params.Tier)))
	}

	if params.IsActive != nil {
		whereClauses = append(whereClauses, qm.Where("is_active = ?", *params.IsActive))
	}

	if params.IsHidden != nil {
		// is_hidden 对应数据库的 is_visible 字段（反向）
		whereClauses = append(whereClauses, qm.Where("is_visible = ?", !*params.IsHidden))
	}

	if params.Search != nil && *params.Search != "" {
		searchTerm := "%" + *params.Search + "%"
		whereClauses = append(whereClauses, qm.Where("(class_name ILIKE ? OR class_code ILIKE ? OR description ILIKE ?)", searchTerm, searchTerm, searchTerm))
	}

	// 计算总数
	totalCount, err := entity.Classes(whereClauses...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	// 添加排序
	sortBy := "display_order"
	sortOrder := "ASC"
	if params.SortBy != "" {
		switch params.SortBy {
		case "name":
			sortBy = "class_name"
		case "code":
			sortBy = "class_code"
		case "tier":
			sortBy = "tier"
		case "display_order":
			sortBy = "display_order"
		case "created_at":
			sortBy = "created_at"
		}
	}
	if params.SortOrder != "" && strings.ToUpper(params.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	orderBy := fmt.Sprintf("%s %s", sortBy, sortOrder)
	whereClauses = append(whereClauses, qm.OrderBy(orderBy))

	// 添加分页
	offset := (params.Page - 1) * params.PageSize
	whereClauses = append(whereClauses, qm.Limit(params.PageSize), qm.Offset(offset))

	// 执行查询
	classes, err := entity.Classes(whereClauses...).All(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	return classes, totalCount, nil
}

// GetHeroStats 获取职业英雄统计
func (r *classRepository) GetHeroStats(ctx context.Context, classID uuid.UUID) (*query.ClassHeroStats, error) {
	var stats query.ClassHeroStats

	queryStr := `
		SELECT
			class_id,
			total_heroes,
			active_heroes,
			average_level,
			max_level
		FROM class_hero_stats
		WHERE class_id = $1
	`

	err := queries.Raw(queryStr, classID.String()).Bind(ctx, r.db, &stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// CreateAttributeBonus 创建属性加成
func (r *classRepository) CreateAttributeBonus(ctx context.Context, bonus *entity.ClassAttributeBonuse) error {
	return bonus.Insert(ctx, r.db, boil.Infer())
}

// GetAttributeBonuses 获取职业属性加成列表
func (r *classRepository) GetAttributeBonuses(ctx context.Context, classID uuid.UUID) ([]*query.ClassAttributeBonusWithDetails, error) {
	queryStr := `
		SELECT
			cab.id,
			cab.class_id,
			cab.attribute_id,
			hat.attribute_code,
			hat.attribute_name,
			cab.base_bonus_value as base_bonus,
			cab.per_level_bonus_value as per_level_bonus,
			cab.created_at,
			cab.updated_at
		FROM class_attribute_bonuses cab
		INNER JOIN hero_attribute_type hat ON hat.id = cab.attribute_id AND hat.deleted_at IS NULL
		WHERE cab.class_id = $1
		ORDER BY hat.display_order, hat.attribute_name
	`

	var bonuses []*query.ClassAttributeBonusWithDetails
	err := queries.Raw(queryStr, classID.String()).Bind(ctx, r.db, &bonuses)
	return bonuses, err
}

// GetAttributeBonus 获取单个属性加成
func (r *classRepository) GetAttributeBonus(ctx context.Context, classID, attributeID uuid.UUID) (*entity.ClassAttributeBonuse, error) {
	return entity.ClassAttributeBonuses(
		qm.Where("class_id = ? AND attribute_id = ?", classID.String(), attributeID.String()),
	).One(ctx, r.db)
}

// UpdateAttributeBonus 更新属性加成
func (r *classRepository) UpdateAttributeBonus(ctx context.Context, bonus *entity.ClassAttributeBonuse) error {
	_, err := bonus.Update(ctx, r.db, boil.Infer())
	return err
}

// DeleteAttributeBonus 删除属性加成
func (r *classRepository) DeleteAttributeBonus(ctx context.Context, id uuid.UUID) error {
	_, err := entity.ClassAttributeBonuses(qm.Where("id = ?", id.String())).DeleteAll(ctx, r.db)
	return err
}

// BatchCreateAttributeBonuses 批量创建属性加成
func (r *classRepository) BatchCreateAttributeBonuses(ctx context.Context, bonuses []*entity.ClassAttributeBonuse) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, bonus := range bonuses {
		if err := bonus.Insert(ctx, tx, boil.Infer()); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CreateAdvancementRequirement 创建进阶要求
func (r *classRepository) CreateAdvancementRequirement(ctx context.Context, req *entity.ClassAdvancedRequirement) error {
	return req.Insert(ctx, r.db, boil.Infer())
}

// GetAdvancementRequirement 获取进阶要求
func (r *classRepository) GetAdvancementRequirement(ctx context.Context, fromClassID, toClassID uuid.UUID) (*entity.ClassAdvancedRequirement, error) {
	return entity.ClassAdvancedRequirements(
		qm.Where("from_class_id = ? AND to_class_id = ? AND deleted_at IS NULL", fromClassID.String(), toClassID.String()),
	).One(ctx, r.db)
}

// GetAdvancementRequirements 获取职业相关的进阶要求
func (r *classRepository) GetAdvancementRequirements(ctx context.Context, classID uuid.UUID) ([]*query.ClassAdvancementWithDetails, error) {
	queryStr := `
		SELECT
			requirement_id as id,
			from_class_id,
			from_class_name,
			to_class_id,
			to_class_name,
			required_level,
			required_honor,
			required_job_change_count,
			required_attributes,
			required_skills,
			is_active,
			display_order,
			created_at,
			updated_at
		FROM class_advancement_paths
		WHERE from_class_id = $1 OR to_class_id = $1
		ORDER BY display_order, created_at
	`

	var requirements []*query.ClassAdvancementWithDetails
	err := queries.Raw(queryStr, classID.String()).Bind(ctx, r.db, &requirements)
	return requirements, err
}

// GetAdvancementPaths 获取进阶路径
func (r *classRepository) GetAdvancementPaths(ctx context.Context, fromClassID uuid.UUID) ([]*query.ClassAdvancementWithDetails, error) {
	queryStr := `
		SELECT
			requirement_id as id,
			from_class_id,
			from_class_name,
			to_class_id,
			to_class_name,
			required_level,
			required_honor,
			required_job_change_count,
			required_attributes,
			required_skills,
			is_active,
			display_order,
			created_at,
			updated_at
		FROM class_advancement_paths
		WHERE from_class_id = $1 AND is_active = true
		ORDER BY display_order, to_class_name
	`

	var paths []*query.ClassAdvancementWithDetails
	err := queries.Raw(queryStr, fromClassID.String()).Bind(ctx, r.db, &paths)
	return paths, err
}

// GetAdvancementSources 获取进阶来源
func (r *classRepository) GetAdvancementSources(ctx context.Context, toClassID uuid.UUID) ([]*query.ClassAdvancementWithDetails, error) {
	queryStr := `
		SELECT
			requirement_id as id,
			from_class_id,
			from_class_name,
			to_class_id,
			to_class_name,
			required_level,
			required_honor,
			required_job_change_count,
			required_attributes,
			required_skills,
			is_active,
			display_order,
			created_at,
			updated_at
		FROM class_advancement_paths
		WHERE to_class_id = $1 AND is_active = true
		ORDER BY display_order, from_class_name
	`

	var sources []*query.ClassAdvancementWithDetails
	err := queries.Raw(queryStr, toClassID.String()).Bind(ctx, r.db, &sources)
	return sources, err
}

// UpdateAdvancementRequirement 更新进阶要求
func (r *classRepository) UpdateAdvancementRequirement(ctx context.Context, req *entity.ClassAdvancedRequirement) error {
	_, err := req.Update(ctx, r.db, boil.Infer())
	return err
}

// DeleteAdvancementRequirement 删除进阶要求
func (r *classRepository) DeleteAdvancementRequirement(ctx context.Context, id uuid.UUID) error {
	req, err := entity.FindClassAdvancedRequirement(ctx, r.db, id.String())
	if err != nil {
		return err
	}
	req.DeletedAt = null.TimeFrom(req.UpdatedAt.Time)
	_, err = req.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at"))
	return err
}

// GetClassTags 获取职业标签
func (r *classRepository) GetClassTags(ctx context.Context, classID uuid.UUID) ([]*query.ClassTagWithDetails, error) {
	queryStr := `
		SELECT
			t.id,
			t.tag_code as code,
			t.tag_name as name,
			t.description,
			t.color,
			t.icon,
			t.display_order,
			ctv.relation_created_at as created_at
		FROM class_tags_view ctv
		INNER JOIN tags t ON t.id = ctv.tag_id
		WHERE ctv.class_id = $1
		ORDER BY t.display_order, t.tag_name
	`

	var tags []*query.ClassTagWithDetails
	err := queries.Raw(queryStr, classID.String()).Bind(ctx, r.db, &tags)
	return tags, err
}

// AddClassTag 添加职业标签
func (r *classRepository) AddClassTag(ctx context.Context, classID, tagID uuid.UUID) error {
	relation := &entity.ClassTagRelation{
		ClassID: classID.String(),
		TagID:   tagID.String(),
	}
	return relation.Insert(ctx, r.db, boil.Infer())
}

// RemoveClassTag 移除职业标签
func (r *classRepository) RemoveClassTag(ctx context.Context, classID, tagID uuid.UUID) error {
	_, err := entity.ClassTagRelations(
		qm.Where("class_id = ? AND tag_id = ?", classID.String(), tagID.String()),
	).DeleteAll(ctx, r.db)
	return err
}

// GetAllTags 获取所有标签
func (r *classRepository) GetAllTags(ctx context.Context, tagType *string) ([]*entity.Tag, error) {
	whereClauses := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
		qm.Where("is_active = true"),
	}

	if tagType != nil && *tagType != "" {
		whereClauses = append(whereClauses, qm.Where("tag_type = ?", *tagType))
	}

	whereClauses = append(whereClauses, qm.OrderBy("display_order, tag_name"))

	return entity.Tags(whereClauses...).All(ctx, r.db)
}