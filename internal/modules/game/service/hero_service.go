package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// HeroService 英雄服务
type HeroService struct {
	db                         *sql.DB
	heroRepo                   interfaces.HeroRepository
	heroClassHistoryRepo       interfaces.HeroClassHistoryRepository
	heroSkillRepo              interfaces.HeroSkillRepository
	classRepo                  interfaces.ClassRepository
	classSkillPoolRepo         interfaces.ClassSkillPoolRepository
	heroAttributeTypeRepo      interfaces.HeroAttributeTypeRepository
	heroLevelRequirementRepo   interfaces.HeroLevelRequirementRepository
	heroAllocatedAttributeRepo interfaces.HeroAllocatedAttributeRepository
	classAdvancedReqRepo       interfaces.ClassAdvancedRequirementRepository
}

// NewHeroService 创建英雄服务
func NewHeroService(db *sql.DB) *HeroService {
	return &HeroService{
		db:                         db,
		heroRepo:                   impl.NewHeroRepository(db),
		heroClassHistoryRepo:       impl.NewHeroClassHistoryRepository(db),
		heroSkillRepo:              impl.NewHeroSkillRepository(db),
		classRepo:                  impl.NewClassRepository(db),
		classSkillPoolRepo:         impl.NewClassSkillPoolRepository(db),
		heroAttributeTypeRepo:      impl.NewHeroAttributeTypeRepository(db),
		heroLevelRequirementRepo:   impl.NewHeroLevelRequirementRepository(db),
		heroAllocatedAttributeRepo: impl.NewHeroAllocatedAttributeRepository(db),
		classAdvancedReqRepo:       impl.NewClassAdvancedRequirementRepository(db),
	}
}

// CreateHeroRequest 创建英雄请求
type CreateHeroRequest struct {
	UserID      string `json:"user_id"`
	ClassID     string `json:"class_id"`
	HeroName    string `json:"hero_name"`
	Description string `json:"description,omitempty"`
}

// CreateHero 创建英雄
func (s *HeroService) CreateHero(ctx context.Context, req *CreateHeroRequest) (*game_runtime.Hero, error) {
	// 1. 验证用户ID和职业ID
	if req.UserID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "用户ID不能为空")
	}
	if req.ClassID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "职业ID不能为空")
	}
	if req.HeroName == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "英雄名称不能为空")
	}

	// 2. 检查职业是否为基础职业（tier='basic'）
	class, err := s.classRepo.GetByID(ctx, req.ClassID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeResourceNotFound, "职业不存在")
	}
	if class.Tier != "basic" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "只能选择基础职业")
	}

	// 3. 检查英雄名称是否已存在
	exists, err := s.heroRepo.CheckExistsByName(ctx, req.UserID, req.HeroName)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "检查英雄名称失败")
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "英雄名称已存在")
	}

	// 4. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 5. 创建英雄记录（初始等级1，经验0）
	hero := &game_runtime.Hero{
		ID:                  uuid.New().String(),
		UserID:              req.UserID,
		ClassID:             req.ClassID,
		HeroName:            req.HeroName,
		CurrentLevel:        1,
		ExperienceTotal:     0,
		ExperienceAvailable: 0,
		ExperienceSpent:     0,
		Status:              "active",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if req.Description != "" {
		hero.Description.SetValid(req.Description)
	}

	if err := hero.Insert(ctx, tx, boil.Infer()); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建英雄失败")
	}

	// 6. 初始化已分配属性（所有basic类型属性初始值为1）
	if err := s.initializeAllocatedAttributesInTable(ctx, tx, hero.ID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "初始化属性失败")
	}

	// 7. 创建职业历史记录（is_current=true, acquisition_type='initial'）
	classHistory := &game_runtime.HeroClassHistory{
		ID:              uuid.New().String(),
		HeroID:          hero.ID,
		ClassID:         req.ClassID,
		IsCurrent:       true,
		AcquiredAt:      time.Now(),
		AcquisitionType: "initial",
		CreatedAt:       null.TimeFrom(time.Now()),
		UpdatedAt:       null.TimeFrom(time.Now()),
	}
	if err := s.heroClassHistoryRepo.Create(ctx, tx, classHistory); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业历史失败")
	}

	// 8. 获取职业初始技能（is_initial_skill=true）并批量插入
	if err := s.learnInitialSkills(ctx, tx, hero.ID, req.ClassID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "学习初始技能失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return hero, nil
}

// initializeAllocatedAttributesInTable 初始化已分配属性表（所有basic属性初始值为1）
func (s *HeroService) initializeAllocatedAttributesInTable(ctx context.Context, executor interface{}, heroID string) error {
	// 查询所有 basic 类型的属性
	attrs, err := s.heroAttributeTypeRepo.ListByCategory(ctx, "basic")
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询属性类型失败")
	}

	// 为每个活跃的属性创建分配记录
	allocatedAttrs := make([]*game_runtime.HeroAllocatedAttribute, 0, len(attrs))
	for _, attr := range attrs {
		// 只初始化活跃的属性
		if attr.IsActive {
			allocatedAttrs = append(allocatedAttrs, &game_runtime.HeroAllocatedAttribute{
				ID:            uuid.New().String(),
				HeroID:        heroID,
				AttributeCode: attr.AttributeCode,
				Value:         1, // 初始值为 1
				SpentXP:       0, // 初始花费 0
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			})
		}
	}

	// 批量创建分配记录
	if len(allocatedAttrs) > 0 {
		if err := s.heroAllocatedAttributeRepo.BatchCreateForHero(ctx, executor, heroID, allocatedAttrs); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "创建属性分配记录失败")
		}
	}

	return nil
}

// learnInitialSkills 学习初始技能
func (s *HeroService) learnInitialSkills(ctx context.Context, tx *sql.Tx, heroID, classID string) error {
	// 获取职业的初始技能
	initialSkills, err := s.classSkillPoolRepo.GetClassSkillPoolsByClassID(ctx, classID)
	if err != nil {
		return err
	}

	// 批量插入初始技能
	now := time.Now()
	for _, skillPool := range initialSkills {
		if skillPool.IsInitialSkill.IsZero() || !skillPool.IsInitialSkill.Bool {
			continue
		}

		heroSkill := &game_runtime.HeroSkill{
			ID:             uuid.New().String(),
			HeroID:         heroID,
			SkillID:        skillPool.SkillID,
			SkillLevel:     1,
			LearnedMethod:  null.StringFrom("class_unlock"),
			FirstLearnedAt: null.TimeFrom(now),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := s.heroSkillRepo.Create(ctx, tx, heroSkill); err != nil {
			return err
		}
	}

	return nil
}

// GetHeroByID 获取英雄详情
func (s *HeroService) GetHeroByID(ctx context.Context, heroID string) (*game_runtime.Hero, error) {
	return s.heroRepo.GetByID(ctx, heroID)
}

// GetHeroesByUserID 获取用户的英雄列表
func (s *HeroService) GetHeroesByUserID(ctx context.Context, userID string) ([]*game_runtime.Hero, error) {
	return s.heroRepo.GetByUserID(ctx, userID)
}

// AutoLevelUp 自动升级检查（每次增加经验后调用）
func (s *HeroService) AutoLevelUp(ctx context.Context, tx *sql.Tx, heroID string) (leveledUp bool, newLevel int, error error) {
	// 1. 获取英雄信息（带锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroID)
	if err != nil {
		return false, 0, err
	}

	// 2. 循环检查是否可以升级
	canLevelUp, targetLevel, err := s.heroLevelRequirementRepo.CheckCanLevelUp(ctx, int(hero.ExperienceTotal), int(hero.CurrentLevel))
	if err != nil {
		return false, int(hero.CurrentLevel), err
	}

	if !canLevelUp {
		return false, int(hero.CurrentLevel), nil
	}

	// 3. 升级
	hero.CurrentLevel = int16(targetLevel)
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return false, int(hero.CurrentLevel), err
	}

	return true, targetLevel, nil
}

// AdvanceClass 职业进阶
func (s *HeroService) AdvanceClass(ctx context.Context, heroID, targetClassID string) error {
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

	// 2. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 3. 获取当前职业历史
	currentClassHistory, err := s.heroClassHistoryRepo.GetCurrentClass(ctx, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "获取当前职业失败")
	}

	// 4. 验证进阶路径
	advReq, err := s.classAdvancedReqRepo.GetByFromAndTo(ctx, currentClassHistory.ClassID, targetClassID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询进阶路径失败")
	}
	if advReq == nil {
		return xerrors.New(xerrors.CodeOperationNotAllowed, "不存在该进阶路径")
	}

	// 5. 检查等级要求
	if hero.CurrentLevel < int16(advReq.RequiredLevel) {
		return xerrors.New(xerrors.CodeInsufficientLevel, fmt.Sprintf("需要等级 %d", advReq.RequiredLevel))
	}

	// 6. 检查属性要求
	if !advReq.RequiredAttributes.IsZero() {
		var reqAttrs map[string]int
		if err := json.Unmarshal(advReq.RequiredAttributes.JSON, &reqAttrs); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "解析属性要求失败")
		}

		// 获取英雄当前属性
		heroAttrs, err := s.heroAllocatedAttributeRepo.GetByHeroID(ctx, heroID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "获取英雄属性失败")
		}

		// 构建属性值映射
		attrMap := make(map[string]int)
		for _, attr := range heroAttrs {
			attrMap[attr.AttributeCode] = int(attr.Value)
		}

		// 验证每个属性要求
		for attrCode, requiredValue := range reqAttrs {
			currentValue, exists := attrMap[attrCode]
			if !exists || currentValue < requiredValue {
				return xerrors.New(xerrors.CodeInsufficientAttributes, fmt.Sprintf("属性 %s 不足，需要 %d", attrCode, requiredValue))
			}
		}
	}

	// 7. 检查技能要求
	if !advReq.RequiredSkills.IsZero() {
		var reqSkills []string
		if err := json.Unmarshal(advReq.RequiredSkills.JSON, &reqSkills); err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "解析技能要求失败")
		}

		// 获取英雄已学习的技能
		learnedSkills, err := s.heroSkillRepo.GetByHeroID(ctx, heroID)
		if err != nil {
			return xerrors.Wrap(err, xerrors.CodeInternalError, "获取已学习技能失败")
		}

		// 构建已学习技能映射
		learnedSkillMap := make(map[string]bool)
		for _, skill := range learnedSkills {
			learnedSkillMap[skill.SkillID] = true
		}

		// 验证每个技能要求
		for _, requiredSkillID := range reqSkills {
			if !learnedSkillMap[requiredSkillID] {
				return xerrors.New(xerrors.CodeInsufficientSkills, fmt.Sprintf("缺少必需技能: %s", requiredSkillID))
			}
		}
	}

	// 8. 更新职业历史（旧职业 is_current=false，新职业 acquisition_type='advancement'）
	if err := s.heroClassHistoryRepo.SetCurrentClass(ctx, tx, heroID, targetClassID, "advancement"); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新职业历史失败")
	}

	// 9. 更新英雄的 class_id 和 promotion_count
	hero.ClassID = targetClassID
	if hero.PromotionCount.IsZero() {
		hero.PromotionCount = null.Int16From(1)
	} else {
		hero.PromotionCount = null.Int16From(hero.PromotionCount.Int16 + 1)
	}
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 10. 学习新职业的初始技能
	if err := s.learnInitialSkills(ctx, tx, heroID, targetClassID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "学习初始技能失败")
	}

	// 11. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// TransferClass 职业转职
func (s *HeroService) TransferClass(ctx context.Context, heroID, targetClassID string) error {
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

	// 2. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 3. 获取目标职业信息
	targetClass, err := s.classRepo.GetByID(ctx, targetClassID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeClassNotFound, "目标职业不存在")
	}

	// 4. 验证目标职业必须是基础职业（只能转职到 basic tier）
	if targetClass.Tier != "basic" {
		return xerrors.New(xerrors.CodeOperationNotAllowed, "只能转职到基础职业")
	}

	// 5. 获取当前职业
	currentClassHistory, err := s.heroClassHistoryRepo.GetCurrentClass(ctx, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "获取当前职业失败")
	}

	// 6. 验证不能转职到相同职业
	if currentClassHistory.ClassID == targetClassID {
		return xerrors.New(xerrors.CodeOperationNotAllowed, "不能转职到相同职业")
	}

	// 7. 更新职业历史（旧职业 is_current=false，新职业 acquisition_type='transfer'）
	if err := s.heroClassHistoryRepo.SetCurrentClass(ctx, tx, heroID, targetClassID, "transfer"); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新职业历史失败")
	}

	// 8. 更新英雄的 class_id
	hero.ClassID = targetClassID
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 9. 学习新职业的初始技能
	if err := s.learnInitialSkills(ctx, tx, heroID, targetClassID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "学习初始技能失败")
	}

	// 10. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// AddExperience 增加英雄经验
func (s *HeroService) AddExperience(ctx context.Context, heroID string, amount int64) (*game_runtime.Hero, error) {
	if amount <= 0 {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "经验值必须大于0")
	}

	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 2. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 3. 增加经验
	hero.ExperienceTotal += amount
	hero.ExperienceAvailable += amount
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 4. 检查是否可以升级
	_, _, err = s.AutoLevelUp(ctx, tx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "自动升级检查失败")
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	// 6. 返回更新后的英雄信息
	return s.heroRepo.GetByID(ctx, heroID)
}
