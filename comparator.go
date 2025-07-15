package ziptree

type LessFn[T any] func(a, b T) bool

type Comparator[T any] interface {
	LessThan(a, b T) bool
	GreaterThan(a, b T) bool
	LessThanOrEqual(a, b T) bool
	GreaterThanOrEqual(a, b T) bool
	Equals(a, b T) bool
}

func (fn LessFn[T]) LessThan(a, b T) bool {
	// a < b
	return fn(a, b)
}

func (fn LessFn[T]) GreaterThan(a, b T) bool {
	// a > b -> b < a
	return fn(b, a)
}

func (fn LessFn[T]) LessThanOrEqual(a, b T) bool {
	// a <= b -> !(b < a)
	return !fn(b, a)
}

func (fn LessFn[T]) GreaterThanOrEqual(a, b T) bool {
	// a >= b -> !(a < b)
	return !fn(a, b)
}

func (fn LessFn[T]) Equals(a, b T) bool {
	// a == b -> !(a < b) && !(b < a)
	return !fn(a, b) && !fn(b, a)
}
