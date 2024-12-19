package lock

import (
	"context"
	"time"
)

// Lock 分布式锁接口
type Lock interface {
	// Lock 加锁
	Lock(ctx context.Context, key string, value any, ttl time.Duration) error

	// TryLock 尝试加锁,立即返回结果
	TryLock(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)

	// LockWithTimeout 在指定时间内尝试加锁
	LockWithTimeout(ctx context.Context, key string, value any, ttl time.Duration, timeout time.Duration) error

	// Unlock 解锁
	Unlock(ctx context.Context, key string) error

	// Refresh 刷新锁的过期时间
	Refresh(ctx context.Context, key string, ttl time.Duration) error

	// GetLockTTL 获取锁的剩余过期时间
	GetLockTTL(ctx context.Context, key string) (time.Duration, error)
}
