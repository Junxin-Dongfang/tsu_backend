// cmd/auth-server/main.go
package main

import (
	"log/slog"
	"os"
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

	// Consul 注册
	consulAddr := "127.0.0.1:8500"
	if envConsulAddr := os.Getenv("CONSUL_ADDRESS"); envConsulAddr != "" {
		consulAddr = envConsulAddr
	}

	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{consulAddr}
	})
	log.Info("Consul 地址: " + consulAddr)

	//NATS地址
	natsAddr := "127.0.0.1:4222"
	if envNatsAddr := os.Getenv("NATS_ADDRESS"); envNatsAddr != "" {
		natsAddr = envNatsAddr
	}

	nc, err := nats.Connect(natsAddr, nats.MaxReconnects(10000))
	if err != nil {
		log.Error("连接 NATS 失败", err)
		return
	}
	log.Info("NATS 地址: " + natsAddr)
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
