package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ListAdvancedRequirementsParams 进阶要求列表查询参数
type ListAdvancedRequirementsParams struct {
	FromClassID *string
	ToClassID   *string
	IsActive    *bool
	Page        int
	PageSize    int
	SortBy      string
	SortDir     string
}

// ClassAdvancedRequirementRepository 职业进阶要求 Repository
type ClassAdvancedRequirementRepository interface {
	// GetByID 根据ID获取进阶要求
	GetByID(ctx context.Context, id string) (*game_config.ClassAdvancedRequirement, error)

	// GetByFromAndTo 获取进阶路径配置（game 模块使用）
	GetByFromAndTo(ctx context.Context, fromClassID, toClassID string) (*game_config.ClassAdvancedRequirement, error)

	// GetByClassPair 获取进阶路径配置（admin 模块使用，同 GetByFromAndTo）
	GetByClassPair(ctx context.Context, fromClassID, toClassID string) (*game_config.ClassAdvancedRequirement, error)

	// GetAdvancementOptions 获取某职业的所有可进阶选项（game 模块使用）
	GetAdvancementOptions(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error)

	// GetByFromClass 获取指定源职业的所有进阶路径（admin 模块使用）
	GetByFromClass(ctx context.Context, fromClassID string) ([]*game_config.ClassAdvancedRequirement, error)

	// GetByToClass 获取可以进阶到指定职业的所有路径
	GetByToClass(ctx context.Context, toClassID string) ([]*game_config.ClassAdvancedRequirement, error)

	// List 获取进阶要求列表（分页）
	List(ctx context.Context, params ListAdvancedRequirementsParams) ([]*game_config.ClassAdvancedRequirement, int64, error)

	// Create 创建进阶要求
	Create(ctx context.Context, requirement *game_config.ClassAdvancedRequirement) error

	// Update 更新进阶要求
	Update(ctx context.Context, id string, updates map[string]interface{}) error

	// Delete 软删除进阶要求
	Delete(ctx context.Context, id string) error

	// BatchCreate 批量创建进阶要求
	BatchCreate(ctx context.Context, requirements []*game_config.ClassAdvancedRequirement) error

	// GetAdvancementPaths 获取完整进阶路径树
	GetAdvancementPaths(ctx context.Context, fromClassID string, maxDepth int) ([][]*game_config.ClassAdvancedRequirement, error)
}
