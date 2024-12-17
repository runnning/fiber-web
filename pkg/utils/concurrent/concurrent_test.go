package concurrent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	t.Run("basic worker pool functionality", func(t *testing.T) {
		pool := NewPool[int](3, 5)
		pool.Start()
		defer pool.Stop()

		// Submit tasks
		for i := 0; i < 5; i++ {
			i := i // capture loop variable
			err := pool.Submit(func(ctx context.Context) (int, error) {
				return i * 2, nil
			})
			if err != nil {
				t.Fatalf("Failed to submit task: %v", err)
			}
		}

		// Collect results
		results := make(map[int]bool)
		for i := 0; i < 5; i++ {
			result := <-pool.Results()
			if result.Err != nil {
				t.Fatalf("Task returned unexpected error: %v", result.Err)
			}
			results[result.Value] = true
		}

		// Verify results
		for i := 0; i < 5; i++ {
			if !results[i*2] {
				t.Errorf("Expected result %d not found in results: %v", i*2, results)
			}
		}
	})

	t.Run("error handling", func(t *testing.T) {
		var errorCount atomic.Int32
		pool := NewPool[int](2, 5, WithErrorHandler[int](func(err error) {
			errorCount.Add(1)
		}))
		pool.Start()
		defer pool.Stop()

		expectedError := errors.New("test error")
		err := pool.Submit(func(ctx context.Context) (int, error) {
			return 0, expectedError
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}

		result := <-pool.Results()
		if result.Err == nil {
			t.Error("Expected error, got nil")
		} else if !errors.Is(result.Err, expectedError) {
			t.Errorf("Expected error %v, got: %v (type: %T)", expectedError, result.Err, result.Err)
		}

		count := errorCount.Load()
		if count != 1 {
			t.Errorf("Expected error handler to be called once, got %d calls", count)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		pool := NewPool[int](2, 5)
		pool.Start()

		taskStarted := make(chan struct{})
		taskCompleted := make(chan struct{})

		// Submit a long-running task
		err := pool.Submit(func(ctx context.Context) (int, error) {
			close(taskStarted) // 通知任务已开始
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(time.Second * 5):
				close(taskCompleted)
				return 42, nil
			}
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}

		// 等待任务开始执行
		<-taskStarted

		// 确保任务正在运行中
		time.Sleep(time.Millisecond * 100)

		// 停止工作池，这应该会取消正在执行的任务
		pool.Stop()

		// 验证任务被取消
		select {
		case <-taskCompleted:
			t.Error("Task completed when it should have been cancelled")
		default:
			// 检查结果通道
			result := <-pool.Results()
			if result.Err == nil {
				t.Error("Expected error, got nil")
			} else if !errors.Is(result.Err, context.Canceled) {
				t.Errorf("Expected context.Canceled error, got: %v (type: %T)", result.Err, result.Err)
			}
		}
	})
}

func TestParallel(t *testing.T) {
	t.Run("successful parallel execution", func(t *testing.T) {
		ctx := context.Background()
		fns := []func(context.Context) (int, error){
			func(ctx context.Context) (int, error) { return 1, nil },
			func(ctx context.Context) (int, error) { return 2, nil },
			func(ctx context.Context) (int, error) { return 3, nil },
		}

		results, err := Parallel(ctx, fns...)
		if err != nil {
			t.Fatalf("Parallel execution failed: %v", err)
		}

		if len(results) != len(fns) {
			t.Fatalf("Expected %d results, got %d", len(fns), len(results))
		}

		expected := []int{1, 2, 3}
		for i, result := range results {
			if result.Err != nil {
				t.Errorf("Task %d returned unexpected error: %v", i, result.Err)
				continue
			}
			if result.Value != expected[i] {
				t.Errorf("Task %d: expected value %d, got %d", i, expected[i], result.Value)
			}
		}
	})

	t.Run("error cancels other tasks", func(t *testing.T) {
		ctx := context.Background()
		errorTask := errors.New("task error")

		fns := []func(context.Context) (int, error){
			func(ctx context.Context) (int, error) {
				time.Sleep(time.Millisecond * 100)
				return 0, errorTask
			},
			func(ctx context.Context) (int, error) {
				<-ctx.Done()
				return 0, ctx.Err()
			},
		}

		results, err := Parallel(ctx, fns...)
		if err == nil {
			t.Error("Expected error from parallel execution")
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if !errors.Is(results[0].Err, errorTask) {
			t.Errorf("Expected first task to return error %v, got %v", errorTask, results[0].Err)
		}
	})
}

func TestRace(t *testing.T) {
	t.Run("returns first successful result", func(t *testing.T) {
		ctx := context.Background()
		fns := []func(context.Context) (int, error){
			func(ctx context.Context) (int, error) {
				time.Sleep(time.Millisecond * 100)
				return 1, nil
			},
			func(ctx context.Context) (int, error) {
				return 2, nil
			},
		}

		result, err := Race(ctx, fns...)
		if err != nil {
			t.Errorf("Race failed: %v", err)
		}
		if result != 2 {
			t.Errorf("Expected result 2, got %d", result)
		}
	})

	t.Run("returns error when all tasks fail", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("test error")
		fns := []func(context.Context) (int, error){
			func(ctx context.Context) (int, error) {
				return 0, expectedError
			},
			func(ctx context.Context) (int, error) {
				return 0, expectedError
			},
		}

		_, err := Race(ctx, fns...)
		if err == nil {
			t.Error("Expected error from race")
		}
		if !errors.Is(err, expectedError) {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
	})
}

func TestSafeMap(t *testing.T) {
	t.Run("concurrent operations", func(t *testing.T) {
		sm := NewSafeMap[string, int]()
		done := make(chan bool)

		// Writer goroutine
		go func() {
			for i := 0; i < 100; i++ {
				sm.Set("counter", i)
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()

		// Reader goroutine
		go func() {
			for i := 0; i < 100; i++ {
				sm.Get("counter")
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()

		// Wait for both goroutines
		<-done
		<-done

		// Final verification
		if _, exists := sm.Get("counter"); !exists {
			t.Error("Counter should exist in map")
		}
	})

	t.Run("range operation", func(t *testing.T) {
		sm := NewSafeMap[string, int]()
		sm.Set("one", 1)
		sm.Set("two", 2)
		sm.Set("three", 3)

		count := 0
		sum := 0
		sm.Range(func(k string, v int) bool {
			count++
			sum += v
			return true
		})

		if count != 3 {
			t.Errorf("Expected to iterate over 3 items, got %d", count)
		}
		if sum != 6 {
			t.Errorf("Expected sum of values to be 6, got %d", sum)
		}
	})
}

func TestDebounce(t *testing.T) {
	t.Run("debounce function calls", func(t *testing.T) {
		var callCount atomic.Int32
		fn := func(x int) {
			callCount.Add(1)
		}

		debounced := Debounce(fn, time.Millisecond*100)

		// Rapid calls
		for i := 0; i < 5; i++ {
			debounced(i)
		}

		// Wait for debounce
		time.Sleep(time.Millisecond * 200)

		if callCount.Load() != 1 {
			t.Errorf("Expected 1 call, got %d", callCount.Load())
		}
	})
}

func TestThrottle(t *testing.T) {
	t.Run("throttle function calls", func(t *testing.T) {
		var callCount atomic.Int32
		interval := time.Millisecond * 100

		fn := func(x int) {
			callCount.Add(1)
		}

		throttled := Throttle(fn, interval)

		// 执行三次调用，间隔小于节流时间
		throttled(1) // 第一次调用应该执行
		time.Sleep(interval / 2)
		throttled(2) // 应该被跳过
		time.Sleep(interval)
		throttled(3) // 应该执行

		time.Sleep(interval / 2) // 等待最后的调用完成

		if count := callCount.Load(); count != 2 {
			t.Errorf("Expected exactly 2 calls, got %d calls", count)
		}
	})
}
