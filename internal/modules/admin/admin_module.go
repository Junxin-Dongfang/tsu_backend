// File: internal/modules/admin/admin_module.go
package admin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/consul/api"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	_ "github.com/lib/pq" // PostgreSQL driver
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "tsu-self/docs"
	customvalidator "tsu-self/internal/api/model/validator"
	custommiddleware "tsu-self/internal/middleware"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/impl"
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
	app                module.App
	echoServer         *echo.Echo
	respWriter         response.Writer
	logger             log.Logger
	db                 *sqlx.DB
	syncService          *service.SyncService
	transactionService   *service.TransactionService
	userService          *service.UserService
	classService         *service.ClassService
	attributeTypeService *service.AttributeTypeService
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

	// 注册 HTTP 服务到 Consul
	go m.registerHTTPService()

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

	// 初始化 TransactionService
	m.transactionService = service.NewTransactionService(m.db, m.syncService, m.logger)

	// 初始化 UserService
	m.userService = service.NewUserService(m.db, m.logger)

	// 初始化 ClassService
	classRepo := impl.NewClassRepository(m.db.DB)
	m.classService = service.NewClassService(classRepo)

	// 初始化 AttributeTypeService
	attributeTypeRepo := impl.NewAttributeTypeRepository(m.db.DB)
	m.attributeTypeService = service.NewAttributeTypeService(attributeTypeRepo)

	m.app = m.GetApp()

	// 启动事件监听器
	go m.startEventListeners()
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
	m.echoServer.GET("/health", m.healthCheck)
	m.echoServer.GET("/ready", m.readinessCheck)

	// API 路由
	api := m.echoServer.Group("")

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.POST("/login", m.Login)
		auth.POST("/register", m.Register)
	}

	user := api.Group("/user")
	{
		user.GET("/:user_id/profile", m.GetUserProfile)
		user.PUT("/:user_id/profile", m.UpdateUserProfile)
	}

	// 职业管理路由
	admin := api.Group("/admin")
	{
		// 职业基础管理
		classes := admin.Group("/classes")
		{
			classes.GET("", m.ListClasses)             // 获取职业列表
			classes.POST("", m.CreateClass)            // 创建职业
			classes.GET("/:id", m.GetClass)            // 获取职业详情
			classes.PUT("/:id", m.UpdateClass)         // 更新职业
			classes.DELETE("/:id", m.DeleteClass)      // 删除职业
			classes.GET("/:id/basic", m.GetClassBasic) // 获取职业基本信息
			classes.GET("/:id/stats", m.GetClassStats) // 获取职业统计信息

			// 职业属性加成管理
			classes.GET("/:id/attribute-bonuses", m.GetClassAttributeBonuses)                // 获取属性加成列表
			classes.POST("/:id/attribute-bonuses", m.CreateClassAttributeBonus)              // 创建属性加成
			classes.POST("/:id/attribute-bonuses/batch", m.BatchCreateClassAttributeBonuses) // 批量创建属性加成
		}

		// 标签管理
		admin.GET("/classes/tags", m.GetAllClassTags) // 获取所有职业标签

		// 属性类型管理
		attributeTypes := admin.Group("/attribute-types")
		{
			attributeTypes.GET("", m.GetAttributeTypes)          // 获取属性类型列表
			attributeTypes.POST("", m.CreateAttributeType)       // 创建属性类型
			attributeTypes.GET("/options", m.GetAttributeTypeOptions) // 获取属性类型选项
			attributeTypes.GET("/:id", m.GetAttributeType)       // 获取属性类型详情
			attributeTypes.PUT("/:id", m.UpdateAttributeType)    // 更新属性类型
			attributeTypes.DELETE("/:id", m.DeleteAttributeType) // 删除属性类型
		}
	}
}

func (m *AdminModule) setupRPCMethods() {
	// 注册 RPC 方法供其他模块调用
	m.GetServer().RegisterGO("HandleUserRegistered", m.handleUserRegisteredRPC)
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

// 新增方法：注册 HTTP 服务到 Consul
func (m *AdminModule) registerHTTPService() {
	time.Sleep(2 * time.Second) // 等待 HTTP 服务器启动

	// 创建 Consul 客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "consul:8500" // 使用容器名

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		m.logger.Error("创建 Consul 客户端失败", err)
		return
	}

	// 获取容器 IP
	containerIP := m.getContainerIP()
	if containerIP == "" {
		m.logger.Error("无法获取容器 IP", err)
		return
	}
	portInt := 8081 // 默认端口
	if httpPortStr := m.GetModuleSettings().Settings["http_port"].(string); httpPortStr != "" {
		if port, err := strconv.Atoi(httpPortStr); err == nil {
			portInt = port
		}
	}

	// 注册 HTTP 服务
	registration := &api.AgentServiceRegistration{
		ID:      "admin-http",
		Name:    "admin-http",
		Port:    portInt,
		Address: containerIP,
		Tags:    []string{"http", "swagger", "admin"},
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", containerIP, portInt),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		m.logger.Error("注册 HTTP 服务到 Consul 失败", err)
		return
	}

	m.logger.Info("HTTP 服务已注册到 Consul",
		log.String("address", containerIP),
		log.Int("port", portInt))
}

// 获取容器 IP 地址
func (m *AdminModule) getContainerIP() string {
	// 方法1：通过网络接口获取
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// startEventListeners 启动事件监听器 - 使用mqant框架推荐方式
func (m *AdminModule) startEventListeners() {
	// 使用mqant的RPC机制替代直接的NATS订阅
	// 这避免了与框架内部订阅的冲突

	m.logger.Info("使用mqant RPC机制处理事件，无需手动NATS订阅")

	// 如果需要监听其他服务的事件，应该通过RPC调用
	// 而不是直接创建NATS订阅，这样可以避免订阅冲突
}

// handleUserRegisteredRPC 通过RPC处理用户注册事件 - 使用mqant推荐方式
// 当其他服务需要通知用户注册事件时，可以通过RPC调用这个方法
func (m *AdminModule) handleUserRegisteredRPC(ctx context.Context, userID, email, username string) error {
	m.logger.InfoContext(ctx, "通过RPC收到用户注册事件",
		log.String("user_id", userID),
		log.String("email", email),
		log.String("username", username))

	// 同步用户到主数据库
	_, syncErr := m.syncService.CreateBusinessUser(ctx, userID, email, username)
	if syncErr != nil {
		m.logger.ErrorContext(ctx, "同步用户到主数据库失败",
			log.String("user_id", userID),
			log.Any("error", syncErr))
		return syncErr
	}

	m.logger.InfoContext(ctx, "用户同步到主数据库成功", log.String("user_id", userID))
	return nil
}
