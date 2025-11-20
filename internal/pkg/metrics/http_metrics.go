// File: internal/pkg/metrics/http_metrics.go
package metrics

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPMetrics HTTP 性能指标收集器
type HTTPMetrics struct {
	// HTTP 请求总数（按路由模板、方法、状态码分组）
	RequestsTotal *prometheus.CounterVec

	// HTTP 请求延迟直方图（按路由模板分组）
	RequestDuration *prometheus.HistogramVec

	// 当前进行中的请求数（Gauge 类型）
	RequestsInProgress *prometheus.GaugeVec
}

var (
	// DefaultHTTPMetrics 默认的 HTTP 指标实例
	DefaultHTTPMetrics *HTTPMetrics
)

// HTTPBuckets 是针对 HTTP 请求延迟优化的 buckets
// 基于 SLO: p95 < 200ms (0.2s)
// 单位：秒
var HTTPBuckets = []float64{
	0.05, // 50ms
	0.1,  // 100ms
	0.15, // 150ms
	0.2,  // 200ms - SLO 边界
	0.3,  // 300ms
	0.5,  // 500ms
	1,    // 1s
	2,    // 2s
	5,    // 5s
}

// init 初始化默认指标
func init() {
	DefaultHTTPMetrics = NewHTTPMetrics("tsu")
}

// NewHTTPMetrics 创建新的 HTTP 指标收集器
func NewHTTPMetrics(namespace string) *HTTPMetrics {
	return NewHTTPMetricsWithRegistry(namespace, GetRegisterer())
}

// NewHTTPMetricsWithRegistry 创建新的 HTTP 指标收集器（使用自定义注册表）
func NewHTTPMetricsWithRegistry(namespace string, registerer prometheus.Registerer) *HTTPMetrics {
	factory := promauto.With(registerer)

	return &HTTPMetrics{
		RequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests by service, route template, method, and status code",
			},
			[]string{"service", "route", "method", "status_code"},
		),

		RequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency histogram by service and route template",
				Buckets:   HTTPBuckets,
			},
			[]string{"service", "route"},
		),

		RequestsInProgress: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_progress",
				Help:      "Current number of HTTP requests being processed by service",
			},
			[]string{"service"},
		),
	}
}

// RecordRequest 记录 HTTP 请求指标
//
// 参数:
//   - route: 路由模板（如 "/api/heroes/:id"，而非 "/api/heroes/123"）
//   - method: HTTP 方法（GET/POST/PUT/DELETE 等）
//   - statusCode: HTTP 状态码（200/404/500 等）
//   - duration: 请求耗时
//   - service: 服务名称 (game/admin)
func (m *HTTPMetrics) RecordRequest(service, route, method string, statusCode int, duration time.Duration) {
	service = normalizeServiceName(service)
	// 记录请求总数
	statusCodeLabel := strconv.Itoa(statusCode)
	m.RequestsTotal.WithLabelValues(service, route, method, statusCodeLabel).Inc()

	// 记录请求延迟
	m.RequestDuration.WithLabelValues(service, route).Observe(duration.Seconds())
}

// IncInProgress 增加当前进行中的请求数
func (m *HTTPMetrics) IncInProgress(service string) {
	service = normalizeServiceName(service)
	m.RequestsInProgress.WithLabelValues(service).Inc()
}

// DecInProgress 减少当前进行中的请求数
func (m *HTTPMetrics) DecInProgress(service string) {
	service = normalizeServiceName(service)
	m.RequestsInProgress.WithLabelValues(service).Dec()
}

// IsHealthCheckEndpoint 判断是否为健康检查端点
// 这些端点不应被监控，以避免指标噪音
func IsHealthCheckEndpoint(path string) bool {
	healthCheckPaths := []string{
		"/metrics",
		"/health",
		"/healthz",
		"/readyz",
		"/livez",
	}

	for _, hc := range healthCheckPaths {
		if path == hc {
			return true
		}
	}
	return false
}

// NormalizeRoute 规范化路由，防止标签基数爆炸。
func NormalizeRoute(route string) string {
	if route == "" {
		return "unknown"
	}
	return route
}

// PathLimitTracker 路径标签基数限制追踪器
type PathLimitTracker struct {
	mu       sync.RWMutex
	paths    map[string]struct{}
	maxPaths int
}

// NewPathLimitTracker 创建路径限制追踪器
func NewPathLimitTracker(maxPaths int) *PathLimitTracker {
	return &PathLimitTracker{
		paths:    make(map[string]struct{}),
		maxPaths: maxPaths,
	}
}

// TrackPath 追踪路径，如果超出限制返回 "other"
func (t *PathLimitTracker) TrackPath(path string) string {
	if path == "" {
		return "unknown"
	}

	t.mu.RLock()
	if _, exists := t.paths[path]; exists {
		t.mu.RUnlock()
		return path
	}
	// 检查是否超出限制
	if len(t.paths) >= t.maxPaths {
		t.mu.RUnlock()
		return "other"
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双重检查，防止并发情况下重复添加
	if _, exists := t.paths[path]; exists {
		return path
	}
	if len(t.paths) >= t.maxPaths {
		return "other"
	}

	// 添加到追踪列表
	t.paths[path] = struct{}{}
	return path
}

// GetTrackedCount 获取已追踪的路径数量
func (t *PathLimitTracker) GetTrackedCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.paths)
}

// LogWarning 如果接近限制，记录警告日志
func (t *PathLimitTracker) LogWarning() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	count := len(t.paths)
	if count >= t.maxPaths*9/10 { // 90% 阈值
		return fmt.Sprintf("WARNING: Path label cardinality approaching limit (%d/%d)", count, t.maxPaths)
	}
	return ""
}
