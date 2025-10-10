package game

import (
	"fmt"
	"os"
	"time"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/server"
)

type GameModule struct {
	basemodule.BaseModule
	httpServer *echo.Echo
	respWriter response.Writer
}

// GetType returns module type
func (m *GameModule) GetType() string {
	return "admin"
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

	// // 2. Initialize database connection (optional for admin module)
	// if err := m.initDatabase(settings); err != nil {
	// 	fmt.Printf("[Admin Module] Warning: Database initialization failed: %v\n", err)
	// }

	// // 3. Initialize HTTP server
	// m.initHTTPServer(settings)

	// // 4. Initialize handlers
	// m.initHandlers(app)

	// // 5. Setup routes
	// m.setupRoutes()

	// // 6. Start HTTP server in background
	// go m.startHTTPServer(settings)

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
	fmt.Println("[Admin Module] Response writer initialized")
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
			fmt.Printf("[Admin Module] Failed to close HTTP server: %v\n", err)
		} else {
			fmt.Println("[Admin Module] HTTP server closed")
		}
	}

	// // Close database connection
	// if m.db != nil {
	// 	if err := m.db.Close(); err != nil {
	// 		fmt.Printf("[Admin Module] Failed to close database: %v\n", err)
	// 	} else {
	// 		fmt.Println("[Admin Module] Database connection closed")
	// 	}
	// }

	m.BaseModule.OnDestroy()
	fmt.Println("[Admin Module] Destroyed")
}

// Module creates Admin module instance
func Module() module.Module {
	return new(GameModule)
}
