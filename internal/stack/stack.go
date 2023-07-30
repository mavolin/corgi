// Package stack implements a stack built using a linked list.
package stack

import "github.com/mavolin/corgi/internal/list"

// Stack is a stack that uses a linked list to store its elements.
//
// The zero value for a Stack is an empty stack ready to use.
type Stack[T any] struct {
	l list.List[T]
}

func New1[T any](t T) *Stack[T] {
	var s Stack[T]
	s.Push(t)
	return &s
}

// Push puts elem on top of the stack.
func (s *Stack[T]) Push(elem T) {
	s.l.PushBack(elem)
}

// Pop takes the top-most element from the stack and removes it.
//
// If the stack is of length 0, Pop will panic.
func (s *Stack[T]) Pop() T {
	return s.l.Remove(s.l.Back())
}

// Swap swaps the top-most element for elem, returning the element previously
// at the top.
//
// If the stack is of length 0, Swap will panic.
func (s *Stack[T]) Swap(elem T) T {
	old := s.Pop()
	s.Push(elem)
	return old
}

// Peek returns the element at the top of the stack without removing it.
//
// If the stack is of length 0, Peek will panic.
func (s *Stack[T]) Peek() T {
	return s.l.Back().V()
}

// Len returns the length of the stack.
func (s *Stack[T]) Len() int {
	return s.l.Len()
}
