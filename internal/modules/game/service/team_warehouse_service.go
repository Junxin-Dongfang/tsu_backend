package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_runtime"
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
}

// NewTeamWarehouseService 创建团队仓库服务
func NewTeamWarehouseService(db *sql.DB) *TeamWarehouseService {
	return &TeamWarehouseService{
		db:                    db,
		teamMemberRepo:        impl.NewTeamMemberRepository(db),
		teamWarehouseRepo:     impl.NewTeamWarehouseRepository(db),
		teamWarehouseItemRepo: impl.NewTeamWarehouseItemRepository(db),
	}
}

// GetWarehouse 查看团队仓库（队长和管理员可查看）
func (s *TeamWarehouseService) GetWarehouse(ctx context.Context, teamID, heroID string) (*game_runtime.TeamWarehouse, error) {
	// 1. 验证参数
	if teamID == "" || heroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查权限（队长或管理员）
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" && member.Role != "admin" {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以查看仓库")
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
	TeamID          string
	DistributorID   string            // 分配者英雄ID
	Distributions   map[string]int64  // 接收者英雄ID -> 金币数量
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
	for _, amount := range req.Distributions {
		if amount <= 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "分配金额必须大于0")
		}
		totalAmount += amount
	}

	// 5. 检查仓库余额
	if warehouse.GoldAmount < totalAmount {
		return xerrors.New(xerrors.CodeInvalidParams, "仓库金币不足")
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

	// 8. 分配金币给成员（这里需要调用背包服务，暂时省略）
	// TODO: 调用背包服务，将金币发送到成员背包

	// 9. 记录分配历史
	// TODO: 创建分配历史记录

	// 10. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// TODO: 通知接收者

	return nil
}

// DistributeItemsRequest 分配物品请求
type DistributeItemsRequest struct {
	TeamID        string
	DistributorID string                        // 分配者英雄ID
	Distributions map[string]map[string]int     // 接收者英雄ID -> (物品ID -> 数量)
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
	for _, items := range req.Distributions {
		for itemID, quantity := range items {
			if quantity <= 0 {
				return xerrors.New(xerrors.CodeInvalidParams, "分配数量必须大于0")
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
			return xerrors.New(xerrors.CodeInvalidParams, "物品库存不足")
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

	// 8. 分配物品给成员（这里需要调用背包服务，暂时省略）
	// TODO: 调用背包服务，将物品发送到成员背包

	// 9. 记录分配历史
	// TODO: 创建分配历史记录

	// 10. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// TODO: 通知接收者

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
		return xerrors.New(xerrors.CodeInvalidParams, "仓库已满，请先分配物品")
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
			// 达到堆叠上限，记录日志但不阻止入库
			continue
		}

		if err := s.teamWarehouseItemRepo.AddItem(ctx, tx, warehouse.ID, item.ItemID, item.ItemType, item.Quantity, &req.SourceDungeonID); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "添加物品失败")
		}
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// TODO: 通知队长和管理员

	return nil
}

// GetWarehouseItems 获取仓库物品列表
func (s *TeamWarehouseService) GetWarehouseItems(ctx context.Context, teamID, heroID string, limit, offset int) ([]*game_runtime.TeamWarehouseItem, int64, error) {
	// 1. 验证参数
	if teamID == "" || heroID == "" {
		return nil, 0, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	// 2. 检查权限（队长或管理员）
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "您不是该团队成员")
	}
	if member.Role != "leader" && member.Role != "admin" {
		return nil, 0, xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以查看仓库物品")
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
