package cron

import (
	"context"
	"errors"
	"fiber_web/pkg/logger"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusReady   TaskStatus = iota // 就绪
	TaskStatusRunning                   // 运行中
	TaskStatusStopped                   // 已停止
)

// TaskFunc 定时任务函数类型
type TaskFunc func(ctx context.Context) error

// Task 定时任务结构
type Task struct {
	Name     string        // 任务名称
	Spec     string        // cron表达式
	Func     TaskFunc      // 执行的函数
	Timeout  time.Duration // 超时时间
	Status   TaskStatus    // 任务状态
	EntryID  cron.EntryID  // cron任务ID
	LastTime time.Time     // 上次执行时间
	mu       sync.RWMutex  // 读写锁，优化并发访问
	ctx      context.Context
	cancel   context.CancelFunc
}

// Scheduler 调度器
type Scheduler struct {
	cron  *cron.Cron
	tasks map[string]*Task
	log   *logger.Logger
	mu    sync.RWMutex
}

// NewScheduler 创建一个新的调度器
func NewScheduler(logger *logger.Logger) *Scheduler {
	return &Scheduler{
		cron:  cron.New(cron.WithSeconds()),
		tasks: make(map[string]*Task),
		log:   logger,
	}
}

// AddTask 添加定时任务
func (s *Scheduler) AddTask(name, spec string, f TaskFunc, timeout time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[name]; exists {
		return ErrTaskAlreadyExists
	}

	task := &Task{
		Name:    name,
		Spec:    spec,
		Func:    f,
		Timeout: timeout,
		Status:  TaskStatusReady,
	}

	wrappedFunc := func() {
		if err := s.runTask(task); err != nil && !errors.Is(err, ErrTaskStopped) {
			s.log.Error("task execution failed", zap.String("task", name), zap.Error(err))
		}
	}

	entryID, err := s.cron.AddFunc(spec, wrappedFunc)
	if err != nil {
		return err
	}

	task.EntryID = entryID
	s.tasks[name] = task
	s.log.Info("task added successfully", zap.String("task", name), zap.String("spec", spec))
	return nil
}

// runTask 运行任务
func (s *Scheduler) runTask(task *Task) error {
	task.mu.Lock()
	if task.Status == TaskStatusRunning {
		task.mu.Unlock()
		return ErrTaskIsRunning
	}

	// 创建新的上下文和取消函数
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	task.ctx = ctx
	task.cancel = cancel
	task.Status = TaskStatusRunning
	task.mu.Unlock()

	// 确保资源清理
	defer func() {
		task.mu.Lock()
		if task.Status == TaskStatusRunning {
			task.Status = TaskStatusReady
		}
		task.LastTime = time.Now()
		if task.cancel != nil {
			task.cancel()
		}
		task.ctx = nil
		task.cancel = nil
		task.mu.Unlock()
	}()

	// 执行任务
	errCh := make(chan error, 1)
	go func() {
		errCh <- task.Func(ctx)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrTaskTimeout
		}
		return ErrTaskStopped
	}
}

// RemoveTask 移除定时任务
func (s *Scheduler) RemoveTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[name]
	if !exists {
		return ErrTaskNotFound
	}

	s.cron.Remove(task.EntryID)
	delete(s.tasks, name)
	s.log.Info("task removed", zap.String("task", name))
	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
	s.log.Info("scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 停止所有运行中的任务
	for _, task := range s.tasks {
		task.mu.Lock()
		if task.Status == TaskStatusRunning && task.cancel != nil {
			task.cancel()
		}
		task.mu.Unlock()
	}

	s.cron.Stop()
	s.log.Info("scheduler stopped")
}

// GetTask 获取任务信息
func (s *Scheduler) GetTask(name string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[name]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// StopTask 停止正在运行的任务
func (s *Scheduler) StopTask(name string) error {
	s.mu.RLock()
	task, exists := s.tasks[name]
	s.mu.RUnlock()

	if !exists {
		return ErrTaskNotFound
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	if task.Status != TaskStatusRunning {
		return nil
	}

	if task.cancel != nil {
		task.cancel()
	}
	task.Status = TaskStatusStopped

	return nil
}
