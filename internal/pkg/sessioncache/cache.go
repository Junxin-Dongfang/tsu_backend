package sessioncache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/metrics"
)

// Session 描述缓存的登录会话信息。
type Session struct {
	SessionToken string
	UserID       string
	Username     string
	Email        string
}

type entry struct {
	value     Session
	expiresAt time.Time
}

// Cache 提供线程安全的会话缓存,用于复用 Kratos/Keto 返回的身份数据。
type Cache struct {
	ttl     time.Duration
	metrics *metrics.LoginMetrics
	logger  log.Logger
	clock   func() time.Time
	mu      sync.RWMutex
	store   map[string]*entry
}

// New 返回默认 Cache 实例。
func New(ttl time.Duration, m *metrics.LoginMetrics, logger log.Logger) *Cache {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if m == nil {
		m = metrics.DefaultLoginMetrics
	}
	if logger == nil {
		logger = log.GetLogger()
	}
	return &Cache{
		ttl:     ttl,
		metrics: m,
		logger:  logger.With("component", "session_cache"),
		clock:   time.Now,
		store:   make(map[string]*entry),
	}
}

// Get 返回缓存的 Session,命中时会刷新 TTL。
func (c *Cache) Get(ctx context.Context, service, token string) (Session, bool) {
	service = NormalizeService(service)
	if token == "" {
		c.metrics.IncCacheMiss(service)
		return Session{}, false
	}

	c.mu.RLock()
	value, ok := c.store[token]
	c.mu.RUnlock()

	if !ok {
		c.metrics.IncCacheMiss(service)
		c.logger.DebugContext(ctx, "session cache miss",
			log.String("service", service),
			log.String("token_hash", hashToken(token)))
		return Session{}, false
	}

	now := c.clock()
	if now.After(value.expiresAt) {
		c.metrics.IncCacheEvicted(service, "expired")
		c.logger.InfoContext(ctx, "session cache expired",
			log.String("service", service),
			log.String("token_hash", hashToken(token)))
		c.mu.Lock()
		delete(c.store, token)
		c.mu.Unlock()
		return Session{}, false
	}

	// 刷新 TTL
	c.mu.Lock()
	value.expiresAt = now.Add(c.ttl)
	c.mu.Unlock()

	c.metrics.IncCacheHit(service)
	c.logger.DebugContext(ctx, "session cache hit",
		log.String("service", service),
		log.String("token_hash", hashToken(token)))
	return value.value, true
}

// Set 写入或刷新 Session。
func (c *Cache) Set(ctx context.Context, service string, session Session) {
	service = NormalizeService(service)
	if session.SessionToken == "" {
		return
	}
	c.mu.Lock()
	c.store[session.SessionToken] = &entry{
		value:     session,
		expiresAt: c.clock().Add(c.ttl),
	}
	c.mu.Unlock()
	c.logger.DebugContext(ctx, "session cache updated",
		log.String("service", service),
		log.String("token_hash", hashToken(session.SessionToken)))
}

// Delete 主动剔除缓存（例如 logout / session 失效）。
func (c *Cache) Delete(ctx context.Context, service, token, reason string) {
	service = NormalizeService(service)
	if token == "" {
		return
	}
	c.mu.Lock()
	if _, ok := c.store[token]; ok {
		delete(c.store, token)
		c.metrics.IncCacheEvicted(service, reason)
		c.logger.InfoContext(ctx, "session cache evicted",
			log.String("service", service),
			log.String("reason", reason),
			log.String("token_hash", hashToken(token)))
	}
	c.mu.Unlock()
}

// NormalizeService 确保 service label 不为空。
func NormalizeService(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "unknown"
	}
	return service
}

func hashToken(token string) string {
	if token == "" {
		return ""
	}
	h := sha1.Sum([]byte(token))
	return hex.EncodeToString(h[:])[:12]
}
