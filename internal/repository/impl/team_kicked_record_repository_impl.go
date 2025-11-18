package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type teamKickedRecordRepositoryImpl struct {
	db *sql.DB
}

// NewTeamKickedRecordRepository 创建团队踢出记录仓储实例
func NewTeamKickedRecordRepository(db *sql.DB) interfaces.TeamKickedRecordRepository {
	return &teamKickedRecordRepositoryImpl{db: db}
}

// Create 创建踢出记录
func (r *teamKickedRecordRepositoryImpl) Create(ctx context.Context, record *game_runtime.TeamKickedRecord) error {
	// 生成UUID
	if record.ID == "" {
		record.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	record.KickedAt = now
	if record.CooldownUntil.IsZero() {
		record.CooldownUntil = now.Add(24 * time.Hour) // 默认24小时冷却期
	}

	// 插入数据库
	if err := record.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("创建踢出记录失败: %w", err)
	}

	return nil
}

// CheckCooldown 检查冷却期
func (r *teamKickedRecordRepositoryImpl) CheckCooldown(ctx context.Context, teamID, heroID string) (bool, error) {
	// 查询是否存在未过冷却期的记录
	count, err := game_runtime.TeamKickedRecords(
		qm.Where("team_id = ? AND hero_id = ? AND cooldown_until > ?", teamID, heroID, time.Now()),
	).Count(ctx, r.db)

	if err != nil {
		return false, fmt.Errorf("检查冷却期失败: %w", err)
	}

	return count > 0, nil
}

// GetLatestByTeamAndHero 获取最新的踢出记录
func (r *teamKickedRecordRepositoryImpl) GetLatestByTeamAndHero(ctx context.Context, teamID, heroID string) (*game_runtime.TeamKickedRecord, error) {
	record, err := game_runtime.TeamKickedRecords(
		qm.Where("team_id = ? AND hero_id = ?", teamID, heroID),
		qm.OrderBy("kicked_at DESC"),
		qm.Limit(1),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil // 没有记录，返回 nil 而不是错误
	}
	if err != nil {
		return nil, fmt.Errorf("查询踢出记录失败: %w", err)
	}

	return record, nil
}

