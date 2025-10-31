package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// FormulaVariableService 公式变量服务（现在使用通用字典表）
type FormulaVariableService struct {
	dictRepo interfaces.MetadataDictionaryRepository
}

func NewFormulaVariableService(db *sql.DB) *FormulaVariableService {
	return &FormulaVariableService{
		dictRepo: impl.NewMetadataDictionaryRepository(db),
	}
}

// GetAll 获取所有公式变量（向后兼容）
func (s *FormulaVariableService) GetAll(ctx context.Context) ([]*game_config.MetadataDictionary, error) {
	return s.dictRepo.GetFormulaVariables(ctx)
}

// GetByID 根据ID获取公式变量
func (s *FormulaVariableService) GetByID(ctx context.Context, id string) (*game_config.MetadataDictionary, error) {
	return s.dictRepo.GetByID(ctx, id)
}
