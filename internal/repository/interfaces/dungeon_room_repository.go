package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

// DungeonRoomQueryParams 房间查询参数
type DungeonRoomQueryParams struct {
	RoomCode *string // 房间代码（模糊搜索）
	RoomName *string // 房间名称（模糊搜索）
	RoomType *string // 房间类型
	IsActive *bool   // 是否启用
	Limit    int     // 每页数量
	Offset   int     // 偏移量
	OrderBy  string  // 排序字段（created_at, updated_at）
	OrderDesc bool   // 是否降序
}

// DungeonRoomRepository 房间仓储接口
type DungeonRoomRepository interface {
	// GetByID 根据ID获取房间
	GetByID(ctx context.Context, roomID string) (*game_config.DungeonRoom, error)

	// GetByCode 根据代码获取房间
	GetByCode(ctx context.Context, code string) (*game_config.DungeonRoom, error)

	// GetByCodes 根据代码列表批量获取房间
	GetByCodes(ctx context.Context, codes []string) ([]*game_config.DungeonRoom, error)

	// GetByIDs 根据ID列表批量获取房间
	GetByIDs(ctx context.Context, ids []string) ([]*game_config.DungeonRoom, error)

	// List 获取房间列表
	List(ctx context.Context, params DungeonRoomQueryParams) ([]*game_config.DungeonRoom, int64, error)

	// Create 创建房间
	Create(ctx context.Context, room *game_config.DungeonRoom) error

	// Update 更新房间
	Update(ctx context.Context, room *game_config.DungeonRoom) error

	// Delete 软删除房间
	Delete(ctx context.Context, roomID string) error

	// Exists 检查代码是否存在
	Exists(ctx context.Context, code string) (bool, error)

	// ExistsExcludingID 检查代码是否存在（排除指定ID）
	ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error)
}

