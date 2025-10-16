package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// HeroSkillService 英雄技能服务
type HeroSkillService struct {
	db                     *sql.DB
	heroRepo               interfaces.HeroRepository
	heroSkillRepo          interfaces.HeroSkillRepository
	heroClassHistoryRepo   interfaces.HeroClassHistoryRepository
	classSkillPoolRepo     interfaces.ClassSkillPoolRepository
	skillUpgradeCostRepo   interfaces.SkillUpgradeCostRepository
	skillOpRepo            interfaces.HeroSkillOperationRepository
	heroService            *HeroService
}

// NewHeroSkillService 创建英雄技能服务
func NewHeroSkillService(db *sql.DB) *HeroSkillService {
	return &HeroSkillService{
		db:                   db,
		heroRepo:             impl.NewHeroRepository(db),
		heroSkillRepo:        impl.NewHeroSkillRepository(db),
		heroClassHistoryRepo: impl.NewHeroClassHistoryRepository(db),
		classSkillPoolRepo:   impl.NewClassSkillPoolRepository(db),
		skillUpgradeCostRepo: impl.NewSkillUpgradeCostRepository(db),
		skillOpRepo:          impl.NewHeroSkillOperationRepository(db),
		heroService:          NewHeroService(db),
	}
}

// LearnSkillRequest 学习技能请求
type LearnSkillRequest struct {
	HeroID  string `json:"hero_id"`
	SkillID string `json:"skill_id"`
}

// LearnSkill 学习技能
func (s *HeroSkillService) LearnSkill(ctx context.Context, req *LearnSkillRequest) error {
	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "开启事务失败")
	}
	defer tx.Rollback()

	// 2. 验证技能是否在可学习池中
	availableClasses, err := s.heroClassHistoryRepo.GetAvailableClassesForSkills(ctx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "查询可学习技能池失败")
	}

	classIDs := make([]string, len(availableClasses))
	for i, c := range availableClasses {
		classIDs[i] = c.ClassID
	}

	skillPool, err := s.classSkillPoolRepo.GetByClassIDsAndSkillID(ctx, classIDs, req.SkillID)
	if err != nil || skillPool == nil {
		return xerrors.New(xerrors.CodeNotFound, "技能不在可学习池中")
	}

	// 3. 检查是否已经学习
	existingSkill, err := s.heroSkillRepo.GetByHeroAndSkillID(ctx, req.HeroID, req.SkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "查询技能失败")
	}
	if existingSkill != nil {
		return xerrors.New(xerrors.CodeDuplicateResource, "已学习该技能")
	}

	// 4. 检查学习条件（等级、属性、前置技能）
	// TODO: 实现条件检查逻辑

	// 5. 获取学习消耗（skill_upgrade_costs level=1）
	cost, err := s.skillUpgradeCostRepo.GetByLevel(ctx, 1)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "学习消耗配置不存在")
	}

	// 6. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "英雄不存在")
	}

	// 7. 验证资源是否足够
	if hero.ExperienceAvailable < cost.CostXP {
		return xerrors.New(xerrors.CodeInsufficientResource, 
			fmt.Sprintf("经验不足: 需要 %d，当前 %d", cost.CostXP, hero.ExperienceAvailable))
	}

	// 8. 扣除资源
	hero.ExperienceAvailable -= cost.CostXP
	hero.ExperienceSpent += cost.CostXP
	hero.ExperienceTotal += cost.CostXP
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "更新英雄失败")
	}

	// 9. 插入 hero_skills
	now := time.Now()
	heroSkill := &game_runtime.HeroSkill{
		ID:             uuid.New().String(),
		HeroID:         req.HeroID,
		SkillID:        req.SkillID,
		SkillLevel:     1,
		LearnedMethod:  "manual",
		FirstLearnedAt: null.TimeFrom(now),
		CreatedAt:      null.TimeFrom(now),
		UpdatedAt:      null.TimeFrom(now),
	}

	if err := s.heroSkillRepo.Create(ctx, tx, heroSkill); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "创建技能记录失败")
	}

	// 10. 创建操作历史
	operation := &game_runtime.HeroSkillOperation{
		ID:               uuid.New().String(),
		HeroSkillID:      heroSkill.ID,
		LevelsAdded:      1,
		XPSpent:          cost.CostXP,
		GoldSpent:        cost.CostGold,
		MaterialsSpent:   types.JSON(cost.CostMaterials.JSON),
		LevelBefore:      0,
		LevelAfter:       1,
		CreatedAt:        now,
		RollbackDeadline: now.Add(1 * time.Hour),
	}

	if err := s.skillOpRepo.Create(ctx, tx, operation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "创建操作历史失败")
	}

	// 11. 调用 AutoLevelUp
	_, _, err = s.heroService.AutoLevelUp(ctx, tx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "自动升级检查失败")
	}

	// 12. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "提交事务失败")
	}

	return nil
}

// UpgradeSkillRequest 升级技能请求
type UpgradeSkillRequest struct {
	HeroSkillID string `json:"hero_skill_id"`
}

// UpgradeSkill 升级技能
func (s *HeroSkillService) UpgradeSkill(ctx context.Context, req *UpgradeSkillRequest) error {
	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "开启事务失败")
	}
	defer tx.Rollback()

	// 2. 获取技能当前信息（加锁）
	heroSkill, err := s.heroSkillRepo.GetByIDForUpdate(ctx, tx, req.HeroSkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "技能不存在")
	}

	// 3. 验证技能是否可以升级（必须在当前可学习技能池中）
	availableClasses, err := s.heroClassHistoryRepo.GetAvailableClassesForSkills(ctx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "查询可学习技能池失败")
	}

	classIDs := make([]string, len(availableClasses))
	for i, c := range availableClasses {
		classIDs[i] = c.ClassID
	}

	skillPool, err := s.classSkillPoolRepo.GetByClassIDsAndSkillID(ctx, classIDs, heroSkill.SkillID)
	if err != nil || skillPool == nil {
		return xerrors.New(xerrors.CodeForbidden, "当前职业无法升级该技能")
	}

	// 4. 验证是否达到上限
	maxLevel := skillPool.MaxLearnableLevel
	// TODO: 比较 skills.max_level

	if heroSkill.SkillLevel >= maxLevel {
		return xerrors.New(xerrors.CodeInvalidArgument, 
			fmt.Sprintf("技能已达到最高等级 %d", maxLevel))
	}

	// 5. 获取升级消耗
	nextLevel := heroSkill.SkillLevel + 1
	cost, err := s.skillUpgradeCostRepo.GetByLevel(ctx, nextLevel)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "升级消耗配置不存在")
	}

	// 6. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "英雄不存在")
	}

	// 7. 验证资源
	if hero.ExperienceAvailable < cost.CostXP {
		return xerrors.New(xerrors.CodeInsufficientResource, 
			fmt.Sprintf("经验不足: 需要 %d，当前 %d", cost.CostXP, hero.ExperienceAvailable))
	}

	// 8. 扣除资源，增加 experience_total
	hero.ExperienceAvailable -= cost.CostXP
	hero.ExperienceSpent += cost.CostXP
	hero.ExperienceTotal += cost.CostXP
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "更新英雄失败")
	}

	// 9. 更新 hero_skills
	oldLevel := heroSkill.SkillLevel
	heroSkill.SkillLevel = nextLevel
	heroSkill.UpdatedAt = null.TimeFrom(time.Now())

	if err := s.heroSkillRepo.Update(ctx, tx, heroSkill); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "更新技能失败")
	}

	// 10. 创建操作历史
	now := time.Now()
	operation := &game_runtime.HeroSkillOperation{
		ID:               uuid.New().String(),
		HeroSkillID:      req.HeroSkillID,
		LevelsAdded:      1,
		XPSpent:          cost.CostXP,
		GoldSpent:        cost.CostGold,
		MaterialsSpent:   types.JSON(cost.CostMaterials.JSON),
		LevelBefore:      oldLevel,
		LevelAfter:       nextLevel,
		CreatedAt:        now,
		RollbackDeadline: now.Add(1 * time.Hour),
	}

	if err := s.skillOpRepo.Create(ctx, tx, operation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "创建操作历史失败")
	}

	// 11. 调用 AutoLevelUp
	_, _, err = s.heroService.AutoLevelUp(ctx, tx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "自动升级检查失败")
	}

	// 12. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "提交事务失败")
	}

	return nil
}

// RollbackSkillOperation 回退技能操作（栈式）
func (s *HeroSkillService) RollbackSkillOperation(ctx context.Context, heroSkillID string) error {
	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "开启事务失败")
	}
	defer tx.Rollback()

	// 2. 获取最近一次未回退且未过期的操作（栈顶）
	operation, err := s.skillOpRepo.GetLatestRollbackable(ctx, heroSkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "查询可回退操作失败")
	}
	if operation == nil {
		return xerrors.New(xerrors.CodeNotFound, "没有可回退的操作")
	}

	// 3. 验证操作存在且未过期
	if time.Now().After(operation.RollbackDeadline) {
		return xerrors.New(xerrors.CodeOperationExpired, "回退时间已过期")
	}

	// 4. 获取技能信息（加锁）
	heroSkill, err := s.heroSkillRepo.GetByIDForUpdate(ctx, tx, heroSkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "技能不存在")
	}

	// 5. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeNotFound, "英雄不存在")
	}

	// 6. 返还资源
	hero.ExperienceAvailable += operation.XPSpent
	hero.ExperienceSpent -= operation.XPSpent
	hero.ExperienceTotal -= operation.XPSpent
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "更新英雄失败")
	}

	// 7. 回退技能等级
	if operation.LevelBefore == 0 {
		// 如果回退到level=0，删除记录
		if err := s.heroSkillRepo.Delete(ctx, tx, heroSkillID); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternal, "删除技能失败")
		}
	} else {
		heroSkill.SkillLevel = operation.LevelBefore
		heroSkill.UpdatedAt = null.TimeFrom(time.Now())
		if err := s.heroSkillRepo.Update(ctx, tx, heroSkill); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternal, "更新技能失败")
		}
	}

	// 8. 标记操作为已回退
	if err := s.skillOpRepo.MarkAsRolledBack(ctx, tx, operation.ID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "标记回退失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternal, "提交事务失败")
	}

	return nil
}

// GetAvailableSkills 获取可学习技能
func (s *HeroSkillService) GetAvailableSkills(ctx context.Context, heroID string) (interface{}, error) {
	// 1. 获取职业历史
	availableClasses, err := s.heroClassHistoryRepo.GetAvailableClassesForSkills(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "查询可学习技能池失败")
	}

	classIDs := make([]string, len(availableClasses))
	for i, c := range availableClasses {
		classIDs[i] = c.ClassID
	}

	// 2. 查询这些职业的技能池
	// TODO: 实现完整的可学习技能查询逻辑
	// - 排除已学习的技能
	// - 检查学习条件

	return map[string]interface{}{
		"available_class_ids": classIDs,
		"message":            "功能实现中",
	}, nil
}

