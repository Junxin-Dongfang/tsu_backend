package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// HeroSkillService 英雄技能服务
type HeroSkillService struct {
	db                   *sql.DB
	heroRepo             interfaces.HeroRepository
	heroSkillRepo        interfaces.HeroSkillRepository
	heroClassHistoryRepo interfaces.HeroClassHistoryRepository
	classSkillPoolRepo   interfaces.ClassSkillPoolRepository
	skillUpgradeCostRepo interfaces.SkillUpgradeCostRepository
	skillOpRepo          interfaces.HeroSkillOperationRepository
	heroService          *HeroService
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
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 2. 获取英雄当前职业
	currentClass, err := s.heroClassHistoryRepo.GetCurrentClass(ctx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "无法获取英雄职业信息")
	}

	// 3. 验证技能是否在职业技能池中
	skillPool, err := s.classSkillPoolRepo.GetByClassIDAndSkillID(ctx, currentClass.ClassID, req.SkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能池失败")
	}
	if skillPool == nil {
		return xerrors.New(xerrors.CodeSkillNotFound, "该职业无法学习此技能")
	}

	// 4. 检查是否已经学习
	existingSkill, err := s.heroSkillRepo.GetByHeroAndSkillID(ctx, req.HeroID, req.SkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能失败")
	}
	if existingSkill != nil {
		return xerrors.New(xerrors.CodeDuplicateResource, "已学习该技能")
	}

	// 5. 获取学习消耗（skill_upgrade_costs level=1）
	cost, err := s.skillUpgradeCostRepo.GetByLevel(ctx, 1)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "学习消耗配置不存在")
	}

	// 6. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 7. 验证资源是否足够
	costXP := 0
	if !cost.CostXP.IsZero() {
		costXP = int(cost.CostXP.Int)
	}

	if hero.ExperienceAvailable < int64(costXP) {
		return xerrors.New(xerrors.CodeInsufficientExperience,
			fmt.Sprintf("经验不足: 需要 %d，当前 %d", costXP, hero.ExperienceAvailable))
	}

	// 8. 扣除资源（注：experience_total 不应在此增加）
	hero.ExperienceAvailable -= int64(costXP)
	hero.ExperienceSpent += int64(costXP)
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 9. 插入 hero_skills
	now := time.Now()
	heroSkill := &game_runtime.HeroSkill{
		ID:             uuid.New().String(),
		HeroID:         req.HeroID,
		SkillID:        req.SkillID,
		SkillLevel:     1,
		LearnedMethod:  null.StringFrom("manual"),
		FirstLearnedAt: null.TimeFrom(now),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.heroSkillRepo.Create(ctx, tx, heroSkill); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建技能记录失败")
	}

	// 10. 创建操作历史
	costGold := 0
	if !cost.CostGold.IsZero() {
		costGold = int(cost.CostGold.Int)
	}

	operation := &game_runtime.HeroSkillOperation{
		ID:               uuid.New().String(),
		HeroSkillID:      heroSkill.ID,
		LevelsAdded:      1,
		XPSpent:          costXP,
		GoldSpent:        costGold,
		MaterialsSpent:   cost.CostMaterials,
		LevelBefore:      0,
		LevelAfter:       1,
		CreatedAt:        now,
		RollbackDeadline: now.Add(1 * time.Hour),
	}

	if err := s.skillOpRepo.Create(ctx, tx, operation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建操作历史失败")
	}

	// 11. 调用 AutoLevelUp
	_, _, err = s.heroService.AutoLevelUp(ctx, tx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "自动升级检查失败")
	}

	// 12. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// UpgradeSkillRequest 升级技能请求
type UpgradeSkillRequest struct {
	HeroID  string `json:"hero_id"`
	SkillID string `json:"skill_id"`
	Levels  int    `json:"levels"` // 要升级的等级数
}

// UpgradeSkill 升级技能
func (s *HeroSkillService) UpgradeSkill(ctx context.Context, req *UpgradeSkillRequest) error {
	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 2. 获取技能当前信息（加锁）
	heroSkill, err := s.heroSkillRepo.GetByHeroAndSkillIDForUpdate(ctx, tx, req.HeroID, req.SkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeSkillNotFound, "技能不存在")
	}

	// 3. 验证技能是否可以升级（必须在当前可学习技能池中）
	currentClass, err := s.heroClassHistoryRepo.GetCurrentClass(ctx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "无法获取英雄职业信息")
	}

	skillPool, err := s.classSkillPoolRepo.GetByClassIDAndSkillID(ctx, currentClass.ClassID, heroSkill.SkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能池失败")
	}
	if skillPool == nil {
		return xerrors.New(xerrors.CodePermissionDenied, "当前职业无法升级该技能")
	}

	// 4. 验证是否达到上限
	maxLevel := 10 // 默认等级上限
	if !skillPool.MaxLearnableLevel.IsZero() {
		maxLevel = int(skillPool.MaxLearnableLevel.Int)
	}

	if int(heroSkill.SkillLevel) >= maxLevel {
		return xerrors.New(xerrors.CodeInvalidParams,
			fmt.Sprintf("技能已达到最高等级 %d", maxLevel))
	}

	// 5. 获取升级消耗
	nextLevel := int(heroSkill.SkillLevel) + 1
	cost, err := s.skillUpgradeCostRepo.GetByLevel(ctx, nextLevel)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "升级消耗配置不存在")
	}

	// 6. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 7. 验证资源
	costXP := 0
	if !cost.CostXP.IsZero() {
		costXP = int(cost.CostXP.Int)
	}

	if hero.ExperienceAvailable < int64(costXP) {
		return xerrors.New(xerrors.CodeInsufficientExperience,
			fmt.Sprintf("经验不足: 需要 %d，当前 %d", costXP, hero.ExperienceAvailable))
	}

	// 8. 扣除资源（注：experience_total 不应在此增加）
	hero.ExperienceAvailable -= int64(costXP)
	hero.ExperienceSpent += int64(costXP)
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 9. 更新 hero_skills
	oldLevel := heroSkill.SkillLevel
	heroSkill.SkillLevel = nextLevel
	heroSkill.UpdatedAt = time.Now()

	if err := s.heroSkillRepo.Update(ctx, tx, heroSkill); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新技能失败")
	}

	// 10. 创建操作历史
	now := time.Now()
	costGold := 0
	if !cost.CostGold.IsZero() {
		costGold = int(cost.CostGold.Int)
	}

	operation := &game_runtime.HeroSkillOperation{
		ID:               uuid.New().String(),
		HeroSkillID:      heroSkill.ID,
		LevelsAdded:      1,  // 当前实现只支持升级1级（req.Levels 被忽略）
		XPSpent:          costXP,
		GoldSpent:        costGold,
		MaterialsSpent:   cost.CostMaterials,
		LevelBefore:      int(oldLevel),
		LevelAfter:       nextLevel,
		CreatedAt:        now,
		RollbackDeadline: now.Add(1 * time.Hour),
	}

	if err := s.skillOpRepo.Create(ctx, tx, operation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建操作历史失败")
	}

	// 11. 调用 AutoLevelUp
	_, _, err = s.heroService.AutoLevelUp(ctx, tx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "自动升级检查失败")
	}

	// 12. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// RollbackSkillOperation 回退技能操作（栈式）
func (s *HeroSkillService) RollbackSkillOperation(ctx context.Context, heroSkillID string) error {
	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 2. 获取最近一次未回退且未过期的操作（栈顶）
	operation, err := s.skillOpRepo.GetLatestRollbackable(ctx, heroSkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询可回退操作失败")
	}
	if operation == nil {
		return xerrors.New(xerrors.CodeResourceNotFound, "没有可回退的操作")
	}

	// 3. 验证操作存在且未过期
	if time.Now().After(operation.RollbackDeadline) {
		return xerrors.New(xerrors.CodeOperationExpired, "回退时间已过期")
	}

	// 4. 获取技能信息（加锁）
	heroSkill, err := s.heroSkillRepo.GetByIDForUpdate(ctx, tx, heroSkillID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeSkillNotFound, "技能不存在")
	}

	// 5. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroSkill.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 6. 返还资源（注：experience_total 不应改变）
	hero.ExperienceAvailable += int64(operation.XPSpent)
	hero.ExperienceSpent -= int64(operation.XPSpent)
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 7. 回退技能等级
	if operation.LevelBefore == 0 {
		// 如果回退到level=0，删除记录
		if err := s.heroSkillRepo.Delete(ctx, tx, heroSkillID); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "删除技能失败")
		}
	} else {
		heroSkill.SkillLevel = operation.LevelBefore
		heroSkill.UpdatedAt = time.Now()
		if err := s.heroSkillRepo.Update(ctx, tx, heroSkill); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "更新技能失败")
		}
	}

	// 8. 标记操作为已回退
	if err := s.skillOpRepo.MarkAsRolledBack(ctx, tx, operation.ID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "标记回退失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// AvailableSkillInfo 可学习技能信息
type AvailableSkillInfo struct {
	SkillID           string
	SkillName         string
	SkillCode         string
	MaxLevel          int
	MaxLearnableLevel int
	CanLearn          bool
	Requirements      string
}

// GetAvailableSkills 获取可学习技能
func (s *HeroSkillService) GetAvailableSkills(ctx context.Context, heroID string) ([]*AvailableSkillInfo, error) {
	// 1. 获取当前职业
	currentClass, err := s.heroClassHistoryRepo.GetCurrentClass(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "获取当前职业失败")
	}

	// 2. 查询职业的所有技能池
	skillPools, err := s.classSkillPoolRepo.GetClassSkillPoolsByClassID(ctx, currentClass.ClassID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询技能池失败")
	}

	// 3. 获取英雄已学习的技能
	learnedSkills, err := s.heroSkillRepo.GetByHeroID(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询已学习技能失败")
	}

	// 构建已学习技能的 map
	learnedSkillIDs := make(map[string]bool)
	for _, skill := range learnedSkills {
		learnedSkillIDs[skill.SkillID] = true
	}

	// 4. 转换为可学习技能信息（排除已学习的技能）
	// TODO: 需要联接 Skills 表获取 SkillName 和 SkillCode
	availableSkills := make([]*AvailableSkillInfo, 0)
	for _, pool := range skillPools {
		// 跳过已学习的技能
		if learnedSkillIDs[pool.SkillID] {
			continue
		}

		maxLearnableLevel := 10
		if !pool.MaxLearnableLevel.IsZero() {
			maxLearnableLevel = int(pool.MaxLearnableLevel.Int)
		}

		availableSkills = append(availableSkills, &AvailableSkillInfo{
			SkillID:           pool.SkillID,
			SkillName:         "", // 暂时为空，需要查询 Skills 表
			SkillCode:         "", // 暂时为空，需要查询 Skills 表
			MaxLevel:          10, // 默认值
			MaxLearnableLevel: maxLearnableLevel,
			CanLearn:          true, // 简化处理：暂不检查条件
		})
	}

	return availableSkills, nil
}
