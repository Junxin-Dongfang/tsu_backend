package nats

import (
	"context"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// HealthChecker NATS连接健康检查器
type HealthChecker struct {
	conn      *nats.Conn
	isHealthy bool
	mutex     sync.RWMutex
	stopCh    chan struct{}
	interval  time.Duration
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(conn *nats.Conn, checkInterval time.Duration) *HealthChecker {
	if checkInterval <= 0 {
		checkInterval = 10 * time.Second // 默认10秒检查一次
	}

	return &HealthChecker{
		conn:      conn,
		isHealthy: true,
		stopCh:    make(chan struct{}),
		interval:  checkInterval,
	}
}

// Start 启动健康检查
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.checkHealth()
		}
	}
}

// Stop 停止健康检查
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// IsHealthy 检查连接是否健康
func (hc *HealthChecker) IsHealthy() bool {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.isHealthy
}

// checkHealth 执行健康检查
func (hc *HealthChecker) checkHealth() {
	healthy := hc.conn.IsConnected() && !hc.conn.IsClosed()

	hc.mutex.Lock()
	hc.isHealthy = healthy
	hc.mutex.Unlock()
}

// WaitForHealthy 等待连接恢复健康
func (hc *HealthChecker) WaitForHealthy(ctx context.Context, maxWait time.Duration) bool {
	ctx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if hc.IsHealthy() {
			return true
		}

		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			// 继续检查
		}
	}
}
