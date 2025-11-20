// File: internal/pkg/metrics/middleware_test.go
package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func withServiceName(t *testing.T, name string) {
	original := GetServiceName()
	SetServiceName(name)
	t.Cleanup(func() {
		SetServiceName(original)
	})
}

// TestMiddleware_RouteTemplate 验证中间件正确提取路由模板
func TestMiddleware_RouteTemplate(t *testing.T) {
	tests := []struct {
		name           string
		registerRoute  string
		requestPath    string
		expectedPath   string
		expectedMethod string
	}{
		{
			name:           "提取参数化路由模板 - /api/heroes/:id",
			registerRoute:  "/api/heroes/:id",
			requestPath:    "/api/heroes/12345",
			expectedPath:   "/api/heroes/:id",
			expectedMethod: "GET",
		},
		{
			name:           "提取嵌套参数路由 - /api/users/:userId/heroes/:heroId",
			registerRoute:  "/api/users/:userId/heroes/:heroId",
			requestPath:    "/api/users/1/heroes/2",
			expectedPath:   "/api/users/:userId/heroes/:heroId",
			expectedMethod: "POST",
		},
		{
			name:           "静态路由 - /api/login",
			registerRoute:  "/api/login",
			requestPath:    "/api/login",
			expectedPath:   "/api/login",
			expectedMethod: "POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withServiceName(t, "test-service")
			// 创建独立的注册表和指标
			reg := prometheus.NewRegistry()
			metrics := NewHTTPMetricsWithRegistry("test", reg)

			// 临时替换全局指标
			originalMetrics := DefaultHTTPMetrics
			DefaultHTTPMetrics = metrics
			defer func() { DefaultHTTPMetrics = originalMetrics }()

			// 创建 Echo 实例
			e := echo.New()
			e.Use(Middleware())

			// 注册路由
			var capturedPath string
			handler := func(c echo.Context) error {
				capturedPath = c.Path()
				return c.String(http.StatusOK, "ok")
			}

			switch tt.expectedMethod {
			case "GET":
				e.GET(tt.registerRoute, handler)
			case "POST":
				e.POST(tt.registerRoute, handler)
			}

			// 创建请求
			req := httptest.NewRequest(tt.expectedMethod, tt.requestPath, nil)
			rec := httptest.NewRecorder()

			// 执行请求
			e.ServeHTTP(rec, req)

			// 验证路由模板被正确捕获
			assert.Equal(t, tt.expectedPath, capturedPath, "应该使用路由模板而非具体路径")
			assert.Equal(t, tt.expectedPath, rec.Header().Get("X-Route-Pattern"))

			// 验证指标被正确记录
			// 等待一小段时间以确保指标被记录
			time.Sleep(10 * time.Millisecond)

			// 验证 requests_total 指标
			metricName := "test_http_requests_total"
			count, err := testutil.GatherAndCount(reg, metricName)
			assert.NoError(t, err)
			assert.Equal(t, 1, count, "应该有一个 requests_total 指标")

			// 验证 request_duration_seconds 指标（histogram 会生成 _count, _sum, _bucket）
			histogramMetricName := "test_http_request_duration_seconds"
			histCount, err := testutil.GatherAndCount(reg, histogramMetricName)
			assert.NoError(t, err)
			assert.Greater(t, histCount, 0, "应该有 histogram 相关指标")
		})
	}
}

// TestMiddleware_HealthCheckSkip 验证健康检查端点被跳过
func TestMiddleware_HealthCheckSkip(t *testing.T) {
	healthCheckPaths := []string{
		"/metrics",
		"/health",
		"/healthz",
		"/readyz",
		"/livez",
	}

	for _, path := range healthCheckPaths {
		t.Run("跳过健康检查端点: "+path, func(t *testing.T) {
			withServiceName(t, "test-service")
			// 创建独立的注册表和指标
			reg := prometheus.NewRegistry()
			metrics := NewHTTPMetricsWithRegistry("test", reg)

			// 临时替换全局指标
			originalMetrics := DefaultHTTPMetrics
			DefaultHTTPMetrics = metrics
			defer func() { DefaultHTTPMetrics = originalMetrics }()

			// 创建 Echo 实例
			e := echo.New()
			e.Use(Middleware())

			// 注册健康检查路由
			e.GET(path, func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			// 创建请求
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			// 执行请求
			e.ServeHTTP(rec, req)

			// 验证请求成功
			assert.Equal(t, http.StatusOK, rec.Code)

			// 验证指标没有被记录
			metricName := "test_http_requests_total"
			count, err := testutil.GatherAndCount(reg, metricName)
			assert.NoError(t, err)
			assert.Equal(t, 0, count, "健康检查端点不应该被记录到指标中")
		})
	}
}

// TestMiddleware_MetricRecording 验证指标正确记录
func TestMiddleware_MetricRecording(t *testing.T) {
	tests := []struct {
		name       string
		route      string
		method     string
		statusCode int
	}{
		{
			name:       "记录 200 成功响应",
			route:      "/api/test",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
		},
		{
			name:       "记录 404 客户端错误",
			route:      "/api/notfound",
			method:     http.MethodGet,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "记录 500 服务器错误",
			route:      "/api/error",
			method:     http.MethodPost,
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withServiceName(t, "test-service")
			// 创建独立的注册表和指标
			reg := prometheus.NewRegistry()
			metrics := NewHTTPMetricsWithRegistry("test", reg)

			// 临时替换全局指标
			originalMetrics := DefaultHTTPMetrics
			DefaultHTTPMetrics = metrics
			defer func() { DefaultHTTPMetrics = originalMetrics }()

			// 创建 Echo 实例
			e := echo.New()
			e.Use(Middleware())

			// 注册路由
			handler := func(c echo.Context) error {
				return c.String(tt.statusCode, "response")
			}

			switch tt.method {
			case http.MethodGet:
				e.GET(tt.route, handler)
			case http.MethodPost:
				e.POST(tt.route, handler)
			}

			// 创建请求
			req := httptest.NewRequest(tt.method, tt.route, nil)
			rec := httptest.NewRecorder()

			// 执行请求
			e.ServeHTTP(rec, req)

			// 验证响应状态码
			assert.Equal(t, tt.statusCode, rec.Code)

			// 等待指标记录
			time.Sleep(10 * time.Millisecond)

			// 验证指标被记录
			statusCode := strconv.Itoa(tt.statusCode)
			count := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("test-service", tt.route, tt.method, statusCode))
			assert.Equal(t, float64(1), count, "应该记录一个 requests_total 指标")
		})
	}
}

// TestMiddleware_InProgressMetric 验证当前进行中的请求数
func TestMiddleware_InProgressMetric(t *testing.T) {
	withServiceName(t, "test-service")
	// 创建独立的注册表和指标
	reg := prometheus.NewRegistry()
	metrics := NewHTTPMetricsWithRegistry("test", reg)

	// 临时替换全局指标
	originalMetrics := DefaultHTTPMetrics
	DefaultHTTPMetrics = metrics
	defer func() { DefaultHTTPMetrics = originalMetrics }()

	// 创建 Echo 实例
	e := echo.New()
	e.Use(Middleware())

	// 创建一个慢响应的处理器
	e.GET("/api/slow", func(c echo.Context) error {
		time.Sleep(100 * time.Millisecond)
		return c.String(http.StatusOK, "ok")
	})

	// 创建请求
	req := httptest.NewRequest(http.MethodGet, "/api/slow", nil)
	rec := httptest.NewRecorder()

	// 在 goroutine 中执行请求
	done := make(chan bool)
	go func() {
		e.ServeHTTP(rec, req)
		done <- true
	}()

	// 短暂等待，确保请求开始处理
	time.Sleep(10 * time.Millisecond)

	// 此时 in_progress 应该为 1（但由于并发，我们无法精确测试）
	// 等待请求完成
	<-done

	// 请求完成后，in_progress 应该回到 0
	inProgress := testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues("test-service"))
	assert.Equal(t, float64(0), inProgress, "请求完成后 in_progress 应该为 0")
}

// TestMiddleware_PathLimitTracking 验证路径标签基数控制
func TestMiddleware_PathLimitTracking(t *testing.T) {
	withServiceName(t, "test-service")
	// 创建独立的注册表和指标
	reg := prometheus.NewRegistry()
	metrics := NewHTTPMetricsWithRegistry("test", reg)

	// 临时替换全局指标
	originalMetrics := DefaultHTTPMetrics
	DefaultHTTPMetrics = metrics
	defer func() { DefaultHTTPMetrics = originalMetrics }()

	// 创建一个新的 PathLimitTracker 用于测试
	testTracker := NewPathLimitTracker(5) // 限制为 5 个路径

	// 临时替换全局 tracker
	originalTracker := pathLimitTracker
	pathLimitTracker = testTracker
	defer func() { pathLimitTracker = originalTracker }()

	// 创建 Echo 实例
	e := echo.New()
	e.Use(Middleware())

	// 注册多个路由
	for i := 1; i <= 7; i++ {
		path := "/api/test" + string(rune('0'+i))
		e.GET(path, func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})
	}

	// 发送 7 个不同的请求
	for i := 1; i <= 7; i++ {
		path := "/api/test" + string(rune('0'+i))
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}

	// 验证追踪的路径数量不超过限制
	trackedCount := testTracker.GetTrackedCount()
	assert.LessOrEqual(t, trackedCount, 5, "追踪的路径数量不应超过限制")

	// 验证警告信息
	warning := testTracker.LogWarning()
	assert.NotEmpty(t, warning, "接近限制时应该有警告信息")
}

// TestMiddleware_ConcurrentRequests 验证并发请求处理
func TestMiddleware_ConcurrentRequests(t *testing.T) {
	withServiceName(t, "test-service")
	// 创建独立的注册表和指标
	reg := prometheus.NewRegistry()
	metrics := NewHTTPMetricsWithRegistry("test", reg)

	// 临时替换全局指标
	originalMetrics := DefaultHTTPMetrics
	DefaultHTTPMetrics = metrics
	defer func() { DefaultHTTPMetrics = originalMetrics }()

	// 创建 Echo 实例
	e := echo.New()
	e.Use(Middleware())

	// 注册路由
	e.GET("/api/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// 并发发送多个请求
	numRequests := 10
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			done <- true
		}()
	}

	// 等待所有请求完成
	for i := 0; i < numRequests; i++ {
		<-done
	}

	// 等待指标记录
	time.Sleep(50 * time.Millisecond)

	// 验证所有请求都被记录
	metricName := "test_http_requests_total"
	count, err := testutil.GatherAndCount(reg, metricName)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "并发请求应该被正确记录")

	// 验证 in_progress 回到 0
	inProgress := testutil.ToFloat64(metrics.RequestsInProgress.WithLabelValues("test-service"))
	assert.Equal(t, float64(0), inProgress, "所有请求完成后 in_progress 应该为 0")
}

func TestPathLimitTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewPathLimitTracker(10)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				tracker.TrackPath(fmt.Sprintf("/api/%d/%d", id, j))
			}
		}(i)
	}

	wg.Wait()
	assert.LessOrEqual(t, tracker.GetTrackedCount(), 10, "路径追踪数量不应超过上限")
}
