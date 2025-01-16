package concurrent

import "sync"

// Synchronized 提供一个线程安全的值容器
type Synchronized[T any] struct {
	mu    sync.RWMutex
	value T
}

// NewSynchronized 创建一个新的同步值容器
func NewSynchronized[T any](initial T) *Synchronized[T] {
	return &Synchronized[T]{value: initial}
}

// Get 获取当前值
func (s *Synchronized[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set 设置新值
func (s *Synchronized[T]) Set(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = value
}
