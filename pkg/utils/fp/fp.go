package fp

// Pipe chains multiple functions together, passing the output of one to the input of the next
func Pipe[T any](f func(T) T) func(func(T) T) func(T) T {
	return func(g func(T) T) func(T) T {
		return func(x T) T {
			return g(f(x))
		}
	}
}

// Compose composes multiple functions from right to left
func Compose[T any](fns ...func(T) T) func(T) T {
	return func(x T) T {
		result := x
		for _, fn := range fns {
			result = fn(result)
		}
		return result
	}
}

// Curry converts a function that takes multiple arguments into a series of functions that each take a single argument
func Curry[T, U, V any](f func(T, U) V) func(T) func(U) V {
	return func(t T) func(U) V {
		return func(u U) V {
			return f(t, u)
		}
	}
}

// Memoize creates a memoized version of a function
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

// Partial returns a new function with some arguments pre-filled
func Partial[T, U, V any](f func(T, U) V, t T) func(U) V {
	return func(u U) V {
		return f(t, u)
	}
}

// Chain allows method chaining for any type
type Chain[T any] struct {
	value T
}

// NewChain creates a new Chain
func NewChain[T any](initial T) Chain[T] {
	return Chain[T]{value: initial}
}

// Map applies a function to the chain value
func (c Chain[T]) Map(f func(T) T) Chain[T] {
	return Chain[T]{value: f(c.value)}
}

// Value returns the final value in the chain
func (c Chain[T]) Value() T {
	return c.value
}

// Either represents a value that can be one of two types
type Either[L, R any] struct {
	left  *L
	right *R
}

// Left creates a new Either with a left value
func Left[L, R any](l L) Either[L, R] {
	return Either[L, R]{left: &l}
}

// Right creates a new Either with a right value
func Right[L, R any](r R) Either[L, R] {
	return Either[L, R]{right: &r}
}

// IsLeft returns true if the Either contains a left value
func (e Either[L, R]) IsLeft() bool {
	return e.left != nil
}

// IsRight returns true if the Either contains a right value
func (e Either[L, R]) IsRight() bool {
	return e.right != nil
}

// Match pattern matches on an Either
func (e Either[L, R]) Match(leftFn func(L), rightFn func(R)) {
	if e.IsLeft() {
		leftFn(*e.left)
	} else {
		rightFn(*e.right)
	}
}

// Option represents an optional value
type Option[T any] struct {
	value *T
}

// Some creates a new Option with a value
func Some[T any](t T) Option[T] {
	return Option[T]{value: &t}
}

// None creates a new Option with no value
func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

// IsSome returns true if the Option contains a value
func (o Option[T]) IsSome() bool {
	return o.value != nil
}

// IsNone returns true if the Option contains no value
func (o Option[T]) IsNone() bool {
	return o.value == nil
}

// Unwrap returns the contained value or panics if none
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("called unwrap on None value")
	}
	return *o.value
}

// UnwrapOr returns the contained value or a default
func (o Option[T]) UnwrapOr(default_ T) T {
	if o.IsNone() {
		return default_
	}
	return *o.value
}
