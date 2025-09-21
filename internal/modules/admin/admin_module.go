// File: internal/modules/admin/admin_module.go
package admin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "tsu-self/docs"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
)

var Module = func() module.Module {
	this := new(AdminModule)
	return this
}

type AdminModule struct {
	basemodule.BaseModule
	echoServer  *echo.Echo
	authService *service.AuthService
	userService *service.UserService
	respWriter  response.Writer
	logger      log.Logger
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

	// 初始化服务
	m.initServices()

	// 初始化 Echo HTTP 服务器
	m.echoServer = echo.New()
	m.echoServer.HideBanner = true
	m.echoServer.HidePort = true

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

	m.BaseModule.OnDestroy()
	m.logger.Info("Admin 模块已关闭")
}

func (m *AdminModule) initServices() {
	// 从配置中获取参数
	settings := m.GetModuleSettings().Settings
	log.Info("初始化服务", log.Any("settings", settings))
	kratosPublicURL := settings["kratos_public_url"].(string)
	kratosAdminURL := settings["kratos_admin_url"].(string)
	environment := settings["environment"].(string)

	// 初始化响应处理器
	m.respWriter = response.NewResponseHandler(m.logger, environment)

	// 初始化认证服务
	var err error
	m.authService, err = service.NewAuthService(kratosPublicURL, kratosAdminURL, m.logger)
	if err != nil {
		m.logger.Error("初始化认证服务失败", err)
		panic(err)
	}

	// 初始化用户服务
	m.userService, err = service.NewUserService(kratosAdminURL, m.logger)
	if err != nil {
		m.logger.Error("初始化用户服务失败", err)
		panic(err)
	}
}

func (m *AdminModule) setupMiddleware() {
	// 恢复中间件
	m.echoServer.Use(middleware.Recover())

	// 日志中间件
	m.echoServer.Use(middleware.Logger())

	// CORS 中间件
	m.echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{"*"},
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
		auth.POST("/logout", m.Logout)
		auth.GET("/session", m.GetSession)
		auth.POST("/recovery", m.InitRecovery)
		auth.POST("/recovery/submit", m.SubmitRecovery)
	}

	// 用户管理路由
	users := api.Group("/users")
	{
		users.GET("", m.ListUsers)
		users.GET("/:id", m.GetUser)
		users.PUT("/:id", m.UpdateUser)
		users.DELETE("/:id", m.DeleteUser)
		users.POST("/:id/disable", m.DisableUser)
		users.POST("/:id/enable", m.EnableUser)
	}

	// 管理员路由
	admin := api.Group("/admin")
	{
		admin.GET("/identities", m.ListIdentities)
		admin.POST("/identities", m.CreateIdentity)
		admin.GET("/identities/:id", m.GetIdentity)
		admin.PUT("/identities/:id", m.UpdateIdentity)
		admin.DELETE("/identities/:id", m.DeleteIdentity)
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
	// if httpPortStr := m.GetModuleSettings().Settings["http_port"].(string); httpPortStr != "" {
	// 	if port, err := strconv.Atoi(httpPortStr); err == nil {
	// 		portInt = port
	// 	}
	// }

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
