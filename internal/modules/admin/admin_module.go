package admin

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	custommiddleware "tsu-self/internal/middleware"
	"tsu-self/internal/modules/admin/handler"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/i18n"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/metrics"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/security"
	"tsu-self/internal/pkg/trace"
	"tsu-self/internal/pkg/validation"
	"tsu-self/internal/pkg/validator"

	_ "tsu-self/docs/admin" // Swagger 生成的文档

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/server"
	_ "github.com/lib/pq"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Module Admin module
type AdminModule struct {
	basemodule.BaseModule
	db                          *sql.DB
	httpServer                  *echo.Echo
	authHandler                 *handler.AuthHandler
	passwordRecoveryHandler     *handler.PasswordRecoveryHandler
	permissionHandler           *handler.PermissionHandler
	userHandler                 *handler.UserHandler
	classHandler                *handler.ClassHandler
	skillCategoryHandler        *handler.SkillCategoryHandler
	actionCategoryHandler       *handler.ActionCategoryHandler
	damageTypeHandler           *handler.DamageTypeHandler
	heroAttributeTypeHandler    *handler.HeroAttributeTypeHandler
	tagHandler                  *handler.TagHandler
	tagRelationHandler          *handler.TagRelationHandler
	itemConfigHandler           *handler.ItemConfigHandler
	equipmentSlotHandler        *handler.EquipmentSlotHandler
	equipmentSetHandler         *handler.EquipmentSetHandler
	dropPoolHandler             *handler.DropPoolHandler
	worldDropHandler            *handler.WorldDropHandler
	effectTypeDefinitionHandler *handler.EffectTypeDefinitionHandler
	formulaVariableHandler      *handler.FormulaVariableHandler
	rangeConfigRuleHandler      *handler.RangeConfigRuleHandler
	actionTypeDefinitionHandler *handler.ActionTypeDefinitionHandler
	skillHandler                *handler.SkillHandler
	skillUpgradeCostHandler     *handler.SkillUpgradeCostHandler
	advancedRequirementHandler  *handler.ClassAdvancedRequirementHandler
	effectHandler               *handler.EffectHandler
	buffHandler                 *handler.BuffHandler
	buffEffectHandler           *handler.BuffEffectHandler
	actionFlagHandler           *handler.ActionFlagHandler
	actionHandler               *handler.ActionHandler
	actionEffectHandler         *handler.ActionEffectHandler
	skillUnlockActionHandler    *handler.SkillUnlockActionHandler
	classSkillPoolHandler       *handler.ClassSkillPoolHandler
	monsterHandler              *handler.MonsterHandler
	dungeonHandler              *handler.DungeonHandler
	dungeonRoomHandler          *handler.DungeonRoomHandler
	dungeonBattleHandler        *handler.DungeonBattleHandler
	dungeonEventHandler         *handler.DungeonEventHandler
	teamAdminHandler            *handler.TeamAdminHandler
	toolsHandler                *handler.ToolsHandler
	respWriter                  response.Writer
}

// GetType returns module type
func (m *AdminModule) GetType() string {
	return "admin"
}

// Version returns module version
func (m *AdminModule) Version() string {
	return "1.0.0"
}

// OnAppConfigurationLoaded 当App初始化时调用
func (m *AdminModule) OnAppConfigurationLoaded(app module.App) {
	m.BaseModule.OnAppConfigurationLoaded(app)
}

// OnInit module initialization
func (m *AdminModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	metrics.SetServiceName("admin")
	// 按照 mqant 官方推荐：在每个模块的 OnInit 中配置服务注册参数
	// TTL = 30s, 心跳间隔 = 15s (TTL 必须大于心跳间隔)
	m.BaseModule.OnInit(m, app, settings,
		server.RegisterInterval(15*time.Second),
		server.RegisterTTL(30*time.Second),
	)

	// 1. Initialize response writer
	m.initResponseWriter()

	// 2. Initialize database connection (optional for admin module)
	if err := m.initDatabase(settings); err != nil {
		fmt.Printf("[Admin Module] Warning: Database initialization failed: %v\n", err)
	}

	// 3. Initialize HTTP server
	m.initHTTPServer()

	// 4. Initialize handlers
	m.initHandlers()

	// 5. Setup routes
	m.setupRoutes()

	// 6. Start HTTP server in background
	go m.startHTTPServer(settings)

	m.GetServer().Options()
}

// initDatabase initializes database connection
func (m *AdminModule) initDatabase(settings *conf.ModuleSettings) error {
	// Read from environment variable first
	dbURL := os.Getenv("TSU_ADMIN_DATABASE_URL")
	if dbURL == "" {
		// Fallback to config file
		if settings != nil && settings.Settings != nil {
			dbURLInterface, ok := settings.Settings["database_url"]
			if ok {
				dbURL, _ = dbURLInterface.(string)
			}
		}
	}

	if dbURL == "" {
		return fmt.Errorf("database URL not configured")
	}

	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	m.db = db
	fmt.Println("[Admin Module] Database connected successfully")

	// 启动数据库连接池监控
	go m.startDBPoolMonitoring(db)

	return nil
}

// initHTTPServer initializes HTTP server
func (m *AdminModule) initHTTPServer() {
	m.httpServer = echo.New()

	// Hide banner
	m.httpServer.HideBanner = true
	m.httpServer.HidePort = true

	// Register validator
	m.httpServer.Validator = validator.New()

	// 获取全局 logger
	logger := log.GetLogger()

	// 获取环境变量
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	// ========== 中间件配置（顺序很重要！） ==========

	// 1. TraceID 中间件 - 最先执行，生成或提取 TraceID
	m.httpServer.Use(trace.Middleware())

	// 2. Metrics 中间件 - 记录 HTTP 方法到 context（用于 Prometheus）
	m.httpServer.Use(metrics.Middleware())

	// 3. i18n 中间件 - 语言检测和设置
	m.httpServer.Use(i18n.Middleware())

	// 4. Logging 中间件 - 记录请求日志（依赖 TraceID）
	loggingConfig := custommiddleware.DefaultLoggingConfig()
	if environment == "development" {
		// 开发环境启用详细日志
		loggingConfig.DetailedLog = true
		loggingConfig.LogRequestBody = true // 可以记录请求体
	}
	m.httpServer.Use(custommiddleware.LoggingMiddlewareWithConfig(logger, loggingConfig))

	// 5. Recovery 中间件 - 捕获 panic
	m.httpServer.Use(custommiddleware.RecoveryMiddleware(m.respWriter, logger))

	// 6. Error 中间件 - 统一错误处理
	m.httpServer.Use(custommiddleware.ErrorMiddleware(m.respWriter, logger))

	// 7. CORS 中间件
	m.httpServer.Use(security.CORSMiddleware())

	fmt.Println("[Admin Module] HTTP middlewares configured:")
	fmt.Println("  ✓ TraceID (自动生成追踪ID)")
	fmt.Println("  ✓ Metrics (Prometheus 指标收集)")
	fmt.Println("  ✓ i18n (国际化支持)")
	fmt.Printf("  ✓ Logging (日志记录 - %s)\n", environment)
	fmt.Println("  ✓ Recovery (Panic 恢复)")
	fmt.Println("  ✓ Error (统一错误处理)")
	fmt.Println("  ✓ CORS (跨域支持)")
}

// initResponseWriter initializes response writer
func (m *AdminModule) initResponseWriter() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	// 使用全局 logger
	logger := log.GetLogger()
	m.respWriter = response.NewResponseHandler(logger, environment)
	fmt.Println("[Admin Module] Response writer initialized")
}

// initHandlers initializes HTTP handlers
func (m *AdminModule) initHandlers() {
	m.authHandler = handler.NewAuthHandler(m, m.respWriter)
	m.passwordRecoveryHandler = handler.NewPasswordRecoveryHandler(m, m.respWriter)
	m.permissionHandler = handler.NewPermissionHandler(m, m.respWriter)
	m.userHandler = handler.NewUserHandler(m, m.respWriter)
	m.classHandler = handler.NewClassHandler(m.db, m.respWriter)
	m.skillCategoryHandler = handler.NewSkillCategoryHandler(m.db, m.respWriter)
	m.actionCategoryHandler = handler.NewActionCategoryHandler(m.db, m.respWriter)
	m.damageTypeHandler = handler.NewDamageTypeHandler(m.db, m.respWriter)
	m.heroAttributeTypeHandler = handler.NewHeroAttributeTypeHandler(m.db, m.respWriter)
	m.tagHandler = handler.NewTagHandler(m.db, m.respWriter)
	m.tagRelationHandler = handler.NewTagRelationHandler(m.db, m.respWriter)
	m.itemConfigHandler = handler.NewItemConfigHandler(m.db, m.respWriter)
	m.equipmentSlotHandler = handler.NewEquipmentSlotHandler(m.db, m.respWriter)
	m.equipmentSetHandler = handler.NewEquipmentSetHandler(m.db, m.respWriter)
	m.dropPoolHandler = handler.NewDropPoolHandler(m.db, m.respWriter)
	m.worldDropHandler = handler.NewWorldDropHandler(m.db, m.respWriter)
	m.effectTypeDefinitionHandler = handler.NewEffectTypeDefinitionHandler(m.db, m.respWriter)
	m.formulaVariableHandler = handler.NewFormulaVariableHandler(m.db, m.respWriter)
	m.rangeConfigRuleHandler = handler.NewRangeConfigRuleHandler(m.db, m.respWriter)
	m.actionTypeDefinitionHandler = handler.NewActionTypeDefinitionHandler(m.db, m.respWriter)
	m.skillHandler = handler.NewSkillHandler(m.db, m.respWriter)
	m.skillUpgradeCostHandler = handler.NewSkillUpgradeCostHandler(m.db, m.respWriter)
	m.advancedRequirementHandler = handler.NewClassAdvancedRequirementHandler(m.db, m.respWriter)
	m.effectHandler = handler.NewEffectHandler(m.db, m.respWriter)
	m.buffHandler = handler.NewBuffHandler(m.db, m.respWriter)
	m.buffEffectHandler = handler.NewBuffEffectHandler(m.db, m.respWriter)
	m.actionFlagHandler = handler.NewActionFlagHandler(m.db, m.respWriter)
	m.actionHandler = handler.NewActionHandler(m.db, m.respWriter)
	m.actionEffectHandler = handler.NewActionEffectHandler(m.db, m.respWriter)
	m.skillUnlockActionHandler = handler.NewSkillUnlockActionHandler(m.db, m.respWriter)
	m.classSkillPoolHandler = handler.NewClassSkillPoolHandler(m.db, m.respWriter)
	m.monsterHandler = handler.NewMonsterHandler(m.db, m.respWriter)
	m.dungeonHandler = handler.NewDungeonHandler(m.db, m.respWriter)
	m.dungeonRoomHandler = handler.NewDungeonRoomHandler(m.db, m.respWriter)
	m.dungeonBattleHandler = handler.NewDungeonBattleHandler(m.db, m.respWriter)
	m.dungeonEventHandler = handler.NewDungeonEventHandler(m.db, m.respWriter)
	m.toolsHandler = handler.NewToolsHandler(service.NewToolsService(m.db), m.respWriter)

	// 团队管理Handler (通过 RPC 调用 Game Server)
	m.teamAdminHandler = handler.NewTeamAdminHandler(m, m.respWriter)
}

// setupRoutes sets up HTTP routes
func (m *AdminModule) setupRoutes() {
	// 获取全局 logger
	logger := log.GetLogger()

	requirePerm := func(code string) echo.MiddlewareFunc {
		return custommiddleware.RequirePermission(m.App, m, m.respWriter, logger, code)
	}

	// API v1 group
	v1 := m.httpServer.Group("/api/v1")

	// Admin routes - 所有 admin 相关接口统一使用 /admin 前缀
	admin := v1.Group("/admin")

	// Auth routes (公开访问，不需要认证)
	auth := admin.Group("/auth")
	{
		auth.POST("/register", m.authHandler.Register)
		auth.POST("/login", m.authHandler.Login)
		auth.POST("/logout", m.authHandler.Logout)
		auth.GET("/users/:user_id", m.authHandler.GetUser)

		// 密码重置 (公开访问)
		auth.POST("/recovery/initiate", m.passwordRecoveryHandler.InitiateRecovery)
		auth.POST("/recovery/verify", m.passwordRecoveryHandler.VerifyRecoveryCode)
		auth.POST("/password/reset", m.passwordRecoveryHandler.ResetPassword)
		auth.POST("/password/reset-with-code", m.passwordRecoveryHandler.ResetPasswordWithCode) // 验证码重置密码
	}

	// Admin protected routes (需要认证，应用认证中间件)
	// 这些路由的请求必须经过 Oathkeeper 验证，并且会从 Header 中提取用户信息
	adminProtected := admin.Group("")
	adminProtected.Use(custommiddleware.AuthMiddleware(m.respWriter, logger))
	adminProtected.Use(validation.UUIDValidationMiddleware(m.respWriter))
	userRead := requirePerm("user:read")
	userUpdate := requirePerm("user:update")
	userBan := requirePerm("user:ban")
	userDelete := requirePerm("user:delete")
	roleRead := requirePerm("role:read")
	roleCreate := requirePerm("role:create")
	roleUpdate := requirePerm("role:update")
	roleDelete := requirePerm("role:delete")
	roleAssign := requirePerm("role:assign")
	permRead := requirePerm("permission:read")
	permAssign := requirePerm("permission:assign")
	permGrantUser := requirePerm("permission:grant_user")
	classManage := requirePerm("class:manage")
	skillManage := requirePerm("skill:manage")
	systemConfig := requirePerm("system:config")
	worldDropItemManage := requirePerm("world-drop:manage-items")
	// TODO: 团队管理路由尚未开放，先保留权限定义以防后续接入
	teamRead := requirePerm("team:read")
	teamModerate := requirePerm("team:moderate")
	_ = teamRead
	_ = teamModerate
	{
		// 用户管理
		adminProtected.GET("/users/me", m.userHandler.GetCurrentUserProfile, userRead) // 🆕 示例：获取当前登录用户信息
		adminProtected.GET("/users", m.userHandler.GetUsers, userRead)
		adminProtected.GET("/users/:id", m.userHandler.GetUser, userRead)
		adminProtected.PUT("/users/:id", m.userHandler.UpdateUser, userUpdate)
		adminProtected.POST("/users/:id/ban", m.userHandler.BanUser, userBan)
		adminProtected.POST("/users/:id/unban", m.userHandler.UnbanUser, userBan)
		adminProtected.DELETE("/users/:user_id", m.authHandler.DeleteUser, userDelete) // 删除用户

		// 角色管理
		adminProtected.GET("/roles", m.permissionHandler.GetRoles, roleRead)
		adminProtected.POST("/roles", m.permissionHandler.CreateRole, roleCreate)
		adminProtected.PUT("/roles/:id", m.permissionHandler.UpdateRole, roleUpdate)
		adminProtected.DELETE("/roles/:id", m.permissionHandler.DeleteRole, roleDelete)

		// 角色-权限管理
		adminProtected.GET("/roles/:id/permissions", m.permissionHandler.GetRolePermissions, roleRead)
		adminProtected.POST("/roles/:id/permissions", m.permissionHandler.AssignPermissionsToRole, permAssign)

		// 权限管理
		adminProtected.GET("/permissions", m.permissionHandler.GetPermissions, permRead)
		adminProtected.GET("/permission-groups", m.permissionHandler.GetPermissionGroups, permRead)

		// 用户-角色管理
		adminProtected.GET("/users/:user_id/roles", m.permissionHandler.GetUserRoles, roleRead)
		adminProtected.POST("/users/:user_id/roles", m.permissionHandler.AssignRolesToUser, roleAssign)
		adminProtected.DELETE("/users/:user_id/roles", m.permissionHandler.RevokeRolesFromUser, roleAssign)

		// 用户-权限管理
		adminProtected.GET("/users/:user_id/permissions", m.permissionHandler.GetUserPermissions, permRead)
		adminProtected.POST("/users/:user_id/permissions", m.permissionHandler.GrantPermissionsToUser, permGrantUser)
		adminProtected.DELETE("/users/:user_id/permissions", m.permissionHandler.RevokePermissionsFromUser, permGrantUser)

		// 职业管理
		adminProtected.GET("/classes", m.classHandler.GetClasses, classManage)
		adminProtected.POST("/classes", m.classHandler.CreateClass, classManage)
		adminProtected.GET("/classes/:id", m.classHandler.GetClass, classManage)
		adminProtected.PUT("/classes/:id", m.classHandler.UpdateClass, classManage)
		adminProtected.DELETE("/classes/:id", m.classHandler.DeleteClass, classManage)

		// 职业属性加成管理
		adminProtected.GET("/classes/:id/attribute-bonuses", m.classHandler.GetClassAttributeBonuses, classManage)
		adminProtected.POST("/classes/:id/attribute-bonuses", m.classHandler.CreateAttributeBonus, classManage)
		adminProtected.POST("/classes/:id/attribute-bonuses/batch", m.classHandler.BatchSetAttributeBonuses, classManage)
		adminProtected.PUT("/classes/:id/attribute-bonuses/:bonus_id", m.classHandler.UpdateAttributeBonus, classManage)
		adminProtected.DELETE("/classes/:id/attribute-bonuses/:bonus_id", m.classHandler.DeleteAttributeBonus, classManage)

		// 职业进阶路径查询（嵌套在职业下）
		adminProtected.GET("/classes/:id/advancement", m.classHandler.GetClassAdvancement, classManage)
		adminProtected.GET("/classes/:id/advancement-paths", m.classHandler.GetClassAdvancementPaths, classManage)
		adminProtected.GET("/classes/:id/advancement-sources", m.classHandler.GetClassAdvancementSources, classManage)

		// 职业进阶要求管理（独立接口）
		adminProtected.GET("/advancement-requirements", m.advancedRequirementHandler.GetAdvancedRequirements, classManage)
		adminProtected.POST("/advancement-requirements", m.advancedRequirementHandler.CreateAdvancedRequirement, classManage)
		adminProtected.POST("/advancement-requirements/batch", m.advancedRequirementHandler.BatchCreateAdvancedRequirements, classManage)
		adminProtected.GET("/advancement-requirements/:id", m.advancedRequirementHandler.GetAdvancedRequirement, classManage)
		adminProtected.PUT("/advancement-requirements/:id", m.advancedRequirementHandler.UpdateAdvancedRequirement, classManage)
		adminProtected.DELETE("/advancement-requirements/:id", m.advancedRequirementHandler.DeleteAdvancedRequirement, classManage)

		// 职业技能池管理（嵌套在职业下）
		adminProtected.GET("/classes/:class_id/skill-pools", m.classSkillPoolHandler.GetClassSkillPoolsByClassID, classManage)

		// 职业技能池管理（独立接口）
		adminProtected.GET("/class-skill-pools", m.classSkillPoolHandler.GetClassSkillPools, classManage)
		adminProtected.POST("/class-skill-pools", m.classSkillPoolHandler.CreateClassSkillPool, classManage)
		adminProtected.GET("/class-skill-pools/:id", m.classSkillPoolHandler.GetClassSkillPool, classManage)
		adminProtected.PUT("/class-skill-pools/:id", m.classSkillPoolHandler.UpdateClassSkillPool, classManage)
		adminProtected.DELETE("/class-skill-pools/:id", m.classSkillPoolHandler.DeleteClassSkillPool, classManage)

		// 技能类别管理
		adminProtected.GET("/skill-categories", m.skillCategoryHandler.GetSkillCategories, skillManage)
		adminProtected.POST("/skill-categories", m.skillCategoryHandler.CreateSkillCategory, skillManage)
		adminProtected.GET("/skill-categories/:id", m.skillCategoryHandler.GetSkillCategory, skillManage)
		adminProtected.PUT("/skill-categories/:id", m.skillCategoryHandler.UpdateSkillCategory, skillManage)
		adminProtected.DELETE("/skill-categories/:id", m.skillCategoryHandler.DeleteSkillCategory, skillManage)

		// 动作类别管理
		adminProtected.GET("/action-categories", m.actionCategoryHandler.GetActionCategories, skillManage)
		adminProtected.POST("/action-categories", m.actionCategoryHandler.CreateActionCategory, skillManage)
		adminProtected.GET("/action-categories/:id", m.actionCategoryHandler.GetActionCategory, skillManage)
		adminProtected.PUT("/action-categories/:id", m.actionCategoryHandler.UpdateActionCategory, skillManage)
		adminProtected.DELETE("/action-categories/:id", m.actionCategoryHandler.DeleteActionCategory, skillManage)

		// 伤害类型管理
		adminProtected.GET("/damage-types", m.damageTypeHandler.GetDamageTypes, systemConfig)
		adminProtected.POST("/damage-types", m.damageTypeHandler.CreateDamageType, systemConfig)
		adminProtected.GET("/damage-types/:id", m.damageTypeHandler.GetDamageType, systemConfig)
		adminProtected.PUT("/damage-types/:id", m.damageTypeHandler.UpdateDamageType, systemConfig)
		adminProtected.DELETE("/damage-types/:id", m.damageTypeHandler.DeleteDamageType, systemConfig)

		// 属性类型管理
		adminProtected.GET("/hero-attribute-types", m.heroAttributeTypeHandler.GetHeroAttributeTypes, systemConfig)
		adminProtected.POST("/hero-attribute-types", m.heroAttributeTypeHandler.CreateHeroAttributeType, systemConfig)
		adminProtected.GET("/hero-attribute-types/:id", m.heroAttributeTypeHandler.GetHeroAttributeType, systemConfig)
		adminProtected.PUT("/hero-attribute-types/:id", m.heroAttributeTypeHandler.UpdateHeroAttributeType, systemConfig)
		adminProtected.DELETE("/hero-attribute-types/:id", m.heroAttributeTypeHandler.DeleteHeroAttributeType, systemConfig)

		// 标签管理
		adminProtected.GET("/tags", m.tagHandler.GetTags, systemConfig)
		adminProtected.POST("/tags", m.tagHandler.CreateTag, systemConfig)
		adminProtected.GET("/tags/:id", m.tagHandler.GetTag, systemConfig)
		adminProtected.PUT("/tags/:id", m.tagHandler.UpdateTag, systemConfig)
		adminProtected.DELETE("/tags/:id", m.tagHandler.DeleteTag, systemConfig)

		// 标签关联管理
		adminProtected.GET("/tags/:tag_id/entities", m.tagRelationHandler.GetTagEntities, systemConfig)
		adminProtected.GET("/entities/:entity_type/:entity_id/tags", m.tagRelationHandler.GetEntityTags, systemConfig)
		adminProtected.POST("/entities/:entity_type/:entity_id/tags", m.tagRelationHandler.AddTagToEntity, systemConfig)
		adminProtected.POST("/entities/:entity_type/:entity_id/tags/batch", m.tagRelationHandler.BatchSetEntityTags, systemConfig)
		adminProtected.DELETE("/entities/:entity_type/:entity_id/tags/:tag_id", m.tagRelationHandler.RemoveTagFromEntity, systemConfig)

		// 物品配置管理
		adminProtected.GET("/items", m.itemConfigHandler.ListItems, systemConfig)
		adminProtected.POST("/items", m.itemConfigHandler.CreateItem, systemConfig)
		adminProtected.GET("/items/:id", m.itemConfigHandler.GetItem, systemConfig)
		adminProtected.PUT("/items/:id", m.itemConfigHandler.UpdateItem, systemConfig)
		adminProtected.DELETE("/items/:id", m.itemConfigHandler.DeleteItem, systemConfig)
		adminProtected.GET("/items/:id/tags", m.itemConfigHandler.GetItemTags, systemConfig)
		adminProtected.POST("/items/:id/tags", m.itemConfigHandler.AddItemTags, systemConfig)
		adminProtected.PUT("/items/:id/tags", m.itemConfigHandler.UpdateItemTags, systemConfig)
		adminProtected.DELETE("/items/:id/tags/:tag_id", m.itemConfigHandler.RemoveItemTag, systemConfig)
		// 物品职业关联管理
		adminProtected.POST("/items/:id/classes", m.itemConfigHandler.AddItemClasses, systemConfig)
		adminProtected.GET("/items/:id/classes", m.itemConfigHandler.GetItemClasses, systemConfig)
		adminProtected.PUT("/items/:id/classes", m.itemConfigHandler.UpdateItemClasses, systemConfig)
		adminProtected.DELETE("/items/:id/classes/:class_id", m.itemConfigHandler.RemoveItemClass, systemConfig)

		// 装备槽位配置管理
		adminProtected.GET("/equipment-slots", m.equipmentSlotHandler.GetSlotList, systemConfig)
		adminProtected.POST("/equipment-slots", m.equipmentSlotHandler.CreateSlot, systemConfig)
		adminProtected.GET("/equipment-slots/:id", m.equipmentSlotHandler.GetSlot, systemConfig)
		adminProtected.PUT("/equipment-slots/:id", m.equipmentSlotHandler.UpdateSlot, systemConfig)
		adminProtected.DELETE("/equipment-slots/:id", m.equipmentSlotHandler.DeleteSlot, systemConfig)

		// 装备套装配置管理
		adminProtected.POST("/equipment-sets", m.equipmentSetHandler.CreateSet, systemConfig)
		adminProtected.GET("/equipment-sets", m.equipmentSetHandler.GetSetList, systemConfig)
		adminProtected.GET("/equipment-sets/unassigned-items", m.equipmentSetHandler.GetUnassignedItems, systemConfig)
		adminProtected.GET("/equipment-sets/:id", m.equipmentSetHandler.GetSet, systemConfig)
		adminProtected.PUT("/equipment-sets/:id", m.equipmentSetHandler.UpdateSet, systemConfig)
		adminProtected.DELETE("/equipment-sets/:id", m.equipmentSetHandler.DeleteSet, systemConfig)
		adminProtected.GET("/equipment-sets/:id/items", m.equipmentSetHandler.GetSetItems, systemConfig)
		adminProtected.POST("/equipment-sets/:set_id/items/batch-assign", m.equipmentSetHandler.BatchAssignItems, systemConfig)
		adminProtected.POST("/equipment-sets/:set_id/items/batch-remove", m.equipmentSetHandler.BatchRemoveItems, systemConfig)
		adminProtected.DELETE("/equipment-sets/:set_id/items/:item_id", m.equipmentSetHandler.RemoveItem, systemConfig)

		// 掉落池配置管理
		adminProtected.GET("/drop-pools", m.dropPoolHandler.GetDropPoolList, systemConfig)
		adminProtected.POST("/drop-pools", m.dropPoolHandler.CreateDropPool, systemConfig)
		adminProtected.GET("/drop-pools/:id", m.dropPoolHandler.GetDropPool, systemConfig)
		adminProtected.PUT("/drop-pools/:id", m.dropPoolHandler.UpdateDropPool, systemConfig)
		adminProtected.DELETE("/drop-pools/:id", m.dropPoolHandler.DeleteDropPool, systemConfig)

		// 掉落池物品管理
		adminProtected.POST("/drop-pools/:pool_id/items", m.dropPoolHandler.AddDropPoolItem, systemConfig)
		adminProtected.GET("/drop-pools/:pool_id/items", m.dropPoolHandler.GetDropPoolItems, systemConfig)
		adminProtected.GET("/drop-pools/:pool_id/items/:item_id", m.dropPoolHandler.GetDropPoolItem, systemConfig)
		adminProtected.PUT("/drop-pools/:pool_id/items/:item_id", m.dropPoolHandler.UpdateDropPoolItem, systemConfig)
		adminProtected.DELETE("/drop-pools/:pool_id/items/:item_id", m.dropPoolHandler.RemoveDropPoolItem, systemConfig)

		// 世界掉落配置管理
		adminProtected.GET("/world-drops", m.worldDropHandler.GetWorldDropList, systemConfig)
		adminProtected.POST("/world-drops", m.worldDropHandler.CreateWorldDrop, systemConfig)
		adminProtected.GET("/world-drops/:id", m.worldDropHandler.GetWorldDrop, systemConfig)
		adminProtected.PUT("/world-drops/:id", m.worldDropHandler.UpdateWorldDrop, systemConfig)
		adminProtected.DELETE("/world-drops/:id", m.worldDropHandler.DeleteWorldDrop, systemConfig)
		adminProtected.GET("/world-drops/:id/items", m.worldDropHandler.ListWorldDropItems, systemConfig)
		adminProtected.POST("/world-drops/:id/items", m.worldDropHandler.CreateWorldDropItem, systemConfig, worldDropItemManage)
		adminProtected.PUT("/world-drops/:id/items/:item_id", m.worldDropHandler.UpdateWorldDropItem, systemConfig, worldDropItemManage)
		adminProtected.DELETE("/world-drops/:id/items/:item_id", m.worldDropHandler.DeleteWorldDropItem, systemConfig, worldDropItemManage)

		// 元数据管理 (需要认证)
		metadata := adminProtected.Group("/metadata", systemConfig)
		{
			// 元效果类型定义
			metadata.GET("/effect-type-definitions", m.effectTypeDefinitionHandler.GetEffectTypeDefinitions)
			metadata.GET("/effect-type-definitions/all", m.effectTypeDefinitionHandler.GetAllEffectTypeDefinitions)
			metadata.GET("/effect-type-definitions/:id", m.effectTypeDefinitionHandler.GetEffectTypeDefinition)

			// 公式变量（向后兼容，数据来自 metadata_dictionary 表）
			metadata.GET("/formula-variables", m.formulaVariableHandler.GetFormulaVariables)
			metadata.GET("/formula-variables/all", m.formulaVariableHandler.GetAllFormulaVariables)
			metadata.GET("/formula-variables/:id", m.formulaVariableHandler.GetFormulaVariable)

			// 范围配置规则
			metadata.GET("/range-config-rules", m.rangeConfigRuleHandler.GetRangeConfigRules)
			metadata.GET("/range-config-rules/all", m.rangeConfigRuleHandler.GetAllRangeConfigRules)
			metadata.GET("/range-config-rules/:id", m.rangeConfigRuleHandler.GetRangeConfigRule)

			// 动作类型定义
			metadata.GET("/action-type-definitions", m.actionTypeDefinitionHandler.GetActionTypeDefinitions)
			metadata.GET("/action-type-definitions/all", m.actionTypeDefinitionHandler.GetAllActionTypeDefinitions)
			metadata.GET("/action-type-definitions/:id", m.actionTypeDefinitionHandler.GetActionTypeDefinition)
		}

		// 技能管理
		adminProtected.GET("/skills", m.skillHandler.GetSkills, skillManage)
		adminProtected.POST("/skills", m.skillHandler.CreateSkill, skillManage)
		adminProtected.GET("/skills/:id", m.skillHandler.GetSkill, skillManage)
		adminProtected.PUT("/skills/:id", m.skillHandler.UpdateSkill, skillManage)
		adminProtected.DELETE("/skills/:id", m.skillHandler.DeleteSkill, skillManage)

		// 全局技能升级消耗管理
		adminProtected.GET("/skill-upgrade-costs", m.skillUpgradeCostHandler.GetSkillUpgradeCosts, skillManage)
		adminProtected.POST("/skill-upgrade-costs", m.skillUpgradeCostHandler.CreateSkillUpgradeCost, skillManage)
		adminProtected.GET("/skill-upgrade-costs/level/:level", m.skillUpgradeCostHandler.GetSkillUpgradeCostByLevel, skillManage)
		adminProtected.GET("/skill-upgrade-costs/:id", m.skillUpgradeCostHandler.GetSkillUpgradeCost, skillManage)
		adminProtected.PUT("/skill-upgrade-costs/:id", m.skillUpgradeCostHandler.UpdateSkillUpgradeCost, skillManage)
		adminProtected.DELETE("/skill-upgrade-costs/:id", m.skillUpgradeCostHandler.DeleteSkillUpgradeCost, skillManage)

		// 效果管理
		adminProtected.GET("/effects", m.effectHandler.GetEffects, skillManage)
		adminProtected.POST("/effects", m.effectHandler.CreateEffect, skillManage)
		adminProtected.GET("/effects/:id", m.effectHandler.GetEffect, skillManage)
		adminProtected.PUT("/effects/:id", m.effectHandler.UpdateEffect, skillManage)
		adminProtected.DELETE("/effects/:id", m.effectHandler.DeleteEffect, skillManage)

		// Buff管理
		adminProtected.GET("/buffs", m.buffHandler.GetBuffs, skillManage)
		adminProtected.POST("/buffs", m.buffHandler.CreateBuff, skillManage)
		adminProtected.GET("/buffs/:id", m.buffHandler.GetBuff, skillManage)
		adminProtected.PUT("/buffs/:id", m.buffHandler.UpdateBuff, skillManage)
		adminProtected.DELETE("/buffs/:id", m.buffHandler.DeleteBuff, skillManage)

		// Buff效果关联管理
		adminProtected.GET("/buffs/:buff_id/effects", m.buffEffectHandler.GetBuffEffects, skillManage)
		adminProtected.POST("/buffs/:buff_id/effects", m.buffEffectHandler.AddBuffEffect, skillManage)
		adminProtected.POST("/buffs/:buff_id/effects/batch", m.buffEffectHandler.BatchSetBuffEffects, skillManage)
		adminProtected.DELETE("/buffs/:buff_id/effects/:effect_id", m.buffEffectHandler.RemoveBuffEffect, skillManage)

		// 动作Flag管理
		adminProtected.GET("/action-flags", m.actionFlagHandler.GetActionFlags, skillManage)
		adminProtected.POST("/action-flags", m.actionFlagHandler.CreateActionFlag, skillManage)
		adminProtected.GET("/action-flags/:id", m.actionFlagHandler.GetActionFlag, skillManage)
		adminProtected.PUT("/action-flags/:id", m.actionFlagHandler.UpdateActionFlag, skillManage)
		adminProtected.DELETE("/action-flags/:id", m.actionFlagHandler.DeleteActionFlag, skillManage)

		// 动作管理
		adminProtected.GET("/actions", m.actionHandler.GetActions, skillManage)
		adminProtected.POST("/actions", m.actionHandler.CreateAction, skillManage)
		adminProtected.GET("/actions/:id", m.actionHandler.GetAction, skillManage)
		adminProtected.PUT("/actions/:id", m.actionHandler.UpdateAction, skillManage)
		adminProtected.DELETE("/actions/:id", m.actionHandler.DeleteAction, skillManage)

		// 动作效果关联管理
		adminProtected.GET("/actions/:action_id/effects", m.actionEffectHandler.GetActionEffects, skillManage)
		adminProtected.POST("/actions/:action_id/effects", m.actionEffectHandler.AddActionEffect, skillManage)
		adminProtected.POST("/actions/:action_id/effects/batch", m.actionEffectHandler.BatchSetActionEffects, skillManage)
		adminProtected.DELETE("/actions/:action_id/effects/:effect_id", m.actionEffectHandler.RemoveActionEffect, skillManage)

		// 技能解锁动作管理
		adminProtected.GET("/skills/:skill_id/unlock-actions", m.skillUnlockActionHandler.GetSkillUnlockActions, skillManage)
		adminProtected.POST("/skills/:skill_id/unlock-actions", m.skillUnlockActionHandler.AddSkillUnlockAction, skillManage)
		adminProtected.PUT("/skills/:skill_id/unlock-actions/:unlock_action_id", m.skillUnlockActionHandler.UpdateSkillUnlockAction, skillManage)
		adminProtected.POST("/skills/:skill_id/unlock-actions/batch", m.skillUnlockActionHandler.BatchSetSkillUnlockActions, skillManage)
		adminProtected.DELETE("/skills/:skill_id/unlock-actions/:unlock_action_id", m.skillUnlockActionHandler.RemoveSkillUnlockAction, skillManage)
		// 获取动作的可配置属性列表
		adminProtected.GET("/actions/:action_id/scalable-attributes", m.skillUnlockActionHandler.GetActionScalableAttributes, skillManage)

		// 怪物配置管理
		adminProtected.GET("/monsters", m.monsterHandler.GetMonsters, systemConfig)
		adminProtected.POST("/monsters", m.monsterHandler.CreateMonster, systemConfig)
		adminProtected.GET("/monsters/:id", m.monsterHandler.GetMonster, systemConfig)
		adminProtected.PUT("/monsters/:id", m.monsterHandler.UpdateMonster, systemConfig)
		adminProtected.DELETE("/monsters/:id", m.monsterHandler.DeleteMonster, systemConfig)

		// 怪物技能管理
		adminProtected.GET("/monsters/:id/skills", m.monsterHandler.GetMonsterSkills, systemConfig)
		adminProtected.POST("/monsters/:id/skills", m.monsterHandler.AddMonsterSkill, systemConfig)
		adminProtected.PUT("/monsters/:id/skills/:skill_id", m.monsterHandler.UpdateMonsterSkill, systemConfig)
		adminProtected.DELETE("/monsters/:id/skills/:skill_id", m.monsterHandler.RemoveMonsterSkill, systemConfig)

		// 怪物掉落管理
		adminProtected.GET("/monsters/:id/drops", m.monsterHandler.GetMonsterDrops, systemConfig)
		adminProtected.POST("/monsters/:id/drops", m.monsterHandler.AddMonsterDrop, systemConfig)
		adminProtected.PUT("/monsters/:id/drops/:drop_pool_id", m.monsterHandler.UpdateMonsterDrop, systemConfig)
		adminProtected.DELETE("/monsters/:id/drops/:drop_pool_id", m.monsterHandler.RemoveMonsterDrop, systemConfig)

		// 地城配置管理
		adminProtected.GET("/dungeons", m.dungeonHandler.GetDungeons, systemConfig)
		adminProtected.POST("/dungeons", m.dungeonHandler.CreateDungeon, systemConfig)
		adminProtected.GET("/dungeons/:id", m.dungeonHandler.GetDungeon, systemConfig)
		adminProtected.PUT("/dungeons/:id", m.dungeonHandler.UpdateDungeon, systemConfig)
		adminProtected.DELETE("/dungeons/:id", m.dungeonHandler.DeleteDungeon, systemConfig)

		// 地城房间管理
		adminProtected.GET("/dungeon-rooms", m.dungeonRoomHandler.GetRooms, systemConfig)
		adminProtected.POST("/dungeon-rooms", m.dungeonRoomHandler.CreateRoom, systemConfig)
		adminProtected.GET("/dungeon-rooms/:id", m.dungeonRoomHandler.GetRoom, systemConfig)
		adminProtected.PUT("/dungeon-rooms/:id", m.dungeonRoomHandler.UpdateRoom, systemConfig)
		adminProtected.DELETE("/dungeon-rooms/:id", m.dungeonRoomHandler.DeleteRoom, systemConfig)

		// 地城战斗配置管理
		adminProtected.POST("/dungeon-battles", m.dungeonBattleHandler.CreateBattle, systemConfig)
		adminProtected.GET("/dungeon-battles/:id", m.dungeonBattleHandler.GetBattle, systemConfig)
		adminProtected.PUT("/dungeon-battles/:id", m.dungeonBattleHandler.UpdateBattle, systemConfig)
		adminProtected.DELETE("/dungeon-battles/:id", m.dungeonBattleHandler.DeleteBattle, systemConfig)

		// 地城事件配置管理
		adminProtected.POST("/dungeon-events", m.dungeonEventHandler.CreateEvent, systemConfig)
		adminProtected.GET("/dungeon-events/:id", m.dungeonEventHandler.GetEvent, systemConfig)
		adminProtected.PUT("/dungeon-events/:id", m.dungeonEventHandler.UpdateEvent, systemConfig)
		adminProtected.DELETE("/dungeon-events/:id", m.dungeonEventHandler.DeleteEvent, systemConfig)

		// 工具：发放物品（仅测试环境启用，需 system:config）
		adminProtected.POST("/tools/grant-item", m.toolsHandler.GrantItem, systemConfig)

		// 团队管理（后台）
		// adminProtected.GET("/teams", m.teamAdminHandler.ListTeams, teamRead)                         // 查询团队列表
		// adminProtected.GET("/teams/:team_id", m.teamAdminHandler.GetTeam, teamRead)                  // 查询团队详情
		// adminProtected.POST("/teams/:team_id/disband", m.teamAdminHandler.DisbandTeam, teamModerate) // 强制解散团队
	}

	// Swagger UI
	m.httpServer.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check
	m.httpServer.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "ok",
			"module": "admin",
		})
	})

	// Prometheus metrics endpoint
	m.httpServer.GET("/metrics", metrics.EchoHandler())

	fmt.Println("[Admin Module] Routes configured successfully")
	fmt.Println("[Admin Module] Swagger UI available at http://localhost:8071/swagger/index.html")
	fmt.Println("[Admin Module] Prometheus metrics available at http://localhost:8071/metrics")
}

// startHTTPServer starts HTTP server
func (m *AdminModule) startHTTPServer(settings *conf.ModuleSettings) {
	// Read HTTP port from environment variable first
	port := os.Getenv("ADMIN_HTTP_PORT")
	if port == "" {
		// Fallback to config file
		if settings != nil && settings.Settings != nil {
			portInterface, ok := settings.Settings["http_port"]
			if ok {
				port, _ = portInterface.(string)
			}
		}
	}

	if port == "" {
		port = "8071" // Default port
	}

	fmt.Printf("[Admin Module] Starting HTTP server on port %s\n", port)

	if err := m.httpServer.Start(":" + port); err != nil {
		fmt.Printf("[Admin Module] HTTP server error: %v\n", err)
	}
}

// Run module run
func (m *AdminModule) Run(closeSig chan bool) {
	fmt.Println("[Admin Module] Started successfully")
	<-closeSig
}

// OnDestroy module destroy
func (m *AdminModule) OnDestroy() {
	// Close HTTP server
	if m.httpServer != nil {
		if err := m.httpServer.Close(); err != nil {
			fmt.Printf("[Admin Module] Failed to close HTTP server: %v\n", err)
		} else {
			fmt.Println("[Admin Module] HTTP server closed")
		}
	}

	// Close database connection
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			fmt.Printf("[Admin Module] Failed to close database: %v\n", err)
		} else {
			fmt.Println("[Admin Module] Database connection closed")
		}
	}

	m.BaseModule.OnDestroy()
	fmt.Println("[Admin Module] Destroyed")
}

// startDBPoolMonitoring 启动数据库连接池监控
// 每 30 秒报告一次连接池统计信息到 Prometheus
func (m *AdminModule) startDBPoolMonitoring(db *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := db.Stats()

		// 记录数据库连接池指标
		metrics.DefaultResourceMetrics.RecordDBPoolStats(
			metrics.GetServiceName(),
			"postgres",            // 数据库名称
			stats.OpenConnections, // 当前打开的连接数
			stats.InUse,           // 正在使用的连接数
			stats.Idle,            // 空闲连接数
			25,                    // 最大连接数（与 SetMaxOpenConns 保持一致）
			stats.WaitCount,       // 等待连接的总次数
			stats.WaitDuration,    // 等待连接的总时长
		)
	}
}

// Module creates Admin module instance
func Module() module.Module {
	return new(AdminModule)
}
