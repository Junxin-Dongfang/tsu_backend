// File: internal/modules/admin/admin_module.go
package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	_ "github.com/lib/pq" // PostgreSQL driver
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "tsu-self/docs"
	custommiddleware "tsu-self/internal/middleware"
	customvalidator "tsu-self/internal/model/validator"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
)

// CustomValidator 自定义验证器
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var Module = func() module.Module {
	this := new(AdminModule)
	return this
}

type AdminModule struct {
	basemodule.BaseModule
	app         module.App
	echoServer  *echo.Echo
	respWriter  response.Writer
	logger      log.Logger
	db          *sqlx.DB
	syncService *service.SyncService
	userService *service.UserService
}

func (m *AdminModule) GetType() string {
	return "admin"
}

func (m *AdminModule) Version() string {
	return "1.0.0"
}

func (m *AdminModule) OnAppConfigurationLoaded(app module.App) {
	//当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
	m.BaseModule.OnAppConfigurationLoaded(app)
}

func (m *AdminModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)

	m.logger = log.GetLogger()
	m.logger.Info("初始化 Admin 模块...")

	// 初始化数据库连接
	if err := m.initDatabase(); err != nil {
		panic("初始化数据库失败: " + err.Error())
	}

	// 初始化服务
	m.initServices()

	// 初始化 Echo HTTP 服务器
	m.echoServer = echo.New()
	m.echoServer.HideBanner = true
	m.echoServer.HidePort = true

	// 设置验证器
	v := validator.New()
	customvalidator.RegisterAuthValidators(v)
	m.echoServer.Validator = &CustomValidator{validator: v}

	// 设置HTTP响应头中间件（支持UTF-8）
	m.echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
			return next(c)
		}
	})

	// 设置中间件
	m.setupMiddleware()

	// 设置路由
	m.setupRoutes()

	// 注册 RPC 方法
	m.setupRPCMethods()

	// 启动 HTTP 服务器
	go m.startHTTPServer()

	m.logger.Info("Admin 模块初始化完成")
}

func (m *AdminModule) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", m.GetType())
	<-closeSig
	log.Info("%v模块已停止...", m.GetType())
}

func (m *AdminModule) OnDestroy() {
	m.logger.Info("正在关闭 Admin 模块...")

	if m.echoServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := m.echoServer.Shutdown(ctx); err != nil {
			m.logger.Error("关闭 HTTP 服务器失败", err)
		}
	}

	if m.db != nil {
		m.db.Close()
	}

	m.BaseModule.OnDestroy()
	m.logger.Info("Admin 模块已关闭")
}

func (m *AdminModule) initDatabase() error {
	settings := m.GetModuleSettings().Settings
	databaseURL := settings["database_url"].(string)
	if databaseURL == "" {
		return fmt.Errorf("database_url 配置缺失")
	}

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return err
	}

	// 设置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return err
	}

	m.db = db
	m.logger.Info("数据库连接初始化成功")
	return nil
}

func (m *AdminModule) initServices() {
	// 从配置中获取参数
	settings := m.GetModuleSettings().Settings
	log.Info("初始化服务", log.Any("settings", settings))
	environment := settings["environment"].(string)

	// 初始化响应处理器
	m.respWriter = response.NewResponseHandler(m.logger, environment)

	// 初始化 SyncService
	m.syncService = service.NewSyncService(m.db, m.logger)

	// 初始化 UserService
	m.userService = service.NewUserService(m.db, m.logger)

	m.app = m.GetApp()
}

func (m *AdminModule) setupMiddleware() {
	// 恢复中间件
	m.echoServer.Use(middleware.Recover())

	// 使用项目统一的日志中间件，过滤健康检查请求
	m.echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		loggingMiddleware := custommiddleware.LoggingMiddleware(m.logger)
		return func(c echo.Context) error {
			// 跳过健康检查请求的日志
			if c.Request().URL.Path == "/health" {
				return next(c)
			}

			// 设置必要的context值，如果不存在的话
			if c.Get("trace_id") == nil {
				c.Set("trace_id", "")
			}
			if c.Get("request_id") == nil {
				c.Set("request_id", "")
			}

			return loggingMiddleware(next)(c)
		}
	})

	// CORS 中间件
	m.echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions, // 重要：支持预检请求
		},
		AllowHeaders: []string{
			"*",
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Session-Token",
			"X-User-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,  // 允许携带认证信息
		MaxAge:           86400, // 预检请求缓存时间
	}))
}

func (m *AdminModule) setupRoutes() {
	// Swagger 文档 (开发环境)
	environment := m.GetModuleSettings().Settings["environment"].(string)
	if environment == "development" {
		m.echoServer.GET("/swagger/*", echoSwagger.WrapHandler)
		m.echoServer.GET("/docs", func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})
	}

	// 健康检查
	// m.echoServer.GET("/health", m.healthCheck)
	// m.echoServer.GET("/ready", m.readinessCheck)

	// API 路由
	api := m.echoServer.Group("")

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.POST("/login", m.Login)
		auth.POST("/register", m.Register)
	}

}

func (m *AdminModule) setupRPCMethods() {
	// 注册 RPC 方法供其他模块调用
	// m.GetServer().RegisterGO("ValidateSession", m.rpcValidateSession)
	// m.GetServer().RegisterGO("GetUserInfo", m.rpcGetUserInfo)
	// m.GetServer().RegisterGO("Login", m.rpcLogin)
}

func (m *AdminModule) startHTTPServer() {
	httpPort := m.GetModuleSettings().Settings["http_port"].(string)
	m.logger.Info("启动 HTTP 服务器", log.String("port", httpPort))

	if err := m.echoServer.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
		m.logger.Error("HTTP 服务器启动失败", err)
		panic(err)
	}
}
