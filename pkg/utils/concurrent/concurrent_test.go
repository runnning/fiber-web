package concurrent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	t.Run("基本工作池功能", func(t *testing.T) {
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
				t.Fatalf("提交任务失败: %v", err)
			}
		}

		// Collect results
		results := make(map[int]bool)
		for i := 0; i < 5; i++ {
			result := <-pool.Results()
			if result.Err != nil {
				t.Fatalf("任务返回意外错误: %v", result.Err)
			}
			results[result.Value] = true
		}

		// Verify results
		for i := 0; i < 5; i++ {
			if !results[i*2] {
				t.Errorf("未找到期望的结果 %d: %v", i*2, results)
			}
		}
	})

	t.Run("错误处理", func(t *testing.T) {
		var errorCount atomic.Int32
		pool := NewPool[int](2, 5, WithErrorHandler[int](func(err error) {
			errorCount.Add(1)
		}))
		pool.Start()
		defer pool.Stop()

		expectedError := errors.New("测试错误")
		err := pool.Submit(func(ctx context.Context) (int, error) {
			return 0, expectedError
		})
		if err != nil {
			t.Fatalf("提交任务失败: %v", err)
		}

		result := <-pool.Results()
		if result.Err == nil {
			t.Error("期望得到错误，但得到了 nil")
		} else if !errors.Is(result.Err, expectedError) {
			t.Errorf("期望错误 %v，实际得到: %v (类型: %T)", expectedError, result.Err, result.Err)
		}

		count := errorCount.Load()
		if count != 1 {
			t.Errorf("期望错误处理器被调用一次，实际被调用 %d 次", count)
		}
	})

	t.Run("上下文取消", func(t *testing.T) {
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
			t.Fatalf("提交任务失败: %v", err)
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
			t.Error("任务完成了，但应该被取消")
		default:
			// 检查结果通道
			result := <-pool.Results()
			if result.Err == nil {
				t.Error("期望得到错误，但得到了 nil")
			} else if !errors.Is(result.Err, context.Canceled) {
				t.Errorf("期望得到 context.Canceled 错误，实际得到: %v (类型: %T)", result.Err, result.Err)
			}
		}
	})
}

func TestParallel(t *testing.T) {
	t.Run("并行执行成功", func(t *testing.T) {
		ctx := context.Background()
		fns := []func(context.Context) (int, error){
			func(ctx context.Context) (int, error) { return 1, nil },
			func(ctx context.Context) (int, error) { return 2, nil },
			func(ctx context.Context) (int, error) { return 3, nil },
		}

		results, err := Parallel(ctx, fns...)
		if err != nil {
			t.Fatalf("并行执行失败: %v", err)
		}

		if len(results) != len(fns) {
			t.Fatalf("期望得到 %d 个结果，实际得到 %d 个", len(fns), len(results))
		}

		expected := []int{1, 2, 3}
		for i, result := range results {
			if result.Err != nil {
				t.Errorf("任务 %d 返回意外错误: %v", i, result.Err)
				continue
			}
			if result.Value != expected[i] {
				t.Errorf("任务 %d: 期望值 %d，实际得到 %d", i, expected[i], result.Value)
			}
		}
	})

	t.Run("错误会取消其他任务", func(t *testing.T) {
		ctx := context.Background()
		errorTask := errors.New("任务错误")

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
			t.Error("期望并行执行出错")
		}

		if len(results) != 2 {
			t.Errorf("期望得到 2 个结果，实际得到 %d 个", len(results))
		}

		if !errors.Is(results[0].Err, errorTask) {
			t.Errorf("期望第一个任务返回错误 %v，实际得到 %v", errorTask, results[0].Err)
		}
	})
}

func TestRace(t *testing.T) {
	t.Run("返回第一个成功的结果", func(t *testing.T) {
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
			t.Errorf("竞争执行失败: %v", err)
		}
		if result != 2 {
			t.Errorf("期望得到结果 2，实际得到 %d", result)
		}
	})

	t.Run("所有任务都失败时返回错误", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.New("测试错误")
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
			t.Error("期望竞争执行出错")
		}
		if !errors.Is(err, expectedError) {
			t.Errorf("期望错误 %v，实际得到 %v", expectedError, err)
		}
	})
}

func TestSafeMap(t *testing.T) {
	t.Run("并发操作", func(t *testing.T) {
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
