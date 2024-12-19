package redis

import (
	"context"
	"errors"
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrNil = errors.New("redis: nil")
)

// Client wraps the Redis client
type Client struct {
	client *redis.Client
}

// Options Redis配置选项
type Options struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int // 连接池大小
	MinIdleConns int // 最小空闲连接数
	MaxRetries   int // 最大重试次数
}

// NewClient creates a new Redis client
func NewClient(cfg *config.Config) (*Client, error) {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     50, // 默认连接池大小
		MinIdleConns: 10, // 最小空闲连接
		MaxRetries:   3,  // 最大重试次数
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, fmt.Errorf("connect to redis failed: %w", err)
	}

	logger.Info("Successfully connected to Redis")
	return &Client{client: client}, nil
}

// Set stores a key-value pair with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := c.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNil
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

// Delete removes keys
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// Exists checks if keys exist
func (c *Client) Exists(ctx context.Context, keys ...string) (bool, error) {
	n, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return n > 0, nil
}

// Incr increments the key
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr failed: %w", err)
	}
	return val, nil
}

// SetNX sets a key-value pair if the key doesn't exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	ok, err := c.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}
	return ok, nil
}

// Eval executes a Lua script
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return c.client.Eval(ctx, script, keys, args...)
}

// Pipeline 创建管道
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// Watch 监视keys的变化
func (c *Client) Watch(ctx context.Context, fn func(tx *redis.Tx) error, keys ...string) error {
	return c.client.Watch(ctx, fn, keys...)
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("redis close failed: %w", err)
	}
	return nil
}

// HSet sets hash fields
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	if err := c.client.HSet(ctx, key, values...).Err(); err != nil {
		return fmt.Errorf("redis hset failed: %w", err)
	}
	return nil
}

// HGet gets hash field
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.client.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNil
	}
	if err != nil {
		return "", fmt.Errorf("redis hget failed: %w", err)
	}
	return val, nil
}

// HGetAll gets all fields and values in a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	val, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis hgetall failed: %w", err)
	}
	return val, nil
}

// Expire sets key expiration
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := c.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}

// TTL gets key time to live
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	return ttl, nil
}
