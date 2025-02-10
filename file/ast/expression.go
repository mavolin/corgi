package ast

import "slices"

// ============================================================================
// Expression
// ======================================================================================

type Expression struct {
	Code Code
}

var _ Node = (*Expression)(nil)

func (e *Expression) Start() Position {
	for _, n := range e.Code {
		if n != nil {
			return n.Start()
		}
	}
	return Position{}
}
func (e *Expression) End() Position {
	for _, n := range slices.Backward(e.Code) {
		if n != nil {
			return n.End()
		}
	}
	return Position{}
}
func (e *Expression) Walk(w func(Node)) {
	if e.Code != nil {
		w(e.Code)
	}
}

func (*Expression) _node() {}
