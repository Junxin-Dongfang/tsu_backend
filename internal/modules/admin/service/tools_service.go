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
}

func NewToolsService(db *sql.DB) *ToolsService {
	return &ToolsService{
		db:                    db,
		itemRepo:              impl.NewItemRepository(db),
		playerItemRepo:        impl.NewPlayerItemRepository(db),
		teamWarehouseRepo:     impl.NewTeamWarehouseRepository(db),
		teamWarehouseItemRepo: impl.NewTeamWarehouseItemRepository(db),
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
