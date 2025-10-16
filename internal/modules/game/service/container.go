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
	heroRepo                        interfaces.HeroRepository
	heroClassHistoryRepo            interfaces.HeroClassHistoryRepository
	heroSkillRepo                   interfaces.HeroSkillRepository
	classRepo                       interfaces.ClassRepository
	classSkillPoolRepo              interfaces.ClassSkillPoolRepository
	classAdvancedReqRepo            interfaces.ClassAdvancedRequirementRepository
	skillRepo                       interfaces.SkillRepository
	heroAttributeTypeRepo           interfaces.HeroAttributeTypeRepository
	heroLevelRequirementRepo        interfaces.HeroLevelRequirementRepository
	attributeUpgradeCostRepo        interfaces.AttributeUpgradeCostRepository
	attributeOpRepo                 interfaces.HeroAttributeOperationRepository
	heroAllocatedAttributeRepo      interfaces.HeroAllocatedAttributeRepository
	skillUpgradeCostRepo            interfaces.SkillUpgradeCostRepository
	skillOpRepo                     interfaces.HeroSkillOperationRepository

	// 所有 Service（共享实例）
	HeroService           *HeroService
	HeroAttributeService  *HeroAttributeService
	HeroSkillService      *HeroSkillService
	ClassService          *ClassService
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

	// 初始化 HeroService（依赖 repository）
	c.HeroService = &HeroService{
		db:                           db,
		heroRepo:                     c.heroRepo,
		heroClassHistoryRepo:         c.heroClassHistoryRepo,
		heroSkillRepo:                c.heroSkillRepo,
		classRepo:                    c.classRepo,
		classSkillPoolRepo:           c.classSkillPoolRepo,
		heroAttributeTypeRepo:        c.heroAttributeTypeRepo,
		heroLevelRequirementRepo:     c.heroLevelRequirementRepo,
		heroAllocatedAttributeRepo:   c.heroAllocatedAttributeRepo,
		classAdvancedReqRepo:         c.classAdvancedReqRepo,
	}

	// 初始化 HeroAttributeService（依赖 repository 和 HeroService）
	c.HeroAttributeService = &HeroAttributeService{
		db:                           db,
		heroRepo:                     c.heroRepo,
		attributeUpgradeCostRepo:     c.attributeUpgradeCostRepo,
		attributeOpRepo:              c.attributeOpRepo,
		heroAttributeTypeRepo:        c.heroAttributeTypeRepo,
		heroAllocatedAttributeRepo:   c.heroAllocatedAttributeRepo,
		heroService:                  c.HeroService,
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
