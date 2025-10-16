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

type classAdvancedRequirementRepository struct {
	db *sql.DB
}

// NewClassAdvancedRequirementRepository 创建职业进阶要求 Repository
func NewClassAdvancedRequirementRepository(db *sql.DB) interfaces.ClassAdvancedRequirementRepository {
	return &classAdvancedRequirementRepository{db: db}
}

// GetByFromAndTo 获取进阶路径配置
func (r *classAdvancedRequirementRepository) GetByFromAndTo(ctx context.Context, fromClassID, toClassID string) (*game_config.ClassAdvancedRequirement, error) {
	req, err := game_config.ClassAdvancedRequirements(
		qm.Where("from_class_id = ?", fromClassID),
		qm.And("to_class_id = ?", toClassID),
		qm.And("is_active = ?", true),
		qm.And("deleted_at IS NULL"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return req, err
}

// GetAdvancementOptions 获取某职业的所有可进阶选项（game 模块）
func (r *classAdvancedRequirementRepository) GetAdvancementOptions(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	reqs, err := game_config.ClassAdvancedRequirements(
		qm.Where("from_class_id = ?", fromClassID),
		qm.And("is_active = ?", true),
		qm.And("deleted_at IS NULL"),
		qm.OrderBy("display_order ASC"),
	).All(ctx, r.db)

	return reqs, err
}

// GetByID 根据ID获取进阶要求
func (r *classAdvancedRequirementRepository) GetByID(ctx context.Context, id string) (*game_config.ClassAdvancedRequirement, error) {
	req, err := game_config.ClassAdvancedRequirements(
		qm.Where("id = ?", id),
		qm.And("deleted_at IS NULL"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("进阶要求不存在: %s", id)
	}

	return req, err
}

// GetByClassPair 获取进阶路径配置（admin 模块，同 GetByFromAndTo）
func (r *classAdvancedRequirementRepository) GetByClassPair(ctx context.Context, fromClassID, toClassID string) (*game_config.ClassAdvancedRequirement, error) {
	return r.GetByFromAndTo(ctx, fromClassID, toClassID)
}

// GetByFromClass 获取指定源职业的所有进阶路径（admin 模块）
func (r *classAdvancedRequirementRepository) GetByFromClass(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	reqs, err := game_config.ClassAdvancedRequirements(
		qm.Where("from_class_id = ?", fromClassID),
		qm.And("deleted_at IS NULL"),
		qm.OrderBy("display_order ASC"),
	).All(ctx, r.db)

	return reqs, err
}

// GetByToClass 获取可以进阶到指定职业的所有路径
func (r *classAdvancedRequirementRepository) GetByToClass(ctx context.Context, toClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	reqs, err := game_config.ClassAdvancedRequirements(
		qm.Where("to_class_id = ?", toClassID),
		qm.And("deleted_at IS NULL"),
		qm.OrderBy("display_order ASC"),
	).All(ctx, r.db)

	return reqs, err
}

// List 获取进阶要求列表（分页）
func (r *classAdvancedRequirementRepository) List(ctx context.Context, params interfaces.ListAdvancedRequirementsParams) ([]*game_config.ClassAdvancedRequirement, int64, error) {
	queries := []qm.QueryMod{
		qm.Where("deleted_at IS NULL"),
	}

	// 筛选条件
	if params.FromClassID != nil {
		queries = append(queries, qm.And("from_class_id = ?", *params.FromClassID))
	}

	if params.ToClassID != nil {
		queries = append(queries, qm.And("to_class_id = ?", *params.ToClassID))
	}

	if params.IsActive != nil {
		queries = append(queries, qm.And("is_active = ?", *params.IsActive))
	}

	// 排序
	sortBy := "display_order"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	sortDir := "ASC"
	if params.SortDir == "DESC" {
		sortDir = "DESC"
	}
	queries = append(queries, qm.OrderBy(fmt.Sprintf("%s %s", sortBy, sortDir)))

	// 分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		queries = append(queries, qm.Limit(params.PageSize), qm.Offset(offset))
	}

	// 查询数据
	reqs, err := game_config.ClassAdvancedRequirements(queries...).All(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	// 查询总数
	countQueries := []qm.QueryMod{qm.Where("deleted_at IS NULL")}
	if params.FromClassID != nil {
		countQueries = append(countQueries, qm.And("from_class_id = ?", *params.FromClassID))
	}
	if params.ToClassID != nil {
		countQueries = append(countQueries, qm.And("to_class_id = ?", *params.ToClassID))
	}
	if params.IsActive != nil {
		countQueries = append(countQueries, qm.And("is_active = ?", *params.IsActive))
	}

	total, err := game_config.ClassAdvancedRequirements(countQueries...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	return reqs, total, nil
}

// Create 创建进阶要求
func (r *classAdvancedRequirementRepository) Create(ctx context.Context, requirement *game_config.ClassAdvancedRequirement) error {
	now := time.Now()
	requirement.ID = uuid.New().String()
	requirement.CreatedAt = now
	requirement.UpdatedAt = now

	return requirement.Insert(ctx, r.db, boil.Infer())
}

// Update 更新进阶要求
func (r *classAdvancedRequirementRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	_, err := game_config.ClassAdvancedRequirements(
		qm.Where("id = ?", id),
		qm.And("deleted_at IS NULL"),
	).UpdateAll(ctx, r.db, updates)

	return err
}

// Delete 软删除进阶要求
func (r *classAdvancedRequirementRepository) Delete(ctx context.Context, id string) error {
	_, err := game_config.ClassAdvancedRequirements(
		qm.Where("id = ?", id),
	).UpdateAll(ctx, r.db, map[string]interface{}{
		"deleted_at": time.Now(),
		"updated_at": time.Now(),
	})

	return err
}

// BatchCreate 批量创建进阶要求
func (r *classAdvancedRequirementRepository) BatchCreate(ctx context.Context, requirements []*game_config.ClassAdvancedRequirement) error {
	now := time.Now()

	for _, req := range requirements {
		req.ID = uuid.New().String()
		req.CreatedAt = now
		req.UpdatedAt = now

		if err := req.Insert(ctx, r.db, boil.Infer()); err != nil {
			return err
		}
	}

	return nil
}

// GetAdvancementPaths 获取完整进阶路径树（递归查询）
func (r *classAdvancedRequirementRepository) GetAdvancementPaths(ctx context.Context, fromClassID string, maxDepth int) ([][]*game_config.ClassAdvancedRequirement, error) {
	if maxDepth <= 0 {
		maxDepth = 5 // 默认最大深度
	}

	var paths [][]*game_config.ClassAdvancedRequirement
	var currentPath []*game_config.ClassAdvancedRequirement

	// 递归查找路径
	var findPaths func(classID string, depth int, visited map[string]bool)
	findPaths = func(classID string, depth int, visited map[string]bool) {
		if depth >= maxDepth {
			return
		}

		// 获取当前职业的所有可进阶选项
		reqs, err := r.GetByFromClass(ctx, classID)
		if err != nil || len(reqs) == 0 {
			// 如果没有更多进阶选项，保存当前路径
			if len(currentPath) > 0 {
				pathCopy := make([]*game_config.ClassAdvancedRequirement, len(currentPath))
				copy(pathCopy, currentPath)
				paths = append(paths, pathCopy)
			}
			return
		}

		for _, req := range reqs {
			// 避免循环
			if visited[req.ToClassID] {
				continue
			}

			// 添加到当前路径
			currentPath = append(currentPath, req)
			visited[req.ToClassID] = true

			// 递归查找下一级
			findPaths(req.ToClassID, depth+1, visited)

			// 回溯
			currentPath = currentPath[:len(currentPath)-1]
			delete(visited, req.ToClassID)
		}
	}

	visited := make(map[string]bool)
	visited[fromClassID] = true
	findPaths(fromClassID, 0, visited)

	return paths, nil
}
