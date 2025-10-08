package interfaces

import (
	"context"

	"tsu-self/internal/entity/game_config"
)

type ActionTypeDefinitionQueryParams struct {
	IsActive *bool
	Limit    int
	Offset   int
}

type ActionTypeDefinitionRepository interface {
	GetByID(ctx context.Context, id string) (*game_config.ActionTypeDefinition, error)
	GetByActionType(ctx context.Context, actionType string) (*game_config.ActionTypeDefinition, error)
	List(ctx context.Context, params ActionTypeDefinitionQueryParams) ([]*game_config.ActionTypeDefinition, int64, error)
	GetAll(ctx context.Context) ([]*game_config.ActionTypeDefinition, error)
}
