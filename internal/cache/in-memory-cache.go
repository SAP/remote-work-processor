package cache

import (
	"sync"
)

type InMemoryCache[K comparable, V any] struct {
	sync.RWMutex
	entries map[K]V
}

func NewInMemoryCache[K comparable, V any]() *InMemoryCache[K, V] {
	return &InMemoryCache[K, V]{
		entries: make(map[K]V),
	}
}

func (c *InMemoryCache[K, V]) FromMap(m map[K]V) MapCache[K, V] {
	if m == nil {
		c.entries = map[K]V{}
	} else {
		for k, v := range m {
			c.entries[k] = v
		}
	}

	return c
}

func (c *InMemoryCache[K, V]) ToMap() map[K]V {
	return c.entries
}

func (c *InMemoryCache[K, V]) Read(k K) V {
	c.RLock()
	defer c.RUnlock()

	v, ok := c.entries[k]
	if !ok {
		return *new(V)
	}

	return v
}

func (c *InMemoryCache[K, V]) Write(k K, v V) V {
	c.Lock()
	defer c.Unlock()

	c.entries[k] = v
	return v
}

func (c *InMemoryCache[K, V]) Remove(k K) {
	c.Lock()
	defer c.Unlock()

	delete(c.entries, k)
}

func (c *InMemoryCache[K, V]) Size() int {
	return len(c.entries)
}
