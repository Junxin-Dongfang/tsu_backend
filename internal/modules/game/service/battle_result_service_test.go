package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type fakeBattleReportRepo struct {
	reports []*interfaces.BattleReport
	err     error
}

func (f *fakeBattleReportRepo) Create(ctx context.Context, report *interfaces.BattleReport) error {
	if f.err != nil {
		return f.err
	}
	f.reports = append(f.reports, report)
	return nil
}

type fakeDungeonCompleter struct {
	called  bool
	lastReq *CompleteDungeonRequest
	result  *game_runtime.TeamDungeonProgress
	err     error
}

func (f *fakeDungeonCompleter) CompleteDungeon(ctx context.Context, req *CompleteDungeonRequest) (*game_runtime.TeamDungeonProgress, error) {
	f.called = true
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}
	if f.result == nil {
		f.result = &game_runtime.TeamDungeonProgress{ID: "progress", TeamID: req.TeamID, DungeonID: req.DungeonID}
	}
	return f.result, nil
}

func TestBattleResultServiceRecordsNonVictory(t *testing.T) {
	repo := &fakeBattleReportRepo{}
	dungeon := &fakeDungeonCompleter{}
	svc := NewBattleResultService(repo, dungeon)

	input := &BattleResultInput{
		BattleID:     "battle-1",
		ResultStatus: "defeat",
		TeamID:       "team",
		DungeonID:    "dungeon",
		Participants: []map[string]string{{"hero_id": "h1"}},
		Loot:         LootData{Gold: 5},
	}
	progress, err := svc.RecordAndComplete(context.Background(), input)
	require.NoError(t, err)
	require.Nil(t, progress)
	require.Len(t, repo.reports, 1)
	require.False(t, dungeon.called, "defeat 不应触发 CompleteDungeon")
	require.Equal(t, "battle-1", repo.reports[0].BattleID)
	require.EqualValues(t, 5, repo.reports[0].LootGold)
}

func TestBattleResultServiceCompletesDungeonOnVictory(t *testing.T) {
	repo := &fakeBattleReportRepo{}
	dungeon := &fakeDungeonCompleter{}
	svc := NewBattleResultService(repo, dungeon)

	input := &BattleResultInput{
		BattleID:     "battle-2",
		BattleCode:   "room-1",
		ResultStatus: "victory",
		TeamID:       "team-1",
		DungeonID:    "dungeon-1",
		HeroID:       "hero-1",
		Participants: []map[string]string{{"hero_id": "hero-1"}},
		Events:       []map[string]interface{}{{"action": "cast", "value": 100}},
		Loot: LootData{
			Gold: 20,
			Items: []LootItem{
				{ItemID: "item-1", ItemType: "equipment", Quantity: 1},
			},
		},
	}

	progress, err := svc.RecordAndComplete(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, progress)
	require.True(t, dungeon.called)
	require.Equal(t, "team-1", dungeon.lastReq.TeamID)
	require.Equal(t, "dungeon-1", dungeon.lastReq.DungeonID)
	require.Equal(t, "hero-1", dungeon.lastReq.HeroID)
	require.Len(t, repo.reports, 1)
	require.Contains(t, string(repo.reports[0].Participants), "hero-1")
	require.Contains(t, string(repo.reports[0].LootItems), "item-1")
}

func TestBattleResultServiceRequiresBattleID(t *testing.T) {
	repo := &fakeBattleReportRepo{}
	svc := NewBattleResultService(repo, &fakeDungeonCompleter{})

	_, err := svc.RecordAndComplete(context.Background(), nil)
	require.Error(t, err)

	_, err = svc.RecordAndComplete(context.Background(), &BattleResultInput{})
	require.Error(t, err)
	require.Empty(t, repo.reports)
}

func TestBattleResultServiceHandlesMarshalError(t *testing.T) {
	repo := &fakeBattleReportRepo{}
	svc := NewBattleResultService(repo, &fakeDungeonCompleter{})

	input := &BattleResultInput{
		BattleID:     "battle-3",
		ResultStatus: "victory",
		TeamID:       "t",
		DungeonID:    "d",
		Participants: []interface{}{make(chan int)},
	}
	_, err := svc.RecordAndComplete(context.Background(), input)
	require.Error(t, err)
	require.Empty(t, repo.reports)
}

func TestBattleResultServiceHandlesEventMarshalError(t *testing.T) {
	repo := &fakeBattleReportRepo{}
	svc := NewBattleResultService(repo, &fakeDungeonCompleter{})

	input := &BattleResultInput{
		BattleID:     "battle-4",
		ResultStatus: "victory",
		TeamID:       "t",
		DungeonID:    "d",
		Participants: []map[string]string{{"hero_id": "h1"}},
		Events:       []interface{}{make(chan int)},
	}
	_, err := svc.RecordAndComplete(context.Background(), input)
	require.Error(t, err)
	require.Empty(t, repo.reports)
}

func TestBattleResultServiceHandlesRawPayloadMarshalError(t *testing.T) {
	repo := &fakeBattleReportRepo{}
	svc := NewBattleResultService(repo, &fakeDungeonCompleter{})

	input := &BattleResultInput{
		BattleID:     "battle-5",
		ResultStatus: "victory",
		TeamID:       "t",
		DungeonID:    "d",
		Participants: []map[string]string{{"hero_id": "h1"}},
		Events:       []map[string]string{{"action": "cast"}},
		RawPayload:   make(chan int),
	}
	_, err := svc.RecordAndComplete(context.Background(), input)
	require.Error(t, err)
	require.Empty(t, repo.reports)
}
