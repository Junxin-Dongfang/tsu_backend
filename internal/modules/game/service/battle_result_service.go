// Package service 聚合游戏服的业务服务实现，包含战斗结果落地等逻辑。
package service

import (
	"context"
	"encoding/json"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// BattleResultInput 为战斗结果回调提供统一的服务层结构。
type BattleResultInput struct {
	BattleID     string
	BattleCode   string
	TeamID       string
	DungeonID    string
	HeroID       string
	ResultStatus string
	Participants interface{}
	Events       interface{}
	RawPayload   interface{}
	Loot         LootData
}

type dungeonCompleter interface {
	CompleteDungeon(ctx context.Context, req *CompleteDungeonRequest) (*game_runtime.TeamDungeonProgress, error)
}

// BattleResultService 负责记录战斗结果并驱动后续掉落逻辑。
type BattleResultService struct {
	battleReportRepo interfaces.BattleReportRepository
	dungeonService   dungeonCompleter
}

// NewBattleResultService 构造函数。
func NewBattleResultService(repo interfaces.BattleReportRepository, dungeonService dungeonCompleter) *BattleResultService {
	return &BattleResultService{
		battleReportRepo: repo,
		dungeonService:   dungeonService,
	}
}

// RecordAndComplete 记录战斗结果并在满足条件时完成地城。
func (s *BattleResultService) RecordAndComplete(ctx context.Context, input *BattleResultInput) (*game_runtime.TeamDungeonProgress, error) {
	if input == nil || input.BattleID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "battle_id 不能为空")
	}
	if err := s.persistBattleReport(ctx, input); err != nil {
		return nil, err
	}

	// 非地城胜利场景仅记录战报。
	if input.ResultStatus != "victory" || input.TeamID == "" || input.DungeonID == "" {
		return nil, nil
	}

	progress, err := s.dungeonService.CompleteDungeon(ctx, &CompleteDungeonRequest{
		TeamID:    input.TeamID,
		HeroID:    input.HeroID,
		DungeonID: input.DungeonID,
		Loot:      input.Loot,
	})
	if err != nil {
		return nil, err
	}
	return progress, nil
}

func (s *BattleResultService) persistBattleReport(ctx context.Context, input *BattleResultInput) error {
	participantsJSON, err := json.Marshal(input.Participants)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInvalidParams, "解析参与者失败")
	}
	eventsJSON, err := json.Marshal(input.Events)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInvalidParams, "解析事件日志失败")
	}
	rawPayloadJSON, err := json.Marshal(input.RawPayload)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInvalidParams, "序列化战斗 payload 失败")
	}
	lootJSON, err := json.Marshal(input.Loot.Items)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInvalidParams, "序列化战利品失败")
	}

	report := &interfaces.BattleReport{
		BattleID:     input.BattleID,
		BattleCode:   input.BattleCode,
		TeamID:       input.TeamID,
		DungeonID:    input.DungeonID,
		ResultStatus: input.ResultStatus,
		LootGold:     input.Loot.Gold,
		LootItems:    lootJSON,
		Participants: participantsJSON,
		Events:       eventsJSON,
		RawPayload:   rawPayloadJSON,
	}
	if err := s.battleReportRepo.Create(ctx, report); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "写入战斗回调失败")
	}
	return nil
}
