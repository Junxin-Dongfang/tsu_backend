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
