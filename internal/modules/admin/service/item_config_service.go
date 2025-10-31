package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ItemConfigService 物品配置管理服务
type ItemConfigService struct {
	db              *sql.DB
	itemRepo        interfaces.ItemRepository
	tagRelationRepo interfaces.TagRelationRepository
	tagRepo         interfaces.TagRepository
}

// NewItemConfigService 创建物品配置管理服务
func NewItemConfigService(db *sql.DB) *ItemConfigService {
	return &ItemConfigService{
		db:              db,
		itemRepo:        impl.NewItemRepository(db),
		tagRelationRepo: impl.NewTagRelationRepository(db),
		tagRepo:         impl.NewTagRepository(db),
	}
}

// CreateItem 创建物品配置
func (s *ItemConfigService) CreateItem(ctx context.Context, req *dto.CreateItemRequest) (*dto.ItemConfigResponse, error) {
	// 1. 验证item_code唯一性
	exists, err := s.itemRepo.CheckCodeExists(ctx, req.ItemCode, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查物品代码是否存在失败")
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("物品代码已存在: %s", req.ItemCode))
	}

	// 2. 验证装备槽位与物品类型匹配
	if err := s.validateEquipSlot(req.ItemType, req.EquipSlot); err != nil {
		return nil, err
	}

	// 3. 验证标签存在性
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			_, err := s.tagRepo.GetByID(ctx, tagID)
			if err != nil {
				return nil, xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("标签不存在: %s", tagID))
			}
		}
	}

	// 3. 验证JSON格式
	if err := s.validateJSONFields(req.OutOfCombatEffects, req.InCombatEffects, req.UseEffects); err != nil {
		return nil, err
	}

	// 4. 创建物品实体
	item := &game_config.Item{
		ID:          uuid.New().String(),
		ItemCode:    req.ItemCode,
		ItemName:    req.ItemName,
		ItemType:    req.ItemType,
		ItemQuality: req.ItemQuality,
		ItemLevel:   req.ItemLevel,
		IsActive:    true, // 默认启用
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 设置bool字段
	if req.IsTradable != nil {
		item.IsTradable.SetValid(*req.IsTradable)
	}
	if req.IsDroppable != nil {
		item.IsDroppable.SetValid(*req.IsDroppable)
	}

	// 设置可选的字符串字段
	if req.Description != "" {
		item.Description.SetValid(req.Description)
	}
	if req.IconURL != "" {
		item.IconURL.SetValid(req.IconURL)
	}
	if req.EquipSlot != nil {
		item.EquipSlot.SetValid(*req.EquipSlot)
	}
	if req.RequiredLevel != nil {
		item.RequiredLevel.SetValid(*req.RequiredLevel)
	}
	if req.MaterialType != nil {
		item.MaterialType.SetValid(*req.MaterialType)
	}
	if req.MaxDurability != nil {
		item.MaxDurability.SetValid(*req.MaxDurability)
	}
	if req.UniquenessType != nil {
		item.UniquenessType.SetValid(*req.UniquenessType)
	}
	if len(req.OutOfCombatEffects) > 0 {
		item.OutOfCombatEffects.SetValid(req.OutOfCombatEffects)
	}
	if len(req.InCombatEffects) > 0 {
		item.InCombatEffects.SetValid(req.InCombatEffects)
	}
	if len(req.UseEffects) > 0 {
		item.UseEffects.SetValid(req.UseEffects)
	}
	if len(req.ProvidedSkills) > 0 {
		var skills []string
		if err := json.Unmarshal(req.ProvidedSkills, &skills); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInvalidParams, "解析 provided_skills 失败")
		}
		item.ProvidedSkills = skills
	}
	if req.SocketType != nil {
		item.SocketType.SetValid(*req.SocketType)
	}
	if req.SocketCount != nil {
		item.SocketCount.SetValid(*req.SocketCount)
	}
	if req.EnhancementMaterialID != nil {
		item.EnhancementMaterialID.SetValid(*req.EnhancementMaterialID)
	}
	if req.EnhancementCostGold != nil {
		item.EnhancementCostGold.SetValid(*req.EnhancementCostGold)
	}
	if req.GemColor != nil {
		item.GemColor.SetValid(*req.GemColor)
	}
	if req.GemSize != nil {
		item.GemSize.SetValid(*req.GemSize)
	}
	if req.RepairDurabilityAmount != nil {
		item.RepairDurabilityAmount.SetValid(*req.RepairDurabilityAmount)
	}
	if req.RepairApplicableQuality != nil {
		item.RepairApplicableQuality = []string{*req.RepairApplicableQuality}
	}
	if req.RepairMaterialType != nil {
		item.RepairMaterialType.SetValid(*req.RepairMaterialType)
	}
	if req.MaxStackSize != nil {
		item.MaxStackSize.SetValid(int(*req.MaxStackSize))
	}
	if req.BaseValue != nil {
		item.BaseValue.SetValid(*req.BaseValue)
	}

	// 5. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 6. 创建物品
	if err := item.Insert(ctx, tx, boil.Infer()); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建物品配置失败")
	}

	// 7. 创建职业关联
	if len(req.RequiredClassIDs) > 0 {
		for _, classID := range req.RequiredClassIDs {
			// 验证职业是否存在
			if err := s.validateClassExists(ctx, classID); err != nil {
				return nil, err
			}

			relation := &game_config.ItemClassRelation{
				ID:        uuid.New().String(),
				ItemID:    item.ID,
				ClassID:   classID,
				CreatedAt: time.Now(),
			}
			if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
				return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业关联失败")
			}
		}
	}

	// 8. 创建标签关联
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			relation := &game_config.TagsRelation{
				ID:         uuid.New().String(),
				TagID:      tagID,
				EntityType: "item",
				EntityID:   item.ID,
				CreatedAt:  time.Now(),
			}
			if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
				return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建标签关联失败")
			}
		}
	}

	// 10. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 11. 查询并返回完整数据（包含标签和职业）
	return s.GetItemByID(ctx, item.ID)
}

// validateJSONFields 验证JSON字段格式
func (s *ItemConfigService) validateJSONFields(fields ...json.RawMessage) error {
	for _, field := range fields {
		if len(field) > 0 {
			var temp interface{}
			if err := json.Unmarshal(field, &temp); err != nil {
				return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("无效的JSON格式: %v", err))
			}
		}
	}
	return nil
}

// GetItemByID 根据ID获取物品配置
func (s *ItemConfigService) GetItemByID(ctx context.Context, itemID string) (*dto.ItemConfigResponse, error) {
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 查询物品的标签
	tags, err := s.getItemTags(ctx, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品标签失败")
	}

	// 查询物品的职业限制
	classIDs, err := s.getItemClassIDs(ctx, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品职业限制失败")
	}

	// 查询套装信息
	setInfo, err := s.getItemSetInfo(ctx, item)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询套装信息失败")
	}

	return s.toItemConfigResponseWithSet(item, tags, classIDs, setInfo), nil
}

// GetItems 查询物品配置列表
func (s *ItemConfigService) GetItems(ctx context.Context, params interfaces.ListItemParams) ([]*dto.ItemConfigResponse, int64, error) {
	items, total, err := s.itemRepo.List(ctx, params)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品配置列表失败")
	}

	// 批量查询物品的标签和职业限制
	responses := make([]*dto.ItemConfigResponse, 0, len(items))
	for _, item := range items {
		tags, err := s.getItemTags(ctx, item.ID)
		if err != nil {
			return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品标签失败")
		}

		classIDs, err := s.getItemClassIDs(ctx, item.ID)
		if err != nil {
			return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品职业限制失败")
		}

		setInfo, err := s.getItemSetInfo(ctx, item)
		if err != nil {
			return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询套装信息失败")
		}

		responses = append(responses, s.toItemConfigResponseWithSet(item, tags, classIDs, setInfo))
	}

	return responses, total, nil
}

// UpdateItem 更新物品配置
func (s *ItemConfigService) UpdateItem(ctx context.Context, itemID string, req *dto.UpdateItemRequest) (*dto.ItemConfigResponse, error) {
	// 1. 查询物品是否存在
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 验证item_code唯一性（如果要更新）
	if req.ItemCode != nil {
		exists, err := s.itemRepo.CheckCodeExists(ctx, *req.ItemCode, &itemID)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查物品代码是否存在失败")
		}
		if exists {
			return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("物品代码已存在: %s", *req.ItemCode))
		}
	}

	// 3. 验证装备槽位与物品类型匹配（如果要更新）
	itemType := item.ItemType
	if req.ItemType != nil {
		itemType = *req.ItemType
	}
	equipSlot := item.EquipSlot.Ptr()
	if req.EquipSlot != nil {
		equipSlot = req.EquipSlot
	}
	if err := s.validateEquipSlot(itemType, equipSlot); err != nil {
		return nil, err
	}

	// 4. 验证JSON格式
	if err := s.validateJSONFields(req.OutOfCombatEffects, req.InCombatEffects, req.UseEffects, req.ProvidedSkills); err != nil {
		return nil, err
	}

	// 5. 更新字段
	if req.ItemCode != nil {
		item.ItemCode = *req.ItemCode
	}
	if req.ItemName != nil {
		item.ItemName = *req.ItemName
	}
	if req.ItemType != nil {
		item.ItemType = *req.ItemType
	}
	if req.ItemQuality != nil {
		item.ItemQuality = *req.ItemQuality
	}
	if req.ItemLevel != nil {
		item.ItemLevel = *req.ItemLevel
	}
	if req.Description != nil {
		item.Description.SetValid(*req.Description)
	}
	if req.IconURL != nil {
		item.IconURL.SetValid(*req.IconURL)
	}
	if req.EquipSlot != nil {
		item.EquipSlot.SetValid(*req.EquipSlot)
	}
	if req.RequiredLevel != nil {
		item.RequiredLevel.SetValid(*req.RequiredLevel)
	}
	if req.MaterialType != nil {
		item.MaterialType.SetValid(*req.MaterialType)
	}
	if req.MaxDurability != nil {
		item.MaxDurability.SetValid(*req.MaxDurability)
	}
	if req.UniquenessType != nil {
		item.UniquenessType.SetValid(*req.UniquenessType)
	}
	if len(req.OutOfCombatEffects) > 0 {
		item.OutOfCombatEffects.SetValid(req.OutOfCombatEffects)
	}
	if len(req.InCombatEffects) > 0 {
		item.InCombatEffects.SetValid(req.InCombatEffects)
	}
	if len(req.UseEffects) > 0 {
		item.UseEffects.SetValid(req.UseEffects)
	}
	if len(req.ProvidedSkills) > 0 {
		var skills []string
		if err := json.Unmarshal(req.ProvidedSkills, &skills); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInvalidParams, "解析 provided_skills 失败")
		}
		item.ProvidedSkills = skills
	}
	if req.SocketType != nil {
		item.SocketType.SetValid(*req.SocketType)
	}
	if req.SocketCount != nil {
		item.SocketCount.SetValid(*req.SocketCount)
	}
	if req.EnhancementMaterialID != nil {
		item.EnhancementMaterialID.SetValid(*req.EnhancementMaterialID)
	}
	if req.EnhancementCostGold != nil {
		item.EnhancementCostGold.SetValid(*req.EnhancementCostGold)
	}
	if req.GemColor != nil {
		item.GemColor.SetValid(*req.GemColor)
	}
	if req.GemSize != nil {
		item.GemSize.SetValid(*req.GemSize)
	}
	if req.RepairDurabilityAmount != nil {
		item.RepairDurabilityAmount.SetValid(*req.RepairDurabilityAmount)
	}
	if req.RepairApplicableQuality != nil {
		item.RepairApplicableQuality = []string{*req.RepairApplicableQuality}
	}
	if req.RepairMaterialType != nil {
		item.RepairMaterialType.SetValid(*req.RepairMaterialType)
	}
	if req.MaxStackSize != nil {
		item.MaxStackSize.SetValid(int(*req.MaxStackSize))
	}
	if req.BaseValue != nil {
		item.BaseValue.SetValid(*req.BaseValue)
	}
	if req.IsTradable != nil {
		item.IsTradable.SetValid(*req.IsTradable)
	}
	if req.IsDroppable != nil {
		item.IsDroppable.SetValid(*req.IsDroppable)
	}

	// 套装关联
	if req.SetID != nil {
		// 验证套装分配
		if err := s.validateSetAssignment(ctx, itemID, *req.SetID); err != nil {
			return nil, err
		}
		item.SetID.SetValid(*req.SetID)
	}

	item.UpdatedAt = time.Now()

	// 6. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 7. 更新物品
	_, err = item.Update(ctx, tx, boil.Infer())
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新物品配置失败")
	}

	// 8. 如果指定了职业限制更新
	if req.RequiredClassIDs != nil {
		// 删除所有现有关联
		if err := s.deleteItemClassRelations(ctx, tx, itemID); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "删除职业关联失败")
		}

		// 创建新的关联
		for _, classID := range *req.RequiredClassIDs {
			// 验证职业是否存在
			if err := s.validateClassExists(ctx, classID); err != nil {
				return nil, err
			}

			relation := &game_config.ItemClassRelation{
				ID:        uuid.New().String(),
				ItemID:    itemID,
				ClassID:   classID,
				CreatedAt: time.Now(),
			}
			if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
				return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业关联失败")
			}
		}
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 10. 查询并返回完整数据
	return s.GetItemByID(ctx, itemID)
}

// DeleteItem 删除物品配置(软删除)
func (s *ItemConfigService) DeleteItem(ctx context.Context, itemID string) error {
	// 1. 查询物品是否存在
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 3. 删除标签关联
	if err := s.tagRelationRepo.DeleteByEntity(ctx, "item", itemID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除标签关联失败")
	}

	// 4. 软删除物品
	if err := s.itemRepo.Delete(ctx, itemID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除物品配置失败")
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// AddItemTags 为物品添加标签
func (s *ItemConfigService) AddItemTags(ctx context.Context, itemID string, tagIDs []string) error {
	// 1. 验证物品存在性
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 验证标签存在性
	for _, tagID := range tagIDs {
		_, err := s.tagRepo.GetByID(ctx, tagID)
		if err != nil {
			return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("标签不存在: %s", tagID))
		}
	}

	// 3. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 4. 创建标签关联
	for _, tagID := range tagIDs {
		// 检查是否已存在
		exists, err := s.tagRelationRepo.Exists(ctx, tagID, "item", itemID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "检查标签关联是否存在失败")
		}
		if exists {
			continue // 已存在则跳过
		}

		relation := &game_config.TagsRelation{
			ID:         uuid.New().String(),
			TagID:      tagID,
			EntityType: "item",
			EntityID:   itemID,
			CreatedAt:  time.Now(),
		}
		if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "创建标签关联失败")
		}
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// GetItemTags 查询物品的所有标签
func (s *ItemConfigService) GetItemTags(ctx context.Context, itemID string) ([]dto.TagResponse, error) {
	// 验证物品存在性
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	return s.getItemTags(ctx, itemID)
}

// UpdateItemTags 批量更新物品标签(替换所有标签)
func (s *ItemConfigService) UpdateItemTags(ctx context.Context, itemID string, tagIDs []string) error {
	// 1. 验证物品存在性
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 验证标签存在性
	for _, tagID := range tagIDs {
		_, err := s.tagRepo.GetByID(ctx, tagID)
		if err != nil {
			return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("标签不存在: %s", tagID))
		}
	}

	// 3. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 4. 删除现有标签关联
	if err := s.tagRelationRepo.DeleteByEntity(ctx, "item", itemID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除标签关联失败")
	}

	// 5. 创建新的标签关联
	for _, tagID := range tagIDs {
		relation := &game_config.TagsRelation{
			ID:         uuid.New().String(),
			TagID:      tagID,
			EntityType: "item",
			EntityID:   itemID,
			CreatedAt:  time.Now(),
		}
		if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "创建标签关联失败")
		}
	}

	// 6. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// RemoveItemTag 移除物品标签
func (s *ItemConfigService) RemoveItemTag(ctx context.Context, itemID string, tagID string) error {
	// 1. 验证物品存在性
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 删除标签关联
	if err := s.tagRelationRepo.DeleteByTagAndEntity(ctx, tagID, "item", itemID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除标签关联失败")
	}

	return nil
}

// getItemTags 查询物品的标签（内部方法）
func (s *ItemConfigService) getItemTags(ctx context.Context, itemID string) ([]dto.TagResponse, error) {
	entityType := "item"
	params := interfaces.TagRelationQueryParams{
		EntityType: &entityType,
		EntityID:   &itemID,
	}

	relations, _, err := s.tagRelationRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	tags := make([]dto.TagResponse, 0, len(relations))
	for _, relation := range relations {
		tag, err := s.tagRepo.GetByID(ctx, relation.TagID)
		if err != nil {
			continue // 忽略不存在的标签
		}
		tags = append(tags, dto.TagResponse{
			ID:           tag.ID,
			TagCode:      tag.TagCode,
			TagName:      tag.TagName,
			Category:     tag.Category,
			Description:  tag.Description.String,
			Icon:         tag.Icon.String,
			Color:        tag.Color.String,
			DisplayOrder: tag.DisplayOrder,
			IsActive:     tag.IsActive,
			CreatedAt:    tag.CreatedAt,
			UpdatedAt:    tag.UpdatedAt,
		})
	}

	return tags, nil
}

// toItemConfigResponse 转换为ItemConfigResponse
func (s *ItemConfigService) toItemConfigResponse(item *game_config.Item, tags []dto.TagResponse, classIDs []string) *dto.ItemConfigResponse {
	resp := &dto.ItemConfigResponse{
		ID:               item.ID,
		ItemCode:         item.ItemCode,
		ItemName:         item.ItemName,
		ItemType:         item.ItemType,
		ItemQuality:      item.ItemQuality,
		ItemLevel:        item.ItemLevel,
		IsTradable:       item.IsTradable.Bool,
		IsDroppable:      item.IsDroppable.Bool,
		IsActive:         item.IsActive,
		Tags:             tags,
		RequiredClassIDs: classIDs, // 职业限制列表
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}

	// 设置可选的字符串字段
	if item.Description.Valid {
		resp.Description = item.Description.String
	}
	if item.IconURL.Valid {
		resp.IconURL = item.IconURL.String
	}

	if item.EquipSlot.Valid {
		resp.EquipSlot = &item.EquipSlot.String
	}
	if item.RequiredLevel.Valid {
		resp.RequiredLevel = &item.RequiredLevel.Int16
	}
	if item.MaterialType.Valid {
		resp.MaterialType = &item.MaterialType.String
	}
	if item.MaxDurability.Valid {
		resp.MaxDurability = &item.MaxDurability.Int
	}
	if item.UniquenessType.Valid {
		resp.UniquenessType = &item.UniquenessType.String
	}
	if item.OutOfCombatEffects.Valid {
		resp.OutOfCombatEffects = item.OutOfCombatEffects.JSON
	}
	if item.InCombatEffects.Valid {
		resp.InCombatEffects = item.InCombatEffects.JSON
	}
	if item.UseEffects.Valid {
		resp.UseEffects = item.UseEffects.JSON
	}
	// ProvidedSkills是types.StringArray，直接赋值
	if len(item.ProvidedSkills) > 0 {
		skillsJSON, _ := json.Marshal(item.ProvidedSkills)
		resp.ProvidedSkills = skillsJSON
	}
	if item.SocketType.Valid {
		resp.SocketType = &item.SocketType.String
	}
	if item.SocketCount.Valid {
		resp.SocketCount = &item.SocketCount.Int16
	}
	if item.EnhancementMaterialID.Valid {
		resp.EnhancementMaterialID = &item.EnhancementMaterialID.String
	}
	if item.EnhancementCostGold.Valid {
		resp.EnhancementCostGold = &item.EnhancementCostGold.Int
	}
	if item.GemColor.Valid {
		resp.GemColor = &item.GemColor.String
	}
	if item.GemSize.Valid {
		resp.GemSize = &item.GemSize.String
	}
	if item.RepairDurabilityAmount.Valid {
		resp.RepairDurabilityAmount = &item.RepairDurabilityAmount.Int
	}
	// RepairApplicableQuality是types.StringArray，需要特殊处理
	if len(item.RepairApplicableQuality) > 0 {
		qualityStr := item.RepairApplicableQuality[0] // 取第一个值
		resp.RepairApplicableQuality = &qualityStr
	}
	if item.RepairMaterialType.Valid {
		resp.RepairMaterialType = &item.RepairMaterialType.String
	}
	if item.MaxStackSize.Valid {
		maxStack := int16(item.MaxStackSize.Int)
		resp.MaxStackSize = &maxStack
	}
	if item.BaseValue.Valid {
		resp.BaseValue = &item.BaseValue.Int
	}

	return resp
}

// SetInfo 套装信息
type SetInfo struct {
	SetID   *string
	SetName *string
	SetCode *string
}

// getItemSetInfo 获取物品的套装信息
func (s *ItemConfigService) getItemSetInfo(ctx context.Context, item *game_config.Item) (*SetInfo, error) {
	setInfo := &SetInfo{}

	// 如果物品没有关联套装，返回空信息
	if !item.SetID.Valid {
		return setInfo, nil
	}

	// 查询套装信息
	setConfig, err := game_config.EquipmentSetConfigs(
		game_config.EquipmentSetConfigWhere.ID.EQ(item.SetID.String),
		game_config.EquipmentSetConfigWhere.DeletedAt.IsNull(),
	).One(ctx, s.db)
	if err != nil {
		if err == sql.ErrNoRows {
			// 套装不存在或已删除，返回空信息
			return setInfo, nil
		}
		return nil, err
	}

	setInfo.SetID = &setConfig.ID
	setInfo.SetName = &setConfig.SetName
	setInfo.SetCode = &setConfig.SetCode

	return setInfo, nil
}

// toItemConfigResponseWithSet 转换为物品配置响应（包含套装信息）
func (s *ItemConfigService) toItemConfigResponseWithSet(item *game_config.Item, tags []dto.TagResponse, classIDs []string, setInfo *SetInfo) *dto.ItemConfigResponse {
	resp := s.toItemConfigResponse(item, tags, classIDs)

	// 添加套装信息
	if setInfo != nil {
		resp.SetID = setInfo.SetID
		resp.SetName = setInfo.SetName
		resp.SetCode = setInfo.SetCode
	}

	return resp
}

// validateEquipSlot 验证装备槽位与物品类型匹配
func (s *ItemConfigService) validateEquipSlot(itemType string, equipSlot *string) error {
	// 如果不是装备类型，不能设置装备槽位
	if itemType != "equipment" {
		if equipSlot != nil && *equipSlot != "" {
			return xerrors.New(xerrors.CodeInvalidParams, "非装备类型物品不能设置装备槽位")
		}
		return nil
	}

	// 装备类型必须设置装备槽位
	if equipSlot == nil || *equipSlot == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "装备类型物品必须设置装备槽位")
	}

	// 验证槽位值是否有效（匹配数据库的slot_type_enum）
	validSlots := map[string]bool{
		"head":         true,
		"eyes":         true,
		"ears":         true,
		"neck":         true,
		"cloak":        true,
		"chest":        true,
		"belt":         true,
		"shoulder":     true,
		"wrist":        true,
		"gloves":       true,
		"legs":         true,
		"feet":         true,
		"ring":         true,
		"badge":        true,
		"coat":         true,
		"pocket":       true,
		"summon_mount": true,
		"mainhand":     true,
		"offhand":      true,
		"twohand":      true,
		"special":      true,
	}

	if !validSlots[*equipSlot] {
		return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("无效的装备槽位: %s", *equipSlot))
	}

	return nil
}

// getItemClassIDs 获取物品的职业限制列表
func (s *ItemConfigService) getItemClassIDs(ctx context.Context, itemID string) ([]string, error) {
	relations, err := game_config.ItemClassRelations(
		game_config.ItemClassRelationWhere.ItemID.EQ(itemID),
	).All(ctx, s.db)
	if err != nil {
		return nil, err
	}

	classIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		classIDs = append(classIDs, relation.ClassID)
	}

	return classIDs, nil
}

// validateClassExists 验证职业是否存在
func (s *ItemConfigService) validateClassExists(ctx context.Context, classID string) error {
	exists, err := game_config.Classes(
		game_config.ClassWhere.ID.EQ(classID),
	).Exists(ctx, s.db)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业失败")
	}
	if !exists {
		return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("职业不存在: %s", classID))
	}
	return nil
}

// validateSetAssignment 验证套装分配
func (s *ItemConfigService) validateSetAssignment(ctx context.Context, itemID, setID string) error {
	// 如果setID为空字符串，表示移除套装关联，直接返回
	if setID == "" {
		return nil
	}

	// 1. 验证套装是否存在
	setExists, err := game_config.EquipmentSetConfigs(
		game_config.EquipmentSetConfigWhere.ID.EQ(setID),
		game_config.EquipmentSetConfigWhere.DeletedAt.IsNull(),
	).Exists(ctx, s.db)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询套装失败")
	}
	if !setExists {
		return xerrors.New(xerrors.CodeResourceNotFound, "套装不存在或已删除")
	}

	// 2. 验证物品类型是否为装备
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品失败")
	}
	if item.ItemType != "equipment" {
		return xerrors.New(xerrors.CodeInvalidParams, "只有装备类型的物品可以分配到套装")
	}

	return nil
}

// deleteItemClassRelations 删除物品的所有职业关联
func (s *ItemConfigService) deleteItemClassRelations(ctx context.Context, exec boil.ContextExecutor, itemID string) error {
	_, err := game_config.ItemClassRelations(
		game_config.ItemClassRelationWhere.ItemID.EQ(itemID),
	).DeleteAll(ctx, exec)
	return err
}

// AddItemClasses 为物品添加职业限制
func (s *ItemConfigService) AddItemClasses(ctx context.Context, itemID string, classIDs []string) error {
	// 1. 验证物品是否存在
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 3. 添加职业关联
	for _, classID := range classIDs {
		// 验证职业是否存在
		if err := s.validateClassExists(ctx, classID); err != nil {
			return err
		}

		// 检查是否已存在关联
		exists, err := game_config.ItemClassRelations(
			game_config.ItemClassRelationWhere.ItemID.EQ(itemID),
			game_config.ItemClassRelationWhere.ClassID.EQ(classID),
		).Exists(ctx, tx)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业关联失败")
		}
		if exists {
			continue // 已存在，跳过
		}

		// 创建关联
		relation := &game_config.ItemClassRelation{
			ID:        uuid.New().String(),
			ItemID:    itemID,
			ClassID:   classID,
			CreatedAt: time.Now(),
		}
		if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业关联失败")
		}
	}

	// 4. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// GetItemClasses 查询物品的职业限制
func (s *ItemConfigService) GetItemClasses(ctx context.Context, itemID string) ([]string, error) {
	// 1. 验证物品是否存在
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 查询职业关联
	return s.getItemClassIDs(ctx, itemID)
}

// UpdateItemClasses 批量更新物品职业限制
func (s *ItemConfigService) UpdateItemClasses(ctx context.Context, itemID string, classIDs []string) error {
	// 1. 验证物品是否存在
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 3. 删除所有现有关联
	if err := s.deleteItemClassRelations(ctx, tx, itemID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除职业关联失败")
	}

	// 4. 创建新的关联
	for _, classID := range classIDs {
		// 验证职业是否存在
		if err := s.validateClassExists(ctx, classID); err != nil {
			return err
		}

		relation := &game_config.ItemClassRelation{
			ID:        uuid.New().String(),
			ItemID:    itemID,
			ClassID:   classID,
			CreatedAt: time.Now(),
		}
		if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业关联失败")
		}
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// RemoveItemClass 移除物品职业限制
func (s *ItemConfigService) RemoveItemClass(ctx context.Context, itemID, classID string) error {
	// 1. 验证物品是否存在
	_, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品配置失败")
	}

	// 2. 删除关联
	_, err = game_config.ItemClassRelations(
		game_config.ItemClassRelationWhere.ItemID.EQ(itemID),
		game_config.ItemClassRelationWhere.ClassID.EQ(classID),
	).DeleteAll(ctx, s.db)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除职业关联失败")
	}

	return nil
}
