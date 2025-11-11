package redis

import (
	"context"
	"fmt"
	"time"

	"tsu-self/internal/pkg/metrics"

	"github.com/redis/go-redis/v9"
)

// Config Redis 配置
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Client Redis 客户端封装
type Client struct {
	*redis.Client
	service string
}

// NewClient 创建 Redis 客户端
func NewClient(cfg Config, service string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %w", err)
	}

	if service == "" {
		service = metrics.GetServiceName()
	}

	return &Client{
		Client:  rdb,
		service: service,
	}, nil
}

// SetWithTTL 设置键值对，带过期时间
func (c *Client) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	err := c.Set(ctx, key, value, ttl).Err()
	duration := time.Since(start)

	// 记录 Redis 操作指标
	metrics.DefaultResourceMetrics.RecordRedisOperation("SET", err == nil, duration, c.service)
	if err != nil {
		metrics.DefaultResourceMetrics.RecordRedisError("operation_error", c.service)
	}

	return err
}

// GetString 获取字符串值
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := c.Get(ctx, key).Result()
	duration := time.Since(start)

	// 记录 Redis 操作指标
	metrics.DefaultResourceMetrics.RecordRedisOperation("GET", err == nil, duration, c.service)
	if err != nil && err != redis.Nil {
		metrics.DefaultResourceMetrics.RecordRedisError("operation_error", c.service)
	} else if err == redis.Nil {
		metrics.DefaultResourceMetrics.RecordRedisError("nil", c.service)
	}

	return result, err
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	n, err := c.Client.Exists(ctx, key).Result()
	duration := time.Since(start)

	// 记录 Redis 操作指标
	metrics.DefaultResourceMetrics.RecordRedisOperation("EXISTS", err == nil, duration, c.service)
	if err != nil {
		metrics.DefaultResourceMetrics.RecordRedisError("operation_error", c.service)
	}

	return n > 0, err
}

// DeleteKey 删除键
func (c *Client) DeleteKey(ctx context.Context, keys ...string) error {
	start := time.Now()
	err := c.Del(ctx, keys...).Err()
	duration := time.Since(start)

	// 记录 Redis 操作指标
	metrics.DefaultResourceMetrics.RecordRedisOperation("DEL", err == nil, duration, c.service)
	if err != nil {
		metrics.DefaultResourceMetrics.RecordRedisError("operation_error", c.service)
	}

	return err
}
