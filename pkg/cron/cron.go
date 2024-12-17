package cron

import (
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// TaskFunc 定时任务函数类型
type TaskFunc func() error

// Task 定时任务结构
type Task struct {
	Name     string        // 任务名称
	Spec     string        // cron表达式
	Func     TaskFunc      // 执行的函数
	Timeout  time.Duration // 超时时间
	Running  bool          // 是否正在运行
	EntryID  cron.EntryID  // cron任务ID
	LastTime time.Time     // 上次执行时间
	mu       sync.Mutex    // 互斥锁
}

// Scheduler 调度器
type Scheduler struct {
	cron  *cron.Cron
	tasks map[string]*Task
	log   *zap.Logger
	mu    sync.RWMutex
}

// NewScheduler 创建一个新的调度器
func NewScheduler(logger *zap.Logger) *Scheduler {
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
	}

	// 包装任务函数，添加超时控制和错误处理
	wrappedFunc := func() {
		if err := s.runTask(task); err != nil {
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
	if task.Running {
		task.mu.Unlock()
		return ErrTaskIsRunning
	}
	task.Running = true
	task.mu.Unlock()

	// 确保任务结束时重置状态
	defer func() {
		task.mu.Lock()
		task.Running = false
		task.LastTime = time.Now()
		task.mu.Unlock()
	}()

	// 创建一个带超时的context
	done := make(chan error, 1)
	go func() {
		done <- task.Func()
	}()

	// 等待任务完成或超时
	select {
	case err := <-done:
		return err
	case <-time.After(task.Timeout):
		return ErrTaskTimeout
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
