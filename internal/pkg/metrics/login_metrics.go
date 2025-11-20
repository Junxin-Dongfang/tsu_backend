package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// LoginMetrics 追踪登录链路的核心指标。
type LoginMetrics struct {
	Duration   *prometheus.HistogramVec
	CacheHit   *prometheus.CounterVec
	CacheMiss  *prometheus.CounterVec
	CacheEvict *prometheus.CounterVec
}

var (
	// DefaultLoginMetrics 全局共享实例。
	DefaultLoginMetrics *LoginMetrics

	loginDurationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.2, 0.3, 0.5, 1, 2}
)

func init() {
	DefaultLoginMetrics = NewLoginMetrics("tsu")
}

// NewLoginMetricsWithRegistry 创建 LoginMetrics,允许 tests 注入自定义 registry。
func NewLoginMetricsWithRegistry(namespace string, reg prometheus.Registerer) *LoginMetrics {
	if reg == nil {
		reg = GetRegisterer()
	}
	factory := promauto.With(reg)

	return &LoginMetrics{
		Duration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "login_duration_seconds",
				Help:      "Latency histogram for admin/game login endpoints",
				Buckets:   loginDurationBuckets,
			},
			[]string{"service", "outcome"},
		),

		CacheHit: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "login_cache_hits_total",
				Help:      "Count of login cache hits by service",
			},
			[]string{"service"},
		),

		CacheMiss: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "login_cache_miss_total",
				Help:      "Count of login cache misses by service",
			},
			[]string{"service"},
		),

		CacheEvict: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "login_cache_evict_total",
				Help:      "Count of login cache evictions grouped by service and reason",
			},
			[]string{"service", "reason"},
		),
	}
}

// NewLoginMetrics 创建默认 registry 的 LoginMetrics。
func NewLoginMetrics(namespace string) *LoginMetrics {
	return NewLoginMetricsWithRegistry(namespace, GetRegisterer())
}

// ObserveDuration 记录登录耗时。
func (m *LoginMetrics) ObserveDuration(service, outcome string, duration time.Duration) {
	if m == nil {
		return
	}
	service = normalizeServiceName(service)
	if outcome == "" {
		outcome = "success"
	}
	m.Duration.WithLabelValues(service, outcome).Observe(duration.Seconds())
}

// IncCacheHit 增加缓存命中次数。
func (m *LoginMetrics) IncCacheHit(service string) {
	if m == nil {
		return
	}
	m.CacheHit.WithLabelValues(normalizeServiceName(service)).Inc()
}

// IncCacheMiss 增加缓存未命中次数。
func (m *LoginMetrics) IncCacheMiss(service string) {
	if m == nil {
		return
	}
	m.CacheMiss.WithLabelValues(normalizeServiceName(service)).Inc()
}

// IncCacheEvicted 记录缓存剔除次数。
func (m *LoginMetrics) IncCacheEvicted(service, reason string) {
	if m == nil {
		return
	}
	if reason == "" {
		reason = "unknown"
	}
	m.CacheEvict.WithLabelValues(normalizeServiceName(service), reason).Inc()
}
