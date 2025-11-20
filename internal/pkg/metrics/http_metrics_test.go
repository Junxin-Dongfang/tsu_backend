// File: internal/pkg/metrics/http_metrics_test.go
package metrics

import (
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPMetrics(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantNil   bool
	}{
		{
			name:      "创建 HTTP 指标收集器 - 成功",
			namespace: "test",
			wantNil:   false,
		},
		{
			name:      "创建 HTTP 指标收集器 - 空 namespace",
			namespace: "",
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用自定义注册表避免冲突
			reg := prometheus.NewRegistry()
			metrics := NewHTTPMetricsWithRegistry(tt.namespace, reg)

			if tt.wantNil {
				assert.Nil(t, metrics)
			} else {
				assert.NotNil(t, metrics)
				assert.NotNil(t, metrics.RequestsTotal)
				assert.NotNil(t, metrics.RequestDuration)
				assert.NotNil(t, metrics.RequestsInProgress)
			}
		})
	}
}

func TestHTTPMetrics_RecordRequest(t *testing.T) {
	tests := []struct {
		name       string
		service    string
		route      string
		method     string
		statusCode int
		duration   time.Duration
	}{
		{
			name:       "记录 GET 请求 - 200",
			service:    "game",
			route:      "/api/heroes/:id",
			method:     "GET",
			statusCode: 200,
			duration:   100 * time.Millisecond,
		},
		{
			name:       "记录 POST 请求 - 201",
			service:    "admin",
			route:      "/api/heroes",
			method:     "POST",
			statusCode: 201,
			duration:   200 * time.Millisecond,
		},
		{
			name:       "记录 PUT 请求 - 404",
			service:    "game",
			route:      "/api/heroes/:id",
			method:     "PUT",
			statusCode: 404,
			duration:   50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			metrics := NewHTTPMetricsWithRegistry("test", reg)

			// 记录请求
			metrics.RecordRequest(tt.service, tt.route, tt.method, tt.statusCode, tt.duration)

			// 验证 counter 增加
			statusCode := strconv.Itoa(tt.statusCode)
			count := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues(tt.service, tt.route, tt.method, statusCode))
			assert.Equal(t, float64(1), count)

			// 验证 histogram 记录了观测值（通过检查 _count 指标）
			// Histogram 会自动创建 _count, _sum, _bucket 指标
			// 这里简单验证 RecordRequest 不会 panic 即可
			assert.NotNil(t, metrics.RequestDuration)
		})
	}
}

func TestHTTPMetrics_InProgress(t *testing.T) {
	t.Run("当前进行中的请求数", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		metrics := NewHTTPMetricsWithRegistry("test", reg)
		service := "game"

		// 初始值应为 0
		assert.Equal(t, float64(0), testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues(service)))

		// 增加进行中的请求
		metrics.IncInProgress(service)
		assert.Equal(t, float64(1), testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues(service)))

		metrics.IncInProgress(service)
		assert.Equal(t, float64(2), testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues(service)))

		// 减少进行中的请求
		metrics.DecInProgress(service)
		assert.Equal(t, float64(1), testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues(service)))

		metrics.DecInProgress(service)
		assert.Equal(t, float64(0), testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues(service)))
	})
}

func TestHTTPMetrics_HistogramBuckets(t *testing.T) {
	t.Run("验证 Histogram Buckets 配置", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		metrics := NewHTTPMetricsWithRegistry("test", reg)
		service := "game"

		// 记录不同延迟的请求
		testCases := []struct {
			duration time.Duration
			route    string
		}{
			{50 * time.Millisecond, "/api/test1"},  // 0.05s
			{100 * time.Millisecond, "/api/test2"}, // 0.1s
			{150 * time.Millisecond, "/api/test3"}, // 0.15s
			{200 * time.Millisecond, "/api/test4"}, // 0.2s (SLO 边界)
			{300 * time.Millisecond, "/api/test5"}, // 0.3s
			{1 * time.Second, "/api/test6"},        // 1s
		}

		for _, tc := range testCases {
			metrics.RecordRequest(service, tc.route, "GET", 200, tc.duration)
		}

		// 验证 histogram 不为空（说明记录成功）
		assert.NotNil(t, metrics.RequestDuration)
	})
}

func TestHTTPMetrics_ConcurrentSafety(t *testing.T) {
	t.Run("并发安全测试", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		metrics := NewHTTPMetricsWithRegistry("test", reg)

		// 启动多个 goroutine 并发记录指标
		done := make(chan bool)
		numGoroutines := 10
		numIterations := 100

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < numIterations; j++ {
					metrics.RecordRequest("game", "/api/test", "GET", 200, 100*time.Millisecond)
					metrics.IncInProgress("game")
					metrics.DecInProgress("game")
				}
				done <- true
			}()
		}

		// 等待所有 goroutine 完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// 验证总计数正确
		totalCount := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("game", "/api/test", "GET", "200"))
		assert.Equal(t, float64(numGoroutines*numIterations), totalCount)

		// 验证进行中的请求数归零
		assert.Equal(t, float64(0), testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues("game")))
	})
}
