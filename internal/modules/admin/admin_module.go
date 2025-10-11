package admin

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	custommiddleware "tsu-self/internal/middleware"
	"tsu-self/internal/modules/admin/handler"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/validator"

	_ "tsu-self/docs" // Swagger 生成的文档

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// Middleware
	m.httpServer.Use(middleware.Logger())
	m.httpServer.Use(middleware.Recover())
	m.httpServer.Use(middleware.CORS())
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
}

// setupRoutes sets up HTTP routes
func (m *AdminModule) setupRoutes() {
	// 获取全局 logger
	logger := log.GetLogger()

	// API v1 group
	v1 := m.httpServer.Group("/api/v1")

	// Auth routes (公开访问，不需要认证)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", m.authHandler.Register)
		auth.POST("/login", m.authHandler.Login)
		auth.POST("/logout", m.authHandler.Logout)
		auth.GET("/users/:user_id", m.authHandler.GetUser)

		// 密码重置 (公开访问)
		auth.POST("/recovery/initiate", m.passwordRecoveryHandler.InitiateRecovery)
		auth.POST("/recovery/verify", m.passwordRecoveryHandler.VerifyRecoveryCode)
		auth.POST("/password/reset", m.passwordRecoveryHandler.ResetPassword)
	}

	// Admin routes (需要认证，应用认证中间件)
	// 这些路由的请求必须经过 Oathkeeper 验证，并且会从 Header 中提取用户信息
	admin := v1.Group("/admin")
	admin.Use(custommiddleware.AuthMiddleware(m.respWriter, logger))
	admin.Use(custommiddleware.UUIDValidationMiddleware(m.respWriter))
	{
		// 用户管理
		admin.GET("/users/me", m.userHandler.GetCurrentUserProfile) // 🆕 示例：获取当前登录用户信息
		admin.GET("/users", m.userHandler.GetUsers)
		admin.GET("/users/:id", m.userHandler.GetUser)
		admin.PUT("/users/:id", m.userHandler.UpdateUser)
		admin.POST("/users/:id/ban", m.userHandler.BanUser)
		admin.POST("/users/:id/unban", m.userHandler.UnbanUser)
		admin.DELETE("/users/:user_id", m.authHandler.DeleteUser) // 删除用户

		// 用户密码重置 (管理员功能)
		admin.POST("/users/recovery-code", m.passwordRecoveryHandler.AdminCreateRecoveryCode)

		// 角色管理
		admin.GET("/roles", m.permissionHandler.GetRoles)
		admin.POST("/roles", m.permissionHandler.CreateRole)
		admin.PUT("/roles/:id", m.permissionHandler.UpdateRole)
		admin.DELETE("/roles/:id", m.permissionHandler.DeleteRole)

		// 角色-权限管理
		admin.GET("/roles/:id/permissions", m.permissionHandler.GetRolePermissions)
		admin.POST("/roles/:id/permissions", m.permissionHandler.AssignPermissionsToRole)

		// 权限管理
		admin.GET("/permissions", m.permissionHandler.GetPermissions)
		admin.GET("/permission-groups", m.permissionHandler.GetPermissionGroups)

		// 用户-角色管理
		admin.GET("/users/:user_id/roles", m.permissionHandler.GetUserRoles)
		admin.POST("/users/:user_id/roles", m.permissionHandler.AssignRolesToUser)
		admin.DELETE("/users/:user_id/roles", m.permissionHandler.RevokeRolesFromUser)

		// 用户-权限管理
		admin.GET("/users/:user_id/permissions", m.permissionHandler.GetUserPermissions)
		admin.POST("/users/:user_id/permissions", m.permissionHandler.GrantPermissionsToUser)
		admin.DELETE("/users/:user_id/permissions", m.permissionHandler.RevokePermissionsFromUser)

		// 职业管理
		admin.GET("/classes", m.classHandler.GetClasses)
		admin.POST("/classes", m.classHandler.CreateClass)
		admin.GET("/classes/:id", m.classHandler.GetClass)
		admin.PUT("/classes/:id", m.classHandler.UpdateClass)
		admin.DELETE("/classes/:id", m.classHandler.DeleteClass)

		// 职业属性加成管理
		admin.GET("/classes/:id/attribute-bonuses", m.classHandler.GetClassAttributeBonuses)
		admin.POST("/classes/:id/attribute-bonuses", m.classHandler.CreateAttributeBonus)
		admin.POST("/classes/:id/attribute-bonuses/batch", m.classHandler.BatchSetAttributeBonuses)
		admin.PUT("/classes/:id/attribute-bonuses/:bonus_id", m.classHandler.UpdateAttributeBonus)
		admin.DELETE("/classes/:id/attribute-bonuses/:bonus_id", m.classHandler.DeleteAttributeBonus)

		// 职业进阶路径查询（嵌套在职业下）
		admin.GET("/classes/:id/advancement", m.classHandler.GetClassAdvancement)
		admin.GET("/classes/:id/advancement-paths", m.classHandler.GetClassAdvancementPaths)
		admin.GET("/classes/:id/advancement-sources", m.classHandler.GetClassAdvancementSources)

		// 职业进阶要求管理（独立接口）
		admin.GET("/advancement-requirements", m.advancedRequirementHandler.GetAdvancedRequirements)
		admin.POST("/advancement-requirements", m.advancedRequirementHandler.CreateAdvancedRequirement)
		admin.POST("/advancement-requirements/batch", m.advancedRequirementHandler.BatchCreateAdvancedRequirements)
		admin.GET("/advancement-requirements/:id", m.advancedRequirementHandler.GetAdvancedRequirement)
		admin.PUT("/advancement-requirements/:id", m.advancedRequirementHandler.UpdateAdvancedRequirement)
		admin.DELETE("/advancement-requirements/:id", m.advancedRequirementHandler.DeleteAdvancedRequirement)

		// 职业技能池管理（嵌套在职业下）
		admin.GET("/classes/:class_id/skill-pools", m.classSkillPoolHandler.GetClassSkillPoolsByClassID)

		// 职业技能池管理（独立接口）
		admin.GET("/class-skill-pools", m.classSkillPoolHandler.GetClassSkillPools)
		admin.POST("/class-skill-pools", m.classSkillPoolHandler.CreateClassSkillPool)
		admin.GET("/class-skill-pools/:id", m.classSkillPoolHandler.GetClassSkillPool)
		admin.PUT("/class-skill-pools/:id", m.classSkillPoolHandler.UpdateClassSkillPool)
		admin.DELETE("/class-skill-pools/:id", m.classSkillPoolHandler.DeleteClassSkillPool)

		// 技能类别管理
		admin.GET("/skill-categories", m.skillCategoryHandler.GetSkillCategories)
		admin.POST("/skill-categories", m.skillCategoryHandler.CreateSkillCategory)
		admin.GET("/skill-categories/:id", m.skillCategoryHandler.GetSkillCategory)
		admin.PUT("/skill-categories/:id", m.skillCategoryHandler.UpdateSkillCategory)
		admin.DELETE("/skill-categories/:id", m.skillCategoryHandler.DeleteSkillCategory)

		// 动作类别管理
		admin.GET("/action-categories", m.actionCategoryHandler.GetActionCategories)
		admin.POST("/action-categories", m.actionCategoryHandler.CreateActionCategory)
		admin.GET("/action-categories/:id", m.actionCategoryHandler.GetActionCategory)
		admin.PUT("/action-categories/:id", m.actionCategoryHandler.UpdateActionCategory)
		admin.DELETE("/action-categories/:id", m.actionCategoryHandler.DeleteActionCategory)

		// 伤害类型管理
		admin.GET("/damage-types", m.damageTypeHandler.GetDamageTypes)
		admin.POST("/damage-types", m.damageTypeHandler.CreateDamageType)
		admin.GET("/damage-types/:id", m.damageTypeHandler.GetDamageType)
		admin.PUT("/damage-types/:id", m.damageTypeHandler.UpdateDamageType)
		admin.DELETE("/damage-types/:id", m.damageTypeHandler.DeleteDamageType)

		// 属性类型管理
		admin.GET("/hero-attribute-types", m.heroAttributeTypeHandler.GetHeroAttributeTypes)
		admin.POST("/hero-attribute-types", m.heroAttributeTypeHandler.CreateHeroAttributeType)
		admin.GET("/hero-attribute-types/:id", m.heroAttributeTypeHandler.GetHeroAttributeType)
		admin.PUT("/hero-attribute-types/:id", m.heroAttributeTypeHandler.UpdateHeroAttributeType)
		admin.DELETE("/hero-attribute-types/:id", m.heroAttributeTypeHandler.DeleteHeroAttributeType)

		// 标签管理
		admin.GET("/tags", m.tagHandler.GetTags)
		admin.POST("/tags", m.tagHandler.CreateTag)
		admin.GET("/tags/:id", m.tagHandler.GetTag)
		admin.PUT("/tags/:id", m.tagHandler.UpdateTag)
		admin.DELETE("/tags/:id", m.tagHandler.DeleteTag)

		// 标签关联管理
		admin.GET("/tags/:tag_id/entities", m.tagRelationHandler.GetTagEntities)
		admin.GET("/entities/:entity_type/:entity_id/tags", m.tagRelationHandler.GetEntityTags)
		admin.POST("/entities/:entity_type/:entity_id/tags", m.tagRelationHandler.AddTagToEntity)
		admin.POST("/entities/:entity_type/:entity_id/tags/batch", m.tagRelationHandler.BatchSetEntityTags)
		admin.DELETE("/entities/:entity_type/:entity_id/tags/:tag_id", m.tagRelationHandler.RemoveTagFromEntity)

		// 元数据管理 (需要认证)
		metadata := admin.Group("/metadata")
		{
			// 元效果类型定义
			metadata.GET("/effect-type-definitions", m.effectTypeDefinitionHandler.GetEffectTypeDefinitions)
			metadata.GET("/effect-type-definitions/all", m.effectTypeDefinitionHandler.GetAllEffectTypeDefinitions)
			metadata.GET("/effect-type-definitions/:id", m.effectTypeDefinitionHandler.GetEffectTypeDefinition)

			// 公式变量
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
		admin.GET("/skills", m.skillHandler.GetSkills)
		admin.POST("/skills", m.skillHandler.CreateSkill)
		admin.GET("/skills/:id", m.skillHandler.GetSkill)
		admin.PUT("/skills/:id", m.skillHandler.UpdateSkill)
		admin.DELETE("/skills/:id", m.skillHandler.DeleteSkill)

		// 全局技能升级消耗管理
		admin.GET("/skill-upgrade-costs", m.skillUpgradeCostHandler.GetSkillUpgradeCosts)
		admin.POST("/skill-upgrade-costs", m.skillUpgradeCostHandler.CreateSkillUpgradeCost)
		admin.GET("/skill-upgrade-costs/level/:level", m.skillUpgradeCostHandler.GetSkillUpgradeCostByLevel)
		admin.GET("/skill-upgrade-costs/:id", m.skillUpgradeCostHandler.GetSkillUpgradeCost)
		admin.PUT("/skill-upgrade-costs/:id", m.skillUpgradeCostHandler.UpdateSkillUpgradeCost)
		admin.DELETE("/skill-upgrade-costs/:id", m.skillUpgradeCostHandler.DeleteSkillUpgradeCost)

		// 效果管理
		admin.GET("/effects", m.effectHandler.GetEffects)
		admin.POST("/effects", m.effectHandler.CreateEffect)
		admin.GET("/effects/:id", m.effectHandler.GetEffect)
		admin.PUT("/effects/:id", m.effectHandler.UpdateEffect)
		admin.DELETE("/effects/:id", m.effectHandler.DeleteEffect)

		// Buff管理
		admin.GET("/buffs", m.buffHandler.GetBuffs)
		admin.POST("/buffs", m.buffHandler.CreateBuff)
		admin.GET("/buffs/:id", m.buffHandler.GetBuff)
		admin.PUT("/buffs/:id", m.buffHandler.UpdateBuff)
		admin.DELETE("/buffs/:id", m.buffHandler.DeleteBuff)

		// Buff效果关联管理
		admin.GET("/buffs/:buff_id/effects", m.buffEffectHandler.GetBuffEffects)
		admin.POST("/buffs/:buff_id/effects", m.buffEffectHandler.AddBuffEffect)
		admin.POST("/buffs/:buff_id/effects/batch", m.buffEffectHandler.BatchSetBuffEffects)
		admin.DELETE("/buffs/:buff_id/effects/:effect_id", m.buffEffectHandler.RemoveBuffEffect)

		// 动作Flag管理
		admin.GET("/action-flags", m.actionFlagHandler.GetActionFlags)
		admin.POST("/action-flags", m.actionFlagHandler.CreateActionFlag)
		admin.GET("/action-flags/:id", m.actionFlagHandler.GetActionFlag)
		admin.PUT("/action-flags/:id", m.actionFlagHandler.UpdateActionFlag)
		admin.DELETE("/action-flags/:id", m.actionFlagHandler.DeleteActionFlag)

		// 动作管理
		admin.GET("/actions", m.actionHandler.GetActions)
		admin.POST("/actions", m.actionHandler.CreateAction)
		admin.GET("/actions/:id", m.actionHandler.GetAction)
		admin.PUT("/actions/:id", m.actionHandler.UpdateAction)
		admin.DELETE("/actions/:id", m.actionHandler.DeleteAction)

		// 动作效果关联管理
		admin.GET("/actions/:action_id/effects", m.actionEffectHandler.GetActionEffects)
		admin.POST("/actions/:action_id/effects", m.actionEffectHandler.AddActionEffect)
		admin.POST("/actions/:action_id/effects/batch", m.actionEffectHandler.BatchSetActionEffects)
		admin.DELETE("/actions/:action_id/effects/:effect_id", m.actionEffectHandler.RemoveActionEffect)

		// 技能解锁动作管理
		admin.GET("/skills/:skill_id/unlock-actions", m.skillUnlockActionHandler.GetSkillUnlockActions)
		admin.POST("/skills/:skill_id/unlock-actions", m.skillUnlockActionHandler.AddSkillUnlockAction)
		admin.POST("/skills/:skill_id/unlock-actions/batch", m.skillUnlockActionHandler.BatchSetSkillUnlockActions)
		admin.DELETE("/skills/:skill_id/unlock-actions/:action_id", m.skillUnlockActionHandler.RemoveSkillUnlockAction)
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

	fmt.Println("[Admin Module] Routes configured successfully")
	fmt.Println("[Admin Module] Swagger UI available at http://localhost:8071/swagger/index.html")
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

// Module creates Admin module instance
func Module() module.Module {
	return new(AdminModule)
}
