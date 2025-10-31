// Package service 提供Admin模块的业务逻辑服务
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

// DropPoolService 掉落池服务
type DropPoolService struct {
	dropPoolRepo interfaces.DropPoolRepository
	itemRepo     interfaces.ItemRepository
}

// NewDropPoolService 创建掉落池服务
func NewDropPoolService(db *sql.DB) *DropPoolService {
	return &DropPoolService{
		dropPoolRepo: impl.NewDropPoolRepository(db),
		itemRepo:     impl.NewItemRepository(db),
	}
}

// CreateDropPool 创建掉落池
func (s *DropPoolService) CreateDropPool(ctx context.Context, req *dto.CreateDropPoolRequest) (*dto.DropPoolResponse, error) {
	// 1. 验证pool_code唯一性
	existing, err := s.dropPoolRepo.GetByCode(ctx, req.PoolCode)
	if err == nil && existing != nil {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("掉落池代码已存在: %s", req.PoolCode))
	}

	// 2. 验证掉落数量范围
	if req.MinDrops > req.MaxDrops {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小掉落数量不能大于最大掉落数量")
	}
	if req.GuaranteedDrops > req.MaxDrops {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "保底掉落数量不能大于最大掉落数量")
	}

	// 3. 创建掉落池实体
	pool := &game_config.DropPool{
		ID:        uuid.New().String(),
		PoolCode:  req.PoolCode,
		PoolName:  req.PoolName,
		PoolType:  req.PoolType,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	pool.MinDrops.SetValid(req.MinDrops)
	pool.MaxDrops.SetValid(req.MaxDrops)
	pool.GuaranteedDrops.SetValid(req.GuaranteedDrops)

	if req.Description != nil {
		pool.Description.SetValid(*req.Description)
	}

	// 4. 保存到数据库
	if err := s.dropPoolRepo.Create(ctx, pool); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建掉落池失败")
	}

	return s.toDropPoolResponse(pool), nil
}

// GetDropPoolByID 获取掉落池详情
func (s *DropPoolService) GetDropPoolByID(ctx context.Context, poolID string) (*dto.DropPoolResponse, error) {
	pool, err := s.dropPoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询掉落池失败")
	}

	return s.toDropPoolResponse(pool), nil
}

// GetDropPoolList 查询掉落池列表
func (s *DropPoolService) GetDropPoolList(ctx context.Context, params interfaces.ListDropPoolParams) (*dto.DropPoolListResponse, error) {
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

	// 查询掉落池列表
	pools, total, err := s.dropPoolRepo.List(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询掉落池列表失败")
	}

	// 转换为响应
	responses := make([]dto.DropPoolResponse, 0, len(pools))
	for _, pool := range pools {
		responses = append(responses, *s.toDropPoolResponse(pool))
	}

	return &dto.DropPoolListResponse{
		Items:    responses,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// UpdateDropPool 更新掉落池
func (s *DropPoolService) UpdateDropPool(ctx context.Context, poolID string, req *dto.UpdateDropPoolRequest) (*dto.DropPoolResponse, error) {
	// 1. 查询掉落池是否存在
	pool, err := s.dropPoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询掉落池失败")
	}

	// 2. 验证pool_code唯一性（如果要更新）
	if req.PoolCode != nil && *req.PoolCode != pool.PoolCode {
		existing, err := s.dropPoolRepo.GetByCode(ctx, *req.PoolCode)
		if err == nil && existing != nil {
			return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("掉落池代码已存在: %s", *req.PoolCode))
		}
	}

	// 3. 更新字段
	if req.PoolCode != nil {
		pool.PoolCode = *req.PoolCode
	}
	if req.PoolName != nil {
		pool.PoolName = *req.PoolName
	}
	if req.PoolType != nil {
		pool.PoolType = *req.PoolType
	}
	if req.Description != nil {
		pool.Description.SetValid(*req.Description)
	}
	if req.MinDrops != nil {
		pool.MinDrops.SetValid(*req.MinDrops)
	}
	if req.MaxDrops != nil {
		pool.MaxDrops.SetValid(*req.MaxDrops)
	}
	if req.GuaranteedDrops != nil {
		pool.GuaranteedDrops.SetValid(*req.GuaranteedDrops)
	}
	if req.IsActive != nil {
		pool.IsActive = *req.IsActive
	}

	// 4. 验证掉落数量范围
	if pool.MinDrops.Valid && pool.MaxDrops.Valid && pool.MinDrops.Int16 > pool.MaxDrops.Int16 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小掉落数量不能大于最大掉落数量")
	}
	if pool.GuaranteedDrops.Valid && pool.MaxDrops.Valid && pool.GuaranteedDrops.Int16 > pool.MaxDrops.Int16 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "保底掉落数量不能大于最大掉落数量")
	}

	pool.UpdatedAt = time.Now()

	// 5. 保存更新
	if err := s.dropPoolRepo.Update(ctx, pool); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新掉落池失败")
	}

	return s.toDropPoolResponse(pool), nil
}

// DeleteDropPool 删除掉落池（软删除）
func (s *DropPoolService) DeleteDropPool(ctx context.Context, poolID string, cascade bool) error {
	// 1. 查询掉落池是否存在
	_, err := s.dropPoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询掉落池失败")
	}

	// 2. 检查是否有关联的掉落物品
	items, err := s.dropPoolRepo.GetPoolItems(ctx, poolID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询掉落池物品失败")
	}

	if len(items) > 0 && !cascade {
		return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("掉落池中还有%d个物品，请先删除物品或使用级联删除", len(items)))
	}

	// 3. 级联删除物品（如果需要）
	if cascade && len(items) > 0 {
		for _, item := range items {
			if err := s.dropPoolRepo.DeletePoolItem(ctx, poolID, item.ItemID); err != nil {
				return xerrors.Wrap(err, xerrors.CodeInternalError, "删除掉落池物品失败")
			}
		}
	}

	// 4. 删除掉落池
	if err := s.dropPoolRepo.Delete(ctx, poolID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除掉落池失败")
	}

	return nil
}

// toDropPoolResponse 转换为掉落池响应
func (s *DropPoolService) toDropPoolResponse(pool *game_config.DropPool) *dto.DropPoolResponse {
	resp := &dto.DropPoolResponse{
		ID:        pool.ID,
		PoolCode:  pool.PoolCode,
		PoolName:  pool.PoolName,
		PoolType:  pool.PoolType,
		IsActive:  pool.IsActive,
		CreatedAt: pool.CreatedAt,
		UpdatedAt: pool.UpdatedAt,
	}

	if pool.MinDrops.Valid {
		resp.MinDrops = pool.MinDrops.Int16
	}
	if pool.MaxDrops.Valid {
		resp.MaxDrops = pool.MaxDrops.Int16
	}
	if pool.GuaranteedDrops.Valid {
		resp.GuaranteedDrops = pool.GuaranteedDrops.Int16
	}

	if pool.Description.Valid {
		desc := pool.Description.String
		resp.Description = &desc
	}

	if pool.DeletedAt.Valid {
		deletedAt := pool.DeletedAt.Time
		resp.DeletedAt = &deletedAt
	}

	return resp
}

// AddDropPoolItem 添加掉落物品
func (s *DropPoolService) AddDropPoolItem(ctx context.Context, poolID string, req *dto.AddDropPoolItemRequest) (*dto.DropPoolItemResponse, error) {
	// 1. 验证掉落池ID存在性
	_, err := s.dropPoolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "掉落池不存在")
	}

	// 2. 验证物品ID存在性
	item, err := s.itemRepo.GetByID(ctx, req.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品不存在")
	}

	// 3. 验证数量范围
	if req.MinQuantity > req.MaxQuantity {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小数量不能大于最大数量")
	}

	// 4. 验证等级范围
	if req.MinLevel != nil && req.MaxLevel != nil && *req.MinLevel > *req.MaxLevel {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最低等级不能大于最高等级")
	}

	// 5. 验证品质权重JSON格式
	if len(req.QualityWeights) > 0 {
		var weights map[string]interface{}
		if unmarshalErr := json.Unmarshal(req.QualityWeights, &weights); unmarshalErr != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "品质权重JSON格式错误")
		}
	}

	// 6. 检查物品是否已在掉落池中
	existing, err := s.dropPoolRepo.GetPoolItemByID(ctx, poolID, req.ItemID)
	if err == nil && existing != nil {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "物品已在掉落池中")
	}

	// 7. 创建掉落池物品实体
	poolItem := &game_config.DropPoolItem{
		ID:         uuid.New().String(),
		DropPoolID: poolID,
		ItemID:     req.ItemID,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 设置掉落权重（默认为1）
	if req.DropWeight != nil {
		poolItem.DropWeight = *req.DropWeight
	} else {
		poolItem.DropWeight = 1
	}

	// 设置掉落概率
	if req.DropRate != nil {
		dec := new(decimal.Big).SetFloat64(*req.DropRate)
		poolItem.DropRate = types.NewNullDecimal(dec)
	}

	// 设置品质权重
	if len(req.QualityWeights) > 0 {
		poolItem.QualityWeights.SetValid(req.QualityWeights)
	}

	// 设置数量范围
	poolItem.MinQuantity.SetValid(int(req.MinQuantity))
	poolItem.MaxQuantity.SetValid(int(req.MaxQuantity))

	// 设置等级范围
	if req.MinLevel != nil {
		poolItem.MinLevel.SetValid(*req.MinLevel)
	}
	if req.MaxLevel != nil {
		poolItem.MaxLevel.SetValid(*req.MaxLevel)
	}

	// 8. 保存到数据库
	if err := s.dropPoolRepo.CreatePoolItem(ctx, poolItem); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "添加掉落物品失败")
	}

	return s.toDropPoolItemResponse(poolItem, item), nil
}

// GetDropPoolItems 查询掉落池物品列表
func (s *DropPoolService) GetDropPoolItems(ctx context.Context, params interfaces.ListDropPoolItemParams) (*dto.DropPoolItemListResponse, error) {
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

	// 查询掉落池物品列表
	items, total, err := s.dropPoolRepo.ListPoolItems(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询掉落池物品列表失败")
	}

	// 批量查询物品信息
	responses := make([]dto.DropPoolItemResponse, 0, len(items))
	for _, poolItem := range items {
		item, err := s.itemRepo.GetByID(ctx, poolItem.ItemID)
		if err != nil {
			continue // 跳过不存在的物品
		}
		responses = append(responses, *s.toDropPoolItemResponse(poolItem, item))
	}

	return &dto.DropPoolItemListResponse{
		Items:    responses,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// GetDropPoolItem 获取掉落物品详情
func (s *DropPoolService) GetDropPoolItem(ctx context.Context, poolID, itemID string) (*dto.DropPoolItemResponse, error) {
	poolItem, err := s.dropPoolRepo.GetPoolItemByID(ctx, poolID, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询掉落物品失败")
	}

	item, err := s.itemRepo.GetByID(ctx, poolItem.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品信息失败")
	}

	return s.toDropPoolItemResponse(poolItem, item), nil
}

// UpdateDropPoolItem 更新掉落物品
func (s *DropPoolService) UpdateDropPoolItem(ctx context.Context, poolID, itemID string, req *dto.UpdateDropPoolItemRequest) (*dto.DropPoolItemResponse, error) {
	// 1. 查询掉落物品是否存在
	poolItem, err := s.dropPoolRepo.GetPoolItemByID(ctx, poolID, itemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询掉落物品失败")
	}

	// 2. 更新字段
	if req.DropWeight != nil {
		poolItem.DropWeight = *req.DropWeight
	}
	if req.DropRate != nil {
		dec := new(decimal.Big).SetFloat64(*req.DropRate)
		poolItem.DropRate = types.NewNullDecimal(dec)
	}
	if len(req.QualityWeights) > 0 {
		// 验证JSON格式
		var weights map[string]interface{}
		if unmarshalErr := json.Unmarshal(req.QualityWeights, &weights); unmarshalErr != nil {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "品质权重JSON格式错误")
		}
		poolItem.QualityWeights.SetValid(req.QualityWeights)
	}
	if req.MinQuantity != nil {
		poolItem.MinQuantity.SetValid(int(*req.MinQuantity))
	}
	if req.MaxQuantity != nil {
		poolItem.MaxQuantity.SetValid(int(*req.MaxQuantity))
	}
	if req.MinLevel != nil {
		poolItem.MinLevel.SetValid(*req.MinLevel)
	}
	if req.MaxLevel != nil {
		poolItem.MaxLevel.SetValid(*req.MaxLevel)
	}
	if req.IsActive != nil {
		poolItem.IsActive = *req.IsActive
	}

	// 3. 验证数量范围
	if poolItem.MinQuantity.Valid && poolItem.MaxQuantity.Valid && poolItem.MinQuantity.Int > poolItem.MaxQuantity.Int {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最小数量不能大于最大数量")
	}

	// 4. 验证等级范围
	if poolItem.MinLevel.Valid && poolItem.MaxLevel.Valid && poolItem.MinLevel.Int16 > poolItem.MaxLevel.Int16 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "最低等级不能大于最高等级")
	}

	poolItem.UpdatedAt = time.Now()

	// 5. 保存更新
	if updateErr := s.dropPoolRepo.UpdatePoolItem(ctx, poolItem); updateErr != nil {
		return nil, xerrors.Wrap(updateErr, xerrors.CodeInternalError, "更新掉落物品失败")
	}

	item, err := s.itemRepo.GetByID(ctx, poolItem.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询物品信息失败")
	}

	return s.toDropPoolItemResponse(poolItem, item), nil
}

// RemoveDropPoolItem 移除掉落物品
func (s *DropPoolService) RemoveDropPoolItem(ctx context.Context, poolID, itemID string) error {
	// 1. 查询掉落物品是否存在
	_, err := s.dropPoolRepo.GetPoolItemByID(ctx, poolID, itemID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "查询掉落物品失败")
	}

	// 2. 删除掉落物品
	if err := s.dropPoolRepo.DeletePoolItem(ctx, poolID, itemID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "删除掉落物品失败")
	}

	return nil
}

// toDropPoolItemResponse 转换为掉落物品响应
func (s *DropPoolService) toDropPoolItemResponse(poolItem *game_config.DropPoolItem, item *game_config.Item) *dto.DropPoolItemResponse {
	resp := &dto.DropPoolItemResponse{
		ID:         poolItem.ID,
		DropPoolID: poolItem.DropPoolID,
		ItemID:     poolItem.ItemID,
		ItemCode:   item.ItemCode,
		ItemName:   item.ItemName,
		IsActive:   poolItem.IsActive,
		CreatedAt:  poolItem.CreatedAt,
		UpdatedAt:  poolItem.UpdatedAt,
	}

	// 掉落权重（非空字段）
	weight := poolItem.DropWeight
	resp.DropWeight = &weight

	// 掉落概率
	if !poolItem.DropRate.IsZero() {
		rate, _ := poolItem.DropRate.Float64()
		resp.DropRate = &rate
	}

	// 品质权重
	if poolItem.QualityWeights.Valid {
		resp.QualityWeights = poolItem.QualityWeights.JSON
	}

	// 数量范围
	if poolItem.MinQuantity.Valid {
		minQty := int16(poolItem.MinQuantity.Int)
		resp.MinQuantity = minQty
	}
	if poolItem.MaxQuantity.Valid {
		maxQty := int16(poolItem.MaxQuantity.Int)
		resp.MaxQuantity = maxQty
	}

	// 等级范围
	if poolItem.MinLevel.Valid {
		level := poolItem.MinLevel.Int16
		resp.MinLevel = &level
	}
	if poolItem.MaxLevel.Valid {
		level := poolItem.MaxLevel.Int16
		resp.MaxLevel = &level
	}

	// 删除时间
	if poolItem.DeletedAt.Valid {
		deletedAt := poolItem.DeletedAt.Time
		resp.DeletedAt = &deletedAt
	}

	return resp
}
