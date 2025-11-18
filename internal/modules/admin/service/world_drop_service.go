package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"

	"github.com/lib/pq"
)

const dropRateEpsilon = 1e-6

// WorldDropService 世界掉落服务
type WorldDropService struct {
	worldDropRepo     interfaces.WorldDropConfigRepository
	worldDropItemRepo interfaces.WorldDropItemRepository
	itemRepo          interfaces.ItemRepository
}

// NewWorldDropService 创建世界掉落服务
func NewWorldDropService(db *sql.DB) *WorldDropService {
	return &WorldDropService{
		worldDropRepo:     impl.NewWorldDropConfigRepository(db),
		worldDropItemRepo: impl.NewWorldDropItemRepository(db),
		itemRepo:          impl.NewItemRepository(db),
	}
}

// CreateWorldDrop 创建世界掉落配置
func (s *WorldDropService) CreateWorldDrop(ctx context.Context, req *dto.CreateWorldDropRequest) (*dto.WorldDropResponse, error) {
	// 1. 验证物品ID存在性
	item, err := s.itemRepo.GetByID(ctx, req.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品不存在")
	}

	// 2. 验证基础掉落概率范围
	if req.BaseDropRate <= 0 || req.BaseDropRate > 1 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "基础掉落概率必须在(0, 1]范围内")
	}

	// 3. 检查物品是否已被其他世界掉落配置使用
	inUse, err := s.worldDropItemRepo.ExistsActiveItem(ctx, req.ItemID, "")
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查物品占用失败")
	}
	if inUse {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("物品已绑定到其他世界掉落: %s", req.ItemID))
	}

	// 4. 验证触发条件JSON格式
	if len(req.TriggerConditions) > 0 {
		var conditions map[string]interface{}
		if err := json.Unmarshal(req.TriggerConditions, &conditions); err != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "触发条件JSON格式错误")
		}
	}

	// 5. 验证概率修正因子JSON格式
	if len(req.DropRateModifiers) > 0 {
		var modifiers map[string]interface{}
		if err := json.Unmarshal(req.DropRateModifiers, &modifiers); err != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "概率修正因子JSON格式错误")
		}
	}

	// 6. 验证掉落间隔范围
	if req.MinDropInterval != nil && req.MaxDropInterval != nil && *req.MinDropInterval > *req.MaxDropInterval {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小掉落间隔不能大于最大掉落间隔")
	}

	// 7. 创建世界掉落配置实体
	config := &game_config.WorldDropConfig{
		ID:        uuid.New().String(),
		ItemID:    req.ItemID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置基础掉落概率
	dec := new(decimal.Big).SetFloat64(req.BaseDropRate)
	config.BaseDropRate = types.NewDecimal(dec)

	// 设置限制
	if req.TotalDropLimit != nil {
		config.TotalDropLimit.SetValid(*req.TotalDropLimit)
	}
	if req.DailyDropLimit != nil {
		config.DailyDropLimit.SetValid(*req.DailyDropLimit)
	}
	if req.HourlyDropLimit != nil {
		config.HourlyDropLimit.SetValid(*req.HourlyDropLimit)
	}
	if req.MinDropInterval != nil {
		config.MinDropInterval.SetValid(*req.MinDropInterval)
	}
	if req.MaxDropInterval != nil {
		config.MaxDropInterval.SetValid(*req.MaxDropInterval)
	}

	// 设置JSON字段
	if len(req.TriggerConditions) > 0 {
		config.TriggerConditions.SetValid(req.TriggerConditions)
	}
	if len(req.DropRateModifiers) > 0 {
		config.DropRateModifiers.SetValid(req.DropRateModifiers)
	}

	// 7. 保存到数据库
	if err := s.worldDropRepo.Create(ctx, config); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建世界掉落配置失败")
	}

	// 8. 创建默认掉落物品记录
	dropRate := req.BaseDropRate
	itemEntry := interfaces.WorldDropItem{
		WorldDropConfigID: config.ID,
		ItemID:            req.ItemID,
		DropRate:          &dropRate,
		MinQuantity:       1,
		MaxQuantity:       1,
		GuaranteedDrop:    false,
	}
	if err := s.worldDropItemRepo.Create(ctx, &itemEntry); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建世界掉落物品失败")
	}

	return s.toWorldDropResponse(config, item), nil
}

// GetWorldDropByID 获取世界掉落详情
func (s *WorldDropService) GetWorldDropByID(ctx context.Context, configID string) (*dto.WorldDropResponse, error) {
	config, err := s.worldDropRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询世界掉落配置失败")
	}

	item, err := s.itemRepo.GetByID(ctx, config.ItemID)
	if err != nil {
		// 若主物品缺失,尝试回退到当前配置下的第一条物品记录
		if fallback, _, listErr := s.worldDropItemRepo.ListByConfig(ctx, interfaces.ListWorldDropItemParams{WorldDropConfigID: config.ID, Page: 1, PageSize: 1}); listErr == nil && len(fallback) > 0 {
			config.ItemID = fallback[0].ItemID
			config.UpdatedAt = time.Now()
			_ = s.worldDropRepo.Update(ctx, config)
			item, err = s.itemRepo.GetByID(ctx, config.ItemID)
		}
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品信息失败")
		}
	}

	return s.toWorldDropResponse(config, item), nil
}

// GetWorldDropList 查询世界掉落列表
func (s *WorldDropService) GetWorldDropList(ctx context.Context, params interfaces.ListWorldDropConfigParams) (*dto.WorldDropListResponse, error) {
	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// 查询世界掉落列表
	configs, total, err := s.worldDropRepo.List(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询世界掉落列表失败")
	}

	// 批量查询物品信息
	responses := make([]dto.WorldDropResponse, 0, len(configs))
	for _, config := range configs {
		item, err := s.itemRepo.GetByID(ctx, config.ItemID)
		if err != nil {
			continue // 跳过不存在的物品
		}
		responses = append(responses, *s.toWorldDropResponse(config, item))
	}

	return &dto.WorldDropListResponse{
		Items:    responses,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// UpdateWorldDrop 更新世界掉落配置
func (s *WorldDropService) UpdateWorldDrop(ctx context.Context, configID string, req *dto.UpdateWorldDropRequest) (*dto.WorldDropResponse, error) {
	// 1. 查询世界掉落配置是否存在
	config, err := s.worldDropRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询世界掉落配置失败")
	}

	// 2. 更新字段
	if req.TotalDropLimit != nil {
		config.TotalDropLimit.SetValid(*req.TotalDropLimit)
	}
	if req.DailyDropLimit != nil {
		config.DailyDropLimit.SetValid(*req.DailyDropLimit)
	}
	if req.HourlyDropLimit != nil {
		config.HourlyDropLimit.SetValid(*req.HourlyDropLimit)
	}
	if req.MinDropInterval != nil {
		config.MinDropInterval.SetValid(*req.MinDropInterval)
	}
	if req.MaxDropInterval != nil {
		config.MaxDropInterval.SetValid(*req.MaxDropInterval)
	}
	if req.BaseDropRate != nil {
		// 验证概率范围
		if *req.BaseDropRate <= 0 || *req.BaseDropRate > 1 {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "基础掉落概率必须在(0, 1]范围内")
		}
		dec := new(decimal.Big).SetFloat64(*req.BaseDropRate)
		config.BaseDropRate = types.NewDecimal(dec)
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	// 3. 验证并更新JSON字段
	if len(req.TriggerConditions) > 0 {
		var conditions map[string]interface{}
		if unmarshalErr := json.Unmarshal(req.TriggerConditions, &conditions); unmarshalErr != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "触发条件JSON格式错误")
		}
		config.TriggerConditions.SetValid(req.TriggerConditions)
	}
	if len(req.DropRateModifiers) > 0 {
		var modifiers map[string]interface{}
		if unmarshalErr := json.Unmarshal(req.DropRateModifiers, &modifiers); unmarshalErr != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "概率修正因子JSON格式错误")
		}
		config.DropRateModifiers.SetValid(req.DropRateModifiers)
	}

	// 4. 验证掉落间隔范围
	if config.MinDropInterval.Valid && config.MaxDropInterval.Valid && config.MinDropInterval.Int > config.MaxDropInterval.Int {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小掉落间隔不能大于最大掉落间隔")
	}

	config.UpdatedAt = time.Now()

	// 5. 保存更新
	if updateErr := s.worldDropRepo.Update(ctx, config); updateErr != nil {
		return nil, xerrors.Wrap(updateErr, xerrors.CodeInternalError, "更新世界掉落配置失败")
	}

	item, err := s.itemRepo.GetByID(ctx, config.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品信息失败")
	}

	return s.toWorldDropResponse(config, item), nil
}

// DeleteWorldDrop 删除世界掉落配置（软删除）
func (s *WorldDropService) DeleteWorldDrop(ctx context.Context, configID string) error {
	// 1. 查询世界掉落配置是否存在
	_, err := s.worldDropRepo.GetByID(ctx, configID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询世界掉落配置失败")
	}

	// 2. 删除世界掉落配置
	if err := s.worldDropRepo.Delete(ctx, configID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除世界掉落配置失败")
	}

	return nil
}

// toWorldDropResponse 转换为世界掉落响应
func (s *WorldDropService) toWorldDropResponse(config *game_config.WorldDropConfig, item *game_config.Item) *dto.WorldDropResponse {
	resp := &dto.WorldDropResponse{
		ID:        config.ID,
		ItemID:    config.ItemID,
		ItemCode:  item.ItemCode,
		ItemName:  item.ItemName,
		IsActive:  config.IsActive,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	// 基础掉落概率
	rate, _ := config.BaseDropRate.Float64()
	resp.BaseDropRate = rate

	// 限制
	if config.TotalDropLimit.Valid {
		limit := config.TotalDropLimit.Int
		resp.TotalDropLimit = &limit
	}
	if config.DailyDropLimit.Valid {
		limit := config.DailyDropLimit.Int
		resp.DailyDropLimit = &limit
	}
	if config.HourlyDropLimit.Valid {
		limit := config.HourlyDropLimit.Int
		resp.HourlyDropLimit = &limit
	}
	if config.MinDropInterval.Valid {
		interval := config.MinDropInterval.Int
		resp.MinDropInterval = &interval
	}
	if config.MaxDropInterval.Valid {
		interval := config.MaxDropInterval.Int
		resp.MaxDropInterval = &interval
	}

	// JSON字段
	if config.TriggerConditions.Valid {
		resp.TriggerConditions = config.TriggerConditions.JSON
	}
	if config.DropRateModifiers.Valid {
		resp.DropRateModifiers = config.DropRateModifiers.JSON
	}

	// 删除时间
	if config.DeletedAt.Valid {
		deletedAt := config.DeletedAt.Time
		resp.DeletedAt = &deletedAt
	}

	return resp
}

// ListWorldDropItems 查询世界掉落物品列表
func (s *WorldDropService) ListWorldDropItems(ctx context.Context, configID string, page, pageSize int) (*dto.WorldDropItemListResponse, error) {
	if _, err := s.worldDropRepo.GetByID(ctx, configID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询世界掉落配置失败")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	params := interfaces.ListWorldDropItemParams{WorldDropConfigID: configID, Page: page, PageSize: pageSize}
	items, total, err := s.worldDropItemRepo.ListByConfig(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询世界掉落物品失败")
	}

	responses := make([]dto.WorldDropItemResponse, 0, len(items))
	for _, entry := range items {
		responses = append(responses, s.toWorldDropItemResponse(entry))
	}

	return &dto.WorldDropItemListResponse{
		Items:    responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CreateWorldDropItem 创建世界掉落物品
func (s *WorldDropService) CreateWorldDropItem(ctx context.Context, configID string, req *dto.CreateWorldDropItemRequest) (*dto.WorldDropItemResponse, error) {
	logger := log.GetLogger()
	config, err := s.worldDropRepo.GetByID(ctx, configID)
	if err != nil {
		logger.WarnContext(ctx, "world drop config missing when creating item", log.String("config_id", configID), log.Any("error", err))
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "世界掉落配置不存在或已删除")
	}
	if err := validateWorldDropItemPayload(req.DropRate, req.DropWeight, req.MinQuantity, req.MaxQuantity, req.MinLevel, req.MaxLevel); err != nil {
		return nil, err
	}
	if _, err := s.itemRepo.GetByID(ctx, req.ItemID); err != nil {
		logger.WarnContext(ctx, "world drop item source missing", log.String("item_id", req.ItemID), log.Any("error", err))
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品不存在或已禁用")
	}

	existsInSameConfig, err := s.worldDropItemRepo.HasItemInConfig(ctx, configID, req.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "校验物品唯一性失败")
	}
	if existsInSameConfig {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "该物品已在当前世界掉落中配置")
	}

	existsElsewhere, err := s.worldDropItemRepo.ExistsActiveItem(ctx, req.ItemID, configID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "校验物品占用失败")
	}
	if existsElsewhere {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "该物品已绑定到其他世界掉落")
	}

	if req.DropRate != nil {
		if err := s.ensureDropRateWithinLimit(ctx, configID, *req.DropRate, nil); err != nil {
			return nil, err
		}
	}

	entry := interfaces.WorldDropItem{
		WorldDropConfigID: config.ID,
		ItemID:            req.ItemID,
		DropRate:          req.DropRate,
		DropWeight:        req.DropWeight,
		MinQuantity:       req.MinQuantity,
		MaxQuantity:       req.MaxQuantity,
		MinLevel:          req.MinLevel,
		MaxLevel:          req.MaxLevel,
		GuaranteedDrop:    req.GuaranteedDrop,
		Metadata:          req.Metadata,
	}

	if err := s.worldDropItemRepo.Create(ctx, &entry); err != nil {
		if isUniqueViolation(err) {
			return nil, xerrors.New(xerrors.CodeDuplicateResource, "该物品已在当前世界掉落中配置")
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建世界掉落物品失败")
	}

	created, err := s.worldDropItemRepo.GetByID(ctx, config.ID, entry.ID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询创建后的世界掉落物品失败")
	}

	resp := s.toWorldDropItemResponse(*created)
	return &resp, nil
}

// UpdateWorldDropItem 更新世界掉落物品
func (s *WorldDropService) UpdateWorldDropItem(ctx context.Context, configID, itemEntryID string, req *dto.UpdateWorldDropItemRequest) (*dto.WorldDropItemResponse, error) {
	logger := log.GetLogger()
	config, err := s.worldDropRepo.GetByID(ctx, configID)
	if err != nil {
		logger.WarnContext(ctx, "world drop config missing when updating item", log.String("config_id", configID), log.Any("error", err))
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "世界掉落配置不存在或已删除")
	}
	entry, err := s.worldDropItemRepo.GetByID(ctx, configID, itemEntryID)
	if err != nil {
		logger.WarnContext(ctx, "world drop item entry missing", log.String("config_id", configID), log.String("item_entry_id", itemEntryID), log.Any("error", err))
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "世界掉落物品不存在")
	}

	newItemID := entry.ItemID
	if req.ItemID != nil {
		newItemID = *req.ItemID
	}

	newDropRate := entry.DropRate
	if req.DropRate != nil {
		newDropRate = req.DropRate
	}
	newDropWeight := entry.DropWeight
	if req.DropWeight != nil {
		newDropWeight = req.DropWeight
	}

	minQty := entry.MinQuantity
	if req.MinQuantity != nil {
		minQty = *req.MinQuantity
	}
	maxQty := entry.MaxQuantity
	if req.MaxQuantity != nil {
		maxQty = *req.MaxQuantity
	}
	minLevel := entry.MinLevel
	if req.MinLevel != nil {
		minLevel = req.MinLevel
	}
	maxLevel := entry.MaxLevel
	if req.MaxLevel != nil {
		maxLevel = req.MaxLevel
	}
	guaranteed := entry.GuaranteedDrop
	if req.GuaranteedDrop != nil {
		guaranteed = *req.GuaranteedDrop
	}

	if err := validateWorldDropItemPayload(newDropRate, newDropWeight, minQty, maxQty, minLevel, maxLevel); err != nil {
		return nil, err
	}

	if newItemID != entry.ItemID {
		if _, err := s.itemRepo.GetByID(ctx, newItemID); err != nil {
			logger.WarnContext(ctx, "world drop item update source missing", log.String("item_id", newItemID), log.Any("error", err))
			return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品不存在或已禁用")
		}
		inConfig, err := s.worldDropItemRepo.HasItemInConfig(ctx, configID, newItemID)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "校验物品唯一性失败")
		}
		if inConfig {
			return nil, xerrors.New(xerrors.CodeDuplicateResource, "该物品已在当前世界掉落中配置")
		}
		inOtherConfig, err := s.worldDropItemRepo.ExistsActiveItem(ctx, newItemID, configID)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "校验物品占用失败")
		}
		if inOtherConfig {
			return nil, xerrors.New(xerrors.CodeDuplicateResource, "该物品已绑定到其他世界掉落")
		}
	}

	if newDropRate != nil {
		if err := s.ensureDropRateWithinLimit(ctx, configID, *newDropRate, &itemEntryID); err != nil {
			return nil, err
		}
	}

	metadata := entry.Metadata
	if len(req.Metadata) > 0 {
		metadata = req.Metadata
	}

	update := interfaces.WorldDropItem{
		ID:                itemEntryID,
		WorldDropConfigID: configID,
		ItemID:            newItemID,
		DropRate:          newDropRate,
		DropWeight:        newDropWeight,
		MinQuantity:       minQty,
		MaxQuantity:       maxQty,
		MinLevel:          minLevel,
		MaxLevel:          maxLevel,
		GuaranteedDrop:    guaranteed,
		Metadata:          metadata,
	}

	if err := s.worldDropItemRepo.Update(ctx, &update); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新世界掉落物品失败")
	}

	if config.ItemID == entry.ItemID && newItemID != entry.ItemID {
		config.ItemID = newItemID
		config.UpdatedAt = time.Now()
		if err := s.worldDropRepo.Update(ctx, config); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "同步世界掉落主物品失败")
		}
	}

	updated, err := s.worldDropItemRepo.GetByID(ctx, configID, itemEntryID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询世界掉落物品失败")
	}

	resp := s.toWorldDropItemResponse(*updated)
	return &resp, nil
}

// DeleteWorldDropItem 删除世界掉落物品
func (s *WorldDropService) DeleteWorldDropItem(ctx context.Context, configID, itemEntryID string) error {
	config, err := s.worldDropRepo.GetByID(ctx, configID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询世界掉落配置失败")
	}
	entry, err := s.worldDropItemRepo.GetByID(ctx, configID, itemEntryID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "世界掉落物品不存在")
	}

	_, total, err := s.worldDropItemRepo.ListByConfig(ctx, interfaces.ListWorldDropItemParams{WorldDropConfigID: configID, Page: 1, PageSize: 1})
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询世界掉落物品失败")
	}
	if total <= 1 {
		return xerrors.New(xerrors.CodeInvalidParams, "至少保留一条世界掉落物品记录")
	}

	if err := s.worldDropItemRepo.SoftDelete(ctx, configID, itemEntryID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除世界掉落物品失败")
	}

	if config.ItemID == entry.ItemID {
		next, _, listErr := s.worldDropItemRepo.ListByConfig(ctx, interfaces.ListWorldDropItemParams{WorldDropConfigID: configID, Page: 1, PageSize: 1})
		if listErr != nil {
			return xerrors.Wrap(listErr, xerrors.CodeInternalError, "查询剩余世界掉落物品失败")
		}
		if len(next) > 0 {
			config.ItemID = next[0].ItemID
			config.UpdatedAt = time.Now()
			if err := s.worldDropRepo.Update(ctx, config); err != nil {
				return xerrors.Wrap(err, xerrors.CodeInternalError, "同步世界掉落主物品失败")
			}
		}
	}

	return nil
}

func (s *WorldDropService) toWorldDropItemResponse(entry interfaces.WorldDropItemWithItem) dto.WorldDropItemResponse {
	resp := dto.WorldDropItemResponse{
		ID:                entry.ID,
		WorldDropConfigID: entry.WorldDropConfigID,
		ItemID:            entry.ItemID,
		ItemCode:          entry.ItemCode,
		ItemName:          entry.ItemName,
		ItemQuality:       entry.ItemQuality,
		MinQuantity:       entry.MinQuantity,
		MaxQuantity:       entry.MaxQuantity,
		GuaranteedDrop:    entry.GuaranteedDrop,
		Metadata:          entry.Metadata,
		CreatedAt:         entry.CreatedAt,
		UpdatedAt:         entry.UpdatedAt,
	}
	resp.DropRate = entry.DropRate
	resp.DropWeight = entry.DropWeight
	resp.MinLevel = entry.MinLevel
	resp.MaxLevel = entry.MaxLevel
	return resp
}

func validateWorldDropItemPayload(dropRate *float64, dropWeight *int, minQuantity, maxQuantity int, minLevel, maxLevel *int) error {
	if err := ensureWorldDropItemModeValid(dropRate, dropWeight); err != nil {
		return err
	}
	if minQuantity <= 0 || maxQuantity <= 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "掉落数量必须大于0")
	}
	if minQuantity > maxQuantity {
		return xerrors.New(xerrors.CodeInvalidParams, "最小掉落数量不能大于最大数量")
	}
	if minLevel != nil && maxLevel != nil && *minLevel > *maxLevel {
		return xerrors.New(xerrors.CodeInvalidParams, "最低等级不能大于最高等级")
	}
	return nil
}

func ensureWorldDropItemModeValid(dropRate *float64, dropWeight *int) error {
	if dropRate == nil && dropWeight == nil {
		return xerrors.New(xerrors.CodeInvalidParams, "必须提供掉落概率或权重")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

func (s *WorldDropService) ensureDropRateWithinLimit(ctx context.Context, configID string, candidate float64, excludeItemEntryID *string) error {
	sum, err := s.worldDropItemRepo.SumDropRates(ctx, configID, excludeItemEntryID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "统计世界掉落概率失败")
	}
	if sum+candidate > 1+dropRateEpsilon {
		return xerrors.New(xerrors.CodeInvalidParams, "世界掉落概率总和不可超过1")
	}
	return nil
}
