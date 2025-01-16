package fp

// Pipe 将多个函数串联在一起，将一个函数的输出作为下一个函数的输入
func Pipe[T any](f func(T) T) func(func(T) T) func(T) T {
	return func(g func(T) T) func(T) T {
		return func(x T) T {
			return g(f(x))
		}
	}
}

// Compose 从右到左组合多个函数
func Compose[T any](fns ...func(T) T) func(T) T {
	return func(x T) T {
		result := x
		for _, fn := range fns {
			result = fn(result)
		}
		return result
	}
}

// Curry 将接受多个参数的函数转换为一系列只接受单个参数的函数
func Curry[T, U, V any](f func(T, U) V) func(T) func(U) V {
	return func(t T) func(U) V {
		return func(u U) V {
			return f(t, u)
		}
	}
}

// Memoize 创建函数的记忆化版本，缓存计算结果
func Memoize[T comparable, V any](f func(T) V) func(T) V {
	cache := make(map[T]V)
	return func(t T) V {
		if v, ok := cache[t]; ok {
			return v
		}
		result := f(t)
		cache[t] = result
		return result
	}
}

// Partial 返回一个新函数，其中一些参数已预先填充
func Partial[T, U, V any](f func(T, U) V, t T) func(U) V {
	return func(u U) V {
		return f(t, u)
	}
}

// Chain 允许对任意类型进行方法链式调用
type Chain[T any] struct {
	value T
}

// NewChain 创建一个新的链式调用对象
func NewChain[T any](initial T) Chain[T] {
	return Chain[T]{value: initial}
}

// Map 对链式调用中的值应用函数
func (c Chain[T]) Map(f func(T) T) Chain[T] {
	return Chain[T]{value: f(c.value)}
}

// Value 返回链式调用的最终值
func (c Chain[T]) Value() T {
	return c.value
}

// Either 表示可以是两种类型之一的值
type Either[L, R any] struct {
	left  *L
	right *R
}

// Left 创建一个包含左值的 Either
func Left[L, R any](l L) Either[L, R] {
	return Either[L, R]{left: &l}
}

// Right 创建一个包含右值的 Either
func Right[L, R any](r R) Either[L, R] {
	return Either[L, R]{right: &r}
}

// IsLeft 如果 Either 包含左值则返回 true
func (e Either[L, R]) IsLeft() bool {
	return e.left != nil
}

// IsRight 如果 Either 包含右值则返回 true
func (e Either[L, R]) IsRight() bool {
	return e.right != nil
}

// Match 对 Either 进行模式匹配
func (e Either[L, R]) Match(leftFn func(L), rightFn func(R)) {
	if e.IsLeft() {
		leftFn(*e.left)
	} else {
		rightFn(*e.right)
	}
}

// Option 表示一个可选值
type Option[T any] struct {
	value *T
}

// Some 创建一个包含值的 Option
func Some[T any](t T) Option[T] {
	return Option[T]{value: &t}
}

// None 创建一个不包含值的 Option
func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

// IsSome 如果 Option 包含值则返回 true
func (o Option[T]) IsSome() bool {
	return o.value != nil
}

// IsNone 如果 Option 不包含值则返回 true
func (o Option[T]) IsNone() bool {
	return o.value == nil
}

// Unwrap 返回包含的值，如果没有值则触发 panic
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("对空值调用了 Unwrap")
	}
	return *o.value
}

// UnwrapOr 返回包含的值，如果没有值则返回默认值
func (o Option[T]) UnwrapOr(default_ T) T {
	if o.IsNone() {
		return default_
	}
	return *o.value
}
