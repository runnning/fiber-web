package errorx

import (
	"fmt"
	"runtime"
	"strings"
)

// Error represents a custom error with stack trace and additional context
type Error struct {
	Err       error
	Stack     []string
	Context   map[string]interface{}
	Code      int
	Message   string
	Operation string
}

// New creates a new Error
func New(message string) *Error {
	return &Error{
		Message: message,
		Stack:   getStack(),
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	var xerr *Error
	if as(err, &xerr) {
		return &Error{
			Err:     xerr,
			Message: message,
			Stack:   getStack(),
			Context: make(map[string]interface{}),
		}
	}

	return &Error{
		Err:     err,
		Message: message,
		Stack:   getStack(),
		Context: make(map[string]interface{}),
	}
}

// WithCode adds an error code
func (e *Error) WithCode(code int) *Error {
	e.Code = code
	return e
}

// WithOperation adds an operation name
func (e *Error) WithOperation(op string) *Error {
	e.Operation = op
	return e
}

// WithContext adds context information
func (e *Error) WithContext(key string, value interface{}) *Error {
	e.Context[key] = value
	return e
}

// Error implements the error interface
func (e *Error) Error() string {
	var b strings.Builder
	if e.Operation != "" {
		b.WriteString(fmt.Sprintf("[%s] ", e.Operation))
	}
	if e.Code != 0 {
		b.WriteString(fmt.Sprintf("(%d) ", e.Code))
	}
	b.WriteString(e.Message)
	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

// StackTrace Stack returns the call stack as a string
func (e *Error) StackTrace() string {
	return strings.Join(e.Stack, "\n")
}

// getStack returns the call stack
func getStack() []string {
	var stack []string
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stack
}

// as is a helper function that wraps errors.As
func as(err error, target interface{}) bool {
	return fmt.Sprintf("%T", err) == fmt.Sprintf("%T", target)
}

// Try executes a function and returns an error if it panics
func Try[T any](f func() T) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	result = f()
	return result, nil
}

// Must panics if err is not nil
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
