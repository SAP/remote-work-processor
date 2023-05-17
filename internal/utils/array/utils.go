package array

import "github.com/SAP/remote-work-processor/internal/functional"

func Map[T any, R any](arr []T, m functional.Function[T, R]) (res []R) {
	res = make([]R, len(arr))
	for i, e := range arr {
		res[i] = m(e)
	}

	return
}

func Filter[T any](arr []T, p functional.Predicate[T]) (filtered []T) {
	filtered = []T{}
	for _, e := range arr {
		if p(e) {
			filtered = append(filtered, e)
		}
	}

	return
}

func Contains[T comparable](arr []T, searched T) bool {
	for _, e := range arr {
		if e == searched {
			return true
		}
	}

	return false
}
