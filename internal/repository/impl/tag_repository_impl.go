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

type tagRepositoryImpl struct {
	db *sql.DB
}

// NewTagRepository 创建标签仓储实例
func NewTagRepository(db *sql.DB) interfaces.TagRepository {
	return &tagRepositoryImpl{db: db}
}

// GetByID 根据ID获取标签
func (r *tagRepositoryImpl) GetByID(ctx context.Context, tagID string) (*game_config.Tag, error) {
	tag, err := game_config.Tags(
		qm.Where("id = ? AND deleted_at IS NULL", tagID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("标签不存在: %s", tagID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	return tag, nil
}

// GetByCode 根据代码获取标签
func (r *tagRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.Tag, error) {
	tag, err := game_config.Tags(
		qm.Where("tag_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("标签不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	return tag, nil
}

// List 获取标签列表
func (r *tagRepositoryImpl) List(ctx context.Context, params interfaces.TagQueryParams) ([]*game_config.Tag, int64, error) {
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

	// 获取总数 (只使用筛选条件)
	count, err := game_config.Tags(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询标签总数失败: %w", err)
	}

	// 构建完整查询条件 (包含排序和分页)
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序
	queryMods = append(queryMods, qm.OrderBy("display_order ASC, created_at DESC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	tags, err := game_config.Tags(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询标签列表失败: %w", err)
	}

	return tags, count, nil
}

// Create 创建标签
func (r *tagRepositoryImpl) Create(ctx context.Context, tag *game_config.Tag) error {
	// 生成 UUID
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}

	// 设置时间戳（time.Time 类型，不需要 SetValid）
	now := time.Now()
	tag.CreatedAt = now
	tag.UpdatedAt = now

	// 插入数据库
	if err := tag.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建标签失败: %w", err)
	}

	return nil
}

// Update 更新标签
func (r *tagRepositoryImpl) Update(ctx context.Context, tag *game_config.Tag) error {
	// 更新时间戳（time.Time 类型，不需要 SetValid）
	tag.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := tag.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新标签失败: %w", err)
	}

	return nil
}

// Delete 软删除标签
func (r *tagRepositoryImpl) Delete(ctx context.Context, tagID string) error {
	tag, err := r.GetByID(ctx, tagID)
	if err != nil {
		return err
	}

	// 软删除
	now := time.Now()
	tag.DeletedAt.SetValid(now) // DeletedAt 是 null.Time 类型，需要 SetValid
	tag.UpdatedAt = now         // UpdatedAt 是 time.Time 类型，直接赋值

	if _, err := tag.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除标签失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *tagRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.Tags(
		qm.Where("tag_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查标签代码是否存在失败: %w", err)
	}

	return count > 0, nil
}
