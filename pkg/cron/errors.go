package cron

import "errors"

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrTaskAlreadyExists   = errors.New("task already exists")
	ErrTaskIsRunning       = errors.New("task is already running")
	ErrTaskTimeout         = errors.New("task execution timeout")
	ErrTaskNotRunning      = errors.New("task is not running")
	ErrTaskCannotBeStopped = errors.New("task cannot be stopped")
)
