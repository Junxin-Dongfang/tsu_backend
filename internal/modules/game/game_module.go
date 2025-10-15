package game

import (
	"fmt"
	"os"
	"time"

	custommiddleware "tsu-self/internal/middleware"
	"tsu-self/internal/modules/game/handler"
	"tsu-self/internal/pkg/i18n"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/metrics"
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
	echoSwagger "github.com/swaggo/echo-swagger"
)

type GameModule struct {
	basemodule.BaseModule
	httpServer              *echo.Echo
	authHandler             *handler.AuthHandler
	passwordRecoveryHandler *handler.PasswordRecoveryHandler
	respWriter              response.Writer
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
	// 按照 mqant 官方推荐：在每个模块的 OnInit 中配置服务注册参数
	// TTL = 30s, 心跳间隔 = 15s (TTL 必须大于心跳间隔)
	m.BaseModule.OnInit(m, app, settings,
		server.RegisterInterval(15*time.Second),
		server.RegisterTTL(30*time.Second),
	)

	// 1. Initialize response writer
	m.initResponseWriter()

	// 2. Initialize HTTP server
	m.initHTTPServer()

	// 3. Initialize handlers
	m.initHandlers()

	// 4. Setup routes
	m.setupRoutes()

	// 5. Start HTTP server in background
	go m.startHTTPServer(settings)

	m.GetServer().Options()
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

// initHandlers initializes HTTP handlers
func (m *GameModule) initHandlers() {
	m.authHandler = handler.NewAuthHandler(m, m.respWriter)
	m.passwordRecoveryHandler = handler.NewPasswordRecoveryHandler(m, m.respWriter)
	fmt.Println("[Game Module] Handlers initialized")
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

		// 游戏业务路由 (后续添加)
		// TODO: 添加英雄、战斗、背包等游戏业务路由
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
	// Close HTTP server
	if m.httpServer != nil {
		if err := m.httpServer.Close(); err != nil {
			fmt.Printf("[Game Module] Failed to close HTTP server: %v\n", err)
		} else {
			fmt.Println("[Game Module] HTTP server closed")
		}
	}

	m.BaseModule.OnDestroy()
	fmt.Println("[Game Module] Destroyed")
}

// Module creates Game module instance
func Module() module.Module {
	return new(GameModule)
}
