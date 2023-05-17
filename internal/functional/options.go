package functional

type Option[T any] func(t *T)
type OptionWithError[T any] func(t *T) error
