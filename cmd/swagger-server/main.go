package main

import (
	"log"
	_ "tsu-self/internal/model/authmodel" // 确保类型定义被包含
	_ "tsu-self/internal/model/usermodel" // 确保类型定义被包含
	_ "tsu-self/internal/modules/admin"   // 导入 admin 模块以确保 swagger 能找到所有类型
	"tsu-self/internal/modules/swagger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// @title TSU API
// @version 1.0
// @description TSU Server API documentation
// @host localhost:8080
// @BasePath /
func main() {
	e := echo.New()

	// 中间件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 初始化 Swagger 模块
	swaggerModule := swagger.NewSwaggerModule()
	swaggerModule.RegisterRoutes(e)

	// 启动服务器
	log.Println("Swagger server starting on :8080")
	log.Fatal(e.Start(":8080"))
}
