package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// DungeonBattleService 战斗配置服务
type DungeonBattleService struct {
	battleRepo  interfaces.DungeonBattleRepository
	monsterRepo interfaces.MonsterRepository
	buffRepo    interfaces.BuffRepository
	db          *sql.DB
}

// NewDungeonBattleService 创建战斗配置服务
func NewDungeonBattleService(db *sql.DB) *DungeonBattleService {
	return &DungeonBattleService{
		battleRepo:  impl.NewDungeonBattleRepository(db),
		monsterRepo: impl.NewMonsterRepository(db),
		buffRepo:    impl.NewBuffRepository(db),
		db:          db,
	}
}

// GetBattleByID 根据ID获取战斗配置
func (s *DungeonBattleService) GetBattleByID(ctx context.Context, battleID string) (*game_config.DungeonBattle, error) {
	return s.battleRepo.GetByID(ctx, battleID)
}

// GetBattleByCode 根据代码获取战斗配置
func (s *DungeonBattleService) GetBattleByCode(ctx context.Context, code string) (*game_config.DungeonBattle, error) {
	return s.battleRepo.GetByCode(ctx, code)
}

// CreateBattle 创建战斗配置
func (s *DungeonBattleService) CreateBattle(ctx context.Context, req *dto.CreateBattleRequest) (*game_config.DungeonBattle, error) {
	// 验证战斗代码唯一性
	exists, err := s.battleRepo.Exists(ctx, req.BattleCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("战斗配置代码已存在: %s", req.BattleCode))
	}

	// 验证怪物配置
	if err := s.validateMonsterSetup(ctx, req.MonsterSetup); err != nil {
		return nil, err
	}

	// 验证全程Buff
	if err := s.validateGlobalBuffs(ctx, req.GlobalBuffs); err != nil {
		return nil, err
	}

	// 构建战斗配置实体
	battle := &game_config.DungeonBattle{
		BattleCode: req.BattleCode,
		IsActive:   req.IsActive,
	}

	// 序列化场地配置
	if req.LocationConfig != nil {
		locationJSON, err := json.Marshal(req.LocationConfig)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化场地配置失败")
		}
		battle.LocationConfig = types.JSON(locationJSON)
	} else {
		battle.LocationConfig = types.JSON([]byte("{}"))
	}

	// 序列化全程Buff
	globalBuffsJSON, err := json.Marshal(req.GlobalBuffs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化全程Buff失败")
	}
	battle.GlobalBuffs = types.JSON(globalBuffsJSON)

	// 序列化怪物配置
	monsterSetupJSON, err := json.Marshal(req.MonsterSetup)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化怪物配置失败")
	}
	battle.MonsterSetup = types.JSON(monsterSetupJSON)

	// 设置描述文本
	if req.BattleStartDesc != nil {
		battle.BattleStartDesc = null.StringFrom(*req.BattleStartDesc)
	}
	if req.BattleSuccessDesc != nil {
		battle.BattleSuccessDesc = null.StringFrom(*req.BattleSuccessDesc)
	}
	if req.BattleFailureDesc != nil {
		battle.BattleFailureDesc = null.StringFrom(*req.BattleFailureDesc)
	}

	// 创建战斗配置
	if err := s.battleRepo.Create(ctx, battle); err != nil {
		return nil, err
	}

	return battle, nil
}

// UpdateBattle 更新战斗配置
func (s *DungeonBattleService) UpdateBattle(ctx context.Context, battleID string, req *dto.UpdateBattleRequest) (*game_config.DungeonBattle, error) {
	// 获取战斗配置
	battle, err := s.battleRepo.GetByID(ctx, battleID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.LocationConfig != nil {
		locationJSON, err := json.Marshal(req.LocationConfig)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化场地配置失败")
		}
		battle.LocationConfig = types.JSON(locationJSON)
	}

	if req.GlobalBuffs != nil {
		globalBuffsJSON, err := json.Marshal(req.GlobalBuffs)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化全程Buff失败")
		}
		battle.GlobalBuffs = types.JSON(globalBuffsJSON)
	}

	if req.MonsterSetup != nil {
		// 验证怪物配置
		if err := s.validateMonsterSetup(ctx, req.MonsterSetup); err != nil {
			return nil, err
		}

		monsterSetupJSON, err := json.Marshal(req.MonsterSetup)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化怪物配置失败")
		}
		battle.MonsterSetup = types.JSON(monsterSetupJSON)
	}

	if req.BattleStartDesc != nil {
		battle.BattleStartDesc = null.StringFrom(*req.BattleStartDesc)
	}
	if req.BattleSuccessDesc != nil {
		battle.BattleSuccessDesc = null.StringFrom(*req.BattleSuccessDesc)
	}
	if req.BattleFailureDesc != nil {
		battle.BattleFailureDesc = null.StringFrom(*req.BattleFailureDesc)
	}

	if req.IsActive != nil {
		battle.IsActive = *req.IsActive
	}

	// 更新战斗配置
	if err := s.battleRepo.Update(ctx, battle); err != nil {
		return nil, err
	}

	return battle, nil
}

// DeleteBattle 删除战斗配置
func (s *DungeonBattleService) DeleteBattle(ctx context.Context, battleID string) error {
	return s.battleRepo.Delete(ctx, battleID)
}

// validateMonsterSetup 验证怪物配置
func (s *DungeonBattleService) validateMonsterSetup(ctx context.Context, setup []dto.MonsterSetupItem) error {
	if len(setup) == 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "怪物配置不能为空")
	}

	positionMap := make(map[int]bool)
	for _, item := range setup {
		// 验证位置范围
		if item.Position < 1 || item.Position > 21 {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("怪物位置必须在1-21之间: %d", item.Position))
		}

		// 验证位置唯一性
		if positionMap[item.Position] {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("怪物位置重复: %d", item.Position))
		}
		positionMap[item.Position] = true

		// 验证怪物存在性
		_, err := s.monsterRepo.GetByCode(ctx, item.MonsterCode)
		if err != nil {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("怪物不存在: %s", item.MonsterCode))
		}
	}

	return nil
}

// validateGlobalBuffs 验证全程Buff配置
func (s *DungeonBattleService) validateGlobalBuffs(ctx context.Context, buffs []dto.GlobalBuffItem) error {
	for _, buff := range buffs {
		// 验证Buff存在性
		_, err := s.buffRepo.GetByCode(ctx, buff.BuffCode)
		if err != nil {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("Buff不存在: %s", buff.BuffCode))
		}

		// 验证目标类型
		validTargets := map[string]bool{
			"all_heroes":   true,
			"all_monsters": true,
			"all":          true,
		}
		if !validTargets[buff.Target] {
			return xerrors.New(xerrors.CodeInvalidParams,
				fmt.Sprintf("无效的Buff目标类型: %s (有效值: all_heroes, all_monsters, all)", buff.Target))
		}
	}

	return nil
}

