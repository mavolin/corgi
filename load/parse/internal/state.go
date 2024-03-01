package internal

import (
	"github.com/mavolin/corgi/file/ast"
)

type (
	State struct {
		Indentation IndentationState

		Start []ast.Position

		// whether the upcoming tokens must be inline
		Inline bool
	}

	IndentationState struct {
		Target  int
		Current int
	}
)

func newState(c *current) {
	c.state["state"] = State{
		Start: make([]ast.Position, 0, 16),
	}
}

func state(c *current) State {
	return c.state["state"].(State)
}

func editState(c *current, f func(*State)) {
	s := state(c)
	f(&s)
	c.state["state"] = s
}

func pushStart(c *current) {
	editState(c, func(s *State) { s.Start = append(s.Start, pos(c)) })
}

func peekStart(c *current) ast.Position {
	return state(c).Start[len(state(c).Start)-1]
}

func popStart(c *current) ast.Position {
	var start ast.Position
	editState(c, func(s *State) {
		start = s.Start[len(s.Start)-1]
		s.Start = s.Start[:len(s.Start)-1]
	})
	return start
}
