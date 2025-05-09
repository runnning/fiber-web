package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fiber_web/pkg/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNil = errors.New("redis: nil")
)

// RedisManager 管理多个Redis实例
type RedisManager struct {
	clients map[string]*Client
}

// Client 包装Redis客户端
type Client struct {
	client *redis.Client
}

// NewRedisManager 创建Redis管理器
func NewRedisManager(cfg *config.RedisConfig) (*RedisManager, error) {
	manager := &RedisManager{
		clients: make(map[string]*Client),
	}

	if cfg.MultiInstance {
		// 多实例模式
		for name, redisConfig := range cfg.Instances {
			client, err := newRedisClient(&redisConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create redis client %s: %w", name, err)
			}
			manager.clients[name] = client
		}
	} else {
		// 单实例模式
		client, err := newRedisClient(&cfg.Default)
		if err != nil {
			return nil, err
		}
		manager.clients["default"] = client
	}

	return manager, nil
}

// newRedisClient 创建单个Redis客户端
func newRedisClient(cfg *config.RedisInstanceConfig) (*Client, error) {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis failed: %w", err)
	}

	return &Client{client: client}, nil
}

// GetClient 获取指定名称的Redis客户端
func (m *RedisManager) GetClient(name string) (*Client, error) {
	client, exists := m.clients[name]
	if !exists {
		return nil, fmt.Errorf("redis client %s not found", name)
	}
	return client, nil
}

// Close 关闭所有Redis连接
func (m *RedisManager) Close() error {
	var errs []error
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis client %s: %w", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing redis clients: %v", errs)
	}
	return nil
}

// Set stores a key-value pair with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

// Delete 基础操作方法
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (bool, error) {
	n, err := c.client.Exists(ctx, keys...).Result()
	return n > 0, err
}

// HSet Hash 操作
func (c *Client) HSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.client.HSet(ctx, key, field, data).Err()
}

func (c *Client) HGet(ctx context.Context, key, field string, value interface{}) error {
	data, err := c.client.HGet(ctx, key, field).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

// Lock 分布式锁
func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	value := fmt.Sprintf("%d", time.Now().UnixNano())
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

func (c *Client) Unlock(ctx context.Context, key string) error {
	return c.Delete(ctx, key)
}

func (c *Client) Close() error {
	return c.client.Close()
}

// LPush List 操作
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	for i, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value at index %d: %w", i, err)
		}
		values[i] = data
	}
	return c.client.LPush(ctx, key, values...).Err()
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	for i, v := range values {
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value at index %d: %w", i, err)
		}
		values[i] = data
	}
	return c.client.RPush(ctx, key, values...).Err()
}

func (c *Client) LPop(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.LPop(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func (c *Client) RPop(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.RPop(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// SAdd Set 集合操作
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	for i, m := range members {
		data, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("failed to marshal member at index %d: %w", i, err)
		}
		members[i] = data
	}
	return c.client.SAdd(ctx, key, members...).Err()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

func (c *Client) SPop(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.SPop(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// ZAdd Sorted Set 操作
func (c *Client) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	data, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %w", err)
	}
	z := redis.Z{
		Score:  score,
		Member: data,
	}
	return c.client.ZAdd(ctx, key, z).Err()
}

func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.ZRange(ctx, key, start, stop).Result()
}

// Eval executes a Lua script
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return c.client.Eval(ctx, script, keys, args...)
}

// Expire sets key expiration
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL gets key time to live
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// GetOrSet 获取值，如果不存在则通过回调函数生成并设置新值
func (c *Client) GetOrSet(ctx context.Context, key string, value interface{}, fn func() (interface{}, error), expiration time.Duration) error {
	// 先尝试获取
	err := c.Get(ctx, key, value)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrNil) {
		return fmt.Errorf("failed to get value: %w", err)
	}

	// 调用回调函数生成新值
	newValue, err := fn()
	if err != nil {
		return fmt.Errorf("failed to generate value: %w", err)
	}

	// 设置新值
	if err := c.Set(ctx, key, newValue, expiration); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	// 将新值复制到传入的value中
	data, err := json.Marshal(newValue)
	if err != nil {
		return fmt.Errorf("failed to marshal new value: %w", err)
	}
	return json.Unmarshal(data, value)
}

// KeyValue 表示一个键值对
type KeyValue struct {
	Key   string
	Value interface{}
}

const (
	defaultBatchSize = 100 // 默认批处理大小
)

// MSetOptions 定义MSet操作的选项
type MSetOptions struct {
	BatchSize   int           // 每批处理的键值对数量，0表示使用默认值
	Expiration  time.Duration // 过期时间
	SkipOnError bool          // 遇到错误时是否继续处理其他键值对
}

// MSet 批量设置键值对，支持回调函数生成缺失值
func (c *Client) MSet(ctx context.Context, pairs []KeyValue, fn func(key string) (interface{}, error), opts *MSetOptions) error {
	if len(pairs) == 0 {
		return nil
	}

	if opts == nil {
		opts = &MSetOptions{
			BatchSize:   defaultBatchSize,
			Expiration:  0,
			SkipOnError: false,
		}
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = defaultBatchSize
	}

	var errs []error
	for i := 0; i < len(pairs); i += opts.BatchSize {
		end := i + opts.BatchSize
		if end > len(pairs) {
			end = len(pairs)
		}

		if err := c.mSetBatch(ctx, pairs[i:end], fn, opts); err != nil {
			if !opts.SkipOnError {
				return err
			}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors occurred during MSet: %v", errs)
	}
	return nil
}

// mSetBatch 处理一批键值对的设置操作
func (c *Client) mSetBatch(ctx context.Context, pairs []KeyValue, fn func(key string) (interface{}, error), opts *MSetOptions) error {
	pipe := c.client.Pipeline()

	for _, pair := range pairs {
		value := pair.Value
		if value == nil && fn != nil {
			var err error
			value, err = fn(pair.Key)
			if err != nil {
				return fmt.Errorf("failed to generate value for key %s: %w", pair.Key, err)
			}
		}

		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", pair.Key, err)
		}

		pipe.Set(ctx, pair.Key, data, opts.Expiration)
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	for i, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			return fmt.Errorf("failed to set value for key %s: %w", pairs[i].Key, err)
		}
	}

	return nil
}

// MGetOptions 定义MGet操作的选项
type MGetOptions struct {
	BatchSize   int  // 每批处理的键数量，0表示使用默认值
	SkipOnError bool // 遇到错误时是否继续处理其他键
}

// MGet 批量获取键值，支持自定义处理函数
func (c *Client) MGet(ctx context.Context, keys []string, fn func(key string, value interface{}) error, opts *MGetOptions) error {
	if len(keys) == 0 {
		return nil
	}

	if opts == nil {
		opts = &MGetOptions{
			BatchSize:   defaultBatchSize,
			SkipOnError: false,
		}
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = defaultBatchSize
	}

	var errs []error
	for i := 0; i < len(keys); i += opts.BatchSize {
		end := i + opts.BatchSize
		if end > len(keys) {
			end = len(keys)
		}

		if err := c.mGetBatch(ctx, keys[i:end], fn); err != nil {
			if !opts.SkipOnError {
				return err
			}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors occurred during MGet: %v", errs)
	}
	return nil
}

// mGetBatch 处理一批键的获取操作
func (c *Client) mGetBatch(ctx context.Context, keys []string, fn func(key string, value interface{}) error) error {
	pipe := c.client.Pipeline()

	for _, key := range keys {
		pipe.Get(ctx, key)
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	for i, cmd := range cmds {
		key := keys[i]
		strCmd, ok := cmd.(*redis.StringCmd)
		if !ok {
			return fmt.Errorf("unexpected command type for key %s", key)
		}

		data, err := strCmd.Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				if fn != nil {
					if err := fn(key, nil); err != nil {
						return fmt.Errorf("failed to handle nil value for key %s: %w", key, err)
					}
				}
				continue
			}
			return fmt.Errorf("failed to get value for key %s: %w", key, err)
		}

		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}

		if fn != nil {
			if err := fn(key, value); err != nil {
				return fmt.Errorf("failed to handle value for key %s: %w", key, err)
			}
		}
	}

	return nil
}
