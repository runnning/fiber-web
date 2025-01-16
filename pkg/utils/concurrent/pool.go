package concurrent

import "sync"

// ObjectPool 是 Go 标准库 sync.Pool 的强类型版本
type ObjectPool[T any] struct {
	syncPool sync.Pool
}

// NewObjectPool 创建一个新的对象池，使用提供的函数创建新对象
func NewObjectPool[T any](newFn func() T) *ObjectPool[T] {
	pool := &ObjectPool[T]{
		syncPool: sync.Pool{
			New: func() any { return newFn() },
		},
	}

	return pool
}

// Get 从池中获取一个对象
func (p *ObjectPool[T]) Get() T {
	return p.syncPool.Get().(T)
}

// Put 将对象放回池中
func (p *ObjectPool[T]) Put(value T) {
	p.syncPool.Put(value)
}
