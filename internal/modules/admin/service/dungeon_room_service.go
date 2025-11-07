package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// DungeonRoomService 房间服务
type DungeonRoomService struct {
	roomRepo interfaces.DungeonRoomRepository
	db       *sql.DB
}

// NewDungeonRoomService 创建房间服务
func NewDungeonRoomService(db *sql.DB) *DungeonRoomService {
	return &DungeonRoomService{
		roomRepo: impl.NewDungeonRoomRepository(db),
		db:       db,
	}
}

// GetRooms 获取房间列表
func (s *DungeonRoomService) GetRooms(ctx context.Context, params interfaces.DungeonRoomQueryParams) ([]*game_config.DungeonRoom, int64, error) {
	return s.roomRepo.List(ctx, params)
}

// GetRoomByID 根据ID获取房间详情
func (s *DungeonRoomService) GetRoomByID(ctx context.Context, roomID string) (*game_config.DungeonRoom, error) {
	return s.roomRepo.GetByID(ctx, roomID)
}

// GetRoomByCode 根据代码获取房间
func (s *DungeonRoomService) GetRoomByCode(ctx context.Context, code string) (*game_config.DungeonRoom, error) {
	return s.roomRepo.GetByCode(ctx, code)
}

// CreateRoom 创建房间
func (s *DungeonRoomService) CreateRoom(ctx context.Context, req *dto.CreateRoomRequest) (*game_config.DungeonRoom, error) {
	// 验证房间代码唯一性
	exists, err := s.roomRepo.Exists(ctx, req.RoomCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("房间代码已存在: %s", req.RoomCode))
	}

	// 构建房间实体
	room := &game_config.DungeonRoom{
		RoomCode: req.RoomCode,
		RoomType: req.RoomType,
		IsActive: req.IsActive,
	}

	if req.RoomName != nil {
		room.RoomName = null.StringFrom(*req.RoomName)
	}

	if req.TriggerID != nil {
		room.TriggerID = null.StringFrom(*req.TriggerID)
	}

	// 序列化开启条件
	if req.OpenConditions != nil {
		conditionsJSON, err := json.Marshal(req.OpenConditions)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化开启条件失败")
		}
		room.OpenConditions = types.JSON(conditionsJSON)
	} else {
		room.OpenConditions = types.JSON([]byte("{}"))
	}

	// 创建房间
	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// UpdateRoom 更新房间
func (s *DungeonRoomService) UpdateRoom(ctx context.Context, roomID string, req *dto.UpdateRoomRequest) (*game_config.DungeonRoom, error) {
	// 获取房间
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.RoomName != nil {
		room.RoomName = null.StringFrom(*req.RoomName)
	}

	if req.RoomType != nil {
		room.RoomType = *req.RoomType
	}

	if req.TriggerID != nil {
		room.TriggerID = null.StringFrom(*req.TriggerID)
	}

	if req.OpenConditions != nil {
		conditionsJSON, err := json.Marshal(req.OpenConditions)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化开启条件失败")
		}
		room.OpenConditions = types.JSON(conditionsJSON)
	}

	if req.IsActive != nil {
		room.IsActive = *req.IsActive
	}

	// 更新房间
	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// DeleteRoom 删除房间
func (s *DungeonRoomService) DeleteRoom(ctx context.Context, roomID string) error {
	return s.roomRepo.Delete(ctx, roomID)
}

