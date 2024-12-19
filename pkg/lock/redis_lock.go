package lock

import (
	"context"
	"errors"
	"fiber_web/pkg/logger"
	redisClient "fiber_web/pkg/redis"
	"time"

	"go.uber.org/zap"
)

var (
	ErrLockFailed  = errors.New("failed to acquire lock")
	ErrLockTimeout = errors.New("lock timeout")
)

type RedisLock struct {
	client *redisClient.Client
}

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(client *redisClient.Client) Lock {
	return &RedisLock{
		client: client,
	}
}

// Lock 加锁
func (l *RedisLock) Lock(ctx context.Context, key string, value any, ttl time.Duration) error {
	script := `
		if redis.call("exists", KEYS[1]) == 0 then
			return redis.call("set", KEYS[1], ARGV[1], "PX", ARGV[2])
		end
		return false`

	result, err := l.client.Eval(ctx, script, []string{key}, value, ttl.Milliseconds()).Result()
	if err != nil {
		logger.Error("Lock failed", zap.String("key", key), zap.Error(err))
		return err
	}
	if result == false {
		return ErrLockFailed
	}
	return nil
}

// TryLock 尝试加锁
func (l *RedisLock) TryLock(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	script := `
		if redis.call("exists", KEYS[1]) == 0 then
			return redis.call("set", KEYS[1], ARGV[1], "PX", ARGV[2])
		end
		return false`

	result, err := l.client.Eval(ctx, script, []string{key}, value, ttl.Milliseconds()).Result()
	if err != nil {
		logger.Error("TryLock failed", zap.String("key", key), zap.Error(err))
		return false, err
	}
	return result != false, nil
}

// LockWithTimeout 在指定时间内尝试加锁
func (l *RedisLock) LockWithTimeout(ctx context.Context, key string, value any, ttl time.Duration, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return ErrLockTimeout
		case <-ticker.C:
			ok, err := l.TryLock(ctx, key, value, ttl)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

// Unlock 解锁
func (l *RedisLock) Unlock(ctx context.Context, key string) error {
	script := `
		if redis.call("exists", KEYS[1]) == 1 then
			return redis.call("del", KEYS[1])
		end
		return false`

	_, err := l.client.Eval(ctx, script, []string{key}).Result()
	if err != nil {
		logger.Error("Unlock failed", zap.String("key", key), zap.Error(err))
	}
	return err
}

// Refresh 刷新锁的过期时间
func (l *RedisLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	err := l.client.Expire(ctx, key, ttl)
	if err != nil {
		logger.Error("Refresh lock failed", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// GetLockTTL 获取锁的剩余过期时间
func (l *RedisLock) GetLockTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := l.client.TTL(ctx, key)
	if err != nil {
		logger.Error("Get lock TTL failed", zap.String("key", key), zap.Error(err))
		return 0, err
	}
	return ttl, nil
}
