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

// DungeonEventService 事件配置服务
type DungeonEventService struct {
	eventRepo    interfaces.DungeonEventRepository
	dropPoolRepo interfaces.DropPoolRepository
	itemRepo     interfaces.ItemRepository
	buffRepo     interfaces.BuffRepository
	db           *sql.DB
}

// NewDungeonEventService 创建事件配置服务
func NewDungeonEventService(db *sql.DB) *DungeonEventService {
	return &DungeonEventService{
		eventRepo:    impl.NewDungeonEventRepository(db),
		dropPoolRepo: impl.NewDropPoolRepository(db),
		itemRepo:     impl.NewItemRepository(db),
		buffRepo:     impl.NewBuffRepository(db),
		db:           db,
	}
}

// GetEventByID 根据ID获取事件配置
func (s *DungeonEventService) GetEventByID(ctx context.Context, eventID string) (*game_config.DungeonEvent, error) {
	return s.eventRepo.GetByID(ctx, eventID)
}

// GetEventByCode 根据代码获取事件配置
func (s *DungeonEventService) GetEventByCode(ctx context.Context, code string) (*game_config.DungeonEvent, error) {
	return s.eventRepo.GetByCode(ctx, code)
}

// CreateEvent 创建事件配置
func (s *DungeonEventService) CreateEvent(ctx context.Context, req *dto.CreateEventRequest) (*game_config.DungeonEvent, error) {
	// 验证事件代码唯一性
	exists, err := s.eventRepo.Exists(ctx, req.EventCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("事件配置代码已存在: %s", req.EventCode))
	}

	// 验证经验值
	if req.RewardExp < 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "经验奖励不能为负数")
	}

	// 验证施加效果中的Buff
	if err := s.validateApplyEffects(ctx, req.ApplyEffects); err != nil {
		return nil, err
	}

	// 验证掉落配置
	if err := s.validateDropConfig(ctx, req.DropConfig); err != nil {
		return nil, err
	}

	// 构建事件配置实体
	event := &game_config.DungeonEvent{
		EventCode: req.EventCode,
		RewardExp: null.IntFrom(req.RewardExp),
		IsActive:  req.IsActive,
	}

	if req.EventDescription != nil {
		event.EventDescription = null.StringFrom(*req.EventDescription)
	}

	// 序列化施加效果
	applyEffectsJSON, err := json.Marshal(req.ApplyEffects)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化施加效果失败")
	}
	event.ApplyEffects = types.JSON(applyEffectsJSON)

	// 序列化掉落配置
	dropConfigJSON, err := json.Marshal(req.DropConfig)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化掉落配置失败")
	}
	event.DropConfig = types.JSON(dropConfigJSON)

	if req.EventEndDesc != nil {
		event.EventEndDesc = null.StringFrom(*req.EventEndDesc)
	}

	// 创建事件配置
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// UpdateEvent 更新事件配置
func (s *DungeonEventService) UpdateEvent(ctx context.Context, eventID string, req *dto.UpdateEventRequest) (*game_config.DungeonEvent, error) {
	// 获取事件配置
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.EventDescription != nil {
		event.EventDescription = null.StringFrom(*req.EventDescription)
	}

	if req.ApplyEffects != nil {
		applyEffectsJSON, err := json.Marshal(req.ApplyEffects)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化施加效果失败")
		}
		event.ApplyEffects = types.JSON(applyEffectsJSON)
	}

	if req.DropConfig != nil {
		dropConfigJSON, err := json.Marshal(req.DropConfig)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化掉落配置失败")
		}
		event.DropConfig = types.JSON(dropConfigJSON)
	}

	if req.RewardExp != nil {
		if *req.RewardExp < 0 {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "经验奖励不能为负数")
		}
		event.RewardExp = null.IntFrom(*req.RewardExp)
	}

	if req.EventEndDesc != nil {
		event.EventEndDesc = null.StringFrom(*req.EventEndDesc)
	}

	if req.IsActive != nil {
		event.IsActive = *req.IsActive
	}

	// 更新事件配置
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// DeleteEvent 删除事件配置
func (s *DungeonEventService) DeleteEvent(ctx context.Context, eventID string) error {
	return s.eventRepo.Delete(ctx, eventID)
}

// validateApplyEffects 验证施加效果配置
func (s *DungeonEventService) validateApplyEffects(ctx context.Context, effects []dto.ApplyEffectItem) error {
	for _, effect := range effects {
		// 验证Buff存在性
		_, err := s.buffRepo.GetByCode(ctx, effect.BuffCode)
		if err != nil {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("Buff不存在: %s", effect.BuffCode))
		}

		// 验证目标类型
		validTargets := map[string]bool{
			"all_heroes":   true,
			"all_monsters": true,
			"all":          true,
			"random_hero":  true,
		}
		if !validTargets[effect.Target] {
			return xerrors.New(xerrors.CodeInvalidParams,
				fmt.Sprintf("无效的效果目标类型: %s", effect.Target))
		}
	}

	return nil
}

// validateDropConfig 验证掉落配置
func (s *DungeonEventService) validateDropConfig(ctx context.Context, config dto.DropConfig) error {
	// 验证掉落池存在性
	if config.DropPoolID != nil && *config.DropPoolID != "" {
		_, err := s.dropPoolRepo.GetByID(ctx, *config.DropPoolID)
		if err != nil {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("掉落池不存在: %s", *config.DropPoolID))
		}
	}

	// 验证保底物品存在性
	for _, item := range config.GuaranteedItems {
		_, err := s.itemRepo.GetByCode(ctx, item.ItemCode)
		if err != nil {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("物品不存在: %s", item.ItemCode))
		}

		// 验证数量
		if item.Quantity <= 0 {
			return xerrors.New(xerrors.CodeInvalidParams,
				fmt.Sprintf("物品数量必须大于0: %s", item.ItemCode))
		}
	}

	return nil
}

