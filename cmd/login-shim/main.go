// File: cmd/login-shim/main.go
package main

import (
	"log/slog"
	"net/http"

	// 引入服务内部的包
	"tsu-self/cmd/login-shim/internal/adapter"
	"tsu-self/cmd/login-shim/internal/controller"

	// 引入项目级的共享包
	"tsu-self/internal/middleware"
	"tsu-self/internal/pkg/log"
)

func main() {
	// --- 1. 初始化 ---
	// 在程序的最开始初始化日志系统
	log.Init(slog.LevelDebug)

	// --- 2. 依赖注入 (DI) ---
	// 创建最底层的依赖：Kratos 适配器
	kratosAdapter := adapter.NewKratosAdapter("http://kratos:4433")

	// 创建控制器，并将 adapter 注入进去
	authController := controller.NewAuthController(kratosAdapter)

	// --- 3. 路由和中间件 ---
	// 创建一个新的 ServeMux 作为我们的主路由器
	mux := http.NewServeMux()

	// 定义认证相关的路由，并将请求指向我们刚刚创建的 controller 的方法
	mux.HandleFunc("POST /auth/login", authController.Login)
	mux.HandleFunc("POST /auth/register", authController.Register)

	// 定义一个简单的健康检查路由
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// --- 4. 包裹中间件 ---
	// 将我们的主路由器包裹在中间件中，注意包裹顺序
	// TraceID 在最外层，这样内层的 Logger 才能获取到 trace_id
	var handler http.Handler = mux
	handler = middleware.Logger(handler)  // Logger 中间件在内层
	handler = middleware.TraceID(handler) // TraceID 中间件在外层

	log.Info("Login API Shim service is starting...", "address", ":8090") // context.TODO() used at startup

	server := &http.Server{
		Addr:    ":8090",
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("Login API Shim service failed to start", err)
	}
}
