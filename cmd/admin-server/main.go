package main

import (
	"fmt"
	"os"
	"time"

	docs "tsu-self/docs/admin"
	"tsu-self/internal/modules/admin"
	"tsu-self/internal/modules/auth"

	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	"github.com/nats-io/nats.go"
)

// @title           TSU Game Admin API
// @version         1.0
// @description     DnD RPG 游戏管理后台 API - 基于 mqant 微服务架构
// @termsOfService  http://swagger.io/terms/

// @contact.name   TSU API Support
// @contact.url    https://github.com/your-org/tsu-server
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}

func main() {
	fmt.Println("==============================================")
	fmt.Println("  TSU Admin Server")
	fmt.Println("  Version: 1.0.0")
	fmt.Println("==============================================")
	fmt.Println()

	// Consul address
	consulAddr := os.Getenv("CONSUL_ADDRESS")
	if consulAddr == "" {
		consulAddr = "localhost:8500"
	}
	fmt.Printf("[Main] Consul address: %s\n", consulAddr)

	// NATS address
	natsAddr := os.Getenv("NATS_ADDRESS")
	if natsAddr == "" {
		natsAddr = "localhost:4222"
	}
	fmt.Printf("[Main] NATS address: %s\n", natsAddr)

	// Connect to NATS
	nc, err := nats.Connect("nats://"+natsAddr,
		nats.MaxReconnects(10),
		nats.ReconnectWait(1*time.Second),
	)
	if err != nil {
		fmt.Printf("[Main] Failed to connect to NATS: %v\n", err)
		return
	}
	fmt.Println("[Main] Connected to NATS successfully")

	// Configure Swagger to follow current request origin
	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Create Consul registry
	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{consulAddr}
	})

	// Create mqant app with configuration
	// 注意：RegisterTTL 和 RegisterInterval 应该在各个模块的 OnInit 中配置
	// 而不是在全局 app 配置中设置（参考 mqant 官方文档）
	app := mqant.CreateApp(
		module.Configure("./configs/server/admin-server.json"),
		module.Debug(false),
		module.Nats(nc),
		module.Registry(rs),
	)

	fmt.Println("[Main] Configuration loaded")

	// Run with modules
	app.Run(
		auth.Module(),
		admin.Module(),
	)
}
