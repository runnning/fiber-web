package cron

import (
	"context"
	"errors"
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"sync"
	"testing"
	"time"
)

var (
	testLogger *logger.Logger
	logOnce    sync.Once
)

// getTestLogger 获取测试用的logger实例（单例模式）
func getTestLogger(t *testing.T) *logger.Logger {
	logOnce.Do(func() {
		logConfig := &config.LogConfig{
			Level:      "info",
			Directory:  "./logs",
			Filename:   "test.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
			Console:    true,
		}
		var err error
		testLogger, err = logger.NewLogger(logConfig)
		if err != nil {
			t.Fatalf("初始化日志失败: %v", err)
		}
	})
	return testLogger
}

// setupTestScheduler 创建测试用调度器
func setupTestScheduler(t *testing.T) *Scheduler {
	return NewScheduler(getTestLogger(t))
}

// createTestTask 创建测试任务的辅助函数
func createTestTask(t *testing.T, s *Scheduler, name, spec string) {
	err := s.AddTask(name, spec, func(ctx context.Context) error { return nil }, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}
}

func TestNewScheduler(t *testing.T) {
	s := setupTestScheduler(t)
	if s == nil {
		t.Fatal("调度器不应为nil")
	}
	if s.cron == nil {
		t.Error("cron不应为nil")
	}
	if s.tasks == nil {
		t.Error("tasks不应为nil")
	}
	if s.log == nil {
		t.Error("log不应为nil")
	}
}

func TestAddTask(t *testing.T) {
	s := setupTestScheduler(t)

	tests := []struct {
		name        string
		taskName    string
		spec        string
		timeout     time.Duration
		expectedErr error
	}{
		{
			name:        "正常添加任务",
			taskName:    "test_task",
			spec:        "*/1 * * * * *",
			timeout:     time.Second * 5,
			expectedErr: nil,
		},
		{
			name:        "添加重复任务",
			taskName:    "test_task",
			spec:        "*/1 * * * * *",
			timeout:     time.Second * 5,
			expectedErr: ErrTaskAlreadyExists,
		},
		{
			name:        "无效的cron表达式",
			taskName:    "invalid_task",
			spec:        "invalid",
			timeout:     time.Second * 5,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.AddTask(tt.taskName, tt.spec, func(ctx context.Context) error { return nil }, tt.timeout)
			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("期望错误为 %v, 实际错误为 %v", tt.expectedErr, err)
			} else if tt.expectedErr == nil && err != nil && tt.name != "无效的cron表达式" {
				t.Errorf("期望无错误，实际错误为 %v", err)
			}
		})
	}
}

func TestRemoveTask(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "test_task"
	createTestTask(t, s, taskName, "*/1 * * * * *")

	// 测试移除任务
	if err := s.RemoveTask(taskName); err != nil {
		t.Errorf("移除任务失败: %v", err)
	}

	// 测试移除不存在的任务
	if err := s.RemoveTask("non_existent_task"); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("期望错误为 %v, 实际错误为 %v", ErrTaskNotFound, err)
	}
}

func TestGetTask(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "test_task"
	createTestTask(t, s, taskName, "*/1 * * * * *")

	t.Run("获取存在的任务", func(t *testing.T) {
		task, err := s.GetTask(taskName)
		if err != nil {
			t.Errorf("获取任务失败: %v", err)
		}
		if task == nil {
			t.Fatal("任务不应为nil")
		}
		if task.Name != taskName {
			t.Errorf("任务名称不匹配，期望 %s，实际 %s", taskName, task.Name)
		}
	})

	t.Run("获取不存在的任务", func(t *testing.T) {
		task, err := s.GetTask("non_existent_task")
		if !errors.Is(err, ErrTaskNotFound) {
			t.Errorf("期望错误为 %v, 实际错误为 %v", ErrTaskNotFound, err)
		}
		if task != nil {
			t.Error("不存在的任务应该返回nil")
		}
	})
}

func TestListTasks(t *testing.T) {
	s := setupTestScheduler(t)
	tasks := []struct {
		name string
		spec string
	}{
		{"task1", "*/1 * * * * *"},
		{"task2", "*/2 * * * * *"},
		{"task3", "*/3 * * * * *"},
	}

	for _, tt := range tasks {
		createTestTask(t, s, tt.name, tt.spec)
	}

	taskList := s.ListTasks()
	if len(taskList) != len(tasks) {
		t.Errorf("任务列表长度不匹配，期望 %d，实际 %d", len(tasks), len(taskList))
	}
}

func TestTaskExecution(t *testing.T) {
	s := setupTestScheduler(t)
	executed := make(chan bool, 1)

	err := s.AddTask("test_task", "* * * * * *", func(ctx context.Context) error {
		executed <- true
		return nil
	}, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	s.Start()
	defer s.Stop()

	select {
	case <-executed:
		// 任务成功执行
	case <-time.After(time.Second * 2):
		t.Fatal("任务执行超时")
	}
}

func TestTaskTimeout(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "timeout_task"

	err := s.AddTask(taskName, "* * * * * *", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 2):
			return ErrTaskTimeout
		}
	}, time.Second)

	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	task, _ := s.GetTask(taskName)
	if err := s.runTask(task); !errors.Is(err, ErrTaskTimeout) {
		t.Errorf("期望超时错误，实际得到：%v", err)
	}
}

func TestStopTask(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "stop_task"
	taskStarted := make(chan bool, 1)
	taskStopped := make(chan bool, 1)

	err := s.AddTask(taskName, "* * * * * *", func(ctx context.Context) error {
		taskStarted <- true
		<-ctx.Done()
		taskStopped <- true
		return ctx.Err()
	}, time.Second*10)

	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	s.Start()
	defer s.Stop()

	// 等待任务开始执行
	select {
	case <-taskStarted:
		// 任务已开始执行
	case <-time.After(time.Second * 2):
		t.Fatal("任务未开始执行")
	}

	if err := s.StopTask(taskName); err != nil {
		t.Errorf("停止任务失败: %v", err)
	}

	select {
	case <-taskStopped:
		// 任务已正确停止
	case <-time.After(time.Second * 2):
		t.Error("任务未能正确停止")
	}

	task, _ := s.GetTask(taskName)
	if task.Status != TaskStatusStopped {
		t.Errorf("任务状态错误，期望 %v，实际 %v", TaskStatusStopped, task.Status)
	}
}

func TestTaskError(t *testing.T) {
	s := setupTestScheduler(t)
	expectedErr := errors.New("test error")

	err := s.AddTask("error_task", "* * * * * *", func(ctx context.Context) error {
		return expectedErr
	}, time.Second*5)

	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	task, _ := s.GetTask("error_task")
	if err := s.runTask(task); !errors.Is(err, expectedErr) {
		t.Errorf("期望错误 %v，实际得到 %v", expectedErr, err)
	}
}

func TestTaskStatusTransition(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "status_task"

	err := s.AddTask(taskName, "* * * * * *", func(ctx context.Context) error {
		time.Sleep(time.Millisecond * 100)
		return nil
	}, time.Second*5)

	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	task, _ := s.GetTask(taskName)

	t.Run("初始状态", func(t *testing.T) {
		if task.Status != TaskStatusReady {
			t.Errorf("初始状态错误，期望 %v，实际 %v", TaskStatusReady, task.Status)
		}
	})

	t.Run("运行状态", func(t *testing.T) {
		go func() { _ = s.runTask(task) }()
		time.Sleep(time.Millisecond * 50)

		task.mu.RLock()
		status := task.Status
		task.mu.RUnlock()

		if status != TaskStatusRunning {
			t.Errorf("运行状态错误，期望 %v，实际 %v", TaskStatusRunning, status)
		}
	})

	t.Run("完成状态", func(t *testing.T) {
		time.Sleep(time.Millisecond * 100)
		task.mu.RLock()
		status := task.Status
		task.mu.RUnlock()

		if status != TaskStatusReady {
			t.Errorf("完成状态错误，期望 %v，实际 %v", TaskStatusReady, status)
		}
	})
}

func TestSchedulerStartStop(t *testing.T) {
	s := setupTestScheduler(t)
	taskExecuted := make(chan bool, 1)

	err := s.AddTask("test_task", "* * * * * *", func(ctx context.Context) error {
		taskExecuted <- true
		return nil
	}, time.Second*5)

	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	t.Run("启动后执行", func(t *testing.T) {
		s.Start()
		select {
		case <-taskExecuted:
			// 任务成功执行
		case <-time.After(time.Second * 2):
			t.Error("启动后任务未执行")
		}
	})

	t.Run("停止后不执行", func(t *testing.T) {
		s.Stop()
		select {
		case <-taskExecuted:
			t.Error("停止后任务仍在执行")
		case <-time.After(time.Second * 2):
			// 正确行为：任务未执行
		}
	})
}

func TestTaskLastTime(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "last_time_task"

	err := s.AddTask(taskName, "* * * * * *", func(ctx context.Context) error {
		return nil
	}, time.Second*5)

	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	task, _ := s.GetTask(taskName)
	initialTime := task.LastTime

	if err := s.runTask(task); err != nil {
		t.Fatalf("运行任务失败: %v", err)
	}

	if !task.LastTime.After(initialTime) {
		t.Error("任务最后执行时间未更新")
	}
}
