package game

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	custommiddleware "tsu-self/internal/middleware"
	authclient "tsu-self/internal/modules/auth/client"
	"tsu-self/internal/modules/game/handler"
	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/modules/game/tasks"
	authpb "tsu-self/internal/pb/auth"
	"tsu-self/internal/pkg/i18n"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/metrics"
	redisClient "tsu-self/internal/pkg/redis"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/trace"
	"tsu-self/internal/pkg/validator"

	_ "tsu-self/docs/game" // Swagger 生成的文档

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/server"
	_ "github.com/lib/pq"
	echoSwagger "github.com/swaggo/echo-swagger"
	"google.golang.org/protobuf/proto"
)

type GameModule struct {
	basemodule.BaseModule
	db                            *sql.DB
	redis                         *redisClient.Client
	ketoClient                    *authclient.KetoClient
	httpServer                    *echo.Echo
	serviceContainer              *service.ServiceContainer
	authHandler                   *handler.AuthHandler
	passwordRecoveryHandler       *handler.PasswordRecoveryHandler
	heroHandler                   *handler.HeroHandler
	heroActivationHandler         *handler.HeroActivationHandler
	heroAttributeHandler          *handler.HeroAttributeHandler
	heroSkillHandler              *handler.HeroSkillHandler
	classHandler                  *handler.ClassHandler
	skillDetailHandler            *handler.SkillDetailHandler
	upgradeCostHandler            *handler.UpgradeCostHandler
	equipmentHandler              *handler.EquipmentHandler
	equipmentSetHandler           *handler.EquipmentSetHandler
	inventoryHandler              *handler.InventoryHandler
	teamHandler                   *handler.TeamHandler
	teamMemberHandler             *handler.TeamMemberHandler
	teamWarehouseHandler          *handler.TeamWarehouseHandler
	teamDungeonHandler            *handler.TeamDungeonHandler
	teamRPCHandler                *handler.TeamRPCHandler
	battleResultHandler           *handler.BattleResultHandler
	teamPermissionMW              *custommiddleware.TeamPermissionMiddleware
	cleanupTask                   *tasks.CleanupTask
	teamLeaderTransferTask        *tasks.TeamLeaderTransferTask
	teamInvitationExpireTask      *tasks.TeamInvitationExpireTask
	teamPermissionConsistencyTask *tasks.TeamPermissionConsistencyTask
	respWriter                    response.Writer
}

// GetType returns module type
func (m *GameModule) GetType() string {
	return "game"
}

// Version returns module version
func (m *GameModule) Version() string {
	return "1.0.0"
}

// OnAppConfigurationLoaded 当App初始化时调用
func (m *GameModule) OnAppConfigurationLoaded(app module.App) {
	m.BaseModule.OnAppConfigurationLoaded(app)
}

// OnInit module initialization
func (m *GameModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	metrics.SetServiceName("game")
	// 按照 mqant 官方推荐：在每个模块的 OnInit 中配置服务注册参数
	// TTL = 30s, 心跳间隔 = 15s (TTL 必须大于心跳间隔)
	m.BaseModule.OnInit(m, app, settings,
		server.RegisterInterval(15*time.Second),
		server.RegisterTTL(30*time.Second),
	)

	// 1. Initialize database connection
	if err := m.initDatabase(settings); err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	// 2. Initialize Redis (for permission/cache workloads)
	if err := m.initRedis(settings); err != nil {
		panic(fmt.Sprintf("Failed to initialize Redis: %v", err))
	}

	// 3. Initialize Keto client (optional, for team permissions)
	m.initKetoClient()

	// 4. Initialize response writer
	m.initResponseWriter()

	// 5. Initialize HTTP server
	m.initHTTPServer()

	// 6. Initialize Services and Handlers
	m.initServicesAndHandlers()

	// 7. Setup routes
	m.setupRoutes()

	// 8. Setup RPC methods
	m.setupRPCMethods()

	// 9. Start cron tasks
	m.startCronTasks()

	// 10. Start HTTP server in background
	go m.startHTTPServer(settings)

	m.GetServer().Options()
}

// initDatabase initializes database connection
func (m *GameModule) initDatabase(settings *conf.ModuleSettings) error {
	// Read from environment variable first
	dbURL := os.Getenv("TSU_GAME_DATABASE_URL")
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
		return fmt.Errorf("TSU_GAME_DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	m.db = db
	fmt.Println("[Game Module] Database initialized successfully")

	// 启动数据库连接池监控
	go m.startDBPoolMonitoring(db)

	return nil
}

// initRedis initializes Redis client for caching/perms
func (m *GameModule) initRedis(settings *conf.ModuleSettings) error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 6379
	if portStr := os.Getenv("REDIS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbIndex := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if d, err := strconv.Atoi(dbStr); err == nil {
			dbIndex = d
		}
	}

	client, err := redisClient.NewClient(redisClient.Config{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       dbIndex,
	}, metrics.GetServiceName())
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	m.redis = client
	fmt.Printf("[Game Module] Redis connected successfully (Host: %s:%d, DB: %d)\n", host, port, dbIndex)
	return nil
}

// initKetoClient initializes Keto client for team permissions
func (m *GameModule) initKetoClient() {
	readURL := os.Getenv("KETO_READ_URL")
	writeURL := os.Getenv("KETO_WRITE_URL")

	// Keto 是可选的，如果没有配置则跳过
	if readURL == "" || writeURL == "" {
		fmt.Println("[Game Module] Keto URLs not configured, team permissions will use database fallback")
		return
	}

	// 创建 Keto 客户端
	ketoClient, err := authclient.NewKetoClient(readURL, writeURL)
	if err != nil {
		fmt.Printf("[Game Module] Failed to initialize Keto client: %v, will use database fallback\n", err)
		return
	}

	m.ketoClient = ketoClient
	fmt.Printf("[Game Module] Keto client initialized (read: %s, write: %s)\n", readURL, writeURL)

	if err := m.ensureTeamPermissionsInitializedViaAuth(); err != nil {
		fmt.Printf("[Game Module] Warning: Auth 初始化团队权限失败 (%v)，尝试直接写入 Keto\n", err)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := custommiddleware.InitializeTeamPermissions(ctx, ketoClient); err != nil {
			fmt.Printf("[Game Module] Failed to initialize team permissions via Keto fallback: %v\n", err)
		} else {
			fmt.Println("[Game Module] Team permissions initialized via Keto fallback")
		}
	}
}

func (m *GameModule) ensureTeamPermissionsInitializedViaAuth() error {
	if m.App == nil {
		return fmt.Errorf("app not initialized")
	}

	req := &authpb.InitializeTeamPermissionsRequest{}
	payload, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}

	result, errStr := m.App.Invoke(m, "auth", "InitializeTeamPermissions", payload)
	if errStr != "" {
		return fmt.Errorf("%s", errStr)
	}

	respBytes, ok := result.([]byte)
	if !ok {
		return fmt.Errorf("unexpected RPC response type")
	}

	resp := &authpb.InitializeTeamPermissionsResponse{}
	if err := proto.Unmarshal(respBytes, resp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	if resp.Initialized {
		fmt.Println("[Game Module] Team permissions verified via Auth service")
	} else {
		fmt.Println("[Game Module] Auth service reported missing team permissions entries")
	}

	if len(resp.MissingPermissions) > 0 {
		fmt.Printf("[Game Module] Missing Keto tuples: %v\n", resp.MissingPermissions)
	}

	return nil
}

// initResponseWriter initializes response writer
func (m *GameModule) initResponseWriter() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	// 使用全局 logger
	logger := log.GetLogger()
	m.respWriter = response.NewResponseHandler(logger, environment)
	fmt.Println("[Game Module] Response writer initialized")
}

// initHTTPServer initializes HTTP server
func (m *GameModule) initHTTPServer() {
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
	m.httpServer.Use(middleware.CORS())

	fmt.Println("[Game Module] HTTP middlewares configured:")
	fmt.Println("  ✓ TraceID (自动生成追踪ID)")
	fmt.Println("  ✓ Metrics (Prometheus 指标收集)")
	fmt.Println("  ✓ i18n (国际化支持)")
	fmt.Printf("  ✓ Logging (日志记录 - %s)\n", environment)
	fmt.Println("  ✓ Recovery (Panic 恢复)")
	fmt.Println("  ✓ Error (统一错误处理)")
	fmt.Println("  ✓ CORS (跨域支持)")
}

// initServicesAndHandlers initializes services and HTTP handlers
func (m *GameModule) initServicesAndHandlers() {
	// 创建服务容器（统一管理所有 Repository 和 Service）
	// 传入 ketoClient（可能为 nil，会优雅降级）
	m.serviceContainer = service.NewServiceContainer(m.db, m.ketoClient, m.redis)

	// 初始化 HTTP Handlers（从容器中获取需要的服务）
	m.authHandler = handler.NewAuthHandler(m, m.respWriter)
	m.passwordRecoveryHandler = handler.NewPasswordRecoveryHandler(m, m.respWriter)
	m.heroHandler = handler.NewHeroHandler(m.serviceContainer, m.respWriter)
	m.heroActivationHandler = handler.NewHeroActivationHandler(m.db, m.respWriter)
	m.heroAttributeHandler = handler.NewHeroAttributeHandler(m.serviceContainer, m.respWriter)
	m.heroSkillHandler = handler.NewHeroSkillHandler(m.serviceContainer, m.respWriter)
	m.classHandler = handler.NewClassHandler(m.serviceContainer, m.respWriter)
	m.skillDetailHandler = handler.NewSkillDetailHandler(m.serviceContainer.GetSkillDetailService(), m.respWriter)
	m.upgradeCostHandler = handler.NewUpgradeCostHandler(
		m.serviceContainer.GetHeroLevelRequirementRepo(),
		m.serviceContainer.GetSkillUpgradeCostRepo(),
		m.serviceContainer.GetAttributeUpgradeCostRepo(),
		m.respWriter,
	)
	m.equipmentHandler = handler.NewEquipmentHandler(m.db, m.respWriter)
	m.equipmentSetHandler = handler.NewEquipmentSetHandler(m.serviceContainer.GetEquipmentSetService(), m.respWriter)
	m.inventoryHandler = handler.NewInventoryHandler(m.db, m.respWriter)
	m.teamHandler = handler.NewTeamHandler(m.serviceContainer, m.respWriter)
	m.teamMemberHandler = handler.NewTeamMemberHandler(m.serviceContainer, m.respWriter)
	m.teamWarehouseHandler = handler.NewTeamWarehouseHandler(m.serviceContainer, m.respWriter)
	m.teamDungeonHandler = handler.NewTeamDungeonHandler(m.serviceContainer, m.respWriter)
	m.teamRPCHandler = handler.NewTeamRPCHandler(m.serviceContainer, m.db)
	m.battleResultHandler = handler.NewBattleResultHandler(m.serviceContainer, m.respWriter)

	// 初始化团队权限中间件（基于权限服务，可回退到数据库）
	permissionService := m.serviceContainer.GetTeamPermissionService()
	if permissionService != nil {
		m.teamPermissionMW = custommiddleware.NewTeamPermissionMiddleware(permissionService, m.respWriter)
		fmt.Println("[Game Module] Team permission middleware initialized (cache-enabled)")
	} else {
		fmt.Println("[Game Module] Team permission middleware skipped (service unavailable)")
	}

	fmt.Println("[Game Module] Handlers initialized successfully")
}

// startCronTasks starts cron scheduled tasks
func (m *GameModule) startCronTasks() {
	logger := log.GetLogger()

	// 1. 创建并启动定时清理任务
	m.cleanupTask = tasks.NewCleanupTask(m.db, logger)
	m.cleanupTask.Start()

	// 2. 创建并启动团队相关定时任务
	// 队长自动转移任务
	teamService := m.serviceContainer.GetTeamService()
	m.teamLeaderTransferTask = tasks.NewTeamLeaderTransferTask(m.db, teamService, logger)
	m.teamLeaderTransferTask.Start()

	// 邀请过期任务
	m.teamInvitationExpireTask = tasks.NewTeamInvitationExpireTask(m.db, logger)
	m.teamInvitationExpireTask.Start()

	// 权限一致性检查任务（仅在 Keto 可用时启动）
	if m.serviceContainer.GetTeamPermissionService() != nil {
		m.teamPermissionConsistencyTask = tasks.NewTeamPermissionConsistencyTask(
			m.db,
			m.serviceContainer.GetTeamPermissionService(),
			logger,
		)
		m.teamPermissionConsistencyTask.Start()
		fmt.Println("[Game Module] Team permission consistency task started")
	} else {
		fmt.Println("[Game Module] Team permission consistency task skipped (Keto not available)")
	}

	fmt.Println("[Game Module] Cron tasks started successfully:")
	fmt.Println("  ✓ Cleanup Task (每天凌晨2点)")
	fmt.Println("  ✓ Team Leader Transfer Task (每小时)")
	fmt.Println("  ✓ Team Invitation Expire Task (每小时)")
}

// setupRoutes sets up HTTP routes
func (m *GameModule) setupRoutes() {
	// API v1 group
	v1 := m.httpServer.Group("/api/v1")

	// Game routes (添加 /game 前缀以区分 admin 和 game)
	game := v1.Group("/game")
	{
		// Auth routes (公开访问，不需要认证)
		auth := game.Group("/auth")
		{
			// 用户注册和登录
			auth.POST("/register", m.authHandler.Register)
			auth.POST("/login", m.authHandler.Login)
			auth.POST("/logout", m.authHandler.Logout)
			auth.GET("/users/:user_id", m.authHandler.GetUser)

			// 密码重置 (公开访问)
			auth.POST("/recovery/initiate", m.passwordRecoveryHandler.InitiateRecovery)
			auth.POST("/password/reset-with-code", m.passwordRecoveryHandler.ResetPasswordWithCode)
		}

		// Hero routes (需要认证)
		logger := log.GetLogger()
		heroes := game.Group("/heroes")
		// 应用认证中间件
		heroes.Use(custommiddleware.AuthMiddleware(m.respWriter, logger, m.db))
		{
			// 英雄管理
			heroes.POST("", m.heroHandler.CreateHero)                                 // 创建英雄
			heroes.GET("", m.heroHandler.GetUserHeroes)                               // 获取用户英雄列表
			heroes.GET("/:hero_id", m.heroHandler.GetHero)                            // 获取英雄详情
			heroes.GET("/:hero_id/full", m.heroHandler.GetHeroFull)                   // 获取英雄完整信息（含职业、属性、技能）
			heroes.GET("/:hero_id/advancement-check", m.heroHandler.CheckAdvancement) // 检查职业进阶条件
			heroes.POST("/:hero_id/experience", m.heroHandler.AddExperience)          // 增加经验（测试用）
			heroes.POST("/:hero_id/advance", m.heroHandler.AdvanceClass)              // 职业进阶
			heroes.POST("/:hero_id/transfer", m.heroHandler.TransferClass)            // 职业转职

			// 属性管理
			heroes.GET("/:hero_id/attributes", m.heroAttributeHandler.GetComputedAttributes)       // 获取属性
			heroes.POST("/:hero_id/attributes/allocate", m.heroAttributeHandler.AllocateAttribute) // 属性加点
			heroes.POST("/:hero_id/attributes/rollback", m.heroAttributeHandler.RollbackAttribute) // 回退属性

			// 技能管理
			heroes.GET("/:hero_id/skills/available", m.heroSkillHandler.GetAvailableSkills)      // 获取可学习技能
			heroes.GET("/:hero_id/skills/learned", m.heroSkillHandler.GetLearnedSkills)          // 获取已学习技能列表
			heroes.POST("/:hero_id/skills/learn", m.heroSkillHandler.LearnSkill)                 // 学习技能
			heroes.POST("/:hero_id/skills/:skill_id/upgrade", m.heroSkillHandler.UpgradeSkill)   // 升级技能
			heroes.POST("/:hero_id/skills/:skill_id/rollback", m.heroSkillHandler.RollbackSkill) // 回退技能

			// 英雄激活管理
			heroes.PATCH("/:hero_id/activate", m.heroActivationHandler.ActivateHero)         // 激活英雄
			heroes.PATCH("/:hero_id/deactivate", m.heroActivationHandler.DeactivateHero)     // 停用英雄
			heroes.PATCH("/switch", m.heroActivationHandler.SwitchCurrentHero)               // 切换当前英雄
			heroes.GET("/activated", m.heroActivationHandler.GetActivatedHeroes)             // 获取已激活英雄列表
		}

		// Class routes (公开访问)
		classes := game.Group("/classes")
		{
			classes.GET("/basic", m.classHandler.GetBasicClasses)                               // 获取基础职业列表（创建角色用）
			classes.GET("", m.classHandler.GetClasses)                                          // 获取职业列表
			classes.GET("/:class_id", m.classHandler.GetClass)                                  // 获取职业详情
			classes.GET("/:class_id/advancement-options", m.classHandler.GetAdvancementOptions) // 获取可进阶选项
		}

		// Skill detail routes (需要认证)
		skills := game.Group("/skills")
		skills.Use(custommiddleware.AuthMiddleware(m.respWriter, logger, m.db))
		{
			skills.GET("/basic", m.skillDetailHandler.ListSkillsBasic)               // 获取技能列表（简化版）
			skills.GET("/standard", m.skillDetailHandler.ListSkillsStandard)         // 获取技能列表（标准版）
			skills.GET("/:skill_id/basic", m.skillDetailHandler.GetSkillBasic)       // 获取技能基本信息
			skills.GET("/:skill_id/standard", m.skillDetailHandler.GetSkillStandard) // 获取技能标准信息（含动作）
			skills.GET("/:skill_id/full", m.skillDetailHandler.GetSkillFull)         // 获取技能完整信息（深度关联）
		}

		// Upgrade cost routes (公开访问 - 配置数据)
		costs := game.Group("")
		{
			// 英雄等级需求
			costs.GET("/hero-level-requirements", m.upgradeCostHandler.GetHeroLevelRequirements)       // 获取所有等级需求
			costs.GET("/hero-level-requirements/:level", m.upgradeCostHandler.GetHeroLevelRequirement) // 获取指定等级需求

			// 技能升级消耗
			costs.GET("/skill-upgrade-costs", m.upgradeCostHandler.GetSkillUpgradeCosts)       // 获取所有技能升级消耗
			costs.GET("/skill-upgrade-costs/:level", m.upgradeCostHandler.GetSkillUpgradeCost) // 获取指定等级升级消耗

			// 属性升级消耗
			costs.GET("/attribute-upgrade-costs", m.upgradeCostHandler.GetAttributeUpgradeCosts)              // 获取所有属性升级消耗
			costs.GET("/attribute-upgrade-costs/:point_number", m.upgradeCostHandler.GetAttributeUpgradeCost) // 获取指定点数升级消耗
		}

		// // Equipment routes (需要认证)
		// equipment := game.Group("/equipment")
		// equipment.Use(custommiddleware.AuthMiddleware(m.respWriter, logger))
		// {
		// 	equipment.POST("/equip", m.equipmentHandler.EquipItem)                      // 穿戴装备
		// 	equipment.POST("/unequip", m.equipmentHandler.UnequipItem)                  // 卸下装备
		// 	equipment.GET("/equipped/:hero_id", m.equipmentHandler.GetEquippedItems)    // 查询已装备物品
		// 	equipment.GET("/slots/:hero_id", m.equipmentHandler.GetEquipmentSlots)      // 查询装备槽位
		// 	equipment.GET("/bonus/:hero_id", m.equipmentHandler.GetEquipmentBonus)      // 查询装备属性加成

		// 	// Equipment Set routes (套装相关)
		// 	equipment.GET("/sets", m.equipmentSetHandler.ListSets)                      // 查询可用套装列表
		// 	equipment.GET("/sets/:set_id", m.equipmentSetHandler.GetSetInfo)            // 查询套装详细信息
		// 	equipment.GET("/sets/active/:hero_id", m.equipmentSetHandler.GetActiveSets) // 查询英雄激活的套装
		// }

		// // Inventory routes (需要认证)
		// inventory := game.Group("/inventory")
		// inventory.Use(custommiddleware.AuthMiddleware(m.respWriter, logger))
		// {
		// 	inventory.GET("", m.inventoryHandler.GetInventory)       // 查询背包/仓库
		// 	inventory.POST("/move", m.inventoryHandler.MoveItem)     // 移动物品
		// 	inventory.POST("/discard", m.inventoryHandler.DiscardItem) // 丢弃物品
		// 	inventory.POST("/sort", m.inventoryHandler.SortInventory)  // 整理背包
		// }

		//Team routes (需要认证 + 英雄上下文)
		teams := game.Group("/teams")
		teams.Use(custommiddleware.AuthMiddleware(m.respWriter, logger, m.db))
		teams.Use(custommiddleware.HeroMiddleware(m.db, m.respWriter, logger)) // 自动从数据库获取当前英雄ID
		{
			// 团队管理
			teams.POST("", m.teamHandler.CreateTeam) // 创建团队（任何认证用户都可以）

			// 获取团队详情（需要是团队成员）
			if m.teamPermissionMW != nil {
				teams.GET("/:team_id", m.teamHandler.GetTeam, m.teamPermissionMW.RequireTeamMember)
			} else {
				teams.GET("/:team_id", m.teamHandler.GetTeam)
			}

			// 更新团队信息（只有队长可以）
			if m.teamPermissionMW != nil {
				teams.PUT("/:team_id", m.teamHandler.UpdateTeamInfo, m.teamPermissionMW.RequireTeamLeader)
			} else {
				teams.PUT("/:team_id", m.teamHandler.UpdateTeamInfo)
			}

			// 解散团队（只有队长可以）
			if m.teamPermissionMW != nil {
				teams.POST("/:team_id/disband", m.teamHandler.DisbandTeam, m.teamPermissionMW.RequireTeamLeader)
			} else {
				teams.POST("/:team_id/disband", m.teamHandler.DisbandTeam)
			}

			// 离开团队（需要是团队成员）
			if m.teamPermissionMW != nil {
				teams.POST("/:team_id/leave", m.teamHandler.LeaveTeam, m.teamPermissionMW.RequireTeamMember)
			} else {
				teams.POST("/:team_id/leave", m.teamHandler.LeaveTeam)
			}

			// 成员管理
			teams.POST("/join/apply", m.teamMemberHandler.ApplyToJoin) // 申请加入团队（任何认证用户都可以）

			// 审批加入申请（管理员或队长）
			if m.teamPermissionMW != nil {
				teams.POST("/join/approve", m.teamMemberHandler.ApproveJoinRequest, m.teamPermissionMW.RequireTeamAdmin)
			} else {
				teams.POST("/join/approve", m.teamMemberHandler.ApproveJoinRequest)
			}

			// 邀请成员（任何团队成员都可以）
			if m.teamPermissionMW != nil {
				teams.POST("/invite", m.teamMemberHandler.InviteMember, m.teamPermissionMW.RequireTeamMember)
			} else {
				teams.POST("/invite", m.teamMemberHandler.InviteMember)
			}

			// 审批邀请（管理员或队长）
			if m.teamPermissionMW != nil {
				teams.POST("/invite/approve", m.teamMemberHandler.ApproveInvitation, m.teamPermissionMW.RequireTeamAdmin)
			} else {
				teams.POST("/invite/approve", m.teamMemberHandler.ApproveInvitation)
			}

			teams.POST("/invite/accept", m.teamMemberHandler.AcceptInvitation) // 接受邀请（被邀请人）
			teams.POST("/invite/reject", m.teamMemberHandler.RejectInvitation) // 拒绝邀请（被邀请人）

			// 踢出成员（管理员或队长）
			if m.teamPermissionMW != nil {
				teams.POST("/members/kick", m.teamMemberHandler.KickMember, m.teamPermissionMW.RequireTeamAdmin)
			} else {
				teams.POST("/members/kick", m.teamMemberHandler.KickMember)
			}

			// 任命管理员（只有队长可以）
			if m.teamPermissionMW != nil {
				teams.POST("/members/promote", m.teamMemberHandler.PromoteToAdmin, m.teamPermissionMW.RequireTeamLeader)
			} else {
				teams.POST("/members/promote", m.teamMemberHandler.PromoteToAdmin)
			}

			// 撤销管理员（只有队长可以）
			if m.teamPermissionMW != nil {
				teams.POST("/members/demote", m.teamMemberHandler.DemoteAdmin, m.teamPermissionMW.RequireTeamLeader)
			} else {
				teams.POST("/members/demote", m.teamMemberHandler.DemoteAdmin)
			}

			// 仓库管理

			// 查看仓库（需要是团队成员）
			if m.teamPermissionMW != nil {
				teams.GET("/:team_id/warehouse", m.teamWarehouseHandler.GetWarehouse, m.teamPermissionMW.RequireTeamMember)
			} else {
				teams.GET("/:team_id/warehouse", m.teamWarehouseHandler.GetWarehouse)
			}

			// 分配金币（管理员或队长）
			if m.teamPermissionMW != nil {
				teams.POST("/:team_id/warehouse/distribute-gold", m.teamWarehouseHandler.DistributeGold, m.teamPermissionMW.RequireTeamAdmin)
			} else {
				teams.POST("/:team_id/warehouse/distribute-gold", m.teamWarehouseHandler.DistributeGold)
			}

			// 查看仓库物品（需要是团队成员）
			if m.teamPermissionMW != nil {
				teams.GET("/:team_id/warehouse/items", m.teamWarehouseHandler.GetWarehouseItems, m.teamPermissionMW.RequireTeamMember)
			} else {
				teams.GET("/:team_id/warehouse/items", m.teamWarehouseHandler.GetWarehouseItems)
			}

			// 地城路由
			if m.teamDungeonHandler != nil {
				if m.teamPermissionMW != nil {
					teams.POST("/:team_id/dungeons/select", m.teamDungeonHandler.SelectDungeon, m.teamPermissionMW.RequireTeamAdmin)
					teams.POST("/:team_id/dungeons/enter", m.teamDungeonHandler.EnterDungeon, m.teamPermissionMW.RequireTeamAdmin)
					teams.GET("/:team_id/dungeons/progress", m.teamDungeonHandler.GetDungeonProgress, m.teamPermissionMW.RequireTeamMember)
					teams.POST("/:team_id/dungeons/complete", m.teamDungeonHandler.CompleteDungeon, m.teamPermissionMW.RequireTeamAdmin)
					teams.POST("/:team_id/dungeons/fail", m.teamDungeonHandler.FailDungeon, m.teamPermissionMW.RequireTeamAdmin)
					teams.POST("/:team_id/dungeons/abandon", m.teamDungeonHandler.AbandonDungeon, m.teamPermissionMW.RequireTeamAdmin)
					teams.GET("/:team_id/dungeons/history", m.teamDungeonHandler.GetDungeonHistory, m.teamPermissionMW.RequireTeamMember)
				} else {
					teams.POST("/:team_id/dungeons/select", m.teamDungeonHandler.SelectDungeon)
					teams.POST("/:team_id/dungeons/enter", m.teamDungeonHandler.EnterDungeon)
					teams.GET("/:team_id/dungeons/progress", m.teamDungeonHandler.GetDungeonProgress)
					teams.POST("/:team_id/dungeons/complete", m.teamDungeonHandler.CompleteDungeon)
					teams.POST("/:team_id/dungeons/fail", m.teamDungeonHandler.FailDungeon)
					teams.POST("/:team_id/dungeons/abandon", m.teamDungeonHandler.AbandonDungeon)
					teams.GET("/:team_id/dungeons/history", m.teamDungeonHandler.GetDungeonHistory)
				}
			}
		}
	}

	internalGroup := v1.Group("/internal")
	{
		battles := internalGroup.Group("/battles")
		battles.POST("/result", m.battleResultHandler.ReportResult)
	}

	// Swagger UI
	m.httpServer.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check
	m.httpServer.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "ok",
			"module": "game",
		})
	})

	// Prometheus metrics endpoint
	m.httpServer.GET("/metrics", metrics.EchoHandler())

	fmt.Println("[Game Module] Routes configured successfully")
	fmt.Println("[Game Module] Game API routes: /api/v1/game/auth/*")
	fmt.Println("[Game Module] Swagger UI available at http://localhost:8072/swagger/index.html")
	fmt.Println("[Game Module] Prometheus metrics available at http://localhost:8072/metrics")
}

// startHTTPServer starts HTTP server
func (m *GameModule) startHTTPServer(settings *conf.ModuleSettings) {
	// Read HTTP port from environment variable first
	port := os.Getenv("GAME_HTTP_PORT")
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
		port = "8072" // Default port
	}

	fmt.Printf("[Game Module] Starting HTTP server on port %s\n", port)

	if err := m.httpServer.Start(":" + port); err != nil {
		fmt.Printf("[Game Module] HTTP server error: %v\n", err)
	}
}

// Run module run
func (m *GameModule) Run(closeSig chan bool) {
	fmt.Println("[Game Module] Started successfully")
	<-closeSig
}

// OnDestroy module destroy
func (m *GameModule) OnDestroy() {
	// Stop cron tasks
	if m.cleanupTask != nil {
		m.cleanupTask.Stop()
		fmt.Println("[Game Module] Cron tasks stopped")
	}

	// Close HTTP server
	if m.httpServer != nil {
		if err := m.httpServer.Close(); err != nil {
			fmt.Printf("[Game Module] Failed to close HTTP server: %v\n", err)
		} else {
			fmt.Println("[Game Module] HTTP server closed")
		}
	}

	// Close database connection
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			fmt.Printf("[Game Module] Failed to close database: %v\n", err)
		} else {
			fmt.Println("[Game Module] Database connection closed")
		}
	}

	m.BaseModule.OnDestroy()
	fmt.Println("[Game Module] Destroyed")
}

// Module creates Game module instance
func Module() module.Module {
	return new(GameModule)
}

// startDBPoolMonitoring 启动数据库连接池监控
// 每 30 秒报告一次连接池统计信息到 Prometheus
func (m *GameModule) startDBPoolMonitoring(db *sql.DB) {
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

// setupRPCMethods 注册 RPC 方法
// 供其他模块（如 Admin Server）调用
func (m *GameModule) setupRPCMethods() {
	// ==================== 团队管理 RPC ====================
	m.GetServer().RegisterGO("GetTeamList", m.teamRPCHandler.GetTeamList)
	m.GetServer().RegisterGO("GetTeamDetail", m.teamRPCHandler.GetTeamDetail)
	m.GetServer().RegisterGO("ForceDisbandTeam", m.teamRPCHandler.ForceDisbandTeam)
	m.GetServer().RegisterGO("GetTeamMembers", m.teamRPCHandler.GetTeamMembers)

	fmt.Println("[Game Module] RPC methods registered:")
	fmt.Println("  ✓ GetTeamList - 获取团队列表")
	fmt.Println("  ✓ GetTeamDetail - 获取团队详情")
	fmt.Println("  ✓ ForceDisbandTeam - 强制解散团队")
	fmt.Println("  ✓ GetTeamMembers - 获取团队成员列表")
}
