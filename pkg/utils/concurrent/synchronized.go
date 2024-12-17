package concurrent

import "sync"

type Synchronized[T any] struct {
	mu    sync.Mutex
	value T
}

func NewSynchronized[T any](value T) *Synchronized[T] {
	return &Synchronized[T]{
		value: value,
	}
}

func (s *Synchronized[T]) WithLock(fn func(value T)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.value)
}
