package set

type Set[K comparable] interface {
	Add(k K)
	Contains(k K) bool
	Clear()
}
