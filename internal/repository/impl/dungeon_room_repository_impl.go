package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type dungeonRoomRepositoryImpl struct {
	db *sql.DB
}

// NewDungeonRoomRepository 创建房间仓储实例
func NewDungeonRoomRepository(db *sql.DB) interfaces.DungeonRoomRepository {
	return &dungeonRoomRepositoryImpl{db: db}
}

// GetByID 根据ID获取房间
func (r *dungeonRoomRepositoryImpl) GetByID(ctx context.Context, roomID string) (*game_config.DungeonRoom, error) {
	room, err := game_config.DungeonRooms(
		qm.Where("id = ? AND deleted_at IS NULL", roomID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("房间不存在: %s", roomID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询房间失败: %w", err)
	}

	return room, nil
}

// GetByCode 根据代码获取房间
func (r *dungeonRoomRepositoryImpl) GetByCode(ctx context.Context, code string) (*game_config.DungeonRoom, error) {
	room, err := game_config.DungeonRooms(
		qm.Where("room_code = ? AND deleted_at IS NULL", code),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("房间不存在: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("查询房间失败: %w", err)
	}

	return room, nil
}

// GetByCodes 根据代码列表批量获取房间
func (r *dungeonRoomRepositoryImpl) GetByCodes(ctx context.Context, codes []string) ([]*game_config.DungeonRoom, error) {
	if len(codes) == 0 {
		return []*game_config.DungeonRoom{}, nil
	}

	// 转换为 interface{} 切片
	codeInterfaces := make([]interface{}, len(codes))
	for i, code := range codes {
		codeInterfaces[i] = code
	}

	rooms, err := game_config.DungeonRooms(
		qm.WhereIn("room_code IN ?", codeInterfaces...),
		qm.Where("deleted_at IS NULL"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("批量查询房间失败: %w", err)
	}

	return rooms, nil
}

// GetByIDs 根据ID列表批量获取房间
func (r *dungeonRoomRepositoryImpl) GetByIDs(ctx context.Context, ids []string) ([]*game_config.DungeonRoom, error) {
	if len(ids) == 0 {
		return []*game_config.DungeonRoom{}, nil
	}

	// 转换为 interface{} 切片
	idInterfaces := make([]interface{}, len(ids))
	for i, id := range ids {
		idInterfaces[i] = id
	}

	rooms, err := game_config.DungeonRooms(
		qm.WhereIn("id IN ?", idInterfaces...),
		qm.Where("deleted_at IS NULL"),
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("批量查询房间失败: %w", err)
	}

	return rooms, nil
}

// List 获取房间列表
func (r *dungeonRoomRepositoryImpl) List(ctx context.Context, params interfaces.DungeonRoomQueryParams) ([]*game_config.DungeonRoom, int64, error) {
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.RoomCode != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("room_code ILIKE ?", "%"+*params.RoomCode+"%"))
	}
	if params.RoomName != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("room_name ILIKE ?", "%"+*params.RoomName+"%"))
	}
	if params.RoomType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("room_type = ?", *params.RoomType))
	}
	if params.IsActive != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("is_active = ?", *params.IsActive))
	}

	// 获取总数
	count, err := game_config.DungeonRooms(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询房间总数失败: %w", err)
	}

	// 排序
	orderBy := "created_at"
	if params.OrderBy != "" {
		orderBy = params.OrderBy
	}
	if params.OrderDesc {
		baseQueryMods = append(baseQueryMods, qm.OrderBy(orderBy+" DESC"))
	} else {
		baseQueryMods = append(baseQueryMods, qm.OrderBy(orderBy+" ASC"))
	}

	// 分页
	if params.Limit > 0 {
		baseQueryMods = append(baseQueryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		baseQueryMods = append(baseQueryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	rooms, err := game_config.DungeonRooms(baseQueryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询房间列表失败: %w", err)
	}

	return rooms, count, nil
}

// Create 创建房间
func (r *dungeonRoomRepositoryImpl) Create(ctx context.Context, room *game_config.DungeonRoom) error {
	// 生成UUID
	if room.ID == "" {
		room.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	room.CreatedAt = now
	room.UpdatedAt = now

	// 插入数据库
	if err := room.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建房间失败: %w", err)
	}

	return nil
}

// Update 更新房间
func (r *dungeonRoomRepositoryImpl) Update(ctx context.Context, room *game_config.DungeonRoom) error {
	// 更新时间戳
	room.UpdatedAt = time.Now()

	// 更新数据库
	if _, err := room.Update(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("更新房间失败: %w", err)
	}

	return nil
}

// Delete 软删除房间
func (r *dungeonRoomRepositoryImpl) Delete(ctx context.Context, roomID string) error {
	// 查询房间
	room, err := r.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	// 设置删除时间
	now := time.Now()
	room.DeletedAt = null.TimeFrom(now)
	room.UpdatedAt = now

	// 更新数据库
	if _, err := room.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除房间失败: %w", err)
	}

	return nil
}

// Exists 检查代码是否存在
func (r *dungeonRoomRepositoryImpl) Exists(ctx context.Context, code string) (bool, error) {
	count, err := game_config.DungeonRooms(
		qm.Where("room_code = ? AND deleted_at IS NULL", code),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查房间代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

// ExistsExcludingID 检查代码是否存在（排除指定ID）
func (r *dungeonRoomRepositoryImpl) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	count, err := game_config.DungeonRooms(
		qm.Where("room_code = ? AND id != ? AND deleted_at IS NULL", code, excludeID),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查房间代码是否存在失败: %w", err)
	}

	return count > 0, nil
}

