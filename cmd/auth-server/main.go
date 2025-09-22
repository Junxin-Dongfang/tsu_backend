// cmd/auth-server/main.go
package main

import (
	"log/slog"
	"time"

	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	"github.com/nats-io/nats.go"

	"tsu-self/internal/modules/auth"
	"tsu-self/internal/pkg/log"
)

func main() {
	log.Init(slog.LevelDebug, "development")
	log.Info("启动 Auth Server...")

	// Consul & NATS 配置 (复用你的现有配置)
	consulAddr := "127.0.0.1:8500"
	natsAddr := "127.0.0.1:4222"

	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{consulAddr}
	})

	nc, err := nats.Connect(natsAddr, nats.MaxReconnects(10000))
	if err != nil {
		log.Error("连接 NATS 失败", err)
		return
	}

	// 创建应用
	app := mqant.CreateApp(
		module.Configure("./configs/server/auth-server.json"),
		module.KillWaitTTL(1*time.Minute),
		module.Debug(true),
		module.Nats(nc),
		module.Registry(rs),
		module.RegisterTTL(10*time.Second),
		module.RegisterInterval(10*time.Second),
	)

	// 启动模块
	app.Run(&auth.AuthModule{})
}
