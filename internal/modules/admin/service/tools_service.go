package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ToolsService 提供测试/运维辅助操作（仅在受控环境启用）。
type ToolsService struct {
	db                    *sql.DB
	itemRepo              interfaces.ItemRepository
	playerItemRepo        interfaces.PlayerItemRepository
	teamWarehouseRepo     interfaces.TeamWarehouseRepository
	teamWarehouseItemRepo interfaces.TeamWarehouseItemRepository
	heroRepo              interfaces.HeroRepository
}

func NewToolsService(db *sql.DB) *ToolsService {
	return &ToolsService{
		db:                    db,
		itemRepo:              impl.NewItemRepository(db),
		playerItemRepo:        impl.NewPlayerItemRepository(db),
		teamWarehouseRepo:     impl.NewTeamWarehouseRepository(db),
		teamWarehouseItemRepo: impl.NewTeamWarehouseItemRepository(db),
		heroRepo:              impl.NewHeroRepository(db),
	}
}

// GrantItem 发放物品到玩家背包或团队仓库。
func (s *ToolsService) GrantItem(ctx context.Context, req *dto.GrantItemRequest) (*dto.GrantItemResponse, error) {
	if req.Quantity <= 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "数量必须大于0")
	}

	it, err := s.itemRepo.GetByID(ctx, req.ItemID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "物品不存在")
	}

	itemType := it.ItemType
	switch req.TargetType {
	case "user":
		return s.grantToUser(ctx, req.TargetID, itemType, req.ItemID, req.Quantity)
	case "team_warehouse":
		return s.grantToTeamWarehouse(ctx, req.TargetID, itemType, req.ItemID, req.Quantity)
	default:
		return nil, xerrors.New(xerrors.CodeInvalidParams, "目标类型不支持")
	}
}

func (s *ToolsService) grantToUser(ctx context.Context, userID, itemType, itemID string, quantity int) (*dto.GrantItemResponse, error) {
	if userID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "用户ID不能为空")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	for i := 0; i < quantity; i++ {
		now := time.Now()
		item := &game_runtime.PlayerItem{
			ID:           uuid.New().String(),
			ItemID:       itemID,
			OwnerID:      userID,
			SourceType:   "admin_grant",
			ItemLocation: "backpack",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.playerItemRepo.Create(ctx, tx, item); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "发放到背包失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return &dto.GrantItemResponse{Granted: quantity}, nil
}

func (s *ToolsService) grantToTeamWarehouse(ctx context.Context, teamID, itemType, itemID string, quantity int) (*dto.GrantItemResponse, error) {
	if teamID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "团队ID不能为空")
	}

	warehouse, err := s.teamWarehouseRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	for i := 0; i < quantity; i++ {
		if err := s.teamWarehouseItemRepo.AddItem(ctx, tx, warehouse.ID, itemID, itemType, 1, nil); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "添加仓库物品失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return &dto.GrantItemResponse{Granted: quantity}, nil
}

// GrantGold 向团队仓库添加金币
func (s *ToolsService) GrantGold(ctx context.Context, req *dto.GrantGoldRequest) (*dto.GrantGoldResponse, error) {
	if req.Amount <= 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "金币数量必须大于0")
	}
	wh, err := s.teamWarehouseRepo.GetByTeamID(ctx, req.TeamID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "团队仓库不存在")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	if err := s.teamWarehouseRepo.AddGold(ctx, tx, wh.ID, req.Amount); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "添加金币失败")
	}
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}
	return &dto.GrantGoldResponse{Added: req.Amount}, nil
}

// GrantExperience 向英雄添加经验（可用于测试）
func (s *ToolsService) GrantExperience(ctx context.Context, req *dto.GrantExperienceRequest) (*dto.GrantExperienceResponse, error) {
	if req.Amount <= 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "经验必须大于0")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, req.HeroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "英雄不存在")
	}

	hero.ExperienceTotal += req.Amount
	hero.ExperienceAvailable += req.Amount
	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄经验失败")
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}
	return &dto.GrantExperienceResponse{Added: req.Amount}, nil
}
