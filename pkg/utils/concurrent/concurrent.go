package concurrent

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrWorkerFull = errors.New("worker pool is full")
)

// Pool 表示一个通用的工作池
type Pool[T any] struct {
	workers    int                                   // 工作协程数量
	tasks      chan func(context.Context) (T, error) // 任务通道
	results    chan Result[T]                        // 结果通道
	done       chan struct{}                         // 完成信号通道
	ctx        context.Context                       // 上下文
	cancel     context.CancelFunc                    // 取消函数
	errHandler func(error)                           // 错误处理函数
	wg         sync.WaitGroup                        // 等待组
}

// Result 表示任务执行的结果
type Result[T any] struct {
	Value T     // 结果值
	Err   error // 错误信息
}

// PoolOption 表示配置工作池的选项
type PoolOption[T any] func(*Pool[T])

// WithErrorHandler 设置工作池的错误处理函数
func WithErrorHandler[T any](handler func(error)) PoolOption[T] {
	return func(p *Pool[T]) {
		p.errHandler = handler
	}
}

// NewPool 创建一个新的工作池
func NewPool[T any](workers int, bufferSize int, opts ...PoolOption[T]) *Pool[T] {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool[T]{
		workers:    workers,
		tasks:      make(chan func(context.Context) (T, error), bufferSize),
		results:    make(chan Result[T], bufferSize),
		done:       make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
		errHandler: func(err error) {}, // 默认错误处理函数什么都不做
		wg:         sync.WaitGroup{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Start 启动工作池
func (p *Pool[T]) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					result, err := task(p.ctx)
					if err != nil && p.errHandler != nil {
						p.errHandler(err)
					}
					// 避免在通道已关闭时发送
					select {
					case <-p.ctx.Done():
						return
					case p.results <- Result[T]{Value: result, Err: err}:
					}
				}
			}
		}()
	}

	// 监控工作池状态
	go func() {
		p.wg.Wait()
		close(p.results)
		close(p.done)
	}()
}

// Submit 提交一个任务到工作池
func (p *Pool[T]) Submit(task func(context.Context) (T, error)) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.tasks <- task:
		return nil
	default:
		// 如果通道已满，返回错误
		return ErrWorkerFull
	}
}

// Results 返回结果通道
func (p *Pool[T]) Results() <-chan Result[T] {
	return p.results
}

// Stop 停止工作池
func (p *Pool[T]) Stop() {
	p.cancel()     // 取消上下文
	close(p.tasks) // 关闭任务通道
	p.wg.Wait()    // 等待所有工作协程完成
}

// WaitForCompletion 等待所有任务完成，带超时机制
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

// Parallel 并行执行多个函数，带上下文和错误处理
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

// Race 并行执行多个函数，返回第一个成功的结果
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

// Debounce 创建一个防抖函数，延迟调用直到等待时间结束
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

// Throttle 创建一个节流函数，在指定时间内最多调用一次
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

// SafeMap 是一个线程安全的映射
type SafeMap[K comparable, V any] struct {
	sync.RWMutex
	data map[K]V
}

// NewSafeMap 创建一个新的安全映射
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

// Set 在映射中设置一个值
func (m *SafeMap[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
}

// Get 从映射中获取一个值
func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

// Delete 从映射中删除一个值
func (m *SafeMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
}
