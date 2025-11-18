package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// DungeonService 地城服务
type DungeonService struct {
	dungeonRepo     interfaces.DungeonRepository
	dungeonRoomRepo interfaces.DungeonRoomRepository
	db              *sql.DB
}

// DungeonServiceDeps allows custom dependency injection (for tests).
type DungeonServiceDeps struct {
	DB              *sql.DB
	DungeonRepo     interfaces.DungeonRepository
	DungeonRoomRepo interfaces.DungeonRoomRepository
}

// NewDungeonService 创建地城服务
func NewDungeonService(db *sql.DB) *DungeonService {
	return NewDungeonServiceWithDeps(DungeonServiceDeps{
		DB: db,
	})
}

// NewDungeonServiceWithDeps allows tests to supply custom repositories.
func NewDungeonServiceWithDeps(deps DungeonServiceDeps) *DungeonService {
	svc := &DungeonService{
		dungeonRepo:     deps.DungeonRepo,
		dungeonRoomRepo: deps.DungeonRoomRepo,
		db:              deps.DB,
	}
	if svc.dungeonRepo == nil {
		svc.dungeonRepo = impl.NewDungeonRepository(svc.db)
	}
	if svc.dungeonRoomRepo == nil {
		svc.dungeonRoomRepo = impl.NewDungeonRoomRepository(svc.db)
	}
	return svc
}

// GetDungeons 获取地城列表
func (s *DungeonService) GetDungeons(ctx context.Context, params interfaces.DungeonQueryParams) ([]*game_config.Dungeon, int64, error) {
	return s.dungeonRepo.List(ctx, params)
}

// GetDungeonByID 根据ID获取地城详情
func (s *DungeonService) GetDungeonByID(ctx context.Context, dungeonID string) (*game_config.Dungeon, error) {
	return s.dungeonRepo.GetByID(ctx, dungeonID)
}

// GetDungeonByCode 根据代码获取地城
func (s *DungeonService) GetDungeonByCode(ctx context.Context, code string) (*game_config.Dungeon, error) {
	return s.dungeonRepo.GetByCode(ctx, code)
}

// CreateDungeon 创建地城
func (s *DungeonService) CreateDungeon(ctx context.Context, req *dto.CreateDungeonRequest) (*game_config.Dungeon, error) {
	// 验证地城代码唯一性
	exists, err := s.dungeonRepo.Exists(ctx, req.DungeonCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("地城代码已存在: %s", req.DungeonCode))
	}

	// 验证等级区间
	if req.MinLevel > req.MaxLevel {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小等级不能大于最大等级")
	}

	// 验证限时配置
	if req.IsTimeLimited {
		if req.TimeLimitStart == nil || req.TimeLimitEnd == nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时地城必须设置开始和结束时间")
		}
		startTime, err := time.Parse(time.RFC3339, *req.TimeLimitStart)
		if err != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时开始时间格式错误")
		}
		endTime, err := time.Parse(time.RFC3339, *req.TimeLimitEnd)
		if err != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时结束时间格式错误")
		}
		if startTime.After(endTime) {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时开始时间不能晚于结束时间")
		}
	}

	// 验证挑战次数配置
	if req.RequiresAttempts {
		if req.MaxAttemptsPerDay == nil || *req.MaxAttemptsPerDay <= 0 {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "需要挑战次数限制时必须设置大于0的最大挑战次数")
		}
	}

	// 验证房间序列
	if err := s.validateRoomSequence(ctx, req.RoomSequence); err != nil {
		return nil, err
	}

	// 构建地城实体
	dungeon := &game_config.Dungeon{
		DungeonCode:      req.DungeonCode,
		DungeonName:      req.DungeonName,
		MinLevel:         req.MinLevel,
		MaxLevel:         req.MaxLevel,
		IsTimeLimited:    req.IsTimeLimited,
		RequiresAttempts: req.RequiresAttempts,
		IsActive:         req.IsActive,
	}

	// 设置可选字段
	if req.Description != nil {
		dungeon.Description = null.StringFrom(*req.Description)
	}

	if req.IsTimeLimited && req.TimeLimitStart != nil && req.TimeLimitEnd != nil {
		startTime, _ := time.Parse(time.RFC3339, *req.TimeLimitStart)
		endTime, _ := time.Parse(time.RFC3339, *req.TimeLimitEnd)
		dungeon.TimeLimitStart = null.TimeFrom(startTime)
		dungeon.TimeLimitEnd = null.TimeFrom(endTime)
	}

	if req.RequiresAttempts && req.MaxAttemptsPerDay != nil {
		dungeon.MaxAttemptsPerDay = null.Int16From(*req.MaxAttemptsPerDay)
	}

	// 序列化房间序列
	roomSequenceJSON, err := json.Marshal(req.RoomSequence)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化房间序列失败")
	}
	dungeon.RoomSequence = types.JSON(roomSequenceJSON)

	// 创建地城
	if err := s.dungeonRepo.Create(ctx, dungeon); err != nil {
		return nil, err
	}

	return dungeon, nil
}

// UpdateDungeon 更新地城
func (s *DungeonService) UpdateDungeon(ctx context.Context, dungeonID string, req *dto.UpdateDungeonRequest) (*game_config.Dungeon, error) {
	// 获取地城
	dungeon, err := s.dungeonRepo.GetByID(ctx, dungeonID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.DungeonName != nil {
		dungeon.DungeonName = *req.DungeonName
	}

	if req.MinLevel != nil {
		dungeon.MinLevel = *req.MinLevel
	}

	if req.MaxLevel != nil {
		dungeon.MaxLevel = *req.MaxLevel
	}

	// 验证等级区间
	if dungeon.MinLevel > dungeon.MaxLevel {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小等级不能大于最大等级")
	}

	if req.Description != nil {
		dungeon.Description = null.StringFrom(*req.Description)
	}

	if req.IsTimeLimited != nil {
		dungeon.IsTimeLimited = *req.IsTimeLimited
	}

	if req.TimeLimitStart != nil {
		startTime, err := time.Parse(time.RFC3339, *req.TimeLimitStart)
		if err != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时开始时间格式错误")
		}
		dungeon.TimeLimitStart = null.TimeFrom(startTime)
	}

	if req.TimeLimitEnd != nil {
		endTime, err := time.Parse(time.RFC3339, *req.TimeLimitEnd)
		if err != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时结束时间格式错误")
		}
		dungeon.TimeLimitEnd = null.TimeFrom(endTime)
	}

	// 验证限时配置
	if dungeon.IsTimeLimited {
		if !dungeon.TimeLimitStart.Valid || !dungeon.TimeLimitEnd.Valid {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时地城必须设置开始和结束时间")
		}
		if dungeon.TimeLimitStart.Time.After(dungeon.TimeLimitEnd.Time) {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "限时开始时间不能晚于结束时间")
		}
	}

	if req.RequiresAttempts != nil {
		dungeon.RequiresAttempts = *req.RequiresAttempts
	}

	if req.MaxAttemptsPerDay != nil {
		dungeon.MaxAttemptsPerDay = null.Int16From(*req.MaxAttemptsPerDay)
	}

	// 验证挑战次数配置
	if dungeon.RequiresAttempts {
		if !dungeon.MaxAttemptsPerDay.Valid || dungeon.MaxAttemptsPerDay.Int16 <= 0 {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "需要挑战次数限制时必须设置大于0的最大挑战次数")
		}
	}

	if len(req.RoomSequence) > 0 {
		// 验证房间序列
		if err := s.validateRoomSequence(ctx, req.RoomSequence); err != nil {
			return nil, err
		}

		// 序列化房间序列
		roomSequenceJSON, err := json.Marshal(req.RoomSequence)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化房间序列失败")
		}
		dungeon.RoomSequence = types.JSON(roomSequenceJSON)
	}

	if req.IsActive != nil {
		dungeon.IsActive = *req.IsActive
	}

	// 更新地城
	if err := s.dungeonRepo.Update(ctx, dungeon); err != nil {
		return nil, err
	}

	return dungeon, nil
}

// DeleteDungeon 删除地城
func (s *DungeonService) DeleteDungeon(ctx context.Context, dungeonID string) error {
	return s.dungeonRepo.Delete(ctx, dungeonID)
}

// validateRoomSequence 验证房间序列
func (s *DungeonService) validateRoomSequence(ctx context.Context, sequence []dto.RoomSequenceItem) error {
	if len(sequence) == 0 {
		return newRoomSequenceError("房间序列不能为空")
	}

	// 提取所有房间ID
	roomIDs := make([]string, 0, len(sequence))
	sortMap := make(map[int]bool)
	roomIDMap := make(map[string]bool)
	makeSeqErr := func(format string, args ...interface{}) error {
		return newRoomSequenceError(fmt.Sprintf(format, args...))
	}

	for _, item := range sequence {
		// 检查sort值唯一性
		if sortMap[item.Sort] {
			return makeSeqErr("房间序列中存在重复的sort值: %d", item.Sort)
		}
		sortMap[item.Sort] = true

		// 检查房间ID唯一性
		if roomIDMap[item.RoomID] {
			return makeSeqErr("房间序列中存在重复的房间ID: %s", item.RoomID)
		}
		roomIDMap[item.RoomID] = true

		roomIDs = append(roomIDs, item.RoomID)
	}

	if len(sequence) > 0 {
		if !sortMap[1] {
			return makeSeqErr("房间序列的 sort 必须从 1 开始")
		}
		for expected := 1; expected <= len(sequence); expected++ {
			if !sortMap[expected] {
				return makeSeqErr("房间序列缺少排序值: %d", expected)
			}
		}
	}

	// 批量查询房间是否存在(通过ID查询)
	rooms, err := s.dungeonRoomRepo.GetByIDs(ctx, roomIDs)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeDatabaseError, "查询房间失败")
	}

	// 检查所有房间是否都存在
	foundRoomIDs := make(map[string]bool)
	for _, room := range rooms {
		foundRoomIDs[room.ID] = true
	}

	for _, roomID := range roomIDs {
		if !foundRoomIDs[roomID] {
			return makeSeqErr("房间不存在: %s", roomID)
		}
	}

	// 验证条件跳过/返回的目标房间存在
	for _, item := range sequence {
		if item.ConditionalSkip != nil {
			if targetRoom, ok := item.ConditionalSkip["target_room"].(string); ok && targetRoom != "" {
				if !roomIDMap[targetRoom] {
					return makeSeqErr("条件跳过的目标房间不在序列中: %s", targetRoom)
				}
			}
		}
		if item.ConditionalReturn != nil {
			if targetRoom, ok := item.ConditionalReturn["target_room"].(string); ok && targetRoom != "" {
				if !roomIDMap[targetRoom] {
					return makeSeqErr("条件返回的目标房间不在序列中: %s", targetRoom)
				}
			}
		}
	}

	// 验证无循环引用
	if err := s.detectCycles(sequence); err != nil {
		return err
	}

	return nil
}

// detectCycles 检测房间序列中的循环引用
func (s *DungeonService) detectCycles(sequence []dto.RoomSequenceItem) error {
	// 构建邻接表
	graph := make(map[string][]string)
	for _, item := range sequence {
		graph[item.RoomID] = []string{}

		// 添加条件跳过的边
		if item.ConditionalSkip != nil {
			if targetRoom, ok := item.ConditionalSkip["target_room"].(string); ok && targetRoom != "" {
				graph[item.RoomID] = append(graph[item.RoomID], targetRoom)
			}
		}

		// 添加条件返回的边
		if item.ConditionalReturn != nil {
			if targetRoom, ok := item.ConditionalReturn["target_room"].(string); ok && targetRoom != "" {
				graph[item.RoomID] = append(graph[item.RoomID], targetRoom)
			}
		}
	}

	// 使用DFS检测环
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(roomID string) bool
	dfs = func(roomID string) bool {
		visited[roomID] = true
		recStack[roomID] = true

		for _, neighbor := range graph[roomID] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true // 发现环
			}
		}

		recStack[roomID] = false
		return false
	}

	// 检查每个节点
	for _, item := range sequence {
		if !visited[item.RoomID] {
			if dfs(item.RoomID) {
				return newRoomSequenceError(fmt.Sprintf("房间序列存在循环引用,涉及房间: %s", item.RoomID))
			}
		}
	}

	return nil
}

func newRoomSequenceError(message string) error {
	return xerrors.FromCode(xerrors.CodeInvalidParams).
		WithMetadata("field", "room_sequence").
		WithMetadata("user_message", message)
}
