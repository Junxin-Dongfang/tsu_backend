// File: cmd/login-shim/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tsu-self/internal/app/login-shim/adapter"
	"tsu-self/internal/app/login-shim/controller"
	"tsu-self/internal/middleware"
	pkglog "tsu-self/internal/pkg/log"
)

func main() {
	// 初始化日志系统
	pkglog.Init(slog.LevelInfo)

	// 初始化 Kratos 适配器
	kratosURL := getEnv("KRATOS_PUBLIC_URL", "http://kratos:4433")
	kratosAdapter := adapter.NewKratosAdapter(kratosURL)

	// 初始化控制器
	authController := controller.NewAuthController(kratosAdapter)

	// 设置路由
	mux := setupRoutes(authController)

	// 应用中间件
	handler := applyMiddleware(mux)

	// 启动服务器
	server := &http.Server{
		Addr:         ":8090",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 优雅关闭
	go func() {
		log.Println("Login Shim 服务启动在 :8090")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败: %v", err)
	}
	log.Println("服务已安全关闭")
}

// setupRoutes 设置路由
func setupRoutes(authController *controller.AuthController) *http.ServeMux {
	mux := http.NewServeMux()

	// 认证相关路由
	mux.HandleFunc("POST /auth/login", authController.Login)
	mux.HandleFunc("POST /auth/register", authController.Register)

	// 健康检查
	mux.HandleFunc("GET /health", authController.HealthCheck)
	mux.HandleFunc("GET /api/health", authController.HealthCheck)

	return mux
}

// applyMiddleware 应用中间件
func applyMiddleware(handler http.Handler) http.Handler {
	// 中间件应用顺序很重要
	handler = middleware.CORS()(handler)
	handler = middleware.TraceID(handler)
	handler = middleware.Recovery()(handler)
	return handler
}

// getEnv 获取环境变量，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
