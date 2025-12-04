package service

import (
	"database/sql"

	"tsu-self/internal/modules/auth/client"
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
	equipmentSetRepo           interfaces.EquipmentSetRepository
	equipmentRepo              interfaces.EquipmentRepository
	itemRepo                   interfaces.ItemRepository
	teamRepo                   interfaces.TeamRepository
	teamMemberRepo             interfaces.TeamMemberRepository
	teamJoinRequestRepo        interfaces.TeamJoinRequestRepository
	teamInvitationRepo         interfaces.TeamInvitationRepository
	teamKickedRecordRepo       interfaces.TeamKickedRecordRepository
	teamWarehouseRepo          interfaces.TeamWarehouseRepository
	teamWarehouseItemRepo      interfaces.TeamWarehouseItemRepository
	heroWalletRepo             interfaces.HeroWalletRepository
	teamLootHistoryRepo        interfaces.TeamLootHistoryRepository
	teamWarehouseLootLogRepo   interfaces.TeamWarehouseLootLogRepository
	playerItemRepo             interfaces.PlayerItemRepository
	dungeonRepo                interfaces.DungeonRepository
	teamDungeonProgressRepo    interfaces.TeamDungeonProgressRepository
	teamDungeonRecordRepo      interfaces.TeamDungeonRecordRepository
	battleReportRepo           interfaces.BattleReportRepository

	// 所有 Service（共享实例）
	HeroService           *HeroService
	HeroAttributeService  *HeroAttributeService
	HeroSkillService      *HeroSkillService
	ClassService          *ClassService
	SkillDetailService    *SkillDetailService
	EquipmentSetService   *EquipmentSetService
	TeamService           *TeamService
	TeamMemberService     *TeamMemberService
	TeamWarehouseService  *TeamWarehouseService
	TeamDungeonService    *TeamDungeonService
	TeamPermissionService *TeamPermissionService
	BattleResultService   *BattleResultService
}

// NewServiceContainer 创建服务容器
// ketoClient 与 permissionCache 都是可选依赖
func NewServiceContainer(db *sql.DB, ketoClient *client.KetoClient, permissionCache permissionCacheClient) *ServiceContainer {
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
	c.equipmentSetRepo = impl.NewEquipmentSetRepository(db)
	c.equipmentRepo = impl.NewEquipmentRepository(db)
	c.itemRepo = impl.NewItemRepository(db)
	c.teamRepo = impl.NewTeamRepository(db)
	c.teamMemberRepo = impl.NewTeamMemberRepository(db)
	c.teamJoinRequestRepo = impl.NewTeamJoinRequestRepository(db)
	c.teamInvitationRepo = impl.NewTeamInvitationRepository(db)
	c.teamKickedRecordRepo = impl.NewTeamKickedRecordRepository(db)
	c.teamWarehouseRepo = impl.NewTeamWarehouseRepository(db)
	c.teamWarehouseItemRepo = impl.NewTeamWarehouseItemRepository(db)
	c.heroWalletRepo = impl.NewHeroWalletRepository(db)
	c.teamLootHistoryRepo = impl.NewTeamLootHistoryRepository(db)
	c.teamWarehouseLootLogRepo = impl.NewTeamWarehouseLootLogRepository(db)
	c.playerItemRepo = impl.NewPlayerItemRepository(db)
	c.dungeonRepo = impl.NewDungeonRepository(db)
	c.teamDungeonProgressRepo = impl.NewTeamDungeonProgressRepository(db)
	c.teamDungeonRecordRepo = impl.NewTeamDungeonRecordRepository(db)
	c.battleReportRepo = impl.NewBattleReportRepository(db)

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

	// 初始化 EquipmentSetService（依赖 repository）
	c.EquipmentSetService = NewEquipmentSetService(
		c.equipmentSetRepo,
		c.equipmentRepo,
		c.itemRepo,
	)

	// 初始化 TeamPermissionService（依赖 repository、ketoClient、permissionCache）
	c.TeamPermissionService = NewTeamPermissionService(db, ketoClient, permissionCache)

	// 初始化 TeamService（依赖 repository 和 TeamPermissionService）
	c.TeamService = NewTeamService(db, c.TeamPermissionService)

	// 初始化 TeamMemberService（依赖 repository 和 TeamPermissionService）
	c.TeamMemberService = NewTeamMemberService(db, c.TeamPermissionService)

	// 初始化 TeamWarehouseService（依赖 repository）
	c.TeamWarehouseService = &TeamWarehouseService{
		db:                    db,
		teamMemberRepo:        c.teamMemberRepo,
		teamWarehouseRepo:     c.teamWarehouseRepo,
		teamWarehouseItemRepo: c.teamWarehouseItemRepo,
		heroWalletRepo:        c.heroWalletRepo,
		lootHistoryRepo:       c.teamLootHistoryRepo,
		lootLogRepo:           c.teamWarehouseLootLogRepo,
		itemRepo:              c.itemRepo,
		heroRepo:              c.heroRepo,
		playerItemRepo:        c.playerItemRepo,
	}

	// 初始化 TeamDungeonService（依赖 repository）
	c.TeamDungeonService = NewTeamDungeonService(db, &TeamDungeonDependencies{
		WarehouseService: c.TeamWarehouseService,
		TeamMemberRepo:   c.teamMemberRepo,
		DungeonRepo:      c.dungeonRepo,
		ProgressRepo:     c.teamDungeonProgressRepo,
		RecordRepo:       c.teamDungeonRecordRepo,
		HeroRepo:         c.heroRepo,
	})

	c.BattleResultService = NewBattleResultService(c.battleReportRepo, c.TeamDungeonService)

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

// GetEquipmentSetService 获取装备套装服务
func (c *ServiceContainer) GetEquipmentSetService() *EquipmentSetService {
	return c.EquipmentSetService
}

// GetTeamDungeonService 获取团队地城服务
func (c *ServiceContainer) GetTeamDungeonService() *TeamDungeonService {
	return c.TeamDungeonService
}

// GetTeamService 获取团队服务
func (c *ServiceContainer) GetTeamService() *TeamService {
	return c.TeamService
}

// GetTeamMemberService 获取团队成员服务
func (c *ServiceContainer) GetTeamMemberService() *TeamMemberService {
	return c.TeamMemberService
}

// GetTeamWarehouseService 获取团队仓库服务
func (c *ServiceContainer) GetTeamWarehouseService() *TeamWarehouseService {
	return c.TeamWarehouseService
}

// GetTeamPermissionService 获取团队权限服务
func (c *ServiceContainer) GetTeamPermissionService() *TeamPermissionService {
	return c.TeamPermissionService
}

// GetBattleResultService 获取战斗结果服务
func (c *ServiceContainer) GetBattleResultService() *BattleResultService {
	return c.BattleResultService
}
