package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// EffectTypeDefinitionService 元效果类型定义服务
type EffectTypeDefinitionService struct {
	repo interfaces.EffectTypeDefinitionRepository
}

// NewEffectTypeDefinitionService 创建元效果类型定义服务
func NewEffectTypeDefinitionService(db *sql.DB) *EffectTypeDefinitionService {
	return &EffectTypeDefinitionService{
		repo: impl.NewEffectTypeDefinitionRepository(db),
	}
}

// GetList 获取元效果类型定义列表
func (s *EffectTypeDefinitionService) GetList(ctx context.Context, params interfaces.EffectTypeDefinitionQueryParams) ([]*game_config.EffectTypeDefinition, int64, error) {
	return s.repo.List(ctx, params)
}

// GetByID 根据ID获取元效果类型定义
func (s *EffectTypeDefinitionService) GetByID(ctx context.Context, id string) (*game_config.EffectTypeDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByCode 根据代码获取元效果类型定义
func (s *EffectTypeDefinitionService) GetByCode(ctx context.Context, code string) (*game_config.EffectTypeDefinition, error) {
	return s.repo.GetByCode(ctx, code)
}

// GetAll 获取所有启用的元效果类型定义
func (s *EffectTypeDefinitionService) GetAll(ctx context.Context) ([]*game_config.EffectTypeDefinition, error) {
	return s.repo.GetAll(ctx)
}
