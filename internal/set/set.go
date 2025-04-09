package set

type Set[K comparable] interface {
	Add(k K) bool
	Contains(k K) bool
	Clear()
}
