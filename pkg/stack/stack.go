// Package stack provides a stack datatype built using a resized slice.
package stack

// Stack is a stack using a resized slice to store its elements.
//
// The zero value for a Stack is an empty stack ready to use.
type Stack[T any] struct {
	s []T
}

// New creates a new Stack with the given initial capacity.
func New[T any](initialCap int) Stack[T] {
	return Stack[T]{s: make([]T, 0, initialCap)}
}

// Push puts elem on top of the stack.
func (s *Stack[T]) Push(elem T) {
	s.s = append(s.s, elem)
}

// Pop takes the top most element from the stack and removes it.
//
// If the stack is of length 0, Pop will panic.
func (s *Stack[T]) Pop() T {
	i := len(s.s) - 1

	elem := s.s[i]
	s.s = s.s[:i]
	return elem
}

// Peek returns the element at the top of the stack without removing it.
//
// If the stack is of length 0, Peek will panic.
func (s *Stack[T]) Peek() T {
	return s.s[len(s.s)-1]
}

// Len returns the length of the stack.
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
