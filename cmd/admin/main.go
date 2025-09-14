package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 这个 API 端点将被 Oathkeeper 保护
	http.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		// 读取 Oathkeeper 注入的 header
		userInfo := r.Header.Get("X-Userinfo")
		if userInfo == "" {
			userInfo = "anonymous (header not found)"
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from protected API!", "user": "%s"}`, userInfo)
	})

    // 一个公开的健康检查端点，用于测试服务是否启动
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, `{"status": "ok"}`)
    })

	log.Println("Admin API server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}