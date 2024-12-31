package errorx

import (
	"errors"
	"testing"
)

// TestNew 测试创建新错误
func TestNew(t *testing.T) {
	err := New("test error")
	if err == nil {
		t.Fatal("期望错误不为 nil")
	}
	if err.Message != "test error" {
		t.Errorf("期望错误消息为 'test error'，实际得到 %v", err.Message)
	}
	if len(err.Stack) == 0 {
		t.Error("期望包含堆栈跟踪信息")
	}
}

// TestErrorf 测试格式化创建错误
func TestErrorf(t *testing.T) {
	err := Errorf("test error %d", 42)
	if err.Message != "test error 42" {
		t.Errorf("期望错误消息为 'test error 42'，实际得到 %v", err.Message)
	}
}

// TestWrap 测试错误包装
func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "wrapped message")
	if wrapped.Message != "wrapped message" {
		t.Errorf("期望错误消息为 'wrapped message'，实际得到 %v", wrapped.Message)
	}
	if !errors.Is(original, wrapped.Err) {
		t.Error("期望包装的错误包含原始错误")
	}
}

// TestWrapf 测试格式化错误包装
func TestWrapf(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrapf(original, "wrapped message %d", 42)
	if wrapped.Message != "wrapped message 42" {
		t.Errorf("期望错误消息为 'wrapped message 42'，实际得到 %v", wrapped.Message)
	}
}

// TestErrorChaining 测试错误链式调用
func TestErrorChaining(t *testing.T) {
	err := New("base error").
		WithCode(404).
		WithOperation("GetUser").
		WithContext("userId", 123)

	if err.Code != 404 {
		t.Errorf("期望错误码为 404，实际得到 %d", err.Code)
	}
	if err.Operation != "GetUser" {
		t.Errorf("期望操作名为 'GetUser'，实际得到 %s", err.Operation)
	}
	if val, ok := err.Context["userId"]; !ok || val != 123 {
		t.Errorf("期望上下文中 userId=123，实际得到 %v", val)
	}
}

// TestErrorString 测试错误字符串格式化
func TestErrorString(t *testing.T) {
	err := New("not found").
		WithCode(404).
		WithOperation("GetUser")

	expected := "[GetUser] (404) not found"
	if err.Error() != expected {
		t.Errorf("期望错误字符串为 '%s'，实际得到 '%s'", expected, err.Error())
	}
}

// TestUnwrap 测试错误解包
func TestUnwrap(t *testing.T) {
	inner := New("inner error")
	outer := Wrap(inner, "outer error")

	if !errors.Is(inner, errors.Unwrap(outer)) {
		t.Error("解包后未得到内部错误")
	}
}

// TestIs 测试错误比较
func TestIs(t *testing.T) {
	err1 := New("test error").WithCode(404)
	err2 := New("test error").WithCode(404)
	err3 := New("different error").WithCode(500)

	if !errors.Is(err1, err2) {
		t.Error("期望错误相等")
	}
	if errors.Is(err1, err3) {
		t.Error("期望错误不相等")
	}
}

// TestGetCode 测试获取错误码
func TestGetCode(t *testing.T) {
	err1 := New("error").WithCode(404)
	err2 := Wrap(err1, "wrapped")

	if err2.GetCode() != 404 {
		t.Errorf("期望错误码为 404，实际得到 %d", err2.GetCode())
	}
}

// TestIsCode 测试错误码检查
func TestIsCode(t *testing.T) {
	err := New("error").WithCode(404)
	wrapped := Wrap(err, "wrapped")

	if !IsCode(wrapped, 404) {
		t.Error("期望错误码检查返回 true")
	}
	if IsCode(wrapped, 500) {
		t.Error("期望错误码检查返回 false")
	}
}

// TestTry 测试 Try 函数
func TestTry(t *testing.T) {
	t.Run("成功执行", func(t *testing.T) {
		result, err := Try(func() int {
			return 42
		})
		if err != nil {
			t.Errorf("期望无错误，实际得到 %v", err)
		}
		if result != 42 {
			t.Errorf("期望结果为 42，实际得到 %d", result)
		}
	})

	t.Run("panic 恢复", func(t *testing.T) {
		_, err := Try(func() int {
			panic("test panic")
		})
		if err == nil {
			t.Error("期望从 panic 中得到错误")
		}
		if err.Error() != "test panic" {
			t.Errorf("期望错误消息为 'test panic'，实际得到 '%v'", err)
		}
	})
}

// TestMust 测试 Must 函数
func TestMust(t *testing.T) {
	t.Run("成功执行", func(t *testing.T) {
		result := Must(42, nil)
		if result != 42 {
			t.Errorf("期望结果为 42，实际得到 %d", result)
		}
	})

	t.Run("panic 触发", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望触发 panic")
			}
		}()
		Must(0, errors.New("test error"))
	})
}

// TestStackTrace 测试堆栈跟踪
func TestStackTrace(t *testing.T) {
	err := New("test error")
	stack := err.StackTrace()
	if stack == "" {
		t.Error("期望非空堆栈跟踪")
	}
	if len(err.Stack) == 0 {
		t.Error("期望非空堆栈切片")
	}
}
