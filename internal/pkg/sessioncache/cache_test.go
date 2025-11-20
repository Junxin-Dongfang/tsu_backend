package sessioncache

import (
	"context"
	"testing"
	"time"

	"tsu-self/internal/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestCacheGetSet(t *testing.T) {
	ctx := context.Background()
	reg := prometheus.NewRegistry()
	metrics := metrics.NewLoginMetricsWithRegistry("test", reg)
	c := New(50*time.Millisecond, metrics, nil)
	s := Session{SessionToken: "token-1", UserID: "u1", Username: "user", Email: "u@example.com"}
	c.Set(ctx, "admin", s)
	got, ok := c.Get(ctx, "admin", s.SessionToken)
	require.True(t, ok)
	require.Equal(t, s, got)
}

func TestCacheExpiry(t *testing.T) {
	ctx := context.Background()
	reg := prometheus.NewRegistry()
	metrics := metrics.NewLoginMetricsWithRegistry("test", reg)
	c := New(10*time.Millisecond, metrics, nil)
	s := Session{SessionToken: "token-2", UserID: "u2"}
	c.Set(ctx, "game", s)
	// wait for expiry
	time.Sleep(15 * time.Millisecond)
	_, ok := c.Get(ctx, "game", s.SessionToken)
	require.False(t, ok)
}

func TestCacheDelete(t *testing.T) {
	ctx := context.Background()
	reg := prometheus.NewRegistry()
	metrics := metrics.NewLoginMetricsWithRegistry("test", reg)
	c := New(time.Second, metrics, nil)
	s := Session{SessionToken: "token-3", UserID: "u3"}
	c.Set(ctx, "admin", s)
	c.Delete(ctx, "admin", s.SessionToken, "logout")
	_, ok := c.Get(ctx, "admin", s.SessionToken)
	require.False(t, ok)
}
