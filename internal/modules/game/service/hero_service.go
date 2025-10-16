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

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// HeroService 英雄服务
type HeroService struct {
	db                        *sql.DB
	heroRepo                  interfaces.HeroRepository
	heroClassHistoryRepo      interfaces.HeroClassHistoryRepository
	heroSkillRepo             interfaces.HeroSkillRepository
	classRepo                 interfaces.ClassRepository
	classSkillPoolRepo        interfaces.ClassSkillPoolRepository
	heroAttributeTypeRepo     interfaces.HeroAttributeTypeRepository
	heroLevelRequirementRepo  interfaces.HeroLevelRequirementRepository
}

// NewHeroService 创建英雄服务
func NewHeroService(db *sql.DB) *HeroService {
	return &HeroService{
		db:                       db,
		heroRepo:                 impl.NewHeroRepository(db),
		heroClassHistoryRepo:     impl.NewHeroClassHistoryRepository(db),
		heroSkillRepo:            impl.NewHeroSkillRepository(db),
		classRepo:                impl.NewClassRepository(db),
		classSkillPoolRepo:       impl.NewClassSkillPoolRepository(db),
		heroAttributeTypeRepo:    impl.NewHeroAttributeTypeRepository(db),
		heroLevelRequirementRepo: impl.NewHeroLevelRequirementRepository(db),
	}
}

// CreateHeroRequest 创建英雄请求
type CreateHeroRequest struct {
	UserID    string `json:"user_id"`
	ClassID   string `json:"class_id"`
	HeroName  string `json:"hero_name"`
	Gender    string `json:"gender,omitempty"`
	Backstory string `json:"backstory,omitempty"`
}

// CreateHero 创建英雄
func (s *HeroService) CreateHero(ctx context.Context, req *CreateHeroRequest) (*game_runtime.Hero, error) {
	// 1. 验证用户ID和职业ID
	if req.UserID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidArgument, "用户ID不能为空")
	}
	if req.ClassID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidArgument, "职业ID不能为空")
	}
	if req.HeroName == "" {
		return nil, xerrors.New(xerrors.CodeInvalidArgument, "英雄名称不能为空")
	}

	// 2. 检查职业是否为基础职业（tier='basic'）
	class, err := s.classRepo.GetByID(ctx, req.ClassID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeNotFound, "职业不存在")
	}
	if class.Tier != "basic" {
		return nil, xerrors.New(xerrors.CodeInvalidArgument, "只能选择基础职业")
	}

	// 3. 检查英雄名称是否已存在
	exists, err := s.heroRepo.CheckExistsByName(ctx, req.UserID, req.HeroName)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "检查英雄名称失败")
	}
	if exists {
		return nil, xerrors.New(xerrors.CodeDuplicateResource, "英雄名称已存在")
	}

	// 4. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "开启事务失败")
	}
	defer tx.Rollback()

	// 5. 初始化 allocated_attributes（所有basic类型属性=1）
	allocatedAttrs, err := s.initializeAllocatedAttributes(ctx)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "初始化属性失败")
	}

	// 6. 创建英雄记录（初始等级1，经验0）
	hero := &game_runtime.Hero{
		ID:                  uuid.New().String(),
		UserID:              req.UserID,
		ClassID:             req.ClassID,
		HeroName:            req.HeroName,
		CurrentLevel:        1,
		ExperienceTotal:     0,
		ExperienceAvailable: 0,
		ExperienceSpent:     0,
		AllocatedAttributes: null.JSONFrom(allocatedAttrs),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if req.Gender != "" {
		hero.Gender.SetValid(req.Gender)
	}
	if req.Backstory != "" {
		hero.Backstory.SetValid(req.Backstory)
	}

	if err := hero.Insert(ctx, tx, boil.Infer()); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "创建英雄失败")
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
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "创建职业历史失败")
	}

	// 8. 获取职业初始技能（is_initial_skill=true）并批量插入
	if err := s.learnInitialSkills(ctx, tx, hero.ID, req.ClassID); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "学习初始技能失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternal, "提交事务失败")
	}

	return hero, nil
}

// initializeAllocatedAttributes 初始化属性分配（所有basic属性初始值为1）
func (s *HeroService) initializeAllocatedAttributes(ctx context.Context) ([]byte, error) {
	// 获取所有basic类型的属性
	attributes, err := s.heroAttributeTypeRepo.GetByCategory(ctx, "basic")
	if err != nil {
		return nil, err
	}

	allocatedAttrs := make(map[string]map[string]interface{})
	for _, attr := range attributes {
		allocatedAttrs[attr.AttributeCode] = map[string]interface{}{
			"value":    1,
			"spent_xp": 0,
		}
	}

	return json.Marshal(allocatedAttrs)
}

// learnInitialSkills 学习初始技能
func (s *HeroService) learnInitialSkills(ctx context.Context, tx *sql.Tx, heroID, classID string) error {
	// 获取职业的初始技能
	initialSkills, err := s.classSkillPoolRepo.GetInitialSkillsByClass(ctx, classID)
	if err != nil {
		return err
	}

	// 批量插入初始技能
	now := time.Now()
	for _, skillPool := range initialSkills {
		heroSkill := &game_runtime.HeroSkill{
			ID:              uuid.New().String(),
			HeroID:          heroID,
			SkillID:         skillPool.SkillID,
			SkillLevel:      1,
			LearnedMethod:   "class_unlock",
			FirstLearnedAt:  null.TimeFrom(now),
			CreatedAt:       null.TimeFrom(now),
			UpdatedAt:       null.TimeFrom(now),
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
	canLevelUp, targetLevel, err := s.heroLevelRequirementRepo.CheckCanLevelUp(ctx, hero.ExperienceTotal, hero.CurrentLevel)
	if err != nil {
		return false, hero.CurrentLevel, err
	}

	if !canLevelUp {
		return false, hero.CurrentLevel, nil
	}

	// 3. 升级
	hero.CurrentLevel = targetLevel
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return false, hero.CurrentLevel, err
	}

	return true, targetLevel, nil
}

// AdvanceClass 职业进阶
func (s *HeroService) AdvanceClass(ctx context.Context, heroID, targetClassID string) error {
	// TODO: 实现职业进阶逻辑
	// 1. 获取英雄信息（加锁）
	// 2. 获取当前职业
	// 3. 验证进阶路径（class_advanced_requirements）
	// 4. 检查进阶条件（等级、属性、技能、荣誉值等）
	// 5. 开启事务
	// 6. 扣除所需资源（荣誉值、物品等）
	// 7. 更新职业历史（旧职业 is_current=false，新职业 acquisition_type='advancement'）
	// 8. 更新英雄的 class_id 和 promotion_count
	// 9. 学习新职业的初始技能（is_initial_skill=true, learned_method='class_unlock'）
	// 10. 提交事务
	return xerrors.New(xerrors.CodeUnimplemented, "职业进阶功能尚未实现")
}

// TransferClass 职业转职
func (s *HeroService) TransferClass(ctx context.Context, heroID, targetClassID string) error {
	// TODO: 实现职业转职逻辑
	// 1. 获取英雄信息（加锁）
	// 2. 获取目标职业信息
	// 3. 验证转职条件（等级、tier要求等）
	// 4. 开启事务
	// 5. 更新职业历史（旧职业 is_current=false，新职业 acquisition_type='transfer'）
	// 6. 更新英雄的 class_id
	// 7. 学习新职业的初始技能
	// 8. 提交事务
	return xerrors.New(xerrors.CodeUnimplemented, "职业转职功能尚未实现")
}

