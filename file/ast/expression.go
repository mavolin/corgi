package ast

type Expression struct {
	Nodes Code
}

var _ Node = (*Expression)(nil)

func (e *Expression) Start() Position {
	if e.Nodes != nil {
		if start := e.Nodes.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (e *Expression) End() Position {
	if e.Nodes != nil {
		if end := e.Nodes.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (e *Expression) Walk(w func(Node)) {
	if e.Nodes != nil {
		w(e.Nodes)
	}
}

func (*Expression) _node() {}
