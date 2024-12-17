package concurrent

import (
	"context"
	"sync"
	"time"
)

// Pool represents a generic worker pool
type Pool[T any] struct {
	workers    int
	tasks      chan func(context.Context) (T, error)
	results    chan Result[T]
	done       chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	errHandler func(error)
}

// Result represents the result of a task execution
type Result[T any] struct {
	Value T
	Err   error
}

// PoolOption represents an option for configuring the pool
type PoolOption[T any] func(*Pool[T])

// WithErrorHandler sets the error handler for the pool
func WithErrorHandler[T any](handler func(error)) PoolOption[T] {
	return func(p *Pool[T]) {
		p.errHandler = handler
	}
}

// NewPool creates a new worker pool
func NewPool[T any](workers int, bufferSize int, opts ...PoolOption[T]) *Pool[T] {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool[T]{
		workers:    workers,
		tasks:      make(chan func(context.Context) (T, error), bufferSize),
		results:    make(chan Result[T], bufferSize),
		done:       make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
		errHandler: func(err error) {}, // default error handler does nothing
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Start starts the worker pool
func (p *Pool[T]) Start() {
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range p.tasks {
				select {
				case <-p.ctx.Done():
					return
				default:
					result, err := task(p.ctx)
					if err != nil && p.errHandler != nil {
						p.errHandler(err)
					}
					p.results <- Result[T]{Value: result, Err: err}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(p.results)
		close(p.done)
	}()
}

// Submit submits a task to the pool
func (p *Pool[T]) Submit(task func(context.Context) (T, error)) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.tasks <- task:
		return nil
	}
}

// Results returns the results channel
func (p *Pool[T]) Results() <-chan Result[T] {
	return p.results
}

// Stop stops the worker pool
func (p *Pool[T]) Stop() {
	p.cancel()
	close(p.tasks)
	<-p.done // 等待所有工作完成
}

// WaitForCompletion waits for all tasks to complete with a timeout
func (p *Pool[T]) WaitForCompletion(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-p.done:
		return true
	case <-timer.C:
		return false
	}
}

// Parallel executes functions in parallel with context and error handling
func Parallel[T any](ctx context.Context, fns ...func(context.Context) (T, error)) ([]Result[T], error) {
	var wg sync.WaitGroup
	results := make([]Result[T], len(fns))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, fn := range fns {
		wg.Add(1)
		go func(index int, f func(context.Context) (T, error)) {
			defer wg.Done()
			value, err := f(ctx)
			results[index] = Result[T]{Value: value, Err: err}
			if err != nil {
				cancel() // 如果有错误发生，取消其他正在进行的任务
			}
		}(i, fn)
	}

	wg.Wait()
	return results, ctx.Err()
}

// Race executes functions in parallel and returns the first successful result
func Race[T any](ctx context.Context, fns ...func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultCh := make(chan Result[T], len(fns))
	for _, fn := range fns {
		go func(f func(context.Context) (T, error)) {
			value, err := f(ctx)
			select {
			case <-ctx.Done():
				return
			case resultCh <- Result[T]{Value: value, Err: err}:
			}
		}(fn)
	}

	var lastErr error
	for i := 0; i < len(fns); i++ {
		select {
		case result := <-resultCh:
			if result.Err == nil {
				return result.Value, nil
			}
			lastErr = result.Err
		case <-ctx.Done():
			var zero T
			if ctx.Err() != nil {
				return zero, ctx.Err()
			}
			return zero, lastErr
		}
	}

	var zero T
	return zero, lastErr
}

// Debounce creates a debounced function that delays invoking fn until after wait duration
func Debounce[T any](fn func(T), wait time.Duration) func(T) {
	var mutex sync.Mutex
	var timer *time.Timer

	return func(arg T) {
		mutex.Lock()
		defer mutex.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(wait, func() {
			fn(arg)
		})
	}
}

// Throttle creates a throttled function that only invokes fn at most once per wait duration
func Throttle[T any](fn func(T), wait time.Duration) func(T) {
	var mutex sync.Mutex
	var lastRun time.Time

	return func(arg T) {
		mutex.Lock()
		defer mutex.Unlock()

		now := time.Now()
		if now.Sub(lastRun) >= wait {
			fn(arg)
			lastRun = now
		}
	}
}

// SafeMap is a thread-safe map
type SafeMap[K comparable, V any] struct {
	sync.RWMutex
	data map[K]V
}

// NewSafeMap creates a new SafeMap
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

// Set sets a value in the map
func (m *SafeMap[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
}

// Get gets a value from the map
func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

// Delete deletes a value from the map
func (m *SafeMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
}

// Range iterates over the map
func (m *SafeMap[K, V]) Range(f func(K, V) bool) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}
