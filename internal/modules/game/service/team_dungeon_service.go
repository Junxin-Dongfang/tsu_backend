package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TeamDungeonDependencies 注入 TeamDungeonService 所需依赖
type TeamDungeonDependencies struct {
	WarehouseService *TeamWarehouseService
	TeamMemberRepo   interfaces.TeamMemberRepository
	DungeonRepo      interfaces.DungeonRepository
	ProgressRepo     interfaces.TeamDungeonProgressRepository
	RecordRepo       interfaces.TeamDungeonRecordRepository
	HeroRepo         interfaces.HeroRepository
}

// TeamDungeonService 团队地城服务
type TeamDungeonService struct {
	db                   *sql.DB
	teamMemberRepo       interfaces.TeamMemberRepository
	dungeonRepo          interfaces.DungeonRepository
	progressRepo         interfaces.TeamDungeonProgressRepository
	recordRepo           interfaces.TeamDungeonRecordRepository
	heroRepo             interfaces.HeroRepository
	teamWarehouseService *TeamWarehouseService
}

// NewTeamDungeonService 创建团队地城服务
func NewTeamDungeonService(db *sql.DB, deps *TeamDungeonDependencies) *TeamDungeonService {
	if deps == nil {
		deps = &TeamDungeonDependencies{}
	}
	if deps.TeamMemberRepo == nil {
		deps.TeamMemberRepo = impl.NewTeamMemberRepository(db)
	}
	if deps.DungeonRepo == nil {
		deps.DungeonRepo = impl.NewDungeonRepository(db)
	}
	if deps.ProgressRepo == nil {
		deps.ProgressRepo = impl.NewTeamDungeonProgressRepository(db)
	}
	if deps.RecordRepo == nil {
		deps.RecordRepo = impl.NewTeamDungeonRecordRepository(db)
	}
	if deps.HeroRepo == nil {
		deps.HeroRepo = impl.NewHeroRepository(db)
	}
	if deps.WarehouseService == nil {
		deps.WarehouseService = &TeamWarehouseService{
			db:                    db,
			teamMemberRepo:        deps.TeamMemberRepo,
			teamWarehouseRepo:     impl.NewTeamWarehouseRepository(db),
			teamWarehouseItemRepo: impl.NewTeamWarehouseItemRepository(db),
		}
	}

	return &TeamDungeonService{
		db:                   db,
		teamMemberRepo:       deps.TeamMemberRepo,
		dungeonRepo:          deps.DungeonRepo,
		progressRepo:         deps.ProgressRepo,
		recordRepo:           deps.RecordRepo,
		heroRepo:             deps.HeroRepo,
		teamWarehouseService: deps.WarehouseService,
	}
}

// SelectDungeonRequest 选择地城请求
type SelectDungeonRequest struct {
	TeamID    string
	HeroID    string // 操作者英雄ID
	DungeonID string
}

// EnterDungeonRequest 进入地城请求
type EnterDungeonRequest struct {
	TeamID    string
	HeroID    string
	DungeonID string
}

// CompleteDungeonRequest 完成地城请求
type CompleteDungeonRequest struct {
	TeamID    string
	HeroID    string
	DungeonID string
	Loot      LootData // 战利品
}

// FailDungeonRequest 地城失败请求
type FailDungeonRequest struct {
	TeamID    string
	HeroID    string
	DungeonID string
	Reason    string
}

// AbandonDungeonRequest 放弃地城请求
type AbandonDungeonRequest struct {
	TeamID    string
	HeroID    string
	DungeonID string
	Reason    string
}

// LootData 战利品数据
type LootData struct {
	Gold  int64
	Items []LootItem
}

// SelectDungeon 选择地城（队长或管理员）
func (s *TeamDungeonService) SelectDungeon(ctx context.Context, req *SelectDungeonRequest) (*game_runtime.TeamDungeonProgress, error) {
	if req.TeamID == "" || req.HeroID == "" || req.DungeonID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}

	if _, err := s.ensureLeaderOrAdmin(ctx, req.TeamID, req.HeroID); err != nil {
		return nil, err
	}

	dungeon, err := s.dungeonRepo.GetByID(ctx, req.DungeonID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "地城不存在")
	}
	if !dungeon.IsActive {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "地城暂未开放")
	}
	if err := s.checkDungeonWindow(dungeon); err != nil {
		return nil, err
	}
	if err := s.checkTeamRequirements(ctx, req.TeamID, dungeon); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	if progress, err := s.progressRepo.GetActiveByTeamForUpdate(ctx, tx, req.TeamID); err == nil && progress != nil {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "已有正在进行的地城挑战")
	} else if err != nil && !errors.Is(err, interfaces.ErrTeamDungeonProgressNotFound) {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城进度失败")
	}

	progress := &game_runtime.TeamDungeonProgress{
		TeamID:         req.TeamID,
		DungeonID:      req.DungeonID,
		Status:         "in_progress",
		StartedAt:      time.Now(),
		CompletedRooms: types.JSON([]byte("[]")),
	}
	if firstRoom := firstRoomCode(dungeon); firstRoom != "" {
		progress.CurrentRoomID = null.StringFrom(firstRoom)
	}
	if err := s.progressRepo.Create(ctx, tx, progress); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建地城进度失败")
	}

	record, err := s.recordRepo.GetByTeamAndDungeonForUpdate(ctx, tx, req.TeamID, req.DungeonID)
	if err != nil {
		if errors.Is(err, interfaces.ErrTeamDungeonRecordNotFound) {
			record = &game_runtime.TeamDungeonRecord{
				TeamID:        req.TeamID,
				DungeonID:     req.DungeonID,
				AttemptsCount: 0,
			}
			if dungeon.MaxAttemptsPerDay.Valid {
				record.MaxAttempts = null.IntFrom(int(dungeon.MaxAttemptsPerDay.Int16))
			}
			if err := s.recordRepo.Create(ctx, tx, record); err != nil {
				return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建地城记录失败")
			}
		} else {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城记录失败")
		}
	} else if dungeon.MaxAttemptsPerDay.Valid && (!record.MaxAttempts.Valid || int(dungeon.MaxAttemptsPerDay.Int16) != record.MaxAttempts.Int) {
		record.MaxAttempts = null.IntFrom(int(dungeon.MaxAttemptsPerDay.Int16))
		if err := s.recordRepo.Update(ctx, tx, record, "max_attempts"); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新地城记录失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return progress, nil
}

// EnterDungeon 进入地城
func (s *TeamDungeonService) EnterDungeon(ctx context.Context, req *EnterDungeonRequest) (*game_runtime.TeamDungeonProgress, error) {
	if req.TeamID == "" || req.HeroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if _, err := s.ensureLeaderOrAdmin(ctx, req.TeamID, req.HeroID); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	progress, err := s.progressRepo.GetActiveByTeamForUpdate(ctx, tx, req.TeamID)
	if err != nil {
		if errors.Is(err, interfaces.ErrTeamDungeonProgressNotFound) {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "请先选择地城")
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城进度失败")
	}

	if req.DungeonID != "" && progress.DungeonID != req.DungeonID {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "当前正在挑战其他地城")
	}

	record, err := s.recordRepo.GetByTeamAndDungeonForUpdate(ctx, tx, req.TeamID, progress.DungeonID)
	if err != nil && !errors.Is(err, interfaces.ErrTeamDungeonRecordNotFound) {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城记录失败")
	}
	if record != nil && record.MaxAttempts.Valid && record.AttemptsCount >= record.MaxAttempts.Int {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "挑战次数已达上限")
	}

	progress.StartedAt = time.Now()
	if err := s.progressRepo.Update(ctx, tx, progress, "started_at"); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新地城进度失败")
	}

	if record != nil {
		record.AttemptsCount++
		if err := s.recordRepo.Update(ctx, tx, record, "attempts_count"); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新地城记录失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}
	return progress, nil
}

// GetDungeonProgress 查看地城进度
func (s *TeamDungeonService) GetDungeonProgress(ctx context.Context, teamID, heroID string) (*game_runtime.TeamDungeonProgress, error) {
	if teamID == "" || heroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodePermissionDenied, "您不是该团队成员")
	}

	progress, err := s.progressRepo.GetActiveByTeam(ctx, teamID)
	if err != nil {
		if errors.Is(err, interfaces.ErrTeamDungeonProgressNotFound) {
			return nil, nil
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城进度失败")
	}
	return progress, nil
}

// CompleteDungeon 完成地城
func (s *TeamDungeonService) CompleteDungeon(ctx context.Context, req *CompleteDungeonRequest) (*game_runtime.TeamDungeonProgress, error) {
	if req.TeamID == "" || req.DungeonID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if req.HeroID != "" {
		if _, err := s.ensureLeaderOrAdmin(ctx, req.TeamID, req.HeroID); err != nil {
			return nil, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	progress, err := s.progressRepo.GetActiveByTeamForUpdate(ctx, tx, req.TeamID)
	if err != nil {
		if errors.Is(err, interfaces.ErrTeamDungeonProgressNotFound) {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "没有正在进行的地城")
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城进度失败")
	}
	if progress.DungeonID != req.DungeonID {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "当前地城与请求不一致")
	}

	progress.Status = "completed"
	progress.CompletedAt = null.TimeFrom(time.Now())
	if err := s.progressRepo.Update(ctx, tx, progress, "status", "completed_at"); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新地城进度失败")
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	s.awardLoot(ctx, req)
	return progress, nil
}

// FailDungeon 地城挑战失败
func (s *TeamDungeonService) FailDungeon(ctx context.Context, req *FailDungeonRequest) (*game_runtime.TeamDungeonProgress, error) {
	if req.TeamID == "" || req.DungeonID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if _, err := s.ensureLeaderOrAdmin(ctx, req.TeamID, req.HeroID); err != nil {
		return nil, err
	}
	return s.updateProgressStatus(ctx, req.TeamID, req.DungeonID, "failed")
}

// AbandonDungeon 放弃地城
func (s *TeamDungeonService) AbandonDungeon(ctx context.Context, req *AbandonDungeonRequest) (*game_runtime.TeamDungeonProgress, error) {
	if req.TeamID == "" || req.DungeonID == "" || req.HeroID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if _, err := s.ensureLeaderOrAdmin(ctx, req.TeamID, req.HeroID); err != nil {
		return nil, err
	}
	return s.updateProgressStatus(ctx, req.TeamID, req.DungeonID, "abandoned")
}

// GetChallengeHistory 查看历史
func (s *TeamDungeonService) GetChallengeHistory(ctx context.Context, teamID, heroID string, limit, offset int) ([]*game_runtime.TeamDungeonRecord, int64, error) {
	if teamID == "" || heroID == "" {
		return nil, 0, xerrors.New(xerrors.CodeInvalidParams, "参数不能为空")
	}
	if _, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID); err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodePermissionDenied, "您不是该团队成员")
	}

	records, total, err := s.recordRepo.ListByTeam(ctx, teamID, limit, offset)
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城记录失败")
	}
	return records, total, nil
}

func (s *TeamDungeonService) ensureLeaderOrAdmin(ctx context.Context, teamID, heroID string) (*game_runtime.TeamMember, error) {
	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodePermissionDenied, "您不是该团队成员")
	}
	if member.Role != "leader" && member.Role != "admin" {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "只有队长和管理员可以执行该操作")
	}
	return member, nil
}

func (s *TeamDungeonService) checkDungeonWindow(dungeon *game_config.Dungeon) error {
	if dungeon.IsTimeLimited {
		if !dungeon.TimeLimitStart.Valid || !dungeon.TimeLimitEnd.Valid {
			return xerrors.New(xerrors.CodeInvalidParams, "地城时间配置不正确")
		}
		now := time.Now()
		if now.Before(dungeon.TimeLimitStart.Time) || now.After(dungeon.TimeLimitEnd.Time) {
			return xerrors.New(xerrors.CodeInvalidParams, "当前不在地城开放时间")
		}
	}
	return nil
}

func (s *TeamDungeonService) checkTeamRequirements(ctx context.Context, teamID string, dungeon *game_config.Dungeon) error {
	members, err := s.teamMemberRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询成员失败")
	}
	if len(members) == 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "团队没有成员")
	}

	minMembers := minMembersFromDungeon(dungeon)
	if len(members) < minMembers {
		return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("团队人数不足，至少需要 %d 人", minMembers))
	}

	for _, member := range members {
		hero, err := s.heroRepo.GetByID(ctx, member.HeroID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "查询英雄信息失败")
		}
		if int(hero.CurrentLevel) < int(dungeon.MinLevel) {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("成员 %s 未达到进入等级", member.HeroID))
		}
	}

	return nil
}

func (s *TeamDungeonService) updateProgressStatus(ctx context.Context, teamID, dungeonID, status string) (*game_runtime.TeamDungeonProgress, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer tx.Rollback()

	progress, err := s.progressRepo.GetActiveByTeamForUpdate(ctx, tx, teamID)
	if err != nil {
		if errors.Is(err, interfaces.ErrTeamDungeonProgressNotFound) {
			return nil, xerrors.New(xerrors.CodeInvalidParams, "没有正在进行的地城")
		}
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询地城进度失败")
	}
	if progress.DungeonID != dungeonID {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "当前地城与请求不一致")
	}

	progress.Status = status
	progress.CompletedAt = null.TimeFrom(time.Now())
	if err := s.progressRepo.Update(ctx, tx, progress, "status", "completed_at"); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新地城进度失败")
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}
	return progress, nil
}

func (s *TeamDungeonService) awardLoot(ctx context.Context, req *CompleteDungeonRequest) {
	if s.teamWarehouseService == nil {
		return
	}
	if req.Loot.Gold <= 0 && len(req.Loot.Items) == 0 {
		return
	}

	items := make([]LootItem, 0, len(req.Loot.Items))
	for _, item := range req.Loot.Items {
		if item.ItemID == "" || item.Quantity <= 0 {
			continue
		}
		items = append(items, item)
	}

	addReq := &AddLootToWarehouseRequest{
		TeamID:          req.TeamID,
		SourceDungeonID: req.DungeonID,
		Gold:            req.Loot.Gold,
	}
	for _, item := range items {
		addReq.Items = append(addReq.Items, LootItem{
			ItemID:   item.ItemID,
			ItemType: item.ItemType,
			Quantity: item.Quantity,
		})
	}

	if err := s.teamWarehouseService.AddLootToWarehouse(ctx, addReq); err != nil {
		fmt.Printf("Warning: AddLootToWarehouse failed: %v\n", err)
	}
}

func minMembersFromDungeon(dungeon *game_config.Dungeon) int {
	if len(dungeon.RoomSequence) == 0 {
		return 1
	}

	var rooms []map[string]interface{}
	if err := json.Unmarshal(dungeon.RoomSequence, &rooms); err != nil || len(rooms) == 0 {
		return 1
	}
	if val, ok := rooms[0]["min_members"].(float64); ok && val >= 1 {
		return int(val)
	}
	return 1
}

func firstRoomCode(dungeon *game_config.Dungeon) string {
	if len(dungeon.RoomSequence) == 0 {
		return ""
	}
	var rooms []map[string]interface{}
	if err := json.Unmarshal(dungeon.RoomSequence, &rooms); err != nil || len(rooms) == 0 {
		return ""
	}
	if code, ok := rooms[0]["room_code"].(string); ok {
		return code
	}
	return ""
}
