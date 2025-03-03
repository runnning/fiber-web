package cron

import (
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
type TaskFunc func() error

// Task 定时任务结构
type Task struct {
	Name     string        // 任务名称
	Spec     string        // cron表达式
	Func     TaskFunc      // 执行的函数
	Timeout  time.Duration // 超时时间
	Status   TaskStatus    // 任务状态
	EntryID  cron.EntryID  // cron任务ID
	LastTime time.Time     // 上次执行时间
	mu       sync.Mutex    // 互斥锁
	done     chan struct{} // 用于停止任务的通道
	cancel   chan struct{} // 用于取消任务的通道
	stopping bool          // 标记任务是否正在停止
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

	// 检查任务是否已存在
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

	// 包装任务函数，添加超时控制和错误处理
	wrappedFunc := func() {
		if err := s.runTask(task); err != nil && !errors.Is(err, ErrTaskStopped) {
			s.log.Error("task execution failed",
				zap.String("task", name),
				zap.Error(err),
			)
		}
	}

	// 添加到cron
	entryID, err := s.cron.AddFunc(spec, wrappedFunc)
	if err != nil {
		return err
	}

	task.EntryID = entryID
	s.tasks[name] = task

	s.log.Info("task added successfully",
		zap.String("task", name),
		zap.String("spec", spec),
	)

	return nil
}

// runTask 运行任务
func (s *Scheduler) runTask(task *Task) error {
	task.mu.Lock()
	if task.Status == TaskStatusRunning {
		task.mu.Unlock()
		return ErrTaskIsRunning
	}
	task.Status = TaskStatusRunning
	task.done = make(chan struct{})
	task.cancel = make(chan struct{})
	task.mu.Unlock()

	// 创建一个错误通道
	errCh := make(chan error, 1)

	// 启动任务执行
	go func() {
		select {
		case <-task.cancel: // 任务被取消
			errCh <- ErrTaskStopped
			return
		default:
			errCh <- task.Func()
		}
	}()

	var err error
	// 等待任务完成、超时或被停止
	select {
	case err = <-errCh:
		task.mu.Lock()
		if task.Status == TaskStatusRunning { // 只有在还在运行时才设置为就绪
			task.Status = TaskStatusReady
		}
		task.mu.Unlock()
	case <-time.After(task.Timeout):
		close(task.cancel) // 通知任务 goroutine 退出
		err = ErrTaskTimeout
	case <-task.done:
		close(task.cancel) // 通知任务 goroutine 退出
		err = ErrTaskStopped
	}

	// 清理资源
	task.mu.Lock()
	defer task.mu.Unlock()

	if task.Status == TaskStatusRunning {
		task.Status = TaskStatusStopped
	}
	task.LastTime = time.Now()

	// 只有在通道未关闭时才关闭它们
	select {
	case <-task.done:
		// done 通道已经关闭
	default:
		close(task.done)
	}

	select {
	case <-task.cancel:
		// cancel 通道已经关闭
	default:
		close(task.cancel)
	}

	task.done = nil
	task.cancel = nil

	return err
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

	s.log.Info("task removed",
		zap.String("task", name),
	)

	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
	s.log.Info("scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
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
		return ErrTaskNotRunning
	}

	if task.done != nil {
		close(task.done)
		task.Status = TaskStatusStopped
		task.LastTime = time.Now()
		return nil
	}

	return ErrTaskCannotBeStopped
}
