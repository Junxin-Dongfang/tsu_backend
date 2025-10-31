package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// EquipmentSlotConfigRepository 装备槽位配置Repository接口
type EquipmentSlotConfigRepository interface {
	// Create 创建槽位配置
	Create(ctx context.Context, slot *game_config.EquipmentSlot) error

	// GetByID 根据ID获取槽位配置
	GetByID(ctx context.Context, id string) (*game_config.EquipmentSlot, error)

	// GetByCode 根据槽位代码获取槽位配置
	GetByCode(ctx context.Context, code string) (*game_config.EquipmentSlot, error)

	// List 查询槽位列表（支持分页和筛选）
	List(ctx context.Context, params ListSlotConfigParams) ([]*game_config.EquipmentSlot, error)

	// Count 统计槽位数量（用于分页）
	Count(ctx context.Context, params ListSlotConfigParams) (int64, error)

	// Update 更新槽位配置
	Update(ctx context.Context, slot *game_config.EquipmentSlot) error

	// Delete 删除槽位配置（软删除）
	Delete(ctx context.Context, id string) error
}

// ListSlotConfigParams 槽位配置列表查询参数
type ListSlotConfigParams struct {
	Page      int     // 页码（从1开始）
	PageSize  int     // 每页数量
	SlotType  *string // 槽位类型筛选
	IsActive  *bool   // 激活状态筛选
	Keyword   *string // 关键词搜索（slot_code, slot_name）
	SortBy    string  // 排序字段（默认 display_order）
	SortOrder string  // 排序方向（asc/desc，默认 asc）
}

