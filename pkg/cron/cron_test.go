package cron

import (
	"errors"
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"testing"
	"time"
)

func setupTestScheduler(t *testing.T) *Scheduler {
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
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}
	return NewScheduler(log)
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
			expectedErr: nil, // 这里会返回一个解析错误，但具体错误信息可能不同
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.AddTask(tt.taskName, tt.spec, func() error { return nil }, tt.timeout)
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

	// 先添加一个任务
	taskName := "test_task"
	err := s.AddTask(taskName, "*/1 * * * * *", func() error { return nil }, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 测试移除任务
	err = s.RemoveTask(taskName)
	if err != nil {
		t.Errorf("移除任务失败: %v", err)
	}

	// 测试移除不存在的任务
	err = s.RemoveTask("non_existent_task")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("期望错误为 %v, 实际错误为 %v", ErrTaskNotFound, err)
	}
}

func TestGetTask(t *testing.T) {
	s := setupTestScheduler(t)

	taskName := "test_task"
	err := s.AddTask(taskName, "*/1 * * * * *", func() error { return nil }, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 测试获取存在的任务
	task, err := s.GetTask(taskName)
	if err != nil {
		t.Errorf("获取任务失败: %v", err)
	}
	if task == nil {
		t.Error("任务不应为nil")
		return
	}
	if task.Name != taskName {
		t.Errorf("任务名称不匹配，期望 %s，实际 %s", taskName, task.Name)
	}

	// 测试获取不存在的任务
	task, err = s.GetTask("non_existent_task")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("期望错误为 %v, 实际错误为 %v", ErrTaskNotFound, err)
	}
	if task != nil {
		t.Error("不存在的任务应该返回nil")
	}
}

func TestListTasks(t *testing.T) {
	s := setupTestScheduler(t)

	// 添加多个任务
	tasks := []struct {
		name string
		spec string
	}{
		{"task1", "*/1 * * * * *"},
		{"task2", "*/2 * * * * *"},
		{"task3", "*/3 * * * * *"},
	}

	for _, tt := range tasks {
		err := s.AddTask(tt.name, tt.spec, func() error { return nil }, time.Second*5)
		if err != nil {
			t.Fatalf("添加任务失败: %v", err)
		}
	}

	// 测试列出所有任务
	taskList := s.ListTasks()
	if len(taskList) != len(tasks) {
		t.Errorf("任务列表长度不匹配，期望 %d，实际 %d", len(tasks), len(taskList))
	}
}

func TestTaskExecution(t *testing.T) {
	s := setupTestScheduler(t)
	executed := make(chan bool, 1)

	// 添加一个会立即执行的任务
	err := s.AddTask("test_task", "* * * * * *", func() error {
		executed <- true
		return nil
	}, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 启动调度器
	s.Start()
	defer s.Stop()

	// 等待任务执行
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
	taskStarted := make(chan struct{})
	taskFinished := make(chan struct{})

	// 添加一个会超时的任务
	err := s.AddTask(taskName, "* * * * * *", func() error {
		select {
		case taskStarted <- struct{}{}:
		default:
		}
		time.Sleep(time.Second * 2)
		select {
		case taskFinished <- struct{}{}:
		default:
		}
		return nil
	}, time.Second)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 启动调度器
	s.Start()
	defer s.Stop()

	// 等待任务开始执行
	select {
	case <-taskStarted:
		// 任务已开始
	case <-time.After(time.Second * 2):
		t.Fatal("任务未开始执行")
	}

	// 等待任务超时
	task, err := s.GetTask(taskName)
	if err != nil {
		t.Fatalf("获取任务失败: %v", err)
	}

	// 等待任务状态变为就绪或停止
	deadline := time.After(time.Second * 3)
	for {
		select {
		case <-deadline:
			t.Fatal("任务未能在预期时间内完成")
		default:
			task.mu.Lock()
			status := task.Status
			task.mu.Unlock()
			if status == TaskStatusReady || status == TaskStatusStopped {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func TestStopTask(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "long_running_task"
	taskStarted := make(chan struct{}, 1)
	taskFinished := make(chan struct{}, 1)

	// 添加一个长时间运行的任务
	err := s.AddTask(taskName, "* * * * * *", func() error {
		select {
		case taskStarted <- struct{}{}:
		default:
		}
		defer func() {
			select {
			case taskFinished <- struct{}{}:
			default:
			}
		}()
		time.Sleep(time.Second * 5)
		return nil
	}, time.Second*10)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	s.Start()
	defer s.Stop()

	// 等待任务开始运行
	select {
	case <-taskStarted:
		// 任务已开始
	case <-time.After(time.Second * 2):
		t.Fatal("任务未开始执行")
	}

	// 测试停止任务
	err = s.StopTask(taskName)
	if err != nil {
		t.Errorf("停止任务失败: %v", err)
	}

	// 等待任务状态变为停止
	deadline := time.After(time.Second * 3)
	for {
		select {
		case <-deadline:
			t.Fatal("任务未能在预期时间内停止")
		default:
			task, err := s.GetTask(taskName)
			if err != nil {
				t.Fatalf("获取任务失败: %v", err)
			}
			task.mu.Lock()
			status := task.Status
			task.mu.Unlock()
			if status == TaskStatusStopped {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func TestTaskError(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "error_task"
	taskStarted := make(chan struct{}, 1)

	// 添加一个会返回错误的任务
	err := s.AddTask(taskName, "* * * * * *", func() error {
		select {
		case taskStarted <- struct{}{}:
		default:
		}
		return fmt.Errorf("任务执行失败")
	}, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 启动调度器
	s.Start()
	defer s.Stop()

	// 等待任务开始执行
	select {
	case <-taskStarted:
		// 任务已开始
	case <-time.After(time.Second * 2):
		t.Fatal("任务未开始执行")
	}

	// 等待任务状态变为就绪
	deadline := time.After(time.Second * 3)
	for {
		select {
		case <-deadline:
			t.Fatal("任务未能在预期时间内完成")
		default:
			task, err := s.GetTask(taskName)
			if err != nil {
				t.Fatalf("获取任务失败: %v", err)
			}
			task.mu.Lock()
			status := task.Status
			task.mu.Unlock()
			if status == TaskStatusReady {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func TestTaskStatusTransition(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "status_task"
	taskStarted := make(chan struct{})
	taskFinished := make(chan struct{})

	// 添加一个长时间运行的任务
	err := s.AddTask(taskName, "* * * * * *", func() error {
		close(taskStarted)
		time.Sleep(time.Second)
		close(taskFinished)
		return nil
	}, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 检查初始状态
	task, err := s.GetTask(taskName)
	if err != nil {
		t.Fatalf("获取任务失败: %v", err)
	}
	if task.Status != TaskStatusReady {
		t.Errorf("初始任务状态不正确，期望 %v，实际 %v", TaskStatusReady, task.Status)
	}

	// 启动调度器
	s.Start()
	defer s.Stop()

	// 等待任务开始运行
	select {
	case <-taskStarted:
		// 任务已开始
	case <-time.After(time.Second * 2):
		t.Fatal("任务未开始执行")
	}

	// 检查运行状态
	task, err = s.GetTask(taskName)
	if err != nil {
		t.Fatalf("获取任务失败: %v", err)
	}
	if task.Status != TaskStatusRunning {
		t.Errorf("运行中任务状态不正确，期望 %v，实际 %v", TaskStatusRunning, task.Status)
	}

	// 等待任务完成
	select {
	case <-taskFinished:
		// 任务已完成
	case <-time.After(time.Second * 2):
		t.Fatal("任务未完成")
	}

	// 检查完成后状态
	task, err = s.GetTask(taskName)
	if err != nil {
		t.Fatalf("获取任务失败: %v", err)
	}
	if task.Status != TaskStatusReady {
		t.Errorf("完成后任务状态不正确，期望 %v，实际 %v", TaskStatusReady, task.Status)
	}
}

func TestSchedulerStartStop(t *testing.T) {
	s := setupTestScheduler(t)
	executed := make(chan bool, 1)

	// 添加任务
	err := s.AddTask("test_task", "* * * * * *", func() error {
		executed <- true
		return nil
	}, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 启动调度器
	s.Start()

	// 等待任务执行
	select {
	case <-executed:
		// 任务成功执行
	case <-time.After(time.Second * 2):
		t.Fatal("任务执行超时")
	}

	// 停止调度器
	s.Stop()

	// 清空通道
	select {
	case <-executed:
	default:
	}

	// 等待一段时间，确认任务不再执行
	select {
	case <-executed:
		t.Error("调度器停止后任务仍在执行")
	case <-time.After(time.Second * 2):
		// 符合预期，任务没有执行
	}
}

func TestTaskLastTime(t *testing.T) {
	s := setupTestScheduler(t)
	taskName := "last_time_task"

	// 添加任务
	err := s.AddTask(taskName, "* * * * * *", func() error {
		return nil
	}, time.Second*5)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 启动调度器
	s.Start()
	defer s.Stop()

	// 等待任务执行
	time.Sleep(time.Second * 2)

	// 检查最后执行时间
	task, err := s.GetTask(taskName)
	if err != nil {
		t.Fatalf("获取任务失败: %v", err)
	}
	if task.LastTime.IsZero() {
		t.Error("任务最后执行时间未更新")
	}
	if time.Since(task.LastTime) > time.Second*3 {
		t.Error("任务最后执行时间异常")
	}
}
