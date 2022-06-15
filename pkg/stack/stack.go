// Package stack provides a stack datatype built using a resized slice.
package stack

// Stack is a stack using a resized slice to store its elements.
//
// The zero value for a Stack is an empty stack ready to use.
type Stack[T any] struct {
	s []T
}

func New[T any](initialSize int) Stack[T] {
	return Stack[T]{s: make([]T, 0, initialSize)}
}

func (s *Stack[T]) Push(elem T) {
	s.s = append(s.s, elem)
}

// Swap swaps the top element for the passed element.
func (s *Stack[T]) Swap(elem T) {
	s.s[len(s.s)-1] = elem
}

func (s *Stack[T]) Pop() T {
	i := len(s.s) - 1

	elem := s.s[i]
	s.s = s.s[:i]
	return elem
}

func (s *Stack[T]) Peek() T {
	return s.s[len(s.s)-1]
}

func (s *Stack[T]) Len() int {
	return len(s.s)
}

// Clone clones the stack, returning a Stack with a copy of the underlying
// slice.
func (s *Stack[T]) Clone() Stack[T] {
	cp := make([]T, len(s.s))
	copy(cp, s.s)

	return Stack[T]{s: cp}
}
