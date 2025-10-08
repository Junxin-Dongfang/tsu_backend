package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

type RangeConfigRuleService struct {
	repo interfaces.RangeConfigRuleRepository
}

func NewRangeConfigRuleService(db *sql.DB) *RangeConfigRuleService {
	return &RangeConfigRuleService{
		repo: impl.NewRangeConfigRuleRepository(db),
	}
}

func (s *RangeConfigRuleService) GetList(ctx context.Context, params interfaces.RangeConfigRuleQueryParams) ([]*game_config.RangeConfigRule, int64, error) {
	return s.repo.List(ctx, params)
}

func (s *RangeConfigRuleService) GetByID(ctx context.Context, id string) (*game_config.RangeConfigRule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RangeConfigRuleService) GetAll(ctx context.Context) ([]*game_config.RangeConfigRule, error) {
	return s.repo.GetAll(ctx)
}
