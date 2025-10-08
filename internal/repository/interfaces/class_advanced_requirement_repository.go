package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ClassAdvancedRequirementRepository 职业进阶要求仓储接口
type ClassAdvancedRequirementRepository interface {
	// GetByID 根据ID获取进阶要求
	GetByID(ctx context.Context, id string) (*game_config.ClassAdvancedRequirement, error)

	// GetByFromClass 获取指定源职业的所有进阶路径
	GetByFromClass(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error)

	// GetByToClass 获取可以进阶到指定职业的所有路径
	GetByToClass(ctx context.Context, toClassID string) ([]*game_config.ClassAdvancedRequirement, error)

	// GetByClassPair 获取指定职业对的进阶要求
	GetByClassPair(ctx context.Context, fromClassID, toClassID string) (*game_config.ClassAdvancedRequirement, error)

	// List 获取进阶要求列表（支持分页和筛选）
	List(ctx context.Context, params ListAdvancedRequirementsParams) ([]*game_config.ClassAdvancedRequirement, int64, error)

	// Create 创建进阶要求
	Create(ctx context.Context, requirement *game_config.ClassAdvancedRequirement) error

	// Update 更新进阶要求
	Update(ctx context.Context, id string, updates map[string]interface{}) error

	// Delete 删除进阶要求（软删除）
	Delete(ctx context.Context, id string) error

	// BatchCreate 批量创建进阶要求
	BatchCreate(ctx context.Context, requirements []*game_config.ClassAdvancedRequirement) error

	// GetAdvancementPaths 获取完整进阶路径树（递归查找）
	GetAdvancementPaths(ctx context.Context, fromClassID string, maxDepth int) ([][]*game_config.ClassAdvancedRequirement, error)
}

// ListAdvancedRequirementsParams 职业进阶要求查询参数
type ListAdvancedRequirementsParams struct {
	FromClassID *string // 源职业ID筛选
	ToClassID   *string // 目标职业ID筛选
	IsActive    *bool   // 激活状态筛选
	Page        int     // 页码
	PageSize    int     // 每页数量
	SortBy      string  // 排序字段: display_order, required_level, created_at
	SortDir     string  // 排序方向: asc, desc
}
