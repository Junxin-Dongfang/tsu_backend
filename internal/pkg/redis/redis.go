package redis

import (
	"context"
	"fmt"
	"time"

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
}

// NewClient 创建 Redis 客户端
func NewClient(cfg Config) (*Client, error) {
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

	return &Client{Client: rdb}, nil
}

// SetWithTTL 设置键值对，带过期时间
func (c *Client) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl).Err()
}

// GetString 获取字符串值
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key).Result()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.Client.Exists(ctx, key).Result()
	return n > 0, err
}

// DeleteKey 删除键
func (c *Client) DeleteKey(ctx context.Context, keys ...string) error {
	return c.Del(ctx, keys...).Err()
}
