// File: internal/pkg/metrics/resource_metrics_test.go
package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

// TestNewResourceMetrics 测试创建资源指标收集器
func TestNewResourceMetrics(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
	}{
		{
			name:      "创建资源指标收集器 - 成功",
			namespace: "test",
		},
		{
			name:      "创建资源指标收集器 - 空 namespace",
			namespace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			metrics := NewResourceMetricsWithRegistry(tt.namespace, reg)

			assert.NotNil(t, metrics)
			assert.NotNil(t, metrics.DBConnections)
			assert.NotNil(t, metrics.RedisOperations)
			assert.NotNil(t, metrics.RedisOperationDuration)
		})
	}
}

// TestResourceMetrics_RecordDBPoolStats 测试数据库连接池统计记录
func TestResourceMetrics_RecordDBPoolStats(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewResourceMetricsWithRegistry("test", reg)
	service := "game"

	// 记录数据库连接池统计信息
	metrics.RecordDBPoolStats(
		service,
		"postgres",
		10,            // openConnections
		5,             // inUse
		5,             // idle
		20,            // maxOpen
		100,           // waitCount
		1*time.Second, // waitDuration
	)

	// 验证 open connections
	openConns := testutil.ToFloat64(metrics.DBConnections.WithLabelValues(service, "postgres", "open"))
	assert.Equal(t, float64(10), openConns)

	// 验证 in_use connections
	inUseConns := testutil.ToFloat64(metrics.DBConnections.WithLabelValues(service, "postgres", "in_use"))
	assert.Equal(t, float64(5), inUseConns)

	// 验证 idle connections
	idleConns := testutil.ToFloat64(metrics.DBConnections.WithLabelValues(service, "postgres", "idle"))
	assert.Equal(t, float64(5), idleConns)

	// 验证 max connections
	maxConns := testutil.ToFloat64(metrics.DBMaxConnections.WithLabelValues(service, "postgres"))
	assert.Equal(t, float64(20), maxConns)

	// 验证 wait count
	waitCount := testutil.ToFloat64(metrics.DBWaitCount.WithLabelValues(service, "postgres"))
	assert.Equal(t, float64(100), waitCount)

	// 验证 wait duration
	waitDuration := testutil.ToFloat64(metrics.DBWaitDuration.WithLabelValues(service, "postgres"))
	assert.Equal(t, float64(1), waitDuration)
}

// TestResourceMetrics_RecordRedisOperation 测试 Redis 操作记录
func TestResourceMetrics_RecordRedisOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		success   bool
		duration  time.Duration
		service   string
	}{
		{
			name:      "记录成功的 GET 操作",
			operation: "GET",
			success:   true,
			duration:  10 * time.Millisecond,
			service:   "game",
		},
		{
			name:      "记录失败的 SET 操作",
			operation: "SET",
			success:   false,
			duration:  5 * time.Millisecond,
			service:   "admin",
		},
		{
			name:      "记录成功的 HGET 操作",
			operation: "HGET",
			success:   true,
			duration:  2 * time.Millisecond,
			service:   "game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			metrics := NewResourceMetricsWithRegistry("test", reg)

			// 记录 Redis 操作
			metrics.RecordRedisOperation(tt.operation, tt.success, tt.duration, tt.service)

			// 验证操作计数
			result := "success"
			if !tt.success {
				result = "error"
			}
			count := testutil.ToFloat64(metrics.RedisOperations.WithLabelValues(tt.operation, result, tt.service))
			assert.Equal(t, float64(1), count)

			// 验证 histogram 已创建
			assert.NotNil(t, metrics.RedisOperationDuration)
		})
	}
}

// TestResourceMetrics_RecordRedisError 测试 Redis 错误记录
func TestResourceMetrics_RecordRedisError(t *testing.T) {
	tests := []struct {
		name      string
		errorType string
		service   string
	}{
		{
			name:      "记录超时错误",
			errorType: "timeout",
			service:   "game",
		},
		{
			name:      "记录连接错误",
			errorType: "connection_error",
			service:   "admin",
		},
		{
			name:      "记录 nil 值错误",
			errorType: "nil",
			service:   "game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			metrics := NewResourceMetricsWithRegistry("test", reg)

			// 记录 Redis 错误
			metrics.RecordRedisError(tt.errorType, tt.service)

			// 验证错误计数
			count := testutil.ToFloat64(metrics.RedisErrors.WithLabelValues(tt.errorType, tt.service))
			assert.Equal(t, float64(1), count)
		})
	}
}

// TestResourceMetrics_RecordRedisPoolStats 测试 Redis 连接池统计记录
func TestResourceMetrics_RecordRedisPoolStats(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewResourceMetricsWithRegistry("test", reg)

	// 记录 Redis 连接池统计
	totalConns := 10
	idleConns := 3
	staleConns := 1
	service := "game"

	metrics.RecordRedisPoolStats(totalConns, idleConns, staleConns, service)

	// 验证 total connections
	total := testutil.ToFloat64(metrics.RedisConnectionPool.WithLabelValues("total", service))
	assert.Equal(t, float64(10), total)

	// 验证 idle connections
	idle := testutil.ToFloat64(metrics.RedisConnectionPool.WithLabelValues("idle", service))
	assert.Equal(t, float64(3), idle)

	// 验证 stale connections
	stale := testutil.ToFloat64(metrics.RedisConnectionPool.WithLabelValues("stale", service))
	assert.Equal(t, float64(1), stale)

	// 验证 active connections (total - idle)
	active := testutil.ToFloat64(metrics.RedisConnectionPool.WithLabelValues("active", service))
	assert.Equal(t, float64(7), active)
}

// TestResourceMetrics_RedisOperationBuckets 验证 Redis 操作 Histogram Buckets
func TestResourceMetrics_RedisOperationBuckets(t *testing.T) {
	expectedBuckets := []float64{
		0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5,
	}

	assert.Equal(t, expectedBuckets, RedisOperationBuckets, "Redis 操作 buckets 应该针对快速操作优化")
}

// TestResourceMetrics_MultipleOperations 测试多个操作的累积记录
func TestResourceMetrics_MultipleOperations(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewResourceMetricsWithRegistry("test", reg)

	// 记录多个 GET 操作
	for i := 0; i < 5; i++ {
		metrics.RecordRedisOperation("GET", true, 10*time.Millisecond, "game")
	}

	// 记录多个 SET 操作（部分失败）
	for i := 0; i < 3; i++ {
		metrics.RecordRedisOperation("SET", true, 5*time.Millisecond, "game")
	}
	metrics.RecordRedisOperation("SET", false, 5*time.Millisecond, "game")

	// 验证 GET 操作计数
	getCount := testutil.ToFloat64(metrics.RedisOperations.WithLabelValues("GET", "success", "game"))
	assert.Equal(t, float64(5), getCount)

	// 验证成功的 SET 操作计数
	setSuccessCount := testutil.ToFloat64(metrics.RedisOperations.WithLabelValues("SET", "success", "game"))
	assert.Equal(t, float64(3), setSuccessCount)

	// 验证失败的 SET 操作计数
	setErrorCount := testutil.ToFloat64(metrics.RedisOperations.WithLabelValues("SET", "error", "game"))
	assert.Equal(t, float64(1), setErrorCount)
}

// TestResourceMetrics_ConcurrentSafety 测试并发安全性
func TestResourceMetrics_ConcurrentSafety(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewResourceMetricsWithRegistry("test", reg)

	numGoroutines := 10
	numIterations := 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numIterations; j++ {
				metrics.RecordRedisOperation("GET", true, 10*time.Millisecond, "game")
				metrics.RecordRedisError("timeout", "game")
				metrics.RecordRedisPoolStats(10, 5, 1, "game")
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证并发操作后的计数
	count := testutil.ToFloat64(metrics.RedisOperations.WithLabelValues("GET", "success", "game"))
	expectedCount := float64(numGoroutines * numIterations)
	assert.Equal(t, expectedCount, count)
}
