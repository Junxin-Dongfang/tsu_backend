// File: internal/pkg/metrics/error_metrics.go
package metrics

import (
	"strconv"

	"tsu-self/internal/pkg/xerrors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ErrorMetrics 错误监控指标
type ErrorMetrics struct {
	// 错误总数（按错误码）
	ErrorsByCode *prometheus.CounterVec

	// 错误总数（按分类）
	ErrorsByCategory *prometheus.CounterVec

	// 错误总数（按级别）
	ErrorsByLevel *prometheus.CounterVec

	// HTTP 响应总数（按状态码）
	HTTPResponses *prometheus.CounterVec

	// 可重试错误计数
	RetryableErrors prometheus.Counter

	// 严重错误计数
	CriticalErrors prometheus.Counter

	// 错误响应延迟（直方图）
	ErrorResponseDuration *prometheus.HistogramVec
}

var (
	// DefaultErrorMetrics 默认的错误指标实例
	DefaultErrorMetrics *ErrorMetrics
)

// init 初始化默认指标
func init() {
	DefaultErrorMetrics = NewErrorMetrics("tsu")
}

// NewErrorMetrics 创建新的错误指标收集器
func NewErrorMetrics(namespace string) *ErrorMetrics {
	return &ErrorMetrics{
		ErrorsByCode: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors by error code",
			},
			[]string{"code", "message"},
		),

		ErrorsByCategory: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_by_category_total",
				Help:      "Total number of errors by category (system, authentication, authorization, etc.)",
			},
			[]string{"category"},
		),

		ErrorsByLevel: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_by_level_total",
				Help:      "Total number of errors by level (info, warn, error, critical)",
			},
			[]string{"level"},
		),

		HTTPResponses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_responses_total",
				Help:      "Total number of HTTP responses by status code",
			},
			[]string{"status_code", "method"},
		),

		RetryableErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "retryable_errors_total",
				Help:      "Total number of retryable errors",
			},
		),

		CriticalErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "critical_errors_total",
				Help:      "Total number of critical errors",
			},
		),

		ErrorResponseDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "error_response_duration_seconds",
				Help:      "Error response duration in seconds",
				Buckets:   prometheus.DefBuckets, // 默认 bucket: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
			},
			[]string{"code", "level"},
		),
	}
}

// RecordError 记录错误指标
func (m *ErrorMetrics) RecordError(appErr *xerrors.AppError, statusCode int, method string, duration float64) {
	if appErr == nil {
		return
	}

	// 记录错误码
	m.ErrorsByCode.WithLabelValues(
		strconv.Itoa(appErr.Code.ToInt()),
		appErr.Message,
	).Inc()

	// 记录错误分类
	if appErr.Category != "" {
		m.ErrorsByCategory.WithLabelValues(appErr.Category).Inc()
	}

	// 记录错误级别
	m.ErrorsByLevel.WithLabelValues(appErr.Level.String()).Inc()

	// 记录 HTTP 状态码
	m.HTTPResponses.WithLabelValues(
		strconv.Itoa(statusCode),
		method,
	).Inc()

	// 记录可重试错误
	if appErr.IsRetryable() {
		m.RetryableErrors.Inc()
	}

	// 记录严重错误
	if appErr.IsCritical() {
		m.CriticalErrors.Inc()
	}

	// 记录响应延迟
	if duration > 0 {
		m.ErrorResponseDuration.WithLabelValues(
			strconv.Itoa(appErr.Code.ToInt()),
			appErr.Level.String(),
		).Observe(duration)
	}
}

// RecordHTTPResponse 记录 HTTP 响应指标（成功响应）
func (m *ErrorMetrics) RecordHTTPResponse(statusCode int, method string) {
	m.HTTPResponses.WithLabelValues(
		strconv.Itoa(statusCode),
		method,
	).Inc()
}
