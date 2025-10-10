package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ClassSkillPoolService 职业技能池服务
type ClassSkillPoolService struct {
	repo interfaces.ClassSkillPoolRepository
}

// NewClassSkillPoolService 创建职业技能池服务
func NewClassSkillPoolService(db *sql.DB) *ClassSkillPoolService {
	return &ClassSkillPoolService{
		repo: impl.NewClassSkillPoolRepository(db),
	}
}

// GetClassSkillPools 获取职业技能池列表
func (s *ClassSkillPoolService) GetClassSkillPools(ctx context.Context, params interfaces.ClassSkillPoolQueryParams) ([]*game_config.ClassSkillPool, int64, error) {
	return s.repo.GetClassSkillPools(ctx, params)
}

// GetClassSkillPoolByID 根据ID获取职业技能池
func (s *ClassSkillPoolService) GetClassSkillPoolByID(ctx context.Context, id string) (*game_config.ClassSkillPool, error) {
	return s.repo.GetClassSkillPoolByID(ctx, id)
}

// GetClassSkillPoolsByClassID 获取指定职业的所有技能
func (s *ClassSkillPoolService) GetClassSkillPoolsByClassID(ctx context.Context, classID string) ([]*game_config.ClassSkillPool, error) {
	return s.repo.GetClassSkillPoolsByClassID(ctx, classID)
}

// CreateClassSkillPool 创建职业技能池配置
func (s *ClassSkillPoolService) CreateClassSkillPool(ctx context.Context, pool *game_config.ClassSkillPool) error {
	return s.repo.CreateClassSkillPool(ctx, pool)
}

// UpdateClassSkillPool 更新职业技能池配置
func (s *ClassSkillPoolService) UpdateClassSkillPool(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.repo.UpdateClassSkillPool(ctx, id, updates)
}

// DeleteClassSkillPool 删除职业技能池配置（软删除）
func (s *ClassSkillPoolService) DeleteClassSkillPool(ctx context.Context, id string) error {
	return s.repo.DeleteClassSkillPool(ctx, id)
}
