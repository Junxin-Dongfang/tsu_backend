// File: internal/pkg/metrics/error_metrics.go
package metrics

import (
	"strconv"
	"strings"

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
	RetryableErrors *prometheus.CounterVec

	// 严重错误计数
	CriticalErrors *prometheus.CounterVec

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
	return NewErrorMetricsWithRegistry(namespace, GetRegisterer())
}

// NewErrorMetricsWithRegistry 创建新的错误指标收集器（使用自定义注册表）
func NewErrorMetricsWithRegistry(namespace string, registerer prometheus.Registerer) *ErrorMetrics {
	factory := promauto.With(registerer)

	return &ErrorMetrics{
		ErrorsByCode: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors by error code",
			},
			[]string{"service", "method", "code", "category", "level"},
		),

		ErrorsByCategory: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_by_category_total",
				Help:      "Total number of errors by category (system, authentication, authorization, etc.)",
			},
			[]string{"service", "category"},
		),

		ErrorsByLevel: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_by_level_total",
				Help:      "Total number of errors by level (info, warn, error, critical)",
			},
			[]string{"service", "level"},
		),

		HTTPResponses: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_responses_total",
				Help:      "Total number of HTTP responses by status code",
			},
			[]string{"service", "status_code", "method"},
		),

		RetryableErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "retryable_errors_total",
				Help:      "Total number of retryable errors",
			},
			[]string{"service", "method", "code", "is_retryable"},
		),

		CriticalErrors: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "critical_errors_total",
				Help:      "Total number of critical errors",
			},
			[]string{"service", "method", "code", "is_critical"},
		),

		ErrorResponseDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "error_response_duration_seconds",
				Help:      "Error response duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "method", "code", "level"},
		),
	}
}

// RecordError 记录错误指标
func (m *ErrorMetrics) RecordError(appErr *xerrors.AppError, statusCode int, method, service string, duration float64) {
	if appErr == nil {
		return
	}

	service = normalizeServiceName(service)
	if method == "" {
		method = "UNKNOWN"
	} else {
		method = strings.ToUpper(method)
	}

	code := strconv.Itoa(appErr.Code.ToInt())
	category := appErr.Category
	level := appErr.Level.String()

	// 记录错误码
	m.ErrorsByCode.WithLabelValues(
		service,
		method,
		code,
		category,
		level,
	).Inc()

	// 记录错误分类
	if appErr.Category != "" {
		m.ErrorsByCategory.WithLabelValues(service, appErr.Category).Inc()
	}

	// 记录错误级别
	m.ErrorsByLevel.WithLabelValues(service, level).Inc()

	// 记录 HTTP 状态码
	m.HTTPResponses.WithLabelValues(
		service,
		strconv.Itoa(statusCode),
		method,
	).Inc()

	// 记录可重试错误
	m.RetryableErrors.WithLabelValues(
		service,
		method,
		code,
		strconv.FormatBool(appErr.IsRetryable()),
	).Inc()

	// 记录严重错误
	m.CriticalErrors.WithLabelValues(
		service,
		method,
		code,
		strconv.FormatBool(appErr.IsCritical()),
	).Inc()

	// 记录响应延迟
	if duration > 0 {
		m.ErrorResponseDuration.WithLabelValues(
			service,
			method,
			code,
			level,
		).Observe(duration)
	}
}

// RecordHTTPResponse 记录 HTTP 响应指标（成功响应）
func (m *ErrorMetrics) RecordHTTPResponse(statusCode int, method, service string) {
	service = normalizeServiceName(service)
	if method == "" {
		method = "UNKNOWN"
	} else {
		method = strings.ToUpper(method)
	}

	m.HTTPResponses.WithLabelValues(
		service,
		strconv.Itoa(statusCode),
		method,
	).Inc()
}
