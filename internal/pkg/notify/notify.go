package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

var (
	ncMu sync.RWMutex
	nc   *nats.Conn
)

// SetNatsConn 设置全局 NATS 连接（由 main 提供）
func SetNatsConn(conn *nats.Conn) {
	ncMu.Lock()
	defer ncMu.Unlock()
	nc = conn
}

// PublishWarehouseEvent 发布仓库相关事件
func PublishWarehouseEvent(ctx context.Context, subject string, payload interface{}) error {
	ncMu.RLock()
	conn := nc
	ncMu.RUnlock()
	if conn == nil {
		return nil // 没有连接时静默降级
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal warehouse event failed: %w", err)
	}
	return conn.Publish(subject, data)
}

// Default subjects
const (
	SubjectWarehouseDistributed = "warehouse.distribution"
	SubjectWarehouseLoot        = "warehouse.loot"
)
