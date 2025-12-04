package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"

	"tsu-self/internal/repository/interfaces"
)

type heroWalletRepositoryImpl struct {
	db *sql.DB
}

// NewHeroWalletRepository 创建英雄钱包仓储实例
func NewHeroWalletRepository(db *sql.DB) interfaces.HeroWalletRepository {
	return &heroWalletRepositoryImpl{db: db}
}

func (r *heroWalletRepositoryImpl) AddGold(ctx context.Context, heroID string, amount int64) error {
	return r.AddGoldTx(ctx, r.db, heroID, amount)
}

func (r *heroWalletRepositoryImpl) AddGoldTx(ctx context.Context, execer boil.ContextExecutor, heroID string, amount int64) error {
	if heroID == "" {
		return fmt.Errorf("hero_id 不能为空")
	}
	// 插入或累加，确保不为负
	query := `
INSERT INTO game_runtime.hero_wallets (hero_id, gold_amount)
VALUES ($1, GREATEST($2,0))
ON CONFLICT (hero_id) DO UPDATE
SET gold_amount = GREATEST(game_runtime.hero_wallets.gold_amount + $2, 0),
    updated_at = NOW()
`
	_, err := execer.ExecContext(ctx, query, heroID, amount)
	if err != nil {
		return fmt.Errorf("更新英雄钱包失败: %w", err)
	}
	return nil
}

func (r *heroWalletRepositoryImpl) GetBalance(ctx context.Context, heroID string) (int64, error) {
	if heroID == "" {
		return 0, fmt.Errorf("hero_id 不能为空")
	}
	var balance int64
	err := r.db.QueryRowContext(ctx, `SELECT gold_amount FROM game_runtime.hero_wallets WHERE hero_id = $1`, heroID).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("查询英雄钱包失败: %w", err)
	}
	return balance, nil
}
