package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

type ActionTypeDefinitionService struct {
	repo interfaces.ActionTypeDefinitionRepository
}

func NewActionTypeDefinitionService(db *sql.DB) *ActionTypeDefinitionService {
	return &ActionTypeDefinitionService{
		repo: impl.NewActionTypeDefinitionRepository(db),
	}
}

func (s *ActionTypeDefinitionService) GetList(ctx context.Context, params interfaces.ActionTypeDefinitionQueryParams) ([]*game_config.ActionTypeDefinition, int64, error) {
	return s.repo.List(ctx, params)
}

func (s *ActionTypeDefinitionService) GetByID(ctx context.Context, id string) (*game_config.ActionTypeDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ActionTypeDefinitionService) GetAll(ctx context.Context) ([]*game_config.ActionTypeDefinition, error) {
	return s.repo.GetAll(ctx)
}
