package functional

type Consumer[T any] func(t T)
type Predicate[T any] func(t T) bool
type Supplier[T any] func() T
type Function[T any, R any] func(t T) R
type UnaryOperator[T any] Function[T, T]

type BiConsumer[T any, U any] func(t T, u U)
type BiPredicate[T any, U any] func(t T, u U)
type BiSupplier[T any, U any] func() (T, U)
type BiFunction[T any, U any, R any] func(t T, u U) R
type BinaryOperator[T any] BiFunction[T, T, T]
