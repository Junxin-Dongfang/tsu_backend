package service

import (
	"database/sql"

	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ServiceContainer 游戏服务容器 - 统一管理所有 Repository 和 Service
// 目的：避免重复创建 Repository，简化依赖注入
type ServiceContainer struct {
	// 所有 Repository（共享实例）
	heroRepo                   interfaces.HeroRepository
	heroClassHistoryRepo       interfaces.HeroClassHistoryRepository
	heroSkillRepo              interfaces.HeroSkillRepository
	classRepo                  interfaces.ClassRepository
	classSkillPoolRepo         interfaces.ClassSkillPoolRepository
	classAdvancedReqRepo       interfaces.ClassAdvancedRequirementRepository
	skillRepo                  interfaces.SkillRepository
	heroAttributeTypeRepo      interfaces.HeroAttributeTypeRepository
	heroLevelRequirementRepo   interfaces.HeroLevelRequirementRepository
	attributeUpgradeCostRepo   interfaces.AttributeUpgradeCostRepository
	attributeOpRepo            interfaces.HeroAttributeOperationRepository
	heroAllocatedAttributeRepo interfaces.HeroAllocatedAttributeRepository
	skillUpgradeCostRepo       interfaces.SkillUpgradeCostRepository
	skillOpRepo                interfaces.HeroSkillOperationRepository
	skillUnlockActionRepo      interfaces.SkillUnlockActionRepository
	actionRepo                 interfaces.ActionRepository
	actionEffectRepo           interfaces.ActionEffectRepository
	effectRepo                 interfaces.EffectRepository
	skillCategoryRepo          interfaces.SkillCategoryRepository

	// 所有 Service（共享实例）
	HeroService          *HeroService
	HeroAttributeService *HeroAttributeService
	HeroSkillService     *HeroSkillService
	ClassService         *ClassService
	SkillDetailService   *SkillDetailService
}

// NewServiceContainer 创建服务容器
func NewServiceContainer(db *sql.DB) *ServiceContainer {
	c := &ServiceContainer{}

	// 初始化所有 Repository
	c.heroRepo = impl.NewHeroRepository(db)
	c.heroClassHistoryRepo = impl.NewHeroClassHistoryRepository(db)
	c.heroSkillRepo = impl.NewHeroSkillRepository(db)
	c.classRepo = impl.NewClassRepository(db)
	c.classSkillPoolRepo = impl.NewClassSkillPoolRepository(db)
	c.classAdvancedReqRepo = impl.NewClassAdvancedRequirementRepository(db)
	c.skillRepo = impl.NewSkillRepository(db)
	c.heroAttributeTypeRepo = impl.NewHeroAttributeTypeRepository(db)
	c.heroLevelRequirementRepo = impl.NewHeroLevelRequirementRepository(db)
	c.attributeUpgradeCostRepo = impl.NewAttributeUpgradeCostRepository(db)
	c.attributeOpRepo = impl.NewHeroAttributeOperationRepository(db)
	c.heroAllocatedAttributeRepo = impl.NewHeroAllocatedAttributeRepository(db)
	c.skillUpgradeCostRepo = impl.NewSkillUpgradeCostRepository(db)
	c.skillOpRepo = impl.NewHeroSkillOperationRepository(db)
	c.skillUnlockActionRepo = impl.NewSkillUnlockActionRepository(db)
	c.actionRepo = impl.NewActionRepository(db)
	c.actionEffectRepo = impl.NewActionEffectRepository(db)
	c.effectRepo = impl.NewEffectRepository(db)
	c.skillCategoryRepo = impl.NewSkillCategoryRepository(db)

	// 初始化 HeroService（依赖 repository）
	c.HeroService = &HeroService{
		db:                         db,
		heroRepo:                   c.heroRepo,
		heroClassHistoryRepo:       c.heroClassHistoryRepo,
		heroSkillRepo:              c.heroSkillRepo,
		classRepo:                  c.classRepo,
		classSkillPoolRepo:         c.classSkillPoolRepo,
		heroAttributeTypeRepo:      c.heroAttributeTypeRepo,
		heroLevelRequirementRepo:   c.heroLevelRequirementRepo,
		heroAllocatedAttributeRepo: c.heroAllocatedAttributeRepo,
		classAdvancedReqRepo:       c.classAdvancedReqRepo,
	}

	// 初始化 HeroAttributeService（依赖 repository 和 HeroService）
	c.HeroAttributeService = &HeroAttributeService{
		db:                         db,
		heroRepo:                   c.heroRepo,
		attributeUpgradeCostRepo:   c.attributeUpgradeCostRepo,
		attributeOpRepo:            c.attributeOpRepo,
		heroAttributeTypeRepo:      c.heroAttributeTypeRepo,
		heroAllocatedAttributeRepo: c.heroAllocatedAttributeRepo,
		heroService:                c.HeroService,
	}

	// 初始化 HeroSkillService（依赖 repository 和 HeroService）
	c.HeroSkillService = &HeroSkillService{
		db:                   db,
		heroRepo:             c.heroRepo,
		heroSkillRepo:        c.heroSkillRepo,
		heroClassHistoryRepo: c.heroClassHistoryRepo,
		classSkillPoolRepo:   c.classSkillPoolRepo,
		skillRepo:            c.skillRepo,
		skillUpgradeCostRepo: c.skillUpgradeCostRepo,
		skillOpRepo:          c.skillOpRepo,
		heroService:          c.HeroService,
	}

	// 初始化 ClassService（依赖 repository）
	c.ClassService = NewClassService(c.classRepo, c.classAdvancedReqRepo)

	// 初始化 SkillDetailService（依赖 repository）
	c.SkillDetailService = NewSkillDetailService(
		c.skillRepo,
		c.skillUnlockActionRepo,
		c.actionRepo,
		c.actionEffectRepo,
		c.effectRepo,
		c.skillCategoryRepo,
	)

	return c
}

// GetHeroService 获取英雄服务
func (c *ServiceContainer) GetHeroService() *HeroService {
	return c.HeroService
}

// GetHeroAttributeService 获取英雄属性服务
func (c *ServiceContainer) GetHeroAttributeService() *HeroAttributeService {
	return c.HeroAttributeService
}

// GetHeroSkillService 获取英雄技能服务
func (c *ServiceContainer) GetHeroSkillService() *HeroSkillService {
	return c.HeroSkillService
}

// GetClassService 获取职业服务
func (c *ServiceContainer) GetClassService() *ClassService {
	return c.ClassService
}

// GetSkillDetailService 获取技能详情服务
func (c *ServiceContainer) GetSkillDetailService() *SkillDetailService {
	return c.SkillDetailService
}

// GetHeroLevelRequirementRepo 获取英雄等级需求仓储
func (c *ServiceContainer) GetHeroLevelRequirementRepo() interfaces.HeroLevelRequirementRepository {
	return c.heroLevelRequirementRepo
}

// GetSkillUpgradeCostRepo 获取技能升级消耗仓储
func (c *ServiceContainer) GetSkillUpgradeCostRepo() interfaces.SkillUpgradeCostRepository {
	return c.skillUpgradeCostRepo
}

// GetAttributeUpgradeCostRepo 获取属性升级消耗仓储
func (c *ServiceContainer) GetAttributeUpgradeCostRepo() interfaces.AttributeUpgradeCostRepository {
	return c.attributeUpgradeCostRepo
}
