package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/game_runtime"
)

// HeroActivationService 管理英雄激活和当前操作英雄上下文
type HeroActivationService struct {
	db *sql.DB
}

// NewHeroActivationService 创建新的英雄激活服务实例
func NewHeroActivationService(db *sql.DB) *HeroActivationService {
	return &HeroActivationService{db: db}
}

// ActivateHero 激活指定的英雄
// 如果用户没有当前操作英雄，则自动设置为当前
func (s *HeroActivationService) ActivateHero(ctx context.Context, userID, heroID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 验证英雄属于该用户且未删除
	var count int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_runtime.heroes
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		heroID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("hero not found or doesn't belong to user")
	}

	// 设置英雄为已激活
	_, err = tx.ExecContext(ctx,
		`UPDATE game_runtime.heroes SET is_activated = TRUE WHERE id = $1`,
		heroID)
	if err != nil {
		return err
	}

	// 检查用户是否有当前操作英雄
	var currentHeroID sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT hero_id FROM game_runtime.current_hero_contexts WHERE user_id = $1`,
		userID).Scan(&currentHeroID)

	if err == sql.ErrNoRows {
		// 没有当前操作英雄，将此英雄设为当前
		_, err = tx.ExecContext(ctx,
			`INSERT INTO game_runtime.current_hero_contexts (user_id, hero_id, switched_at)
			 VALUES ($1, $2, NOW())`,
			userID, heroID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return tx.Commit()
}

// DeactivateHero 停用指定的英雄（不能是当前操作英雄）
func (s *HeroActivationService) DeactivateHero(ctx context.Context, userID, heroID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 验证英雄属于该用户且未删除
	var count int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_runtime.heroes
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		heroID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("hero not found or doesn't belong to user")
	}

	// 检查是否是当前操作英雄
	var currentHeroID sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT hero_id FROM game_runtime.current_hero_contexts WHERE user_id = $1`,
		userID).Scan(&currentHeroID)

	if err != sql.ErrNoRows && err != nil {
		return err
	}

	if currentHeroID.Valid && currentHeroID.String == heroID {
		return fmt.Errorf("cannot deactivate current hero, switch to another hero first")
	}

	// 设置英雄为未激活
	_, err = tx.ExecContext(ctx,
		`UPDATE game_runtime.heroes SET is_activated = FALSE WHERE id = $1`,
		heroID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SwitchCurrentHero 切换当前操作英雄（目标英雄必须已激活）
func (s *HeroActivationService) SwitchCurrentHero(ctx context.Context, userID, heroID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 验证英雄属于该用户、已激活且未删除
	var count int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_runtime.heroes
		 WHERE id = $1 AND user_id = $2 AND is_activated = TRUE AND deleted_at IS NULL`,
		heroID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("hero not found, doesn't belong to user, or is not activated")
	}

	// UPSERT current_hero_contexts
	_, err = tx.ExecContext(ctx,
		`INSERT INTO game_runtime.current_hero_contexts (user_id, hero_id, switched_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET hero_id = $2, switched_at = NOW()`,
		userID, heroID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetActivatedHeroes 获取用户已激活的所有英雄，并标记当前操作英雄
func (s *HeroActivationService) GetActivatedHeroes(ctx context.Context, userID string) ([]*game_runtime.Hero, string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT h.id, h.user_id, h.class_id, h.hero_name, h.created_at, h.updated_at, h.deleted_at, h.is_activated
		 FROM game_runtime.heroes h
		 WHERE h.user_id = $1 AND h.is_activated = TRUE AND h.deleted_at IS NULL
		 ORDER BY h.created_at ASC`,
		userID)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	heroes := make([]*game_runtime.Hero, 0)
	for rows.Next() {
		hero := &game_runtime.Hero{}
		err = rows.Scan(&hero.ID, &hero.UserID, &hero.ClassID, &hero.HeroName, &hero.CreatedAt, &hero.UpdatedAt, &hero.DeletedAt, &hero.IsActivated)
		if err != nil {
			return nil, "", err
		}
		heroes = append(heroes, hero)
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	// 获取当前操作英雄
	var currentHeroID sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT hero_id FROM game_runtime.current_hero_contexts WHERE user_id = $1`,
		userID).Scan(&currentHeroID)

	if err != nil && err != sql.ErrNoRows {
		return nil, "", err
	}

	if !currentHeroID.Valid {
		currentHeroID.String = ""
	}

	return heroes, currentHeroID.String, nil
}

// GetCurrentHero 获取用户的当前操作英雄
func (s *HeroActivationService) GetCurrentHero(ctx context.Context, userID string) (*game_runtime.Hero, error) {
	hero := &game_runtime.Hero{}
	err := s.db.QueryRowContext(ctx,
		`SELECT h.id, h.user_id, h.class_id, h.hero_name, h.created_at, h.updated_at, h.deleted_at, h.is_activated
		 FROM game_runtime.heroes h
		 INNER JOIN game_runtime.current_hero_contexts c ON h.id = c.hero_id
		 WHERE c.user_id = $1`,
		userID).Scan(&hero.ID, &hero.UserID, &hero.ClassID, &hero.HeroName, &hero.CreatedAt, &hero.UpdatedAt, &hero.DeletedAt, &hero.IsActivated)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no current hero found for user")
	} else if err != nil {
		return nil, err
	}

	return hero, nil
}
