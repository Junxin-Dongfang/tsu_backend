package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

type RangeConfigRuleQueryParams struct {
	IsActive *bool
	Limit    int
	Offset   int
}

type RangeConfigRuleRepository interface {
	GetByID(ctx context.Context, id string) (*game_config.RangeConfigRule, error)
	List(ctx context.Context, params RangeConfigRuleQueryParams) ([]*game_config.RangeConfigRule, int64, error)
	GetAll(ctx context.Context) ([]*game_config.RangeConfigRule, error)
}
