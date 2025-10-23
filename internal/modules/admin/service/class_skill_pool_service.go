package service

import (
	"context"
	"database/sql"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ClassSkillPoolService 职业技能池服务
type ClassSkillPoolService struct {
	repo      interfaces.ClassSkillPoolRepository
	skillRepo interfaces.SkillRepository
}

// NewClassSkillPoolService 创建职业技能池服务
func NewClassSkillPoolService(db *sql.DB) *ClassSkillPoolService {
	return &ClassSkillPoolService{
		repo:      impl.NewClassSkillPoolRepository(db),
		skillRepo: impl.NewSkillRepository(db),
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

// ValidatePrerequisiteSkills 验证前置技能配置的有效性
func (s *ClassSkillPoolService) ValidatePrerequisiteSkills(ctx context.Context, classID, skillID string, prerequisiteSkillIds []string) error {
	if len(prerequisiteSkillIds) == 0 {
		return nil
	}

	// 1. 验证前置技能ID是否存在于技能表
	var invalidSkillIDs []string
	for _, prereqID := range prerequisiteSkillIds {
		skill, err := s.skillRepo.GetByID(ctx, prereqID)
		if err != nil || skill == nil {
			invalidSkillIDs = append(invalidSkillIDs, prereqID)
		}
	}

	if len(invalidSkillIDs) > 0 {
		log.WarnContext(ctx, "前置技能ID无效",
			"class_id", classID,
			"skill_id", skillID,
			"invalid_prerequisite_skills", invalidSkillIDs)
		return xerrors.New(xerrors.CodeSkillInvalidPrerequisite, "前置技能配置无效")
	}

	// 2. 验证前置技能是否在同一职业的技能池中
	for _, prereqID := range prerequisiteSkillIds {
		poolEntry, err := s.repo.GetByClassIDAndSkillID(ctx, classID, prereqID)
		if err != nil {
			log.ErrorContext(ctx, "查询前置技能池失败",
				"class_id", classID,
				"prerequisite_skill_id", prereqID)
			return xerrors.Wrap(err, xerrors.CodeInternalError, "查询前置技能池失败")
		}
		if poolEntry == nil {
			log.WarnContext(ctx, "前置技能不在职业技能池中",
				"class_id", classID,
				"skill_id", skillID,
				"prerequisite_skill_id", prereqID)
			return xerrors.New(xerrors.CodeSkillInvalidPrerequisite, "前置技能必须在同一职业技能池中")
		}
	}

	// 3. 检测循环依赖（简单检测：前置技能不能依赖当前技能）
	if err := s.detectCircularDependency(ctx, classID, skillID, prerequisiteSkillIds); err != nil {
		return err
	}

	return nil
}

// detectCircularDependency 检测循环依赖
func (s *ClassSkillPoolService) detectCircularDependency(ctx context.Context, classID, skillID string, prerequisiteSkillIds []string) error {
	// 使用 visited 记录已访问的技能，防止无限循环
	visited := make(map[string]bool)
	visited[skillID] = true

	// 递归检查每个前置技能
	for _, prereqID := range prerequisiteSkillIds {
		if err := s.checkDependencyChain(ctx, classID, prereqID, skillID, visited); err != nil {
			return err
		}
	}

	return nil
}

// checkDependencyChain 递归检查依赖链
func (s *ClassSkillPoolService) checkDependencyChain(ctx context.Context, classID, checkSkillID, originalSkillID string, visited map[string]bool) error {
	// 如果已经访问过，说明有循环
	if visited[checkSkillID] {
		log.WarnContext(ctx, "检测到循环依赖",
			"class_id", classID,
			"original_skill_id", originalSkillID,
			"circular_skill_id", checkSkillID)
		return xerrors.New(xerrors.CodeSkillInvalidPrerequisite, "检测到循环依赖")
	}

	// 标记为已访问
	visited[checkSkillID] = true

	// 获取这个技能的前置技能
	poolEntry, err := s.repo.GetByClassIDAndSkillID(ctx, classID, checkSkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能池失败")
	}

	if poolEntry != nil && len(poolEntry.PrerequisiteSkillIds) > 0 {
		// 递归检查这个技能的前置技能
		for _, prereqID := range poolEntry.PrerequisiteSkillIds {
			if err := s.checkDependencyChain(ctx, classID, prereqID, originalSkillID, visited); err != nil {
				return err
			}
		}
	}

	// 回溯时移除标记
	delete(visited, checkSkillID)
	return nil
}
