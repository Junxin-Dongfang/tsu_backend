// Package service 提供管理端业务逻辑服务
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// EquipmentSetService 装备套装管理服务
type EquipmentSetService struct {
	setRepo  interfaces.EquipmentSetRepository
	itemRepo interfaces.ItemRepository
}

// NewEquipmentSetService 创建装备套装管理服务
func NewEquipmentSetService(db *sql.DB) *EquipmentSetService {
	return &EquipmentSetService{
		setRepo:  impl.NewEquipmentSetRepository(db),
		itemRepo: impl.NewItemRepository(db),
	}
}

// CreateSet 创建套装配置
func (s *EquipmentSetService) CreateSet(ctx context.Context, req *dto.CreateEquipmentSetRequest) (*dto.EquipmentSetResponse, error) {
	// 1. 验证 set_code 唯一性
	existing, err := s.setRepo.GetByCode(ctx, req.SetCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "套装代码已存在")
	}

	// 2. 验证套装效果
	if validateErr := s.validateSetEffects(req.SetEffects); validateErr != nil {
		return nil, validateErr
	}

	// 3. 转换为JSON
	setEffectsJSON, err := json.Marshal(req.SetEffects)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "套装效果JSON序列化失败")
	}

	// 4. 创建实体
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	setConfig := &game_config.EquipmentSetConfig{
		ID:         uuid.New().String(),
		SetCode:    req.SetCode,
		SetName:    req.SetName,
		SetEffects: setEffectsJSON,
		IsActive:   isActive,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.Description != nil {
		setConfig.Description.SetValid(*req.Description)
	}

	// 5. 保存到数据库
	if err := s.setRepo.Create(ctx, setConfig); err != nil {
		return nil, err
	}

	// 6. 返回响应
	return s.toSetResponse(setConfig)
}

// GetSetList 查询套装列表
func (s *EquipmentSetService) GetSetList(ctx context.Context, req *dto.ListEquipmentSetsRequest) (*dto.EquipmentSetListResponse, error) {
	// 1. 设置默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 2. 构建查询参数
	listReq := &interfaces.ListEquipmentSetsRequest{
		Page:     page,
		PageSize: pageSize,
	}

	if req.Keyword != nil {
		listReq.Keyword = *req.Keyword
	}
	if req.IsActive != nil {
		listReq.IsActive = req.IsActive
	}
	if req.SortBy != nil {
		listReq.SortBy = *req.SortBy
	}
	if req.SortOrder != nil {
		listReq.SortOrder = *req.SortOrder
	}

	// 3. 查询列表
	sets, total, err := s.setRepo.List(ctx, listReq)
	if err != nil {
		return nil, err
	}

	// 4. 转换为响应
	responses := make([]dto.EquipmentSetResponse, 0, len(sets))
	for _, set := range sets {
		resp, err := s.toSetResponse(set)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}

	return &dto.EquipmentSetListResponse{
		Sets:  responses,
		Total: total,
		Page:  page,
	}, nil
}

// GetSetByID 查询套装详情
func (s *EquipmentSetService) GetSetByID(ctx context.Context, id string) (*dto.EquipmentSetResponse, error) {
	// 1. 查询套装配置
	setConfig, err := s.setRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 转换为响应
	return s.toSetResponse(setConfig)
}

// UpdateSet 更新套装配置
func (s *EquipmentSetService) UpdateSet(ctx context.Context, id string, req *dto.UpdateEquipmentSetRequest) (*dto.EquipmentSetResponse, error) {
	// 1. 查询现有套装配置
	setConfig, err := s.setRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 更新字段
	if req.SetName != nil {
		setConfig.SetName = *req.SetName
	}

	if req.Description != nil {
		setConfig.Description.SetValid(*req.Description)
	}

	if req.SetEffects != nil {
		// 验证套装效果
		if err := s.validateSetEffects(req.SetEffects); err != nil {
			return nil, err
		}

		// 转换为JSON
		setEffectsJSON, err := json.Marshal(req.SetEffects)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "套装效果JSON序列化失败")
		}
		setConfig.SetEffects = setEffectsJSON
	}

	if req.IsActive != nil {
		setConfig.IsActive = *req.IsActive
	}

	setConfig.UpdatedAt = time.Now()

	// 3. 保存更新
	if err := s.setRepo.Update(ctx, setConfig); err != nil {
		return nil, err
	}

	// 4. 返回响应
	return s.toSetResponse(setConfig)
}

// DeleteSet 删除套装配置
func (s *EquipmentSetService) DeleteSet(ctx context.Context, id string) error {
	// 1. 查询套装配置（验证存在性）
	_, err := s.setRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. 软删除
	return s.setRepo.Delete(ctx, id)
}

// GetSetItems 查询套装包含的装备列表
func (s *EquipmentSetService) GetSetItems(ctx context.Context, setID string, page, pageSize int) (*dto.SetItemListResponse, error) {
	// 1. 验证套装存在
	_, err := s.setRepo.GetByID(ctx, setID)
	if err != nil {
		return nil, err
	}

	// 2. 查询装备列表
	items, err := s.setRepo.GetItemsBySetID(ctx, setID)
	if err != nil {
		return nil, err
	}

	// 3. 转换为响应
	responses := make([]dto.SetItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, s.toItemResponse(item))
	}

	return &dto.SetItemListResponse{
		Items: responses,
		Total: int64(len(items)),
		Page:  page,
	}, nil
}

// GetUnassignedItems 查询未关联套装的装备列表
func (s *EquipmentSetService) GetUnassignedItems(ctx context.Context, req *dto.ListEquipmentSetsRequest) (*dto.SetItemListResponse, error) {
	// 1. 设置默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 2. 构建查询参数
	params := interfaces.ListItemParams{
		Page:     page,
		PageSize: pageSize,
	}

	if req.Keyword != nil {
		params.Keyword = req.Keyword
	}

	// 3. 查询未关联装备
	items, total, err := s.itemRepo.GetUnassignedItems(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询未关联套装的装备失败")
	}

	// 4. 转换为响应
	responses := make([]dto.SetItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, s.toItemResponse(item))
	}

	return &dto.SetItemListResponse{
		Items: responses,
		Total: total,
		Page:  page,
	}, nil
}

// validateSetEffects 验证套装效果
func (s *EquipmentSetService) validateSetEffects(effects []dto.SetEffectDTO) error {
	if len(effects) == 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "套装效果不能为空")
	}

	// 验证档位件数递增
	prevPieceCount := 0
	for i, effect := range effects {
		// 验证件数递增
		if effect.PieceCount <= prevPieceCount {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("套装效果档位件数必须递增，第%d个档位件数(%d)不大于前一个(%d)", i+1, effect.PieceCount, prevPieceCount))
		}
		prevPieceCount = effect.PieceCount

		// 验证至少有一种效果
		if len(effect.OutOfCombatEffects) == 0 && len(effect.InCombatEffects) == 0 {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d个档位必须至少包含局外加成或局内加成", i+1))
		}

		// 验证局外加成
		if err := s.validateOutOfCombatEffects(effect.OutOfCombatEffects, i+1); err != nil {
			return err
		}

		// 验证局内加成
		if err := s.validateInCombatEffects(effect.InCombatEffects, i+1); err != nil {
			return err
		}
	}

	return nil
}

// validateOutOfCombatEffects 验证局外加成
func (s *EquipmentSetService) validateOutOfCombatEffects(effects []dto.OutOfCombatEffect, tierIndex int) error {
	for i, effect := range effects {
		// 验证 Data_type
		if effect.DataType != "Status" {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d档位第%d个局外加成的Data_type必须为Status", tierIndex, i+1))
		}

		// 验证 Bouns_type
		if effect.BonusType != "bonus" && effect.BonusType != "percent" {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d档位第%d个局外加成的Bouns_type必须为bonus或percent", tierIndex, i+1))
		}

		// 验证 Bouns_Number 是有效数字
		if _, err := strconv.ParseFloat(effect.BonusNumber, 64); err != nil {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d档位第%d个局外加成的Bouns_Number必须是有效数字", tierIndex, i+1))
		}

		// TODO: 验证 Data_ID 是否是有效的属性ID（需要属性列表）
	}

	return nil
}

// validateInCombatEffects 验证局内加成
func (s *EquipmentSetService) validateInCombatEffects(effects []dto.InCombatEffect, tierIndex int) error {
	for i, effect := range effects {
		// 验证 Trigger_type
		validTriggerTypes := map[string]bool{
			"on_attack": true,
			"on_hit":    true,
			"on_kill":   true,
			"passive":   true,
		}
		if !validTriggerTypes[effect.TriggerType] {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d档位第%d个局内加成的Trigger_type无效", tierIndex, i+1))
		}

		// 验证 Trigger_chance（如果不是passive）
		if effect.TriggerType != "passive" && effect.TriggerChance != nil {
			chance, err := strconv.ParseFloat(*effect.TriggerChance, 64)
			if err != nil {
				return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d档位第%d个局内加成的Trigger_chance必须是有效数字", tierIndex, i+1))
			}
			if chance < 0 || chance > 100 {
				return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("第%d档位第%d个局内加成的Trigger_chance必须在0-100之间", tierIndex, i+1))
			}
		}

		// TODO: 验证 Data_ID 是否是有效的技能/效果ID
	}

	return nil
}

// toSetResponse 转换为套装响应
func (s *EquipmentSetService) toSetResponse(setConfig *game_config.EquipmentSetConfig) (*dto.EquipmentSetResponse, error) {
	// 解析套装效果
	var effects []dto.SetEffectResponse
	if err := json.Unmarshal(setConfig.SetEffects, &effects); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "套装效果JSON解析失败")
	}

	resp := &dto.EquipmentSetResponse{
		ID:         setConfig.ID,
		SetCode:    setConfig.SetCode,
		SetName:    setConfig.SetName,
		SetEffects: effects,
		IsActive:   setConfig.IsActive,
		CreatedAt:  setConfig.CreatedAt,
		UpdatedAt:  setConfig.UpdatedAt,
	}

	if setConfig.Description.Valid {
		desc := setConfig.Description.String
		resp.Description = &desc
	}

	return resp, nil
}

// toItemResponse 转换为装备响应
func (s *EquipmentSetService) toItemResponse(item *game_config.Item) dto.SetItemResponse {
	resp := dto.SetItemResponse{
		ID:          item.ID,
		ItemCode:    item.ItemCode,
		ItemName:    item.ItemName,
		ItemType:    item.ItemType,
		ItemQuality: item.ItemQuality,
	}

	if item.EquipSlot.Valid {
		slot := item.EquipSlot.String
		resp.EquipSlot = &slot
	}

	if item.SetID.Valid {
		setID := item.SetID.String
		resp.SetID = &setID
	}

	return resp
}

// BatchAssignItems 批量分配物品到套装
func (s *EquipmentSetService) BatchAssignItems(ctx context.Context, setID string, req *dto.BatchAssignItemsToSetRequest) (*dto.BatchAssignItemsToSetResponse, error) {
	// 1. 验证套装存在
	setConfig, err := s.setRepo.GetByID(ctx, setID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}
	if setConfig == nil {
		return nil, xerrors.New(xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}

	// 2. 验证所有物品并分配
	assignedCount := 0
	failedItems := []dto.FailedItem{}

	for _, itemID := range req.ItemIDs {
		// 验证物品存在
		item, err := s.itemRepo.GetByID(ctx, itemID)
		if err != nil || item == nil {
			failedItems = append(failedItems, dto.FailedItem{
				ItemID: itemID,
				Reason: "物品不存在或已删除",
			})
			continue
		}

		// 验证物品类型为装备
		if item.ItemType != "equipment" {
			failedItems = append(failedItems, dto.FailedItem{
				ItemID: itemID,
				Reason: "只有装备类型的物品可以分配到套装",
			})
			continue
		}

		// 更新物品的set_id
		item.SetID.SetValid(setID)
		item.UpdatedAt = time.Now()

		if err := s.itemRepo.Update(ctx, item); err != nil {
			failedItems = append(failedItems, dto.FailedItem{
				ItemID: itemID,
				Reason: "更新失败",
			})
			continue
		}

		assignedCount++
	}

	// 如果有失败的物品，返回错误
	if len(failedItems) > 0 {
		return &dto.BatchAssignItemsToSetResponse{
			AssignedCount: assignedCount,
			FailedItems:   failedItems,
		}, xerrors.New(xerrors.CodeInvalidParams, "部分物品分配失败")
	}

	return &dto.BatchAssignItemsToSetResponse{
		AssignedCount: assignedCount,
		FailedItems:   []dto.FailedItem{},
	}, nil
}

// BatchRemoveItems 批量移除物品从套装
func (s *EquipmentSetService) BatchRemoveItems(ctx context.Context, setID string, req *dto.BatchRemoveItemsFromSetRequest) (*dto.BatchRemoveItemsFromSetResponse, error) {
	// 1. 验证套装存在
	setConfig, err := s.setRepo.GetByID(ctx, setID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}
	if setConfig == nil {
		return nil, xerrors.New(xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}

	// 2. 移除所有物品
	removedCount := 0

	for _, itemID := range req.ItemIDs {
		// 验证物品存在
		item, err := s.itemRepo.GetByID(ctx, itemID)
		if err != nil || item == nil {
			continue // 跳过不存在的物品
		}

		// 验证物品属于该套装
		if !item.SetID.Valid || item.SetID.String != setID {
			continue // 跳过不属于该套装的物品
		}

		// 移除套装关联
		item.SetID.SetValid("")
		item.SetID.Valid = false
		item.UpdatedAt = time.Now()

		if err := s.itemRepo.Update(ctx, item); err != nil {
			continue // 跳过更新失败的物品
		}

		removedCount++
	}

	return &dto.BatchRemoveItemsFromSetResponse{
		RemovedCount: removedCount,
	}, nil
}

// RemoveItem 移除单个物品从套装
func (s *EquipmentSetService) RemoveItem(ctx context.Context, setID, itemID string) error {
	// 1. 验证套装存在
	setConfig, err := s.setRepo.GetByID(ctx, setID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}
	if setConfig == nil {
		return xerrors.New(xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}

	// 2. 验证物品存在
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil || item == nil {
		return xerrors.New(xerrors.CodeResourceNotFound, "物品不存在或已删除")
	}

	// 3. 验证物品属于该套装
	if !item.SetID.Valid || item.SetID.String != setID {
		return xerrors.New(xerrors.CodeInvalidParams, "物品不属于该套装")
	}

	// 4. 移除套装关联
	item.SetID.SetValid("")
	item.SetID.Valid = false
	item.UpdatedAt = time.Now()

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新物品失败")
	}

	return nil
}
