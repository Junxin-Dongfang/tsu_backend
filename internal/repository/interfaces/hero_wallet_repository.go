package interfaces

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// HeroWalletRepository 英雄钱包仓储接口
type HeroWalletRepository interface {
	// AddGold 为英雄增加金币（可为负，需确保不小于0）
	AddGold(ctx context.Context, heroID string, amount int64) error
	// AddGoldTx 在事务内为英雄增加金币
	AddGoldTx(ctx context.Context, tx boil.ContextExecutor, heroID string, amount int64) error
	// GetBalance 获取英雄金币余额
	GetBalance(ctx context.Context, heroID string) (int64, error)
}
