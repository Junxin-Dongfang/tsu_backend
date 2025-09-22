// internal/modules/auth/auth_module.go - 完整版本
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/redis/go-redis/v9"

	"tsu-self/internal/modules/auth/service"
	"tsu-self/internal/pkg/log"
)

type AuthModule struct {
	basemodule.BaseModule

	// Services
	kratosService       *service.KratosService
	ketoService         *service.KetoService
	sessionService      *service.SessionService
	syncService         *service.SyncService
	notificationService *service.NotificationService

	// Infrastructure
	db     *sql.DB
	redis  *redis.Client
	logger log.Logger
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

	// 初始化数据库连接
	if err := m.initDatabase(); err != nil {
		panic("初始化数据库失败: " + err.Error())
	}

	// 初始化 Redis
	if err := m.initRedis(); err != nil {
		panic("初始化 Redis 失败: " + err.Error())
	}

	// 初始化服务
	m.initServices()

	m.logger.Info("Auth Module 初始化完成")
}

func (m *AuthModule) Run(closeSig chan bool) {
	m.logger.Info("Auth Module 开始运行")

	// 注册RPC处理器
	rpcHandler := NewAuthRPCHandler(
		m.kratosService,
		m.ketoService,
		m.sessionService,
		m.syncService,
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

	if m.db != nil {
		m.db.Close()
	}

	m.BaseModule.OnDestroy()
}

func (m *AuthModule) initDatabase() error {
	settings := m.GetModuleSettings().Settings
	databaseURL := settings["database_url"].(string)
	if databaseURL == "" {
		return fmt.Errorf("database_url 配置缺失")
	}

	db, err := sql.Open("postgres", databaseURL)
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
	// 初始化 SyncService
	m.syncService = service.NewSyncService(m.db, m.logger)

	// 初始化 KratosService
	kratosPublicURL := m.GetModuleSettings().Settings["kratos_public_url"].(string)
	kratosAdminURL := m.GetModuleSettings().Settings["kratos_admin_url"].(string)

	var err error
	m.kratosService, err = service.NewKratosService(kratosPublicURL, kratosAdminURL, m.syncService, m.logger)
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
