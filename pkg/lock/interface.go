package lock

import (
	"context"
	"time"
)

// Locker represents a distributed lock interface
type Locker interface {
	// TryLock attempts to acquire the lock
	// Returns true if the lock was acquired, false otherwise
	TryLock(ctx context.Context) (bool, error)

	// Lock blocks until the lock is acquired or context is cancelled
	Lock(ctx context.Context) error

	// Unlock releases the lock
	Unlock(ctx context.Context) error

	// Refresh extends the lock's expiration
	Refresh(ctx context.Context) error

	// IsHeld checks if the lock is currently held
	IsHeld(ctx context.Context) (bool, error)

	// AutoRefresh automatically refreshes the lock periodically
	// Returns a cancel function to stop the auto-refresh
	AutoRefresh(ctx context.Context, interval time.Duration) (func(), error)
}
