package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// ListEquipmentSetsRequest 查询套装列表请求
type ListEquipmentSetsRequest struct {
	Page      int
	PageSize  int
	Keyword   string // 搜索关键词（set_code, set_name）
	IsActive  *bool  // 筛选激活状态
	SortBy    string // 排序字段
	SortOrder string // 排序方向（asc/desc）
}

// EquipmentSetRepository 装备套装Repository接口
type EquipmentSetRepository interface {
	// GetByID 根据ID查询套装配置
	GetByID(ctx context.Context, id string) (*game_config.EquipmentSetConfig, error)

	// GetByIDs 根据ID列表批量查询套装配置
	GetByIDs(ctx context.Context, ids []string) ([]*game_config.EquipmentSetConfig, error)

	// GetByCode 根据代码查询套装配置（用于唯一性验证）
	GetByCode(ctx context.Context, code string) (*game_config.EquipmentSetConfig, error)

	// List 查询套装列表
	List(ctx context.Context, req *ListEquipmentSetsRequest) ([]*game_config.EquipmentSetConfig, int64, error)

	// Count 统计套装总数（用于分页）
	Count(ctx context.Context, req *ListEquipmentSetsRequest) (int64, error)

	// Create 创建套装配置
	Create(ctx context.Context, config *game_config.EquipmentSetConfig) error

	// Update 更新套装配置
	Update(ctx context.Context, config *game_config.EquipmentSetConfig) error

	// Delete 软删除套装配置
	Delete(ctx context.Context, id string) error

	// GetItemsBySetID 查询套装包含的装备列表
	GetItemsBySetID(ctx context.Context, setID string) ([]*game_config.Item, error)
}

