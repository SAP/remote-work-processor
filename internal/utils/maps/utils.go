package maps

import "github.com/SAP/remote-work-processor/internal/utils/tuple"

func Pairs[K comparable, V any](m map[K]V) []tuple.Pair[K, V] {
	pairs := make([]tuple.Pair[K, V], len(m))
	var i int32

	for k, v := range m {
		pairs[i] = tuple.PairOf(k, v)
		i++
	}

	return pairs
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	var i int32

	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))
	var i int32

	for _, v := range m {
		values[i] = v
		i++
	}

	return values
}
