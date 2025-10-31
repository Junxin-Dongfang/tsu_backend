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
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// WorldDropService 世界掉落服务
type WorldDropService struct {
	worldDropRepo interfaces.WorldDropConfigRepository
	itemRepo      interfaces.ItemRepository
}

// NewWorldDropService 创建世界掉落服务
func NewWorldDropService(db *sql.DB) *WorldDropService {
	return &WorldDropService{
		worldDropRepo: impl.NewWorldDropConfigRepository(db),
		itemRepo:      impl.NewItemRepository(db),
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

	// 3. 检查物品是否已有世界掉落配置
	existing, err := s.worldDropRepo.GetByItemID(ctx, req.ItemID)
	if err == nil && existing != nil {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("物品已有世界掉落配置: %s", req.ItemID))
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
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品信息失败")
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
