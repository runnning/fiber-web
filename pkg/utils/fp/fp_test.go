package fp

import (
	"errors"
	"testing"
)

func TestPipe(t *testing.T) {
	add1 := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }

	f := Pipe(double)(add1)
	result := f(5)
	if result != 11 {
		t.Errorf("Expected 11, got %v", result)
	}
}

func TestCompose(t *testing.T) {
	add1 := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }

	result := Compose(double, add1)(5) // 先 double(5)=10, 再 add1(10)=11
	if result != 11 {
		t.Errorf("Expected 11, got %v", result)
	}
}

func TestOption(t *testing.T) {
	t.Run("Some value", func(t *testing.T) {
		opt := Some(42)
		if !opt.IsSome() {
			t.Error("Expected Some to be present")
		}
		if opt.IsNone() {
			t.Error("Expected Some not to be None")
		}
		if v := opt.Unwrap(); v != 42 {
			t.Errorf("Expected 42, got %v", v)
		}
	})

	t.Run("None value", func(t *testing.T) {
		var opt Option[int]
		if opt.IsSome() {
			t.Error("Expected None not to be present")
		}
		if !opt.IsNone() {
			t.Error("Expected None to be None")
		}
		if v := opt.UnwrapOr(0); v != 0 {
			t.Errorf("Expected 0, got %v", v)
		}
	})
}

func TestEither(t *testing.T) {
	t.Run("Right value", func(t *testing.T) {
		right := Right[string, int](42)
		if !right.IsRight() {
			t.Error("Expected Right to be right")
		}
		right.Match(
			func(s string) { t.Error("Should not match Left") },
			func(i int) {
				if i != 42 {
					t.Errorf("Expected 42, got %v", i)
				}
			},
		)
	})

	t.Run("Left value", func(t *testing.T) {
		left := Left[string, int]("error")
		if !left.IsLeft() {
			t.Error("Expected Left to be left")
		}
		left.Match(
			func(s string) {
				if s != "error" {
					t.Errorf("Expected 'error', got %v", s)
				}
			},
			func(i int) { t.Error("Should not match Right") },
		)
	})

	t.Run("Error handling", func(t *testing.T) {
		divide := func(x, y int) Either[error, int] {
			if y == 0 {
				return Left[error, int](errors.New("division by zero"))
			}
			return Right[error, int](x / y)
		}

		if result := divide(42, 2); !result.IsRight() {
			t.Error("Expected successful division")
		}

		if result := divide(42, 0); !result.IsLeft() {
			t.Error("Expected division by zero error")
		}
	})
}
