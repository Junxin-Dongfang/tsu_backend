package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// InventoryService 背包管理服务
type InventoryService struct {
	db             *sql.DB
	playerItemRepo interfaces.PlayerItemRepository
	itemRepo       interfaces.ItemRepository
}

// NewInventoryService 创建背包管理服务
func NewInventoryService(db *sql.DB) *InventoryService {
	return &InventoryService{
		db:             db,
		playerItemRepo: impl.NewPlayerItemRepository(db),
		itemRepo:       impl.NewItemRepository(db),
	}
}

// GetInventoryRequest 查询背包请求
type GetInventoryRequest struct {
	OwnerID      string  `json:"owner_id"`
	ItemLocation string  `json:"item_location"` // backpack/warehouse/storage
	ItemType     *string `json:"item_type,omitempty"`
	ItemQuality  *string `json:"item_quality,omitempty"`
	Page         int     `json:"page"`
	PageSize     int     `json:"page_size"`
}

// GetInventoryResponse 查询背包响应
type GetInventoryResponse struct {
	Items      []*PlayerItemDTO `json:"items"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

// GetInventory 查询背包/仓库
func (s *InventoryService) GetInventory(ctx context.Context, req *GetInventoryRequest) (*GetInventoryResponse, error) {
	// 1. 验证参数
	if req.OwnerID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "所有者ID不能为空")
	}
	if req.ItemLocation == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品位置不能为空")
	}

	// 2. 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 3. 查询物品列表
	params := interfaces.GetPlayerItemsParams{
		OwnerID:      req.OwnerID,
		ItemLocation: &req.ItemLocation,
		ItemType:     req.ItemType,
		ItemQuality:  req.ItemQuality,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	items, totalCount, err := s.playerItemRepo.GetByOwnerPaginated(ctx, params)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询背包失败")
	}

	// 转换为DTO
	dtos := make([]*PlayerItemDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, convertItemToDTO(item))
	}

	return &GetInventoryResponse{
		Items:      dtos,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// MoveItemRequest 移动物品请求
type MoveItemRequest struct {
	ItemInstanceID string `json:"item_instance_id"`
	FromLocation   string `json:"from_location"`
	ToLocation     string `json:"to_location"`
}

// MoveItemResponse 移动物品响应
type MoveItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// MoveItem 移动物品(背包↔仓库)
func (s *InventoryService) MoveItem(ctx context.Context, req *MoveItemRequest) (*MoveItemResponse, error) {
	// 1. 验证参数
	if req.ItemInstanceID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品实例ID不能为空")
	}
	if req.FromLocation == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "源位置不能为空")
	}
	if req.ToLocation == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "目标位置不能为空")
	}

	// 2. 验证位置有效性
	validLocations := map[string]bool{
		"backpack":  true,
		"warehouse": true,
		"storage":   true,
	}
	if !validLocations[req.FromLocation] || !validLocations[req.ToLocation] {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "无效的物品位置")
	}

	// 3. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 4. 获取物品实例(带锁)
	item, err := s.playerItemRepo.GetByIDForUpdate(ctx, tx, req.ItemInstanceID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品实例不存在")
	}

	// 5. 验证当前位置
	if item.ItemLocation != req.FromLocation {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品当前位置与源位置不匹配")
	}

	// 6. 验证物品未装备
	if item.ItemLocation == "equipped" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "已装备的物品无法移动,请先卸下")
	}

	// 7. TODO: 验证目标位置有足够空间

	// 8. 更新物品位置
	if err := s.playerItemRepo.UpdateLocation(ctx, tx, req.ItemInstanceID, req.ToLocation); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新物品位置失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return &MoveItemResponse{
		Success: true,
		Message: "物品移动成功",
	}, nil
}

// DiscardItemRequest 丢弃物品请求
type DiscardItemRequest struct {
	ItemInstanceID string `json:"item_instance_id"`
}

// DiscardItemResponse 丢弃物品响应
type DiscardItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DiscardItem 丢弃物品(软删除)
func (s *InventoryService) DiscardItem(ctx context.Context, req *DiscardItemRequest) (*DiscardItemResponse, error) {
	// 1. 验证参数
	if req.ItemInstanceID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品实例ID不能为空")
	}

	// 2. 获取物品实例
	item, err := s.playerItemRepo.GetByID(ctx, req.ItemInstanceID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品实例不存在")
	}

	// 3. 验证物品未装备
	if item.ItemLocation == "equipped" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "已装备的物品无法丢弃,请先卸下")
	}

	// 4. 获取物品配置
	itemConfig, err := s.itemRepo.GetByID(ctx, item.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品配置不存在")
	}

	// 5. 验证物品可丢弃
	if itemConfig.IsDroppable.Valid && !itemConfig.IsDroppable.Bool {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "该物品无法丢弃")
	}

	// 6. 软删除物品
	if err := s.playerItemRepo.Delete(ctx, req.ItemInstanceID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "丢弃物品失败")
	}

	return &DiscardItemResponse{
		Success: true,
		Message: "物品丢弃成功",
	}, nil
}

// SortInventoryRequest 整理背包请求
type SortInventoryRequest struct {
	OwnerID      string `json:"owner_id"`
	ItemLocation string `json:"item_location"` // backpack/warehouse/storage
}

// SortInventoryResponse 整理背包响应
type SortInventoryResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Items   []*PlayerItemDTO `json:"items"`
}

// SortInventory 整理背包(按类型、品质、等级排序)
func (s *InventoryService) SortInventory(ctx context.Context, req *SortInventoryRequest) (*SortInventoryResponse, error) {
	// 1. 验证参数
	if req.OwnerID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "所有者ID不能为空")
	}
	if req.ItemLocation == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品位置不能为空")
	}

	// 2. 查询该位置的所有物品
	items, err := s.playerItemRepo.GetByOwner(ctx, req.OwnerID, &req.ItemLocation)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品失败")
	}

	// 3. TODO: 实现排序逻辑
	// - 按物品类型排序(装备、消耗品、材料等)
	// - 同类型按品质排序
	// - 同品质按等级排序
	// - 更新物品的 location_index

	// 转换为DTO
	dtos := make([]*PlayerItemDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, convertItemToDTO(item))
	}

	return &SortInventoryResponse{
		Success: true,
		Message: "背包整理成功",
		Items:   dtos,
	}, nil
}

// GetInventoryCapacity 查询背包容量
func (s *InventoryService) GetInventoryCapacity(ctx context.Context, ownerID string, location string) (int, int, error) {
	// TODO: 实现容量查询
	// - 查询该位置的物品数量
	// - 查询该位置的最大容量(可能需要新增配置表)
	// - 返回 (已用容量, 最大容量)

	return 0, 100, nil // 临时返回
}

