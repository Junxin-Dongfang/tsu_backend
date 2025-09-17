// File: cmd/admin/main.go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
	"tsu-self/cmd/admin/controller"
	"tsu-self/internal/app/admin/service"
	"tsu-self/internal/middleware"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"

	"github.com/labstack/echo/v4"
)

func main() {
	// 初始化配置
	config := loadConfig()

	// 初始化日志
	log.Init(config.LogLevel, config.Environment)
	logger := log.GetLogger()

	logger.Info("Admin 服务启动中",
		log.String("environment", config.Environment),
		log.String("log_level", config.LogLevel.String()),
		log.String("port", config.Port),
		log.String("kratos_public_url", config.KratosPublicURL),
		log.String("kratos_admin_url", config.KratosAdminURL),
	)

	// 初始化服务
	authService, err := service.NewAuthService(config.KratosPublicURL, config.KratosAdminURL, logger)
	if err != nil {
		logger.Error("初始化认证服务失败", err)
		os.Exit(1)
	}

	userService, err := service.NewUserService(config.KratosAdminURL, logger)
	if err != nil {
		logger.Error("初始化用户服务失败", err)
		os.Exit(1)
	}

	// 初始化响应处理器
	respWriter := response.NewResponseHandler(logger, config.Environment)

	// 初始化控制器
	authController := controller.NewAuthController(authService, respWriter, logger)
	userController := controller.NewUserController(userService, respWriter, logger)

	// 初始化 Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// 注册中间件
	registerMiddleware(e, respWriter, logger, config)

	// 注册路由
	registerRoutes(e, authController, userController)

	// 启动服务器
	startServer(e, logger, config.Port)
}

// Config 应用配置
type Config struct {
	Environment     string
	LogLevel        slog.Level
	Port            string
	KratosPublicURL string
	KratosAdminURL  string
	EnableRateLimit bool
	TrustedProxies  []string
}

// loadConfig 加载配置
func loadConfig() *Config {
	return &Config{
		Environment:     getEnv("ENV", "development"),
		LogLevel:        parseLogLevel(getEnv("LOG_LEVEL", "info")),
		Port:            getEnv("PORT", "8080"),
		KratosPublicURL: getEnv("KRATOS_PUBLIC_URL", "http://kratos:4433"),
		KratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://kratos:4434"),
		EnableRateLimit: getEnv("ENABLE_RATE_LIMIT", "false") == "true",
		TrustedProxies:  []string{"127.0.0.1", "::1"}, // 可以从环境变量读取
	}
}

// registerMiddleware 注册中间件
func registerMiddleware(e *echo.Echo, respWriter response.Writer, logger log.Logger, config *Config) {
	// 设置信任的代理
	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	// 恢复中间件（最先注册）
	e.Use(middleware.RecoveryMiddleware(respWriter, logger))

	// 安全中间件
	e.Use(middleware.SecurityMiddleware())

	// CORS 中间件
	e.Use(middleware.CORSMiddleware())

	// 限流中间件（可选）
	if config.EnableRateLimit {
		e.Use(middleware.RateLimitMiddleware())
	}

	// 链路追踪中间件
	e.Use(middleware.TraceMiddleware())

	// 日志中间件
	e.Use(middleware.LoggingMiddleware(logger))

	// 错误处理中间件（最后注册）
	e.Use(middleware.ErrorMiddleware(respWriter, logger))
}

// registerRoutes 注册路由
func registerRoutes(e *echo.Echo, authController *controller.AuthController, userController *controller.UserController) {
	// 健康检查
	e.GET("/health", healthCheck)
	e.GET("/ready", readinessCheck)

	// API 路由
	api := e.Group("/api/v1")

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		auth.POST("/logout", authController.Logout)
		auth.GET("/session", authController.GetSession)
		auth.POST("/recovery", authController.InitRecovery)
		auth.POST("/recovery/submit", authController.SubmitRecovery)
	}

	// 用户管理路由
	users := api.Group("/users")
	{
		users.GET("", userController.ListUsers)
		users.GET("/:id", userController.GetUser)
		users.PUT("/:id", userController.UpdateUser)
		users.DELETE("/:id", userController.DeleteUser)
		users.POST("/:id/disable", userController.DisableUser)
		users.POST("/:id/enable", userController.EnableUser)
	}

	// 管理员路由
	admin := api.Group("/admin")
	{
		admin.GET("/identities", userController.ListIdentities)
		admin.POST("/identities", userController.CreateIdentity)
		admin.GET("/identities/:id", userController.GetIdentity)
		admin.PUT("/identities/:id", userController.UpdateIdentity)
		admin.DELETE("/identities/:id", userController.DeleteIdentity)
	}

	// 404 处理
	e.RouteNotFound("/*", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "API 端点不存在")
	})
}

// healthCheck 健康检查
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "tsu-admin",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
	})
}

// readinessCheck 就绪检查
func readinessCheck(c echo.Context) error {
	// 这里可以检查依赖服务是否就绪
	// 例如检查 Kratos 是否可达
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ready",
		"service": "tsu-admin",
		"checks": map[string]string{
			"kratos":   "ok",
			"database": "ok",
		},
		"timestamp": time.Now().Unix(),
	})
}

// startServer 启动服务器
func startServer(e *echo.Echo, logger log.Logger, port string) {
	logger.Info("准备启动服务器",
		log.String("port", port),
		log.String("address", "0.0.0.0:"+port),
	)

	// 异步启动服务器
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败", err)
			os.Exit(1)
		}
	}()

	logger.Info("服务器已启动",
		log.String("port", port),
		log.String("health_check", "http://localhost:"+port+"/health"),
		log.String("ready_check", "http://localhost:"+port+"/ready"),
		log.String("api_base", "http://localhost:"+port+"/api/v1"),
	)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("收到关闭信号，正在优雅关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭出错", err)
	} else {
		logger.Info("服务器已成功关闭")
	}
}

// 辅助函数

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
