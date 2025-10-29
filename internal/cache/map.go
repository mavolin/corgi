package cache

import "sync"

type (
	Map[K comparable, V any] struct {
		mut sync.Mutex
		m   map[K]*mapValue[V]
	}

	mapValue[T any] struct {
		v    T
		done chan struct{}
	}
)

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		m: make(map[K]*mapValue[V]),
	}
}

func (c *Map[K, V]) Get(key K, compute func() V) V {
	c.mut.Lock()
	if v, ok := c.m[key]; ok {
		c.mut.Unlock()
		<-v.done
		return v.v
	}

	v := mapValue[V]{done: make(chan struct{})}
	c.m[key] = &v
	c.mut.Unlock()

	v.v = compute()
	close(v.done)
	return v.v
}

func (c *Map[K, V]) Preload(key K, compute func() V) {
	c.mut.Lock()
	if _, ok := c.m[key]; ok {
		c.mut.Unlock()
		return
	}

	v := &mapValue[V]{done: make(chan struct{})}
	c.m[key] = v
	c.mut.Unlock()

	go func() {
		v.v = compute()
		close(v.done)
	}()
}
