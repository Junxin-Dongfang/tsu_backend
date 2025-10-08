package impl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type classRepositoryImpl struct {
	db *sql.DB
}

// NewClassRepository 创建职业仓储实例
func NewClassRepository(db *sql.DB) interfaces.ClassRepository {
	return &classRepositoryImpl{db: db}
}

// GetByID 根据ID获取职业
func (r *classRepositoryImpl) GetByID(ctx context.Context, classID string) (*game_config.Class, error) {
	class, err := game_config.Classes(
		qm.Where("id = ? AND deleted_at IS NULL", classID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("职业不存在: %s", classID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询职业失败: %w", err)
	}

	return class, nil
}

// GetByCode 根据职业代码获取职业
func (r *classRepositoryImpl) GetByCode(ctx context.Context, classCode string) (*game_config.Class, error) {
	class, err := game_config.Classes(
		qm.Where("class_code = ? AND deleted_at IS NULL", classCode),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("职业不存在: %s", classCode)
	}
	if err != nil {
		return nil, fmt.Errorf("查询职业失败: %w", err)
	}

	return class, nil
}

// List 获取职业列表（分页）
func (r *classRepositoryImpl) List(ctx context.Context, params interfaces.ClassQueryParams) ([]*game_config.Class, int64, error) {
	// 构建查询条件
	queries := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 层级筛选
	if params.Tier != "" {
		queries = append(queries, qm.And("tier = ?", params.Tier))
	}

	// 激活状态筛选
	if params.IsActive != nil {
		queries = append(queries, qm.And("is_active = ?", *params.IsActive))
	}

	// 可见性筛选
	if params.IsVisible != nil {
		queries = append(queries, qm.And("is_visible = ?", *params.IsVisible))
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "display_order"
	}
	sortDir := strings.ToUpper(params.SortDir)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "ASC"
	}

	// 获取总数
	total, err := game_config.Classes(queries...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("统计职业总数失败: %w", err)
	}

	// 分页
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize

	queries = append(queries,
		qm.OrderBy(fmt.Sprintf("%s %s", sortBy, sortDir)),
		qm.Limit(params.PageSize),
		qm.Offset(offset),
	)

	// 查询职业列表
	classes, err := game_config.Classes(queries...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询职业列表失败: %w", err)
	}

	return classes, total, nil
}

// Create 创建职业
func (r *classRepositoryImpl) Create(ctx context.Context, class *game_config.Class) error {
	// 生成 UUID
	if class.ID == "" {
		class.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	class.CreatedAt = now
	class.UpdatedAt = now

	// 插入数据库
	if err := class.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建职业失败: %w", err)
	}

	return nil
}

// Update 更新职业信息
func (r *classRepositoryImpl) Update(ctx context.Context, class *game_config.Class) error {
	// 更新时间戳
	class.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := class.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新职业失败: %w", err)
	}

	return nil
}

// Delete 软删除职业
func (r *classRepositoryImpl) Delete(ctx context.Context, classID string) error {
	class, err := r.GetByID(ctx, classID)
	if err != nil {
		return err
	}

	now := time.Now()
	class.DeletedAt.SetValid(now)
	class.UpdatedAt = now

	return r.Update(ctx, class)
}

// Exists 检查职业是否存在
func (r *classRepositoryImpl) Exists(ctx context.Context, classID string) (bool, error) {
	count, err := game_config.Classes(
		qm.Where("id = ? AND deleted_at IS NULL", classID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查职业是否存在失败: %w", err)
	}

	return count > 0, nil
}
