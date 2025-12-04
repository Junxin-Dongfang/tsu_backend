package service

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/ctxkey"
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

	currentUserID := ctxkey.GetString(ctx, ctxkey.UserID)
	if currentUserID == "" {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "未登录或会话失效")
	}
	if item.OwnerID != currentUserID {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "只能操作自己的物品")
	}

	// 5. 验证当前位置
	if item.ItemLocation != req.FromLocation {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "物品当前位置与源位置不匹配")
	}

	// 5.1 验证归属：若物品有 hero_id 则只能由该英雄移动
	if item.HeroID.Valid {
		currentHeroID := ctxkey.GetString(ctx, ctxkey.HeroID)
		if currentHeroID != "" && item.HeroID.String != currentHeroID {
			return nil, xerrors.New(xerrors.CodePermissionDenied, "无权操作该物品")
		}
	}

	// 6. 验证物品未装备
	if item.ItemLocation == "equipped" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "已装备的物品无法移动,请先卸下")
	}

	// 7. 验证目标位置有足够空间（按条目计数近似）
	used, capacity, err := s.GetInventoryCapacity(ctx, item.OwnerID, req.ToLocation)
	if err != nil {
		return nil, err
	}
	if used >= capacity {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "目标位置容量不足")
	}

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
	Quantity       int    `json:"quantity"`
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
	if req.Quantity <= 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "丢弃数量必须大于0")
	}

	// 2. 开启事务并加锁获取
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	item, err := s.playerItemRepo.GetByIDForUpdate(ctx, tx, req.ItemInstanceID)
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

	// 6. 校验数量并更新/删除
	current := 1
	if item.StackCount.Valid && item.StackCount.Int > 0 {
		current = int(item.StackCount.Int)
	}
	if req.Quantity > current {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "丢弃数量超过持有数量")
	}

	now := time.Now()
	if req.Quantity == current {
		item.DeletedAt = null.TimeFrom(now)
		item.UpdatedAt = now
		if _, err := item.Update(ctx, tx, boil.Whitelist("deleted_at", "updated_at")); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "丢弃物品失败")
		}
	} else {
		item.StackCount = null.IntFrom(current - req.Quantity)
		item.UpdatedAt = now
		if _, err := item.Update(ctx, tx, boil.Whitelist("stack_count", "updated_at")); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新物品数量失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
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

	// 3. 实现排序逻辑：类型 > 品质 > 获得时间倒序
	type itemWithMeta struct {
		item  *game_runtime.PlayerItem
		typeV string
		qualV string
	}

	itemMeta := make([]itemWithMeta, 0, len(items))
	for _, it := range items {
		cfg, err := s.itemRepo.GetByID(ctx, it.ItemID)
		if err != nil {
			itemMeta = append(itemMeta, itemWithMeta{item: it})
			continue
		}
		itemMeta = append(itemMeta, itemWithMeta{
			item:  it,
			typeV: cfg.ItemType,
			qualV: cfg.ItemQuality,
		})
	}

	sort.Slice(itemMeta, func(i, j int) bool {
		if itemMeta[i].typeV != itemMeta[j].typeV {
			return itemMeta[i].typeV < itemMeta[j].typeV
		}
		if itemMeta[i].qualV != itemMeta[j].qualV {
			return itemMeta[i].qualV < itemMeta[j].qualV
		}
		return itemMeta[i].item.CreatedAt.After(itemMeta[j].item.CreatedAt)
	})

	for idx, meta := range itemMeta {
		meta.item.LocationIndex = null.IntFrom(int(idx))
		if err := s.playerItemRepo.Update(ctx, s.db, meta.item); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新物品排序失败")
		}
	}

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
	if ownerID == "" || location == "" {
		return 0, 0, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 查询容量配置
	var maxSlots int
	err := s.db.QueryRowContext(ctx, `SELECT max_slots FROM game_config.inventory_capacities WHERE location = $1`, location).Scan(&maxSlots)
	if err == sql.ErrNoRows {
		return 0, 0, xerrors.New(xerrors.CodeInternalError, "未配置容量")
	}
	if err != nil {
		return 0, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询容量配置失败")
	}

	// 已用容量（以记录数近似槽位数）
	var used int
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM game_runtime.player_items WHERE owner_id = $1 AND item_location = $2 AND deleted_at IS NULL`, ownerID, location).Scan(&used)
	if err != nil {
		return 0, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "统计已用容量失败")
	}

	return used, maxSlots, nil
}
