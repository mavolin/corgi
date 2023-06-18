// Package stack implements a stack built using a linked list.
package stack

import "github.com/mavolin/corgi/internal/list"

// Stack is a stack that uses a linked list to store its elements.
//
// The zero value for a Stack is an empty stack ready to use.
type Stack[T any] struct {
	l list.List[T]
}

// Push puts elem on top of the stack.
func (s *Stack[T]) Push(elem T) {
	s.l.PushBack(elem)
}

// Pop takes the top most element from the stack and removes it.
//
// If the stack is of length 0, Pop will panic.
func (s *Stack[T]) Pop() T {
	back := s.l.Back()
	defer s.l.Remove(back)
	return back.V()
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
