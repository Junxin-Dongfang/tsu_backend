// @title Tsu Admin API
// @version 1.0
// @description Tsu 后台管理系统 API
// @contact.name Tsu Team
// @contact.email support@tsu.com
// @host localhost
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token format: Bearer {token}
package main

import (
	"log/slog"
	"os"
	"time"
	"tsu-self/internal/modules/admin"
	"tsu-self/internal/pkg/log"

	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	"github.com/nats-io/nats.go"
)

// @title Tsu Admin API
// @version 1.0
// @description Tsu 后台管理 服务 API 文档
// @contact.name Tsu Team
// @contact.email tsu@tsu.com
// @host localhost
// @BasePath /admin/api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式为 Bearer {token}

func main() {
	log.Init(slog.LevelDebug, "development")

	log.Info("启动 TSU Admin Server...")

	// Consul 注册
	consulAddr := "consul:8500"
	if envConsulAddr := os.Getenv("CONSUL_ADDRESS"); envConsulAddr != "" {
		consulAddr = envConsulAddr
	}

	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{consulAddr}
	})
	log.Info("Consul 地址: " + consulAddr)

	//NATS地址
	natsAddr := "nats:4222"
	if envNatsAddr := os.Getenv("NATS_ADDRESS"); envNatsAddr != "" {
		natsAddr = envNatsAddr
	}

	nc, err := nats.Connect(natsAddr,
		nats.MaxReconnects(10000),
		nats.ReconnectWait(1*time.Second), // 重连等待时间
		nats.PingInterval(30*time.Second), // 心跳间隔
		nats.MaxPingsOutstanding(3),       // 最大未响应心跳数
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Info("NATS 连接断开: " + err.Error())
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS 重新连接成功")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS 连接关闭")
		}),
	)
	if err != nil {
		log.Error("连接 NATS 失败", err)
		return
	}
	log.Info("NATS 地址: " + natsAddr)

	// 创建 mqant 应用
	app := mqant.CreateApp(
		module.Configure("./configs/server/admin-server.json"),
		module.KillWaitTTL(1*time.Minute),
		module.Debug(true),
		module.BILogDir("./logs/admin-server"),
		module.Nats(nc),
		module.Registry(rs),
		module.RegisterTTL(10*time.Second),
		module.RegisterInterval(10*time.Second),
	)

	app.OnConfigurationLoaded(func(app module.App) {
		log.Info("配置文件加载完成")
	})
	log.Info("应用配置加载完成", "settings", app.GetSettings().Settings)

	// 注册模块
	log.Info("注册模块...")
	app.Run(
		&admin.AdminModule{},
	)
}
