package auth

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"tsu-self/internal/modules/auth/client"
	"tsu-self/internal/modules/auth/handler"
	"tsu-self/internal/modules/auth/service"
	redisClient "tsu-self/internal/pkg/redis"

	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/server"
	_ "github.com/lib/pq"
)

// Module Auth module
type AuthModule struct {
	basemodule.BaseModule
	db                   *sql.DB
	redis                *redisClient.Client
	authService          *service.AuthService
	permissionService    *service.PermissionService
	userService          *service.UserService
	rpcHandler           *handler.RPCHandler
	permissionRPCHandler *handler.PermissionRPCHandler
	userRPCHandler       *handler.UserRPCHandler
}

// GetType returns module type
func (m *AuthModule) GetType() string {
	return "auth"
}

// Version returns module version
func (m *AuthModule) Version() string {
	return "1.0.0"
}

// OnAppConfigurationLoaded 当App初始化时调用
func (m *AuthModule) OnAppConfigurationLoaded(app module.App) {
	m.BaseModule.OnAppConfigurationLoaded(app)
}

// OnInit module initialization
func (m *AuthModule) OnInit(app module.App, settings *conf.ModuleSettings) {
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

	// 2. Initialize Redis
	if err := m.initRedis(settings); err != nil {
		panic(fmt.Sprintf("Failed to initialize Redis: %v", err))
	}

	// 3. Initialize Kratos Client
	kratosClient := m.initKratosClient(settings)

	// 4. Initialize Keto Client
	ketoClient := m.initKetoClient(settings)

	// 5. Initialize Services
	m.authService = service.NewAuthService(m.db, kratosClient, m.redis)
	m.permissionService = service.NewPermissionService(m.db, ketoClient)
	m.userService = service.NewUserService(m.db)

	// 5. Initialize RPC Handlers
	m.rpcHandler = handler.NewRPCHandler(m.authService)
	m.permissionRPCHandler = handler.NewPermissionRPCHandler(m.db, m.permissionService)
	m.userRPCHandler = handler.NewUserRPCHandler(m.db, m.userService)

	// 6. Register RPC methods
	m.setupRPCMethods()

	m.GetServer().Options()
}

// initDatabase initializes database connection
func (m *AuthModule) initDatabase(settings *conf.ModuleSettings) error {
	// Read from environment variable first
	dbURL := os.Getenv("TSU_AUTH_DATABASE_URL")
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
		return fmt.Errorf("database URL not configured, please set TSU_AUTH_DATABASE_URL environment variable")
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
	fmt.Println("[Auth Module] Database connected successfully")
	return nil
}

// initRedis initializes Redis client
func (m *AuthModule) initRedis(settings *conf.ModuleSettings) error {
	// Read Redis configuration from environment variables
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // Default value
	}

	portStr := os.Getenv("REDIS_PORT")
	port := 6379 // Default port
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	password := os.Getenv("REDIS_PASSWORD")
	// 密码可以为空（本地开发）

	dbStr := os.Getenv("REDIS_DB")
	db := 0 // Default DB
	if dbStr != "" {
		if d, err := strconv.Atoi(dbStr); err == nil {
			db = d
		}
	}

	// Create Redis client
	redisClient, err := redisClient.NewClient(redisClient.Config{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	m.redis = redisClient
	fmt.Printf("[Auth Module] Redis connected successfully (Host: %s:%d, DB: %d)\n", host, port, db)
	return nil
}

// initKratosClient initializes Kratos client
func (m *AuthModule) initKratosClient(settings *conf.ModuleSettings) *client.KratosClient {
	// Read Admin URL from environment variable first
	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminURL == "" {
		// Fallback to config file
		if settings != nil && settings.Settings != nil {
			kratosAdminURLInterface, ok := settings.Settings["kratos_admin_url"]
			if ok {
				kratosAdminURL, _ = kratosAdminURLInterface.(string)
			}
		}
	}

	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434" // Default value
	}

	// Read Public URL from environment variable
	kratosPublicURL := os.Getenv("KRATOS_PUBLIC_URL")
	if kratosPublicURL == "" {
		// Fallback to config file
		if settings != nil && settings.Settings != nil {
			kratosPublicURLInterface, ok := settings.Settings["kratos_public_url"]
			if ok {
				kratosPublicURL, _ = kratosPublicURLInterface.(string)
			}
		}
	}

	if kratosPublicURL == "" {
		kratosPublicURL = "http://localhost:4433" // Default value
	}

	fmt.Printf("[Auth Module] Kratos Admin URL: %s\n", kratosAdminURL)
	fmt.Printf("[Auth Module] Kratos Public URL: %s\n", kratosPublicURL)

	kratosClient := client.NewKratosClient(kratosAdminURL)
	kratosClient.SetPublicURL(kratosPublicURL)
	return kratosClient
}

// initKetoClient initializes Keto client
func (m *AuthModule) initKetoClient(settings *conf.ModuleSettings) *client.KetoClient {
	// Read from environment variable first
	ketoReadURL := os.Getenv("KETO_READ_URL")
	ketoWriteURL := os.Getenv("KETO_WRITE_URL")

	if ketoReadURL == "" {
		// Fallback to config file
		if settings != nil && settings.Settings != nil {
			ketoReadURLInterface, ok := settings.Settings["keto_read_url"]
			if ok {
				ketoReadURL, _ = ketoReadURLInterface.(string)
			}
		}
	}

	if ketoWriteURL == "" {
		// Fallback to config file
		if settings != nil && settings.Settings != nil {
			ketoWriteURLInterface, ok := settings.Settings["keto_write_url"]
			if ok {
				ketoWriteURL, _ = ketoWriteURLInterface.(string)
			}
		}
	}

	// Default values
	if ketoReadURL == "" {
		ketoReadURL = "localhost:4466"
	}
	if ketoWriteURL == "" {
		ketoWriteURL = "localhost:4467"
	}

	fmt.Printf("[Auth Module] Keto Read URL: %s, Write URL: %s\n", ketoReadURL, ketoWriteURL)

	ketoClient, err := client.NewKetoClient(ketoReadURL, ketoWriteURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Keto client: %v", err))
	}

	return ketoClient
}

// setupRPCMethods registers RPC methods
func (m *AuthModule) setupRPCMethods() {
	// ==================== 用户认证 RPC ====================
	m.GetServer().RegisterGO("Register", m.rpcHandler.Register)
	m.GetServer().RegisterGO("GetUser", m.rpcHandler.GetUser)
	m.GetServer().RegisterGO("UpdateLoginInfo", m.rpcHandler.UpdateLoginInfo)
	m.GetServer().RegisterGO("SyncUserFromKratos", m.rpcHandler.SyncUserFromKratos)
	m.GetServer().RegisterGO("Login", m.rpcHandler.Login)
	m.GetServer().RegisterGO("Logout", m.rpcHandler.Logout)

	// ==================== 密码重置 RPC ====================
	m.GetServer().RegisterGO("InitiateRecovery", m.rpcHandler.InitiateRecovery)
	m.GetServer().RegisterGO("VerifyRecoveryCode", m.rpcHandler.VerifyRecoveryCode)
	m.GetServer().RegisterGO("ResetPassword", m.rpcHandler.ResetPassword)
	m.GetServer().RegisterGO("AdminCreateRecoveryCode", m.rpcHandler.AdminCreateRecoveryCode)

	// ==================== 权限检查 RPC ====================
	m.GetServer().RegisterGO("CheckUserPermission", m.permissionRPCHandler.CheckUserPermission)

	// ==================== 角色管理 RPC ====================
	m.GetServer().RegisterGO("GetRoles", m.permissionRPCHandler.GetRoles)
	m.GetServer().RegisterGO("CreateRole", m.permissionRPCHandler.CreateRole)
	m.GetServer().RegisterGO("UpdateRole", m.permissionRPCHandler.UpdateRole)
	m.GetServer().RegisterGO("DeleteRole", m.permissionRPCHandler.DeleteRole)

	// ==================== 权限管理 RPC ====================
	m.GetServer().RegisterGO("GetPermissions", m.permissionRPCHandler.GetPermissions)
	m.GetServer().RegisterGO("GetPermissionGroups", m.permissionRPCHandler.GetPermissionGroups)

	// ==================== 角色-权限管理 RPC ====================
	m.GetServer().RegisterGO("GetRolePermissions", m.permissionRPCHandler.GetRolePermissions)
	m.GetServer().RegisterGO("AssignPermissionsToRole", m.permissionRPCHandler.AssignPermissionsToRole)

	// ==================== 用户-角色管理 RPC ====================
	m.GetServer().RegisterGO("GetUserRoles", m.permissionRPCHandler.GetUserRoles)
	m.GetServer().RegisterGO("AssignRolesToUser", m.permissionRPCHandler.AssignRolesToUser)
	m.GetServer().RegisterGO("RevokeRolesFromUser", m.permissionRPCHandler.RevokeRolesFromUser)

	// ==================== 用户-权限管理 RPC ====================
	m.GetServer().RegisterGO("GetUserPermissions", m.permissionRPCHandler.GetUserPermissions)
	m.GetServer().RegisterGO("GrantPermissionsToUser", m.permissionRPCHandler.GrantPermissionsToUser)
	m.GetServer().RegisterGO("RevokePermissionsFromUser", m.permissionRPCHandler.RevokePermissionsFromUser)

	// ==================== 用户管理 RPC ====================
	m.GetServer().RegisterGO("GetUsers", m.userRPCHandler.GetUsers)
	m.GetServer().RegisterGO("UpdateUser", m.userRPCHandler.UpdateUser)
	m.GetServer().RegisterGO("BanUser", m.userRPCHandler.BanUser)
	m.GetServer().RegisterGO("UnbanUser", m.userRPCHandler.UnbanUser)

	fmt.Println("[Auth Module] RPC methods registered successfully")
}

// Run module run
func (m *AuthModule) Run(closeSig chan bool) {
	fmt.Println("[Auth Module] Started successfully")
	<-closeSig
}

// OnDestroy module destroy
func (m *AuthModule) OnDestroy() {
	// Close database connection
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			fmt.Printf("[Auth Module] Failed to close database: %v\n", err)
		} else {
			fmt.Println("[Auth Module] Database connection closed")
		}
	}

	// Close Redis connection
	if m.redis != nil {
		if err := m.redis.Close(); err != nil {
			fmt.Printf("[Auth Module] Failed to close Redis: %v\n", err)
		} else {
			fmt.Println("[Auth Module] Redis connection closed")
		}
	}

	m.BaseModule.OnDestroy()
	fmt.Println("[Auth Module] Destroyed")
}

// Module creates Auth module instance
var Module = func() module.Module {
	this := new(AuthModule)
	return this
}
