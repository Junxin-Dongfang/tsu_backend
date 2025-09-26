package impl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	apiReqAdmin "tsu-self/internal/api/model/request/admin"
	"tsu-self/internal/entity"
	"tsu-self/internal/repository/interfaces"
)

type attributeTypeRepositoryImpl struct {
	db *sql.DB
}

// NewAttributeTypeRepository 创建属性类型仓储实现
func NewAttributeTypeRepository(db *sql.DB) interfaces.AttributeTypeRepository {
	return &attributeTypeRepositoryImpl{db: db}
}

// Create 创建属性类型
func (r *attributeTypeRepositoryImpl) Create(ctx context.Context, attributeType *entity.HeroAttributeType) error {
	return attributeType.Insert(ctx, r.db, boil.Infer())
}

// GetByID 根据ID获取属性类型
func (r *attributeTypeRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.HeroAttributeType, error) {
	return entity.FindHeroAttributeType(ctx, r.db, id)
}

// GetByCode 根据属性代码获取属性类型
func (r *attributeTypeRepositoryImpl) GetByCode(ctx context.Context, code string) (*entity.HeroAttributeType, error) {
	return entity.HeroAttributeTypes(
		entity.HeroAttributeTypeWhere.AttributeCode.EQ(code),
		entity.HeroAttributeTypeWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
}

// List 获取属性类型列表
func (r *attributeTypeRepositoryImpl) List(ctx context.Context, req *apiReqAdmin.GetAttributeTypesRequest) ([]*entity.HeroAttributeType, int64, error) {
	// 构建查询条件
	mods := []qm.QueryMod{
		entity.HeroAttributeTypeWhere.DeletedAt.IsNull(),
	}

	// 分类过滤
	if req.Category != "" {
		mods = append(mods, entity.HeroAttributeTypeWhere.Category.EQ(req.Category))
	}

	// 启用状态过滤
	if req.IsActive != nil {
		mods = append(mods, entity.HeroAttributeTypeWhere.IsActive.EQ(null.BoolFrom(*req.IsActive)))
	}

	// 可见状态过滤
	if req.IsVisible != nil {
		mods = append(mods, entity.HeroAttributeTypeWhere.IsVisible.EQ(null.BoolFrom(*req.IsVisible)))
	}

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		mods = append(mods, qm.Where(
			"(attribute_code ILIKE ? OR attribute_name ILIKE ?)",
			keyword, keyword,
		))
	}

	// 获取总数
	totalMods := append([]qm.QueryMod{}, mods...)
	total, err := entity.HeroAttributeTypes(totalMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count attribute types: %w", err)
	}

	// 排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "display_order"
	}
	sortOrder := strings.ToUpper(req.SortOrder)
	if sortOrder != "DESC" {
		sortOrder = "ASC"
	}
	mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s", sortBy, sortOrder)))

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	mods = append(mods, qm.Limit(pageSize), qm.Offset(offset))

	// 查询数据
	attributeTypes, err := entity.HeroAttributeTypes(mods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query attribute types: %w", err)
	}

	return attributeTypes, total, nil
}

// Update 更新属性类型
func (r *attributeTypeRepositoryImpl) Update(ctx context.Context, attributeType *entity.HeroAttributeType) error {
	_, err := attributeType.Update(ctx, r.db, boil.Infer())
	return err
}

// Delete 软删除属性类型
func (r *attributeTypeRepositoryImpl) Delete(ctx context.Context, id string) error {
	attributeType := &entity.HeroAttributeType{ID: id}
	_, err := attributeType.Delete(ctx, r.db, false)
	return err
}

// ExistsByCode 检查属性代码是否存在
func (r *attributeTypeRepositoryImpl) ExistsByCode(ctx context.Context, code string, excludeID ...string) (bool, error) {
	mods := []qm.QueryMod{
		entity.HeroAttributeTypeWhere.AttributeCode.EQ(code),
		entity.HeroAttributeTypeWhere.DeletedAt.IsNull(),
	}

	// 排除特定ID（用于更新时的检查）
	if len(excludeID) > 0 && excludeID[0] != "" {
		mods = append(mods, entity.HeroAttributeTypeWhere.ID.NEQ(excludeID[0]))
	}

	count, err := entity.HeroAttributeTypes(mods...).Count(ctx, r.db)
	if err != nil {
		return false, fmt.Errorf("failed to check attribute code existence: %w", err)
	}

	return count > 0, nil
}

// GetActiveList 获取启用的属性类型列表（用于选项）
func (r *attributeTypeRepositoryImpl) GetActiveList(ctx context.Context, category string) ([]*entity.HeroAttributeType, error) {
	mods := []qm.QueryMod{
		entity.HeroAttributeTypeWhere.DeletedAt.IsNull(),
		entity.HeroAttributeTypeWhere.IsActive.EQ(null.BoolFrom(true)),
		entity.HeroAttributeTypeWhere.IsVisible.EQ(null.BoolFrom(true)),
		qm.OrderBy("display_order ASC, attribute_name ASC"),
	}

	// 分类过滤
	if category != "" {
		mods = append(mods, entity.HeroAttributeTypeWhere.Category.EQ(category))
	}

	return entity.HeroAttributeTypes(mods...).All(ctx, r.db)
}