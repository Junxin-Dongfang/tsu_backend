package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

type FormulaVariableService struct {
	repo interfaces.FormulaVariableRepository
}

func NewFormulaVariableService(db *sql.DB) *FormulaVariableService {
	return &FormulaVariableService{
		repo: impl.NewFormulaVariableRepository(db),
	}
}

func (s *FormulaVariableService) GetList(ctx context.Context, params interfaces.FormulaVariableQueryParams) ([]*game_config.FormulaVariable, int64, error) {
	return s.repo.List(ctx, params)
}

func (s *FormulaVariableService) GetByID(ctx context.Context, id string) (*game_config.FormulaVariable, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *FormulaVariableService) GetAll(ctx context.Context) ([]*game_config.FormulaVariable, error) {
	return s.repo.GetAll(ctx)
}
