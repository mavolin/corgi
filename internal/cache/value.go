package cache

type Value[T any] struct {
	v    T
	done chan struct{}
}

// Preload starts a goroutine to compute the value and then returns a handle
// to that value.
func Preload[T any](compute func() T) *Value[T] {
	v := &Value[T]{
		done: make(chan struct{}),
	}
	go func() {
		v.v = compute()
		close(v.done)
	}()
	return v
}

func (v *Value[T]) Get() T {
	<-v.done
	return v.v
}
