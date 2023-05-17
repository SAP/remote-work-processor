package cache

import "fmt"

type NoSuchElementError[K any] struct {
	key any
}

func NewNoSuchElementError[K any](key K) *NoSuchElementError[K] {
	return &NoSuchElementError[K]{
		key: key,
	}
}

func (e *NoSuchElementError[K]) Error() string {
	return fmt.Sprintf("Value mapped to key '%v' does not exist in cache", e.key)
}
