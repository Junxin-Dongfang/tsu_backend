package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/notify"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TeamWarehouseService 团队仓库服务
type TeamWarehouseService struct {
	db                    *sql.DB
	teamMemberRepo        interfaces.TeamMemberRepository
	teamWarehouseRepo     interfaces.TeamWarehouseRepository
	teamWarehouseItemRepo interfaces.TeamWarehouseItemRepository
	heroWalletRepo        interfaces.HeroWalletRepository
	lootHistoryRepo       interfaces.TeamLootHistoryRepository
	lootLogRepo           interfaces.TeamWarehouseLootLogRepository
	itemRepo              interfaces.ItemRepository
	heroRepo              interfaces.HeroRepository
	playerItemRepo        interfaces.PlayerItemRepository
}

// NewTeamWarehouseService 创建团队仓库服务
func NewTeamWarehouseService(db *sql.DB) *TeamWarehouseService {
	return &TeamWarehouseService{
		db:                    db,
		teamMemberRepo:        impl.NewTeamMemberRepository(db),
		teamWarehouseRepo:     impl.NewTeamWarehouseRepository(db),
		teamWarehouseItemRepo: impl.NewTeamWarehouseItemRepository(db),
		heroWalletRepo:        impl.NewHeroWalletRepository(db),
		lootHistoryRepo:       impl.NewTeamLootHistoryRepository(db),
		lootLogRepo:           impl.NewTeamWarehouseLootLogRepository(db),
		itemRepo:              impl.NewItemRepository(db),
		heroRepo:              impl.NewHeroRepository(db),
		playerItemRepo:        impl.NewPlayerItemRepository(db),
	}
}

// GetWarehouse 查看团队仓库（团队成员可查看）
func (s *TeamWarehouseService) GetWarehouse(ctx context.Context, teamID, heroID string) (*game_runtime.TeamWarehouse, error) {
	// 1. 验证参数
	if teamID == "" || heroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查权限（团队成员即可）
	if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}

	// 3. 获取仓库
	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}

	return warehouse, nil
}

// DistributeGoldRequest 分配金币请求
type DistributeGoldRequest struct {
	TeamID        string
	DistributorID string           // 分配者英雄ID
	Distributions map[string]int64 // 接收者英雄ID -> 金币数量
}

// DistributeGold 分配金币
func (s *TeamWarehouseService) DistributeGold(ctx context.Context, req *DistributeGoldRequest) error {
	// 1. 验证参数
	if req.TeamID == "" || req.DistributorID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if len(req.Distributions) == 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "分配列表不能为空")
	}

	// 2. 检查权限（队长或管理员）
	distributor, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.DistributorID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if distributor.Role != "leader" && distributor.Role != "admin" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以分配金币")
	}

	// 3. 获取仓库
	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, req.TeamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}

	// 4. 计算总分配金额
	var totalAmount int64
	for heroID, amount := range req.Distributions {
		if amount <= 0 {
			msg := fmt.Sprintf("分配给 %s 的金额必须大于0", heroID)
			return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
		}
		totalAmount += amount
	}

	// 5. 检查仓库余额
	if warehouse.GoldAmount < totalAmount {
		msg := fmt.Sprintf("仓库金币不足：余额 %d，试图分配 %d", warehouse.GoldAmount, totalAmount)
		return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
	}

	// 5.1 校验接收者为团队成员
	for heroID := range req.Distributions {
		if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, heroID); err != nil {
			return xerrors.Wrap(err, xerrors.CodePermissionDenied, "接收者不是团队成员")
		}
	}

	// 6. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 7. 扣除仓库金币
	if err := s.teamWarehouseRepo.DeductGold(ctx, tx, warehouse.ID, totalAmount); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "扣除仓库金币失败")
	}

	// 8. 分配金币给成员（写入英雄钱包）并记录分配历史
	for heroID, amount := range req.Distributions {
		if err := s.heroWalletRepo.AddGoldTx(ctx, tx, heroID, amount); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "发放金币到英雄钱包失败")
		}
		if s.lootHistoryRepo != nil {
			if err := s.lootHistoryRepo.CreateDistribution(ctx, tx, &interfaces.TeamLootHistoryCreateReq{
				TeamID:            req.TeamID,
				WarehouseID:       warehouse.ID,
				DistributorHeroID: req.DistributorID,
				RecipientHeroID:   heroID,
				ItemType:          "gold",
				Quantity:          amount,
			}); err != nil {
				return xerrors.Wrap(err, xerrors.CodeInternalError, "写入分配历史失败")
			}
		}
	}

	// 10. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 通知
	event := &DistributionEvent{
		TeamID:      req.TeamID,
		WarehouseID: warehouse.ID,
		Distributor: req.DistributorID,
		Recipients:  req.Distributions,
		Result:      "success",
	}
	_ = notify.PublishWarehouseEvent(ctx, notify.SubjectWarehouseDistributed, event)

	return nil
}

// DistributeItemsRequest 分配物品请求
type DistributeItemsRequest struct {
	TeamID        string
	DistributorID string                    // 分配者英雄ID
	Distributions map[string]map[string]int // 接收者英雄ID -> (物品ID -> 数量)
}

// DistributeItems 分配物品
func (s *TeamWarehouseService) DistributeItems(ctx context.Context, req *DistributeItemsRequest) error {
	// 1. 验证参数
	if req.TeamID == "" || req.DistributorID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if len(req.Distributions) == 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "分配列表不能为空")
	}

	// 2. 检查权限（队长或管理员）
	distributor, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, req.DistributorID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if distributor.Role != "leader" && distributor.Role != "admin" {
		return xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以分配物品")
	}

	// 3. 获取仓库
	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, req.TeamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}

	// 4. 统计每个物品的总分配数量
	itemTotals := make(map[string]int)
	for heroID, items := range req.Distributions {
		for itemID, quantity := range items {
			if quantity <= 0 {
				msg := fmt.Sprintf("分配给 %s 的物品 %s 数量必须大于0", heroID, itemID)
				return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
			}
			itemTotals[itemID] += quantity
		}
	}

	// 5. 检查仓库物品库存
	for itemID, totalQuantity := range itemTotals {
		currentQuantity, err := s.teamWarehouseItemRepo.GetItemCount(ctx, warehouse.ID, itemID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "查询物品库存失败")
		}
		if currentQuantity < totalQuantity {
			msg := fmt.Sprintf("仓库物品不足：%s 现有 %d，需要 %d", itemID, currentQuantity, totalQuantity)
			return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
		}
	}

	// 5.1 校验接收者为团队成员
	for heroID := range req.Distributions {
		if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, req.TeamID, heroID); err != nil {
			return xerrors.Wrap(err, xerrors.CodePermissionDenied, "接收者不是团队成员")
		}
	}

	// 6. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 7. 扣除仓库物品
	for itemID, totalQuantity := range itemTotals {
		if err := s.teamWarehouseItemRepo.DeductItem(ctx, tx, warehouse.ID, itemID, totalQuantity); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "扣除仓库物品失败")
		}
	}

	// 8. 分配物品给成员（发放到英雄背包）
	// 简单容量校验：按背包槽位数量检查
	maxSlots, err := s.getLocationMaxSlots(ctx, "backpack")
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询背包容量失败")
	}

	for heroID, items := range req.Distributions {
		hero, err := s.heroRepo.GetByID(ctx, heroID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "接收者英雄不存在")
		}

		currentCount, err := s.countHeroBackpack(ctx, heroID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "统计背包容量失败")
		}
		newSlots := len(items)
		if currentCount+int64(newSlots) > maxSlots {
			msg := fmt.Sprintf("背包已满：当前已占 %d / %d，本次需新增 %d 个种类", currentCount, maxSlots, newSlots)
			return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
		}

		for itemID, qty := range items {
			itemCfg, err := s.itemRepo.GetByID(ctx, itemID)
			if err != nil {
				return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品配置不存在")
			}
			maxStack := 1
			if itemCfg.MaxStackSize.Valid && itemCfg.MaxStackSize.Int > 0 {
				maxStack = int(itemCfg.MaxStackSize.Int)
			}
			if qty > maxStack {
				msg := fmt.Sprintf("物品堆叠数量超限：%s 最大堆叠 %d，本次分配 %d", itemID, maxStack, qty)
				return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
			}

			playerItem := &game_runtime.PlayerItem{
				ID:           uuid.NewString(),
				ItemID:       itemID,
				OwnerID:      hero.UserID,
				HeroID:       null.StringFrom(heroID),
				SourceType:   "reward",
				ItemLocation: "backpack",
				StackCount:   null.IntFrom(qty),
			}
			if err := playerItem.Insert(ctx, tx, boil.Infer()); err != nil {
				return xerrors.Wrap(err, xerrors.CodeInternalError, "发放物品到背包失败")
			}

			if s.lootHistoryRepo != nil {
				if err := s.lootHistoryRepo.CreateDistribution(ctx, tx, &interfaces.TeamLootHistoryCreateReq{
					TeamID:            req.TeamID,
					WarehouseID:       warehouse.ID,
					DistributorHeroID: req.DistributorID,
					RecipientHeroID:   heroID,
					ItemType:          "item",
					ItemID:            &itemID,
					Quantity:          int64(qty),
				}); err != nil {
					return xerrors.Wrap(err, xerrors.CodeInternalError, "写入物品分配历史失败")
				}
			}
		}
	}

	// 10. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	event := &DistributionEvent{
		TeamID:      req.TeamID,
		WarehouseID: warehouse.ID,
		Distributor: req.DistributorID,
		ItemPayload: req.Distributions,
		Result:      "success",
	}
	_ = notify.PublishWarehouseEvent(ctx, notify.SubjectWarehouseDistributed, event)

	return nil
}

// AddLootToWarehouseRequest 战利品入库请求
type AddLootToWarehouseRequest struct {
	TeamID          string
	SourceDungeonID string
	Gold            int64
	Items           []LootItem
}

// LootItem 战利品物品
type LootItem struct {
	ItemID   string
	ItemType string
	Quantity int
}

// AddLootToWarehouse 战利品入库（地城完成后调用）
func (s *TeamWarehouseService) AddLootToWarehouse(ctx context.Context, req *AddLootToWarehouseRequest) error {
	// 1. 验证参数
	if req.TeamID == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}

	// 2. 获取仓库
	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, req.TeamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}

	// 3. 检查仓库容量
	currentItemCount, err := s.teamWarehouseItemRepo.CountDistinctItems(ctx, warehouse.ID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "统计仓库物品种类失败")
	}

	// 已满直接拒绝
	if currentItemCount >= 100 {
		msg := fmt.Sprintf("仓库已满：当前已有 %d 种，最多 100 种，请先分配/清理", currentItemCount)
		fmt.Printf("[AddLoot] full-existing: %s\n", msg)
		_ = s.logLoot(ctx, warehouse.ID, req, "failed", msg)
		return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
	}

	// 计算新增物品种类数
	newItemTypes := make(map[string]bool)
	for _, item := range req.Items {
		// 检查是否已存在
		existingCount, _ := s.teamWarehouseItemRepo.GetItemCount(ctx, warehouse.ID, item.ItemID)
		if existingCount == 0 {
			newItemTypes[item.ItemID] = true
		}
	}

	if currentItemCount+int64(len(newItemTypes)) > 100 {
		msg := fmt.Sprintf("仓库已满：已有 %d 种，本次新增 %d 种会超出 100 上限", currentItemCount, len(newItemTypes))
		fmt.Printf("[AddLoot] full-new: %s\n", msg)
		_ = s.logLoot(ctx, warehouse.ID, req, "failed", msg)
		return xerrors.New(xerrors.CodeInvalidParams, msg).WithMetadata("user_message", msg)
	}

	// 4. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	// 5. 添加金币
	if req.Gold > 0 {
		if err := s.teamWarehouseRepo.AddGold(ctx, tx, warehouse.ID, req.Gold); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "添加金币失败")
		}
	}

	// 6. 添加物品
	for _, item := range req.Items {
		// 检查堆叠上限
		existingCount, _ := s.teamWarehouseItemRepo.GetItemCount(ctx, warehouse.ID, item.ItemID)
		if existingCount+item.Quantity > 999 {
			tx.Rollback()
			_ = s.logLoot(ctx, warehouse.ID, req, "failed", "仓库物品堆叠超限")
			return xerrors.New(xerrors.CodeInvalidParams, "仓库物品堆叠超限")
		}

		if err := s.teamWarehouseItemRepo.AddItem(ctx, tx, warehouse.ID, item.ItemID, item.ItemType, item.Quantity, &req.SourceDungeonID); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "添加物品失败")
		}
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	_ = s.logLoot(ctx, warehouse.ID, req, "success", "")
	s.notifyLoot(ctx, warehouse.ID, req, "success", "")

	return nil
}

// GetWarehouseItems 获取仓库物品列表（团队成员可查看）
func (s *TeamWarehouseService) GetWarehouseItems(ctx context.Context, teamID, heroID string, limit, offset int) ([]*game_runtime.TeamWarehouseItem, int64, error) {
	// 1. 验证参数
	if teamID == "" || heroID == "" {
		return nil, 0, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查权限（团队成员即可）
	if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID); err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}

	// 3. 获取仓库
	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}

	// 4. 查询物品列表
	items, total, err := s.teamWarehouseItemRepo.ListByWarehouse(ctx, warehouse.ID, limit, offset)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询仓库物品列表失败")
	}

	return items, total, nil
}

// GetDistributionHistory 查看分配历史
func (s *TeamWarehouseService) GetDistributionHistory(ctx context.Context, teamID, heroID string, startAt, endAt *string, limit, offset int) ([]*interfaces.TeamLootHistoryRow, int64, error) {
	if teamID == "" || heroID == "" {
		return nil, 0, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	// 权限：队长或管理员
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" && member.Role != "admin" {
		return nil, 0, xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以查看分配历史")
	}
	if s.lootHistoryRepo == nil {
		return nil, 0, xerrors.New(xerrors.CodeInternalError, "分配历史存储未配置")
	}
	return s.lootHistoryRepo.ListByTeam(ctx, teamID, startAt, endAt, limit, offset)
}

// getLocationMaxSlots 读取容量配置
func (s *TeamWarehouseService) getLocationMaxSlots(ctx context.Context, location string) (int64, error) {
	var maxSlots int64
	err := s.db.QueryRowContext(ctx, `SELECT max_slots FROM game_config.inventory_capacities WHERE location = $1`, location).Scan(&maxSlots)
	if err == sql.ErrNoRows {
		return 0, xerrors.New(xerrors.CodeInternalError, "未配置容量")
	}
	if err != nil {
		return 0, err
	}
	return maxSlots, nil
}

func (s *TeamWarehouseService) countHeroBackpack(ctx context.Context, heroID string) (int64, error) {
	var cnt int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM game_runtime.player_items WHERE hero_id = $1 AND item_location = 'backpack' AND deleted_at IS NULL`, heroID).Scan(&cnt)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

// logLoot 写入入库审计日志（最佳努力）
func (s *TeamWarehouseService) logLoot(ctx context.Context, warehouseID string, req *AddLootToWarehouseRequest, result string, reason string) error {
	if s.lootLogRepo == nil {
		return nil
	}
	itemsJSON, _ := json.Marshal(req.Items)
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	logReq := &interfaces.TeamWarehouseLootLog{
		TeamID:          req.TeamID,
		WarehouseID:     warehouseID,
		SourceDungeonID: &req.SourceDungeonID,
		GoldAmount:      req.Gold,
		ItemsJSON:       string(itemsJSON),
		Result:          result,
		Reason:          reasonPtr,
	}
	return s.lootLogRepo.Log(ctx, nil, logReq)
}

func (s *TeamWarehouseService) notifyLoot(ctx context.Context, warehouseID string, req *AddLootToWarehouseRequest, result string, reason string) {
	event := &LootEvent{
		TeamID:          req.TeamID,
		WarehouseID:     warehouseID,
		SourceDungeonID: req.SourceDungeonID,
		Gold:            req.Gold,
		Items:           req.Items,
		Result:          result,
		Reason:          reason,
	}
	_ = notify.PublishWarehouseEvent(ctx, notify.SubjectWarehouseLoot, event)
}
