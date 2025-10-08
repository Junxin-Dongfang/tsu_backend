package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// ClassAdvancedRequirementRepositoryImpl 职业进阶要求仓储实现
type ClassAdvancedRequirementRepositoryImpl struct {
	db *sql.DB
}

// NewClassAdvancedRequirementRepository 创建职业进阶要求仓储
func NewClassAdvancedRequirementRepository(db *sql.DB) interfaces.ClassAdvancedRequirementRepository {
	return &ClassAdvancedRequirementRepositoryImpl{db: db}
}

// GetByID 根据ID获取进阶要求
func (r *ClassAdvancedRequirementRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.ClassAdvancedRequirement, error) {
	requirement, err := game_config.FindClassAdvancedRequirement(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.NewNotFoundError("ClassAdvancedRequirement", id)
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶要求失败")
	}
	return requirement, nil
}

// GetByFromClass 获取指定源职业的所有进阶路径
func (r *ClassAdvancedRequirementRepositoryImpl) GetByFromClass(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	requirements, err := game_config.ClassAdvancedRequirements(
		game_config.ClassAdvancedRequirementWhere.FromClassID.EQ(fromClassID),
		game_config.ClassAdvancedRequirementWhere.DeletedAt.IsNull(),
		qm.OrderBy("display_order ASC, required_level ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶路径失败")
	}
	return requirements, nil
}

// GetByToClass 获取可以进阶到指定职业的所有路径
func (r *ClassAdvancedRequirementRepositoryImpl) GetByToClass(ctx context.Context, toClassID string) ([]*game_config.ClassAdvancedRequirement, error) {
	requirements, err := game_config.ClassAdvancedRequirements(
		game_config.ClassAdvancedRequirementWhere.ToClassID.EQ(toClassID),
		game_config.ClassAdvancedRequirementWhere.DeletedAt.IsNull(),
		qm.OrderBy("display_order ASC, required_level ASC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶来源失败")
	}
	return requirements, nil
}

// GetByClassPair 获取指定职业对的进阶要求
func (r *ClassAdvancedRequirementRepositoryImpl) GetByClassPair(ctx context.Context, fromClassID, toClassID string) (*game_config.ClassAdvancedRequirement, error) {
	requirement, err := game_config.ClassAdvancedRequirements(
		game_config.ClassAdvancedRequirementWhere.FromClassID.EQ(fromClassID),
		game_config.ClassAdvancedRequirementWhere.ToClassID.EQ(toClassID),
		game_config.ClassAdvancedRequirementWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.NewNotFoundError("ClassAdvancedRequirement", fmt.Sprintf("%s -> %s", fromClassID, toClassID))
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶要求失败")
	}
	return requirement, nil
}

// List 获取进阶要求列表（支持分页和筛选）
func (r *ClassAdvancedRequirementRepositoryImpl) List(ctx context.Context, params interfaces.ListAdvancedRequirementsParams) ([]*game_config.ClassAdvancedRequirement, int64, error) {
	// 构建查询条件
	var queries []qm.QueryMod
	queries = append(queries, game_config.ClassAdvancedRequirementWhere.DeletedAt.IsNull())

	// 筛选条件
	if params.FromClassID != nil {
		queries = append(queries, game_config.ClassAdvancedRequirementWhere.FromClassID.EQ(*params.FromClassID))
	}
	if params.ToClassID != nil {
		queries = append(queries, game_config.ClassAdvancedRequirementWhere.ToClassID.EQ(*params.ToClassID))
	}
	if params.IsActive != nil {
		var activeFilter game_config.ClassAdvancedRequirement
		activeFilter.IsActive.SetValid(*params.IsActive)
		queries = append(queries, game_config.ClassAdvancedRequirementWhere.IsActive.EQ(activeFilter.IsActive))
	}

	// 总数查询
	total, err := game_config.ClassAdvancedRequirements(queries...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶要求总数失败")
	}

	// 排序
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "display_order"
	}
	sortDir := params.SortDir
	if sortDir == "" {
		sortDir = "asc"
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

	// 查询列表
	requirements, err := game_config.ClassAdvancedRequirements(queries...).All(ctx, r.db)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶要求列表失败")
	}

	return requirements, total, nil
}

// Create 创建进阶要求
func (r *ClassAdvancedRequirementRepositoryImpl) Create(ctx context.Context, requirement *game_config.ClassAdvancedRequirement) error {
	if err := requirement.Insert(ctx, r.db, boil.Infer()); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业进阶要求失败")
	}
	return nil
}

// Update 更新进阶要求
func (r *ClassAdvancedRequirementRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	requirement, err := game_config.FindClassAdvancedRequirement(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return xerrors.NewNotFoundError("ClassAdvancedRequirement", id)
		}
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶要求失败")
	}

	// 更新字段
	for key, value := range updates {
		switch key {
		case "from_class_id":
			if v, ok := value.(string); ok {
				requirement.FromClassID = v
			}
		case "to_class_id":
			if v, ok := value.(string); ok {
				requirement.ToClassID = v
			}
		case "required_level":
			if v, ok := value.(int); ok {
				requirement.RequiredLevel = v
			}
		case "required_honor":
			if v, ok := value.(int); ok {
				requirement.RequiredHonor = v
			}
		case "required_job_change_count":
			if v, ok := value.(int); ok {
				requirement.RequiredJobChangeCount = v
			}
		case "required_attributes":
			if v, ok := value.([]byte); ok {
				requirement.RequiredAttributes.UnmarshalJSON(v)
			}
		case "required_skills":
			if v, ok := value.([]byte); ok {
				requirement.RequiredSkills.UnmarshalJSON(v)
			}
		case "required_items":
			if v, ok := value.([]byte); ok {
				requirement.RequiredItems.UnmarshalJSON(v)
			}
		case "is_active":
			if v, ok := value.(bool); ok {
				requirement.IsActive.SetValid(v)
			}
		case "display_order":
			if v, ok := value.(int); ok {
				requirement.DisplayOrder.SetValid(int16(v))
			}
		}
	}

	_, err = requirement.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新职业进阶要求失败")
	}
	return nil
}

// Delete 删除进阶要求（软删除）
func (r *ClassAdvancedRequirementRepositoryImpl) Delete(ctx context.Context, id string) error {
	requirement, err := game_config.FindClassAdvancedRequirement(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return xerrors.NewNotFoundError("ClassAdvancedRequirement", id)
		}
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶要求失败")
	}

	requirement.DeletedAt.SetValid(time.Now())
	_, err = requirement.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at"))
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除职业进阶要求失败")
	}
	return nil
}

// BatchCreate 批量创建进阶要求
func (r *ClassAdvancedRequirementRepositoryImpl) BatchCreate(ctx context.Context, requirements []*game_config.ClassAdvancedRequirement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开始事务失败")
	}
	defer tx.Rollback()

	for _, req := range requirements {
		if err := req.Insert(ctx, tx, boil.Infer()); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "批量创建职业进阶要求失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}
	return nil
}

// GetAdvancementPaths 获取完整进阶路径树（递归查找）
// 使用深度优先搜索（DFS）算法查找所有可能的进阶路径
func (r *ClassAdvancedRequirementRepositoryImpl) GetAdvancementPaths(ctx context.Context, fromClassID string, maxDepth int) ([][]*game_config.ClassAdvancedRequirement, error) {
	if maxDepth <= 0 {
		maxDepth = 5 // 默认最大深度5级
	}

	var paths [][]*game_config.ClassAdvancedRequirement
	var currentPath []*game_config.ClassAdvancedRequirement
	visited := make(map[string]bool) // 防止循环依赖

	// 深度优先搜索
	var dfs func(classID string, depth int) error
	dfs = func(classID string, depth int) error {
		if depth >= maxDepth {
			return nil
		}

		// 获取当前职业的所有进阶路径
		nextRequirements, err := r.GetByFromClass(ctx, classID)
		if err != nil {
			return err
		}

		// 如果没有更多进阶路径，保存当前路径
		if len(nextRequirements) == 0 {
			if len(currentPath) > 0 {
				// 复制当前路径
				pathCopy := make([]*game_config.ClassAdvancedRequirement, len(currentPath))
				copy(pathCopy, currentPath)
				paths = append(paths, pathCopy)
			}
			return nil
		}

		// 遍历所有可能的下一步
		for _, req := range nextRequirements {
			// 防止循环依赖
			if visited[req.ToClassID] {
				continue
			}

			// 标记为已访问
			visited[req.ToClassID] = true
			currentPath = append(currentPath, req)

			// 递归查找
			if err := dfs(req.ToClassID, depth+1); err != nil {
				return err
			}

			// 回溯
			currentPath = currentPath[:len(currentPath)-1]
			delete(visited, req.ToClassID)
		}

		return nil
	}

	// 从起始职业开始搜索
	visited[fromClassID] = true
	if err := dfs(fromClassID, 0); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业进阶路径树失败")
	}

	return paths, nil
}
