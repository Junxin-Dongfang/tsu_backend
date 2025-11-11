// File: internal/pkg/metrics/resource_metrics.go
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ResourceMetrics 系统资源监控指标收集器
type ResourceMetrics struct {
	// 数据库连接池指标
	DBConnections     *prometheus.GaugeVec   // 当前连接数
	DBMaxConnections  *prometheus.GaugeVec   // 最大连接数
	DBIdleConnections *prometheus.GaugeVec   // 空闲连接数
	DBWaitCount       *prometheus.CounterVec // 等待连接的总次数
	DBWaitDuration    *prometheus.CounterVec // 等待连接的总时长

	// Redis 操作指标
	RedisOperations        *prometheus.CounterVec   // Redis 操作总数（按操作类型和结果）
	RedisOperationDuration *prometheus.HistogramVec // Redis 操作延迟（按操作类型）
	RedisConnectionPool    *prometheus.GaugeVec     // Redis 连接池状态
	RedisErrors            *prometheus.CounterVec   // Redis 错误数（按错误类型）
}

var (
	// DefaultResourceMetrics 默认的资源指标实例
	DefaultResourceMetrics *ResourceMetrics
)

// RedisOperationBuckets 是针对 Redis 操作延迟优化的 buckets
// Redis 操作通常非常快，使用更细粒度的 buckets
// 单位：秒
var RedisOperationBuckets = []float64{
	0.001, // 1ms
	0.005, // 5ms
	0.01,  // 10ms
	0.025, // 25ms
	0.05,  // 50ms
	0.1,   // 100ms
	0.25,  // 250ms
	0.5,   // 500ms
	1,     // 1s
	2.5,   // 2.5s
}

// init 初始化默认指标
func init() {
	DefaultResourceMetrics = NewResourceMetrics("tsu")
}

// NewResourceMetrics 创建新的资源指标收集器
func NewResourceMetrics(namespace string) *ResourceMetrics {
	return NewResourceMetricsWithRegistry(namespace, GetRegisterer())
}

// NewResourceMetricsWithRegistry 创建新的资源指标收集器（使用自定义注册表）
func NewResourceMetricsWithRegistry(namespace string, registerer prometheus.Registerer) *ResourceMetrics {
	factory := promauto.With(registerer)

	return &ResourceMetrics{
		// 数据库连接池指标
		DBConnections: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "connections",
				Help:      "Current number of database connections by state (open/in_use/idle)",
			},
			[]string{"service", "database", "state"},
		),

		DBMaxConnections: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "max_connections",
				Help:      "Maximum number of database connections allowed",
			},
			[]string{"service", "database"},
		),

		DBIdleConnections: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "idle_connections",
				Help:      "Number of idle database connections",
			},
			[]string{"service", "database"},
		),

		DBWaitCount: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "wait_count_total",
				Help:      "Total number of connections waited for",
			},
			[]string{"service", "database"},
		),

		DBWaitDuration: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "wait_duration_seconds_total",
				Help:      "Total time blocked waiting for a new connection",
			},
			[]string{"service", "database"},
		),

		// Redis 操作指标
		RedisOperations: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "operations_total",
				Help:      "Total number of Redis operations by type and result (success/error)",
			},
			[]string{"operation", "result", "service"},
		),

		RedisOperationDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "operation_duration_seconds",
				Help:      "Redis operation duration in seconds by operation type",
				Buckets:   RedisOperationBuckets,
			},
			[]string{"operation", "service"},
		),

		RedisConnectionPool: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "connection_pool",
				Help:      "Redis connection pool status (total/idle/stale/active)",
			},
			[]string{"state", "service"},
		),

		RedisErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "errors_total",
				Help:      "Total number of Redis errors by type",
			},
			[]string{"error_type", "service"},
		),
	}
}

// RecordDBPoolStats 记录数据库连接池统计信息
//
// 参数:
//   - database: 数据库名称（如 "postgres"）
//   - openConnections: 当前打开的连接数
//   - inUse: 正在使用的连接数
//   - idle: 空闲连接数
//   - maxOpen: 最大连接数
//   - waitCount: 等待连接的总次数
//   - waitDuration: 等待连接的总时长
func (m *ResourceMetrics) RecordDBPoolStats(
	service string,
	database string,
	openConnections, inUse, idle int,
	maxOpen int,
	waitCount int64,
	waitDuration time.Duration,
) {
	service = normalizeServiceName(service)
	m.DBConnections.WithLabelValues(service, database, "open").Set(float64(openConnections))
	m.DBConnections.WithLabelValues(service, database, "in_use").Set(float64(inUse))
	m.DBConnections.WithLabelValues(service, database, "idle").Set(float64(idle))
	m.DBMaxConnections.WithLabelValues(service, database).Set(float64(maxOpen))
	m.DBIdleConnections.WithLabelValues(service, database).Set(float64(idle))
	m.DBWaitCount.WithLabelValues(service, database).Add(float64(waitCount))
	m.DBWaitDuration.WithLabelValues(service, database).Add(waitDuration.Seconds())
}

// RecordRedisOperation 记录 Redis 操作指标
//
// 参数:
//   - operation: 操作类型（如 "GET", "SET", "HGET", "HSET", "ZADD" 等）
//   - success: 操作是否成功
//   - duration: 操作耗时
//   - service: 服务名称（"game" 或 "admin"）
func (m *ResourceMetrics) RecordRedisOperation(operation string, success bool, duration time.Duration, service string) {
	service = normalizeServiceName(service)
	result := "success"
	if !success {
		result = "error"
	}

	m.RedisOperations.WithLabelValues(operation, result, service).Inc()
	m.RedisOperationDuration.WithLabelValues(operation, service).Observe(duration.Seconds())
}

// RecordRedisError 记录 Redis 错误
//
// 参数:
//   - errorType: 错误类型（如 "timeout", "connection_error", "nil" 等）
//   - service: 服务名称
func (m *ResourceMetrics) RecordRedisError(errorType, service string) {
	service = normalizeServiceName(service)
	m.RedisErrors.WithLabelValues(errorType, service).Inc()
}

// RecordRedisPoolStats 记录 Redis 连接池统计信息
//
// 参数:
//   - totalConns: 总连接数
//   - idleConns: 空闲连接数
//   - staleConns: 过期连接数
//   - service: 服务名称
func (m *ResourceMetrics) RecordRedisPoolStats(totalConns, idleConns, staleConns int, service string) {
	service = normalizeServiceName(service)
	m.RedisConnectionPool.WithLabelValues("total", service).Set(float64(totalConns))
	m.RedisConnectionPool.WithLabelValues("idle", service).Set(float64(idleConns))
	m.RedisConnectionPool.WithLabelValues("stale", service).Set(float64(staleConns))

	// 计算活跃连接数
	activeConns := totalConns - idleConns
	m.RedisConnectionPool.WithLabelValues("active", service).Set(float64(activeConns))
}
