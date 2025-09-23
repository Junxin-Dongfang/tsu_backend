// internal/modules/auth/auth_module.go - 完整版本
package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	"github.com/redis/go-redis/v9"

	custommiddleware "tsu-self/internal/middleware"
	"tsu-self/internal/modules/auth/service"
	"tsu-self/internal/pkg/log"
)

type AuthModule struct {
	basemodule.BaseModule

	// Services
	kratosService       *service.KratosService
	ketoService         *service.KetoService
	sessionService      *service.SessionService
	notificationService *service.NotificationService

	// Infrastructure
	redis      *redis.Client
	logger     log.Logger
	echoServer *echo.Echo
}

func (m *AuthModule) Version() string {
	return "1.0.0"
}

func (m *AuthModule) GetType() string {
	return "auth"
}

func (m *AuthModule) OnAppConfigurationLoaded(app module.App) {
	//当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
	m.BaseModule.OnAppConfigurationLoaded(app)
}

func (m *AuthModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)

	// 初始化日志
	m.logger = log.GetLogger().WithGroup("auth-module")

	// 初始化 Redis
	if err := m.initRedis(); err != nil {
		panic("初始化 Redis 失败: " + err.Error())
	}

	// 初始化服务
	m.initServices()

	// 初始化 Echo 服务器
	m.initEchoServer()

	// 检查是否配置了HTTP端口，如果配置了就启动HTTP服务器
	if httpPort, exists := m.GetModuleSettings().Settings["http_port"]; exists && httpPort != "" {
		// 启动 HTTP 服务器
		go m.startHTTPServer()

		// 注册 HTTP 服务到 Consul
		go m.registerHTTPService()
	}

	m.logger.Info("Auth Module 初始化完成")
}

func (m *AuthModule) Run(closeSig chan bool) {
	m.logger.Info("Auth Module 开始运行")

	// 注册RPC处理器
	rpcHandler := NewAuthRPCHandler(
		m.kratosService,
		m.ketoService,
		m.sessionService,
		m.notificationService,
		m.logger,
	)

	// 注册所有 RPC 方法
	m.GetServer().RegisterGO("Login", rpcHandler.Login)
	m.GetServer().RegisterGO("Register", rpcHandler.Register)
	m.GetServer().RegisterGO("ValidateToken", rpcHandler.ValidateToken)
	m.GetServer().RegisterGO("Logout", rpcHandler.Logout)
	m.GetServer().RegisterGO("CheckPermission", rpcHandler.CheckPermission)
	m.GetServer().RegisterGO("GetUserInfo", rpcHandler.GetUserInfo)
	m.GetServer().RegisterGO("UpdateUserTraits", rpcHandler.UpdateUserTraits)
	m.GetServer().RegisterGO("AssignRole", rpcHandler.AssignRole)
	m.GetServer().RegisterGO("RevokeRole", rpcHandler.RevokeRole)
	m.GetServer().RegisterGO("CreateRole", rpcHandler.CreateRole)

	m.logger.Info("Auth Module RPC 处理器注册完成")

	<-closeSig
}

func (m *AuthModule) OnDestroy() {
	m.logger.Info("Auth Module 正在关闭")

	if m.redis != nil {
		m.redis.Close()
	}

	m.BaseModule.OnDestroy()
}


func (m *AuthModule) initRedis() error {
	settings := m.GetModuleSettings().Settings
	redisAddr := settings["redis_addr"].(string)
	redisPassword := settings["redis_password"].(string)
	redisDB := settings["redis_db"].(float64)

	m.redis = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       int(redisDB),
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.redis.Ping(ctx).Err(); err != nil {
		return err
	}

	m.logger.Info("Redis 连接初始化成功", log.String("addr", redisAddr))
	return nil
}

func (m *AuthModule) initServices() {
	// 初始化 KratosService
	kratosPublicURL := m.GetModuleSettings().Settings["kratos_public_url"].(string)
	kratosAdminURL := m.GetModuleSettings().Settings["kratos_admin_url"].(string)

	var err error
	m.kratosService, err = service.NewKratosService(kratosPublicURL, kratosAdminURL, m.logger)
	if err != nil {
		panic("初始化 Kratos Service 失败: " + err.Error())
	}

	// 初始化 SessionService
	jwtSecret := m.GetModuleSettings().Settings["jwt_secret"].(string)
	tokenTTL := time.Duration(m.GetModuleSettings().Settings["token_ttl_minutes"].(float64)) * time.Minute
	m.sessionService = service.NewSessionService(m.redis, jwtSecret, tokenTTL, m.logger)

	// 初始化 KetoService
	ketoReadURL := m.GetModuleSettings().Settings["keto_read_url"].(string)
	ketoWriteURL := m.GetModuleSettings().Settings["keto_write_url"].(string)
	m.ketoService = service.NewKetoService(ketoReadURL, ketoWriteURL, m.logger)

	// 初始化 NotificationService
	app := m.GetApp()
	m.notificationService = service.NewNotificationService(app.Options().Nats, m.logger)

	m.logger.Info("所有服务初始化完成")
}

// 初始化 Echo 服务器
func (m *AuthModule) initEchoServer() {
	m.echoServer = echo.New()
	m.echoServer.HideBanner = true
	m.echoServer.HidePort = true

	// 添加中间件
	m.echoServer.Use(middleware.Recover())
	m.echoServer.Use(middleware.CORS())

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

	// 注册路由
	m.setupHTTPRoutes()
}

// 设置 HTTP 路由
func (m *AuthModule) setupHTTPRoutes() {
	// 创建健康检查处理器
	healthHandler := &HealthHandler{module: m}

	// 注册路由
	m.echoServer.GET("/health", healthHandler.Health)
}

// 启动 HTTP 服务器
func (m *AuthModule) startHTTPServer() {
	httpPort := m.GetModuleSettings().Settings["http_port"].(string)
	m.logger.Info("启动 HTTP 服务器", log.String("port", httpPort))

	if err := m.echoServer.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
		m.logger.Error("HTTP 服务器启动失败", err)
		panic(err)
	}
}

// 注册 HTTP 服务到 Consul
func (m *AuthModule) registerHTTPService() {
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

	// 获取HTTP端口
	httpPortStr := m.GetModuleSettings().Settings["http_port"].(string)
	portInt := 8082 // 默认端口
	if port, err := strconv.Atoi(httpPortStr); err == nil {
		portInt = port
	}

	// 注册 HTTP 服务
	registration := &api.AgentServiceRegistration{
		ID:      "auth-http",
		Name:    "auth-http",
		Port:    portInt,
		Address: containerIP,
		Tags:    []string{"http", "auth", "authentication"},
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
func (m *AuthModule) getContainerIP() string {
	// 通过网络接口获取
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
