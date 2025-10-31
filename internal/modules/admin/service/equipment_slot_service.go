package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// EquipmentSlotService 装备槽位配置Service
type EquipmentSlotService struct {
	db       *sql.DB
	slotRepo interfaces.EquipmentSlotConfigRepository
}

// NewEquipmentSlotService 创建装备槽位配置Service
func NewEquipmentSlotService(db *sql.DB) *EquipmentSlotService {
	return &EquipmentSlotService{
		db:       db,
		slotRepo: impl.NewEquipmentSlotConfigRepository(db),
	}
}

// CreateSlot 创建槽位配置
func (s *EquipmentSlotService) CreateSlot(ctx context.Context, req *dto.CreateSlotRequest) (*dto.SlotConfigResponse, error) {
	// 1. 验证槽位代码唯一性
	existing, err := s.slotRepo.GetByCode(ctx, req.SlotCode)
	if err != nil && err != sql.ErrNoRows {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询槽位代码失败")
	}
	if existing != nil {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "槽位代码已存在")
	}

	// 2. 创建槽位实体
	slot := &game_config.EquipmentSlot{
		ID:           uuid.New().String(),
		SlotCode:     req.SlotCode,
		SlotName:     req.SlotName,
		SlotType:     req.SlotType,
		DisplayOrder: req.DisplayOrder,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Icon != nil {
		slot.Icon.SetValid(*req.Icon)
	}
	if req.Description != nil {
		slot.Description.SetValid(*req.Description)
	}

	// 3. 保存到数据库
	if err := s.slotRepo.Create(ctx, slot); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建槽位配置失败")
	}

	// 4. 返回响应
	return s.toSlotConfigResponse(slot), nil
}

// GetSlotList 查询槽位列表
func (s *EquipmentSlotService) GetSlotList(ctx context.Context, params interfaces.ListSlotConfigParams) (*dto.SlotListResponse, error) {
	// 1. 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// 2. 查询槽位列表
	slots, err := s.slotRepo.List(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询槽位列表失败")
	}

	// 3. 查询总数
	total, err := s.slotRepo.Count(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询槽位总数失败")
	}

	// 4. 转换为响应
	responses := make([]dto.SlotConfigResponse, 0, len(slots))
	for _, slot := range slots {
		responses = append(responses, *s.toSlotConfigResponse(slot))
	}

	return &dto.SlotListResponse{
		Slots:    responses,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// GetSlotByID 获取槽位详情
func (s *EquipmentSlotService) GetSlotByID(ctx context.Context, id string) (*dto.SlotConfigResponse, error) {
	slot, err := s.slotRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.New(xerrors.CodeResourceNotFound, "槽位配置不存在")
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询槽位配置失败")
	}

	return s.toSlotConfigResponse(slot), nil
}

// UpdateSlot 更新槽位配置
func (s *EquipmentSlotService) UpdateSlot(ctx context.Context, id string, req *dto.UpdateSlotRequest) (*dto.SlotConfigResponse, error) {
	// 1. 查询槽位是否存在
	slot, err := s.slotRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.New(xerrors.CodeResourceNotFound, "槽位配置不存在")
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询槽位配置失败")
	}

	// 2. 更新字段（部分更新）
	if req.SlotName != nil {
		slot.SlotName = *req.SlotName
	}
	if req.DisplayOrder != nil {
		slot.DisplayOrder = *req.DisplayOrder
	}
	if req.Icon != nil {
		slot.Icon.SetValid(*req.Icon)
	}
	if req.Description != nil {
		slot.Description.SetValid(*req.Description)
	}
	if req.IsActive != nil {
		slot.IsActive = *req.IsActive
	}

	slot.UpdatedAt = time.Now()

	// 3. 保存更新
	if err := s.slotRepo.Update(ctx, slot); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新槽位配置失败")
	}

	// 4. 返回响应
	return s.toSlotConfigResponse(slot), nil
}

// DeleteSlot 删除槽位配置（软删除）
func (s *EquipmentSlotService) DeleteSlot(ctx context.Context, id string) error {
	// 1. 查询槽位是否存在
	_, err := s.slotRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return xerrors.New(xerrors.CodeResourceNotFound, "槽位配置不存在")
		}
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询槽位配置失败")
	}

	// 2. 软删除
	if err := s.slotRepo.Delete(ctx, id); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除槽位配置失败")
	}

	return nil
}

// toSlotConfigResponse 转换为SlotConfigResponse
func (s *EquipmentSlotService) toSlotConfigResponse(slot *game_config.EquipmentSlot) *dto.SlotConfigResponse {
	resp := &dto.SlotConfigResponse{
		ID:           slot.ID,
		SlotCode:     slot.SlotCode,
		SlotName:     slot.SlotName,
		SlotType:     slot.SlotType,
		DisplayOrder: slot.DisplayOrder,
		IsActive:     slot.IsActive,
		CreatedAt:    slot.CreatedAt,
		UpdatedAt:    slot.UpdatedAt,
	}

	if slot.Icon.Valid {
		resp.Icon = &slot.Icon.String
	}
	if slot.Description.Valid {
		resp.Description = &slot.Description.String
	}

	return resp
}

