package errorx

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("test error")
	if err == nil {
		t.Fatal("Expected non-nil error")
	}
	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got %v", err.Message)
	}
	if len(err.Stack) == 0 {
		t.Error("Expected stack trace")
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "wrapped message")
	if wrapped.Message != "wrapped message" {
		t.Errorf("Expected message 'wrapped message', got %v", wrapped.Message)
	}
	if wrapped.Err != original {
		t.Error("Expected wrapped error to contain original error")
	}
}

func TestErrorChaining(t *testing.T) {
	err := New("base error").
		WithCode(404).
		WithOperation("GetUser").
		WithContext("userId", 123)

	if err.Code != 404 {
		t.Errorf("Expected code 404, got %d", err.Code)
	}
	if err.Operation != "GetUser" {
		t.Errorf("Expected operation 'GetUser', got %s", err.Operation)
	}
	if val, ok := err.Context["userId"]; !ok || val != 123 {
		t.Errorf("Expected context userId=123, got %v", val)
	}
}

func TestErrorString(t *testing.T) {
	err := New("not found").
		WithCode(404).
		WithOperation("GetUser")

	expected := "[GetUser] (404) not found"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestTry(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		result, err := Try(func() int {
			return 42
		})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("panic recovery", func(t *testing.T) {
		_, err := Try(func() int {
			panic("test panic")
		})
		if err == nil {
			t.Error("Expected error from panic")
		}
		if err.Error() != "test panic" {
			t.Errorf("Expected 'test panic', got '%v'", err)
		}
	})
}

func TestMust(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		result := Must(42, nil)
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("panic on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic")
			}
		}()
		Must(0, errors.New("test error"))
	})
}

func TestStackTrace(t *testing.T) {
	err := New("test error")
	stack := err.StackTrace()
	if stack == "" {
		t.Error("Expected non-empty stack trace")
	}
	if len(err.Stack) == 0 {
		t.Error("Expected non-empty stack slice")
	}
}
