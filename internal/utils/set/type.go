package set

type Set[E comparable] interface {
	Add(e E) bool
	Contains(e E) bool
	Clear()
	IsEmpty() bool
	Size() int
}

type HashSet[E comparable] struct {
	m map[E]bool
}

func Empty[E comparable]() Set[E] {
	return &HashSet[E]{
		map[E]bool{},
	}
}

func New[E comparable](elements ...E) Set[E] {
	s := Empty[E]()

	for _, e := range elements {
		s.Add(e)
	}

	return s
}

func (hs *HashSet[E]) Add(e E) bool {
	return hs.m[e]
}

func (hs *HashSet[E]) Contains(e E) bool {
	_, ok := hs.m[e]
	return ok
}

func (hs *HashSet[E]) Clear() {
	hs.m = map[E]bool{}
}

func (hs *HashSet[E]) IsEmpty() bool {
	return len(hs.m) == 0
}

func (hs *HashSet[E]) Size() int {
	return len(hs.m)
}
