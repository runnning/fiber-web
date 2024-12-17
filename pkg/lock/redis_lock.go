package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fiber_web/pkg/logger"
	redisClient "fiber_web/pkg/redis"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	// ErrLockNotAcquired 获取锁失败
	ErrLockNotAcquired = errors.New("could not acquire lock")
	// ErrLockNotHeld 未持有锁
	ErrLockNotHeld = errors.New("lock not held")
)

// RedisLock represents a distributed lock implementation using Redis
type RedisLock struct {
	client     *redisClient.Client
	key        string
	value      string
	expiration time.Duration
}

// Options contains the configuration for the Redis lock
type Options struct {
	// Key is the lock key
	Key string
	// Expiration is the lock expiration time
	Expiration time.Duration
	// RedisClient is the Redis client instance
	RedisClient *redisClient.Client
}

// NewRedisLock creates a new distributed lock using Redis
func NewRedisLock(opts Options) (*RedisLock, error) {
	if opts.RedisClient == nil {
		return nil, errors.New("redis client is required")
	}
	if opts.Key == "" {
		return nil, errors.New("key is required")
	}
	if opts.Expiration <= 0 {
		return nil, errors.New("expiration must be greater than 0")
	}

	// Generate a random value for the lock
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random value: %v", err)
	}
	value := hex.EncodeToString(b)

	return &RedisLock{
		client:     opts.RedisClient,
		key:        opts.Key,
		value:      value,
		expiration: opts.Expiration,
	}, nil
}

// TryLock attempts to acquire the lock
// Returns true if the lock was acquired, false otherwise
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	// Use SET NX to ensure atomic lock acquisition
	success, err := l.client.SetNX(ctx, l.key, l.value, l.expiration)
	if err != nil {
		logger.Error("Failed to acquire lock",
			zap.String("key", l.key),
			zap.Error(err))
		return false, err
	}

	if success {
		logger.Debug("Lock acquired",
			zap.String("key", l.key),
			zap.String("value", l.value))
	}

	return success, nil
}

// Lock blocks until the lock is acquired or context is cancelled
func (l *RedisLock) Lock(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			acquired, err := l.TryLock(ctx)
			if err != nil {
				return err
			}
			if acquired {
				return nil
			}
			// Wait a bit before trying again
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Unlock releases the lock
func (l *RedisLock) Unlock(ctx context.Context) error {
	// Lua script to ensure we only delete our own lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		logger.Error("Failed to release lock",
			zap.String("key", l.key),
			zap.Error(err))
		return err
	}

	if result.(int64) == 0 {
		logger.Warn("Lock not held",
			zap.String("key", l.key),
			zap.String("value", l.value))
		return ErrLockNotHeld
	}

	logger.Debug("Lock released",
		zap.String("key", l.key),
		zap.String("value", l.value))
	return nil
}

// Refresh extends the lock's expiration
func (l *RedisLock) Refresh(ctx context.Context) error {
	// Lua script to extend expiration only if we hold the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end`

	result, err := l.client.Eval(
		ctx,
		script,
		[]string{l.key},
		l.value,
		l.expiration.Milliseconds(),
	).Result()

	if err != nil {
		logger.Error("Failed to refresh lock",
			zap.String("key", l.key),
			zap.Error(err))
		return err
	}

	if result.(int64) == 0 {
		logger.Warn("Lock not held during refresh",
			zap.String("key", l.key),
			zap.String("value", l.value))
		return ErrLockNotHeld
	}

	logger.Debug("Lock refreshed",
		zap.String("key", l.key),
		zap.Duration("expiration", l.expiration))
	return nil
}

// IsHeld checks if the lock is currently held by us
func (l *RedisLock) IsHeld(ctx context.Context) (bool, error) {
	value, err := l.client.Get(ctx, l.key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return value == l.value, nil
}

// AutoRefresh automatically refreshes the lock periodically
// Returns a cancel function to stop the auto-refresh
func (l *RedisLock) AutoRefresh(ctx context.Context, interval time.Duration) (func(), error) {
	// Check if we hold the lock first
	held, err := l.IsHeld(ctx)
	if err != nil {
		return nil, err
	}
	if !held {
		return nil, ErrLockNotHeld
	}

	refreshCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-refreshCtx.Done():
				return
			case <-ticker.C:
				if err := l.Refresh(refreshCtx); err != nil {
					logger.Error("Failed to auto-refresh lock",
						zap.String("key", l.key),
						zap.Error(err))
					cancel()
					return
				}
			}
		}
	}()

	return cancel, nil
}
