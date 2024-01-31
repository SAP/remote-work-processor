package array

func Contains[T comparable](arr []T, searched T) bool {
	for _, e := range arr {
		if e == searched {
			return true
		}
	}
	return false
}
