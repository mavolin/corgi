// Package stack implements a stack built using a linked list.
package stack

// Stack is a slice-backed stack implementation.
//
// The zero value for a Stack is an empty stack ready to use.
type Stack[T any] struct {
	s []T
}

func New[T any](capacity int) *Stack[T] {
	return &Stack[T]{s: make([]T, 0, capacity)}
}

// Push puts elem on top of the stack.
func (s *Stack[T]) Push(elem T) {
	s.s = append(s.s, elem)
}

// Pop takes the top-most element from the stack and removes it.
//
// If the stack is of length 0, Pop will panic.
func (s *Stack[T]) Pop() T {
	elem := s.s[len(s.s)-1]
	s.s = s.s[:len(s.s)-1]
	return elem
}

// Swap swaps the top-most element for elem, returning the element previously
// at the top.
//
// If the stack is of length 0, Swap will panic.
func (s *Stack[T]) Swap(elem T) T {
	old := s.s[len(s.s)-1]
	s.s[len(s.s)-1] = elem
	return old
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
