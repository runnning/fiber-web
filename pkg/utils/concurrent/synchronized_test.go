package concurrent

import (
	"testing"
)

func TestSynchronized(t *testing.T) {
	t.Run("基本同步操作", func(t *testing.T) {
		sync := NewSynchronized(0)

		// 设置值
		sync.Set(42)

		// 获取值
		if value := sync.Get(); value != 42 {
			t.Errorf("期望得到 42，实际得到 %v", value)
		}
	})

	t.Run("并发安全性", func(t *testing.T) {
		sync := NewSynchronized(0)
		done := make(chan bool)

		// 启动多个 goroutine 进行并发操作
		for i := 0; i < 100; i++ {
			go func() {
				sync.Set(42)
				done <- true
			}()
		}

		// 等待所有 goroutine 完成
		for i := 0; i < 100; i++ {
			<-done
		}

		if value := sync.Get(); value != 42 {
			t.Errorf("期望得到 42，实际得到 %v", value)
		}
	})
}
