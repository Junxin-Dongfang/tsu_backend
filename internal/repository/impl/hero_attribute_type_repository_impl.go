package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type heroAttributeTypeRepositoryImpl struct {
	db *sql.DB
}

// NewHeroAttributeTypeRepository 创建属性类型仓储实例
func NewHeroAttributeTypeRepository(db *sql.DB) interfaces.HeroAttributeTypeRepository {
	return &heroAttributeTypeRepositoryImpl{db: db}
}

// GetByID 根据ID获取属性类型
func (r *heroAttributeTypeRepositoryImpl) GetByID(ctx context.Context, attributeTypeID string) (*game_config.HeroAttributeType, error) {
	attributeType, err := game_config.HeroAttributeTypes(
		qm.Where("id = ? AND deleted_at IS NULL", attributeTypeID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("属性类型不存在: %s", attributeTypeID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询属性类型失败: %w", err)
	}

	return attributeType, nil
}

// GetByCode 根据代码获取属性类型
func (r *heroAttributeTypeRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.HeroAttributeType, error) {
	attributeType, err := game_config.HeroAttributeTypes(
		qm.Where("attribute_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("属性类型不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询属性类型失败: %w", err)
	}

	return attributeType, nil
}

// List 获取属性类型列表
func (r *heroAttributeTypeRepositoryImpl) List(ctx context.Context, params interfaces.HeroAttributeTypeQueryParams) ([]*game_config.HeroAttributeType, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.Category != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("category = ?", *params.Category))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}
	if params.IsVisible != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_visible = ?", *params.IsVisible))
	}

	// 获取总数 (只使用筛选条件)
	count, err := game_config.HeroAttributeTypes(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询属性类型总数失败: %w", err)
	}

	// 构建完整查询条件 (包含排序和分页)
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序 (按类别和显示顺序)
	queryMods = append(queryMods, qm.OrderBy("category ASC, display_order ASC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	attributeTypes, err := game_config.HeroAttributeTypes(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询属性类型列表失败: %w", err)
	}

	return attributeTypes, count, nil
}

// Create 创建属性类型
func (r *heroAttributeTypeRepositoryImpl) Create(ctx context.Context, attributeType *game_config.HeroAttributeType) error {
	// 生成 UUID
	if attributeType.ID == "" {
		attributeType.ID = uuid.New().String()
	}

	// 设置时间戳 (time.Time, 不是 null.Time)
	now := time.Now()
	attributeType.CreatedAt = now
	attributeType.UpdatedAt = now

	// 插入数据库
	if err := attributeType.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建属性类型失败: %w", err)
	}

	return nil
}

// Update 更新属性类型
func (r *heroAttributeTypeRepositoryImpl) Update(ctx context.Context, attributeType *game_config.HeroAttributeType) error {
	// 更新时间戳 (time.Time, 不是 null.Time)
	attributeType.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := attributeType.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新属性类型失败: %w", err)
	}

	return nil
}

// Delete 软删除属性类型
func (r *heroAttributeTypeRepositoryImpl) Delete(ctx context.Context, attributeTypeID string) error {
	attributeType, err := r.GetByID(ctx, attributeTypeID)
	if err != nil {
		return err
	}

	// 软删除
	attributeType.DeletedAt.SetValid(time.Now())
	attributeType.UpdatedAt = time.Now()

	if _, err := attributeType.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除属性类型失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *heroAttributeTypeRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.HeroAttributeTypes(
		qm.Where("attribute_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查属性类型代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
