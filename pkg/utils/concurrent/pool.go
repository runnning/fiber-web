package concurrent

import "sync"

// ObjectPool is a strongly typed version of sync.Pool from go standard library
type ObjectPool[T any] struct {
	syncPool sync.Pool
}

func NewObjectPool[T any](newFn func() T) *ObjectPool[T] {
	pool := &ObjectPool[T]{
		syncPool: sync.Pool{
			New: func() any { return newFn() },
		},
	}

	return pool
}

// Get returns an arbitrary item from the pool.
func (p *ObjectPool[T]) Get() T {
	return p.syncPool.Get().(T)
}

// Put places an item in the pool
func (p *ObjectPool[T]) Put(value T) {
	p.syncPool.Put(value)
}
