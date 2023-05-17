package cache

type Cache[K comparable, V any] interface {
	Read(k K) V
	Write(k K, v V) V
	Remove(k K)
	Size() int
}

type MapCache[K comparable, V any] interface {
	Cache[K, V]
	FromMap(m map[K]V) MapCache[K, V]
	ToMap() map[K]V
}
