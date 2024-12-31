package errorx

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Error 表示一个带有堆栈跟踪和附加上下文的自定义错误
type Error struct {
	Err       error                  // 原始错误
	Stack     []string               // 堆栈跟踪信息
	Context   map[string]interface{} // 错误上下文信息
	Code      int                    // 错误码
	Message   string                 // 错误消息
	Operation string                 // 操作名称
}

// New 创建一个新的 Error
func New(message string) *Error {
	return &Error{
		Message: message,
		Stack:   getStack(),
		Context: make(map[string]interface{}),
	}
}

// Errorf 使用格式化字符串创建新的 Error
func Errorf(format string, args ...interface{}) *Error {
	return New(fmt.Sprintf(format, args...))
}

// Wrap 使用附加上下文包装现有错误
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	var xerr *Error
	if errors.As(err, &xerr) {
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

// Wrapf 使用格式化消息包装现有错误
func Wrapf(err error, format string, args ...interface{}) *Error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// WithCode 添加错误码
func (e *Error) WithCode(code int) *Error {
	e.Code = code
	return e
}

// WithOperation 添加操作名称
func (e *Error) WithOperation(op string) *Error {
	e.Operation = op
	return e
}

// WithContext 添加上下文信息
func (e *Error) WithContext(key string, value interface{}) *Error {
	e.Context[key] = value
	return e
}

// Error 实现 error 接口
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

// Unwrap 实现 errors.Unwrap 接口
func (e *Error) Unwrap() error {
	return e.Err
}

// Is 实现 errors.Is 接口
func (e *Error) Is(target error) bool {
	var t *Error
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Message == t.Message
}

// StackTrace 返回堆栈跟踪信息字符串
func (e *Error) StackTrace() string {
	return strings.Join(e.Stack, "\n")
}

// GetCode 获取错误码，如果需要会遍历包装的错误
func (e *Error) GetCode() int {
	if e.Code != 0 {
		return e.Code
	}
	if e.Err != nil {
		var xerr *Error
		if errors.As(e.Err, &xerr) {
			return xerr.GetCode()
		}
	}
	return 0
}

// getStack 返回调用堆栈信息
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

// Try 执行一个函数并在发生 panic 时返回错误
// 泛型参数 T 表示函数返回值类型
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

// Must 如果 err 不为 nil 则触发 panic
// 泛型参数 T 表示返回值类型
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// IsCode 检查错误（或其包装的错误）是否具有指定的错误码
func IsCode(err error, code int) bool {
	var xerr *Error
	if errors.As(err, &xerr) {
		return xerr.GetCode() == code
	}
	return false
}
