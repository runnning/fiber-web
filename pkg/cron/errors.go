package cron

import "errors"

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrTaskIsRunning     = errors.New("task is already running")
	ErrTaskTimeout       = errors.New("task execution timeout")
	ErrTaskStopped       = errors.New("task stopped")
)
