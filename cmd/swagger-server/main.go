package main

import (
	"log/slog"
	"os"
	"tsu-self/internal/modules/swagger"
	"tsu-self/internal/pkg/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	log.Init(slog.LevelDebug, "development")
	log.Info("启动 TSU Swagger 聚合服务器...")
	log.Info("版本: 1.0.0")

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// 中间件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}))

	// 获取 Consul 地址
	consulAddr := "127.0.0.1:8500"
	if envConsulAddr := os.Getenv("CONSUL_ADDR"); envConsulAddr != "" {
		consulAddr = envConsulAddr
	}

	// 初始化 Swagger 模块
	swaggerModule, err := swagger.NewSwaggerModule(consulAddr)
	if err != nil {
		log.Error("初始化 Swagger 模块失败", err)
		panic(err)
	}

	// 注册路由
	swaggerModule.RegisterRoutes(e)

	// 启动服务器
	port := "8080"
	log.Info("Swagger 服务器启动", log.String("port", port))

	if err := e.Start(":" + port); err != nil {
		log.Error("Swagger 服务器启动失败", err)
		panic(err)
	}
}
