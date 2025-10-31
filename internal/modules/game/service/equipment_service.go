package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// EquipmentService 装备穿戴服务
type EquipmentService struct {
	db                   *sql.DB
	itemRepo             interfaces.ItemRepository
	playerItemRepo       interfaces.PlayerItemRepository
	equipmentSlotRepo    interfaces.EquipmentSlotRepository
	heroRepo             interfaces.HeroRepository
	equipmentRepo        interfaces.EquipmentRepository
	equipmentEffectSvc   *EquipmentEffectService
	equipmentSetSvc      *EquipmentSetService
}

// NewEquipmentService 创建装备穿戴服务
func NewEquipmentService(db *sql.DB) *EquipmentService {
	equipmentRepo := impl.NewEquipmentRepository(db)
	equipmentSetRepo := impl.NewEquipmentSetRepository(db)
	itemRepo := impl.NewItemRepository(db)

	return &EquipmentService{
		db:                 db,
		itemRepo:           itemRepo,
		playerItemRepo:     impl.NewPlayerItemRepository(db),
		equipmentSlotRepo:  impl.NewEquipmentSlotRepository(db),
		heroRepo:           impl.NewHeroRepository(db),
		equipmentRepo:      equipmentRepo,
		equipmentEffectSvc: NewEquipmentEffectService(),
		equipmentSetSvc:    NewEquipmentSetService(equipmentSetRepo, equipmentRepo, itemRepo),
	}
}

// EquipItemRequest 穿戴装备请求
type EquipItemRequest struct {
	HeroID         string `json:"hero_id"`
	ItemInstanceID string `json:"item_instance_id"`
	SlotType       string `json:"slot_type,omitempty"` // 可选,如果不指定则自动选择
}

// EquipmentSlotDTO 装备槽位DTO
type EquipmentSlotDTO struct {
	ID              string
	HeroID          string
	SlotType        string
	SlotIndex       int
	IsUnlocked      bool
	UnlockLevel     *int
	EquippedItemID  *string
	AddedByItemID   *string
}

// PlayerItemDTO 玩家物品DTO
type PlayerItemDTO struct {
	ID                    string
	ItemID                string
	OwnerID               string
	ItemLocation          string
	CurrentDurability     *int
	MaxDurability         *int
	EnhancementLevel      *int
	StackCount            *int
}

// EquipItemResponse 穿戴装备响应
type EquipItemResponse struct {
	Success        bool              `json:"success"`
	Message        string            `json:"message"`
	EquippedSlot   *EquipmentSlotDTO `json:"equipped_slot,omitempty"`
	UnequippedItem *PlayerItemDTO    `json:"unequipped_item,omitempty"` // 如果替换了装备
}

// EquipItem 穿戴装备
func (s *EquipmentService) EquipItem(ctx context.Context, req *EquipItemRequest) (*EquipItemResponse, error) {
	// 1. 验证参数
	if req.HeroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}
	if req.ItemInstanceID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品实例ID不能为空")
	}

	// 2. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 3. 获取装备实例(带锁)
	itemInstance, err := s.playerItemRepo.GetByIDForUpdate(ctx, tx, req.ItemInstanceID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "装备实例不存在")
	}

	// 4. 获取装备配置
	itemConfig, err := s.itemRepo.GetByID(ctx, itemInstance.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "装备配置不存在")
	}

	// 5. 验证物品类型是装备
	if itemConfig.ItemType != "equipment" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "该物品不是装备,无法穿戴")
	}

	// 6. 获取英雄信息
	hero, err := s.heroRepo.GetByID(ctx, req.HeroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "英雄不存在")
	}

	// 7. 验证装备是否属于该玩家
	if itemInstance.OwnerID != hero.UserID {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "该装备不属于您")
	}

	// 8. 验证装备耐久度 > 0
	if itemInstance.CurrentDurability.Valid && itemInstance.CurrentDurability.Int <= 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "装备耐久度为0,无法装备,请先修复")
	}

	// 9. 验证职业要求
	if err := s.validateClassRequirement(itemConfig, hero); err != nil {
		return nil, err
	}

	// 10. 验证等级要求
	if err := s.validateLevelRequirement(itemConfig, hero); err != nil {
		return nil, err
	}

	// 11. 验证装备唯一性
	if err := s.validateUniqueness(ctx, itemConfig, hero); err != nil {
		return nil, err
	}

	// 12. 确定槽位类型
	slotType := req.SlotType
	if slotType == "" {
		if !itemConfig.EquipSlot.Valid {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "装备没有指定槽位类型")
		}
		slotType = itemConfig.EquipSlot.String
	}

	// 13. 查找可用槽位
	slot, err := s.equipmentSlotRepo.FindAvailableSlot(ctx, req.HeroID, slotType)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "没有可用的装备槽位")
	}

	// 14. 如果槽位已有装备,先卸下
	var unequippedItem *game_runtime.PlayerItem
	if slot.EquippedItemID.Valid {
		unequippedItem, err = s.unequipItemInternal(ctx, tx, slot.EquippedItemID.String)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "卸下原装备失败")
		}
	}

	// 15. 更新槽位装备ID
	slot.EquippedItemID.SetValid(req.ItemInstanceID)
	if err := s.equipmentSlotRepo.UpdateSlot(ctx, tx, slot); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新槽位失败")
	}

	// 16. 更新装备位置为"equipped"
	if err := s.playerItemRepo.UpdateLocation(ctx, tx, req.ItemInstanceID, "equipped"); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新装备位置失败")
	}

	// 17. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 18. TODO: 失效英雄属性缓存

	// 18. 转换为DTO
	resp := &EquipItemResponse{
		Success:      true,
		Message:      "装备穿戴成功",
		EquippedSlot: convertSlotToDTO(slot),
	}

	if unequippedItem != nil {
		resp.UnequippedItem = convertItemToDTO(unequippedItem)
	}

	return resp, nil
}

// validateClassRequirement 验证职业要求
func (s *EquipmentService) validateClassRequirement(itemConfig *game_config.Item, hero *game_runtime.Hero) error {
	// TODO: 职业要求通过 ItemClassRelations 关联表实现，需要查询关联表
	// if !itemConfig.RequiredClassID.Valid {
	// 	return nil // 没有职业要求
	// }
	//
	// if itemConfig.RequiredClassID.String != hero.ClassID {
	// 	return xerrors.New(xerrors.CodeInvalidParams, "职业不符合装备要求")
	// }
	_ = itemConfig
	_ = hero

	return nil
}

// validateLevelRequirement 验证等级要求
func (s *EquipmentService) validateLevelRequirement(itemConfig *game_config.Item, hero *game_runtime.Hero) error {
	if !itemConfig.RequiredLevel.Valid {
		return nil // 没有等级要求
	}

	if hero.CurrentLevel < itemConfig.RequiredLevel.Int16 {
		return xerrors.New(
			xerrors.CodeInvalidParams,
			fmt.Sprintf("等级不足,无法装备(需要等级%d)", itemConfig.RequiredLevel.Int16),
		)
	}

	return nil
}

// validateUniqueness 验证装备唯一性
func (s *EquipmentService) validateUniqueness(ctx context.Context, itemConfig *game_config.Item, hero *game_runtime.Hero) error {
	if !itemConfig.UniquenessType.Valid || itemConfig.UniquenessType.String == "none" {
		return nil // 没有唯一性限制
	}

	// TODO: 实现唯一性验证
	// - account: 检查该账户下所有角色是否已装备
	// - character: 检查该角色是否已装备
	// - team: 检查该队伍内其他成员是否已装备
	// - guild: 检查该公会内其他成员是否已装备

	return nil
}

// unequipItemInternal 卸下装备(内部方法,在事务中调用)
func (s *EquipmentService) unequipItemInternal(ctx context.Context, tx *sql.Tx, itemInstanceID string) (*game_runtime.PlayerItem, error) {
	// 获取装备实例
	item, err := s.playerItemRepo.GetByIDForUpdate(ctx, tx, itemInstanceID)
	if err != nil {
		return nil, err
	}

	// 更新装备位置为"backpack"
	if err := s.playerItemRepo.UpdateLocation(ctx, tx, itemInstanceID, "backpack"); err != nil {
		return nil, err
	}

	return item, nil
}

// UnequipItemRequest 卸下装备请求
type UnequipItemRequest struct {
	HeroID string `json:"hero_id"`
	SlotID string `json:"slot_id"`
}

// UnequipItemResponse 卸下装备响应
type UnequipItemResponse struct {
	Success        bool           `json:"success"`
	Message        string         `json:"message"`
	UnequippedItem *PlayerItemDTO `json:"unequipped_item,omitempty"`
}

// UnequipItem 卸下装备
func (s *EquipmentService) UnequipItem(ctx context.Context, req *UnequipItemRequest) (*UnequipItemResponse, error) {
	// 1. 验证参数
	if req.HeroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}
	if req.SlotID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "槽位ID不能为空")
	}

	// 2. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 3. 获取槽位
	slot, err := s.equipmentSlotRepo.GetSlotByID(ctx, req.SlotID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "槽位不存在")
	}

	// 4. 验证槽位属于该英雄
	if slot.HeroID != req.HeroID {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "该槽位不属于该英雄")
	}

	// 5. 验证槽位有装备
	if !slot.EquippedItemID.Valid {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "该槽位没有装备")
	}

	// 6. 卸下装备
	unequippedItem, err := s.unequipItemInternal(ctx, tx, slot.EquippedItemID.String)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "卸下装备失败")
	}

	// 7. 更新槽位装备ID为NULL
	slot.EquippedItemID = null.String{}
	if err := s.equipmentSlotRepo.UpdateSlot(ctx, tx, slot); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新槽位失败")
	}

	// 8. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 9. TODO: 失效英雄属性缓存

	// 9. 转换为DTO
	return &UnequipItemResponse{
		Success:        true,
		Message:        "装备卸下成功",
		UnequippedItem: convertItemToDTO(unequippedItem),
	}, nil
}

// GetEquippedItems 查询已装备物品
func (s *EquipmentService) GetEquippedItems(ctx context.Context, heroID string) ([]*PlayerItemDTO, error) {
	// 1. 验证参数
	if heroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}

	// 2. 获取英雄信息
	hero, err := s.heroRepo.GetByID(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "英雄不存在")
	}

	// 3. 查询已装备物品
	items, err := s.playerItemRepo.GetEquippedItems(ctx, hero.UserID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询已装备物品失败")
	}

	// 4. 转换为DTO
	dtos := make([]*PlayerItemDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, convertItemToDTO(item))
	}

	return dtos, nil
}

// GetEquipmentSlots 查询装备槽位
func (s *EquipmentService) GetEquipmentSlots(ctx context.Context, heroID string) ([]*EquipmentSlotDTO, error) {
	// 1. 验证参数
	if heroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "英雄ID不能为空")
	}

	// 2. 查询槽位
	slots, err := s.equipmentSlotRepo.GetSlots(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询装备槽位失败")
	}

	// 3. 转换为DTO
	dtos := make([]*EquipmentSlotDTO, 0, len(slots))
	for _, slot := range slots {
		dtos = append(dtos, convertSlotToDTO(slot))
	}

	return dtos, nil
}

// GetEquipmentAttributeBonus 获取装备属性加成
func (s *EquipmentService) GetEquipmentAttributeBonus(ctx context.Context, heroID string) (map[string]*AttributeBonus, error) {
	// 1. 查询已装备物品
	equippedItems, err := s.GetEquippedItems(ctx, heroID)
	if err != nil {
		return nil, err
	}

	if len(equippedItems) == 0 {
		return make(map[string]*AttributeBonus), nil
	}

	// 2. 获取所有装备配置
	itemIDs := make([]string, len(equippedItems))
	for i, item := range equippedItems {
		itemIDs[i] = item.ItemID
	}

	itemConfigs, err := s.itemRepo.GetByIDs(ctx, itemIDs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询装备配置失败")
	}

	// 创建配置映射
	configMap := make(map[string]*game_config.Item)
	for _, config := range itemConfigs {
		configMap[config.ID] = config
	}

	// 3. 计算每个装备的属性加成
	bonusesList := make([]map[string]*AttributeBonus, 0, len(equippedItems))

	for _, item := range equippedItems {
		config, exists := configMap[item.ItemID]
		if !exists {
			continue
		}

		// 获取最大耐久度
		maxDurability := 0
		if config.MaxDurability.Valid {
			maxDurability = int(config.MaxDurability.Int)
		}

		// 获取当前耐久度
		currentDurability := maxDurability
		if item.CurrentDurability != nil {
			currentDurability = *item.CurrentDurability
		}

		// 获取强化等级
		enhancementLevel := 0
		if item.EnhancementLevel != nil {
			enhancementLevel = *item.EnhancementLevel
		}

		// 计算属性加成
		bonuses, err := s.equipmentEffectSvc.CalculateAttributeBonuses(
			config.OutOfCombatEffects,
			enhancementLevel,
			currentDurability,
			maxDurability,
		)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "计算装备属性加成失败")
		}

		bonusesList = append(bonusesList, bonuses)
	}

	// 4. 合并所有装备的属性加成
	mergedBonuses := s.equipmentEffectSvc.MergeAttributeBonuses(bonusesList)

	// 5. 计算套装属性加成并合并
	if s.equipmentSetSvc != nil {
		setBonuses, err := s.equipmentSetSvc.CalculateSetBonuses(ctx, heroID)
		if err != nil {
			// 套装加成计算失败不影响装备加成，记录日志后继续
			// TODO: 添加日志记录
			return mergedBonuses, nil
		}

		// 合并套装加成到装备加成
		for attrID, setBonus := range setBonuses {
			if existing, ok := mergedBonuses[attrID]; ok {
				existing.FlatBonus += setBonus.FlatBonus
				existing.PercentBonus += setBonus.PercentBonus
			} else {
				mergedBonuses[attrID] = setBonus
			}
		}
	}

	return mergedBonuses, nil
}

// ==================== Helper Functions ====================

// convertSlotToDTO 转换槽位Entity为DTO
func convertSlotToDTO(slot *game_runtime.HeroEquipmentSlot) *EquipmentSlotDTO {
	if slot == nil {
		return nil
	}

	dto := &EquipmentSlotDTO{
		ID:         slot.ID,
		HeroID:     slot.HeroID,
		SlotType:   slot.SlotType,
		SlotIndex:  int(slot.SlotIndex),
		IsUnlocked: slot.IsUnlocked,
	}

	if slot.UnlockLevel.Valid {
		level := int(slot.UnlockLevel.Int16)
		dto.UnlockLevel = &level
	}

	if slot.EquippedItemID.Valid {
		dto.EquippedItemID = &slot.EquippedItemID.String
	}

	if slot.AddedByItemID.Valid {
		dto.AddedByItemID = &slot.AddedByItemID.String
	}

	return dto
}

// convertItemToDTO 转换物品Entity为DTO
func convertItemToDTO(item *game_runtime.PlayerItem) *PlayerItemDTO {
	if item == nil {
		return nil
	}

	dto := &PlayerItemDTO{
		ID:           item.ID,
		ItemID:       item.ItemID,
		OwnerID:      item.OwnerID,
		ItemLocation: item.ItemLocation,
	}

	if item.CurrentDurability.Valid {
		durability := item.CurrentDurability.Int
		dto.CurrentDurability = &durability
	}

	if item.MaxDurabilityOverride.Valid {
		maxDurability := item.MaxDurabilityOverride.Int
		dto.MaxDurability = &maxDurability
	}

	if item.EnhancementLevel.Valid {
		level := int(item.EnhancementLevel.Int16)
		dto.EnhancementLevel = &level
	}

	if item.StackCount.Valid {
		count := item.StackCount.Int
		dto.StackCount = &count
	}

	return dto
}

