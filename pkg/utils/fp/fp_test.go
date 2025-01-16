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
		t.Errorf("期望得到 11，实际得到 %v", result)
	}
}

func TestCompose(t *testing.T) {
	add1 := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }

	result := Compose(double, add1)(5) // 先执行 double(5)=10，再执行 add1(10)=11
	if result != 11 {
		t.Errorf("期望得到 11，实际得到 %v", result)
	}
}

func TestOption(t *testing.T) {
	t.Run("包含值的情况", func(t *testing.T) {
		opt := Some(42)
		if !opt.IsSome() {
			t.Error("期望 Some 包含值")
		}
		if opt.IsNone() {
			t.Error("期望 Some 不为空")
		}
		if v := opt.Unwrap(); v != 42 {
			t.Errorf("期望得到 42，实际得到 %v", v)
		}
	})

	t.Run("空值的情况", func(t *testing.T) {
		var opt Option[int]
		if opt.IsSome() {
			t.Error("期望 None 不包含值")
		}
		if !opt.IsNone() {
			t.Error("期望 None 为空")
		}
		if v := opt.UnwrapOr(0); v != 0 {
			t.Errorf("期望得到 0，实际得到 %v", v)
		}
	})
}

func TestEither(t *testing.T) {
	t.Run("右值情况", func(t *testing.T) {
		right := Right[string, int](42)
		if !right.IsRight() {
			t.Error("期望是右值")
		}
		right.Match(
			func(s string) { t.Error("不应该匹配到左值") },
			func(i int) {
				if i != 42 {
					t.Errorf("期望得到 42，实际得到 %v", i)
				}
			},
		)
	})

	t.Run("左值情况", func(t *testing.T) {
		left := Left[string, int]("错误")
		if !left.IsLeft() {
			t.Error("期望是左值")
		}
		left.Match(
			func(s string) {
				if s != "错误" {
					t.Errorf("期望得到 '错误'，实际得到 %v", s)
				}
			},
			func(i int) { t.Error("不应该匹配到右值") },
		)
	})

	t.Run("错误处理示例", func(t *testing.T) {
		divide := func(x, y int) Either[error, int] {
			if y == 0 {
				return Left[error, int](errors.New("除数不能为零"))
			}
			return Right[error, int](x / y)
		}

		if result := divide(42, 2); !result.IsRight() {
			t.Error("期望除法运算成功")
		}

		if result := divide(42, 0); !result.IsLeft() {
			t.Error("期望出现除零错误")
		}
	})
}
