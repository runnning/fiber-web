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

const (
	lockScript = `
		if redis.call("exists", KEYS[1]) == 0 then
			return redis.call("set", KEYS[1], ARGV[1], "PX", ARGV[2])
		end
		return false`

	unlockScript = `
		if redis.call("exists", KEYS[1]) == 1 then
			return redis.call("del", KEYS[1])
		end
		return false`
)

const defaultRetryInterval = 100 * time.Millisecond

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(client *redisClient.Client) Lock {
	return &RedisLock{
		client: client,
	}
}

// Lock 加锁
func (l *RedisLock) Lock(ctx context.Context, key string, value any, ttl time.Duration) error {
	ok, err := l.TryLock(ctx, key, value, ttl)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockFailed
	}
	return nil
}

// TryLock 尝试加锁
func (l *RedisLock) TryLock(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	result, err := l.client.Eval(ctx, lockScript, []string{key}, value, ttl.Milliseconds()).Result()
	if err != nil {
		return false, l.handleError("TryLock", key, err)
	}
	return result != false, nil
}

// LockWithTimeout 在指定时间内尝试加锁
func (l *RedisLock) LockWithTimeout(ctx context.Context, key string, value any, ttl time.Duration, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	// 先尝试一次加锁
	ok, err := l.TryLock(ctx, key, value, ttl)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	// 使用动态重试间隔
	retryInterval := defaultRetryInterval
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ok, err := l.TryLock(ctx, key, value, ttl)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}

			// 动态调整重试间隔（可选）
			retryInterval = min(retryInterval*2, timeout/10)
			ticker.Reset(retryInterval)
		}
	}

	return ErrLockTimeout
}

// Unlock 解锁
func (l *RedisLock) Unlock(ctx context.Context, key string) error {
	_, err := l.client.Eval(ctx, unlockScript, []string{key}).Result()
	return l.handleError("Unlock", key, err)
}

// Refresh 刷新锁的过期时间
func (l *RedisLock) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	err := l.client.Expire(ctx, key, ttl)
	return l.handleError("Refresh", key, err)
}

// GetLockTTL 获取锁的剩余过期时间
func (l *RedisLock) GetLockTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := l.client.TTL(ctx, key)
	if err != nil {
		return 0, l.handleError("GetLockTTL", key, err)
	}
	return ttl, nil
}

func (l *RedisLock) handleError(operation string, key string, err error) error {
	if err != nil {
		logger.Error(operation+" failed",
			zap.String("key", key),
			zap.Error(err))
	}
	return err
}
