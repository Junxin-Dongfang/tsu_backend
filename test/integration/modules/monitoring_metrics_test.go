package modules

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMetricsEndpointsExposePrometheus 确保 admin/game /metrics 都可通过 80 端口访问。
func TestMetricsEndpointsExposePrometheus(t *testing.T) {
	ctx, cfg, _, _ := setup(t)
	client := &http.Client{Timeout: 5 * time.Second}
	tests := []struct {
		name string
		path string
	}{
		{name: "admin-metrics", path: "/admin/metrics"},
		{name: "game-metrics", path: "/game/metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.BaseURL+tt.path, http.NoBody)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode, "%s body=%s", tt.path, string(body))
			require.Contains(t, string(body), "tsu_http_requests_total", "%s 未暴露 prom 指标", tt.path)
		})
	}
}
