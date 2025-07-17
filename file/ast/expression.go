package ast

type Expression struct {
	Nodes Code
}

var _ Node = (*Expression)(nil)

func (e *Expression) Start() Position {
	if e.Nodes != nil {
		return e.Nodes.Start()
	}
	return Position{}
}

func (e *Expression) End() Position {
	if e.Nodes != nil {
		return e.End()
	}
	return Position{}
}

func (e *Expression) Walk(w func(Node)) {
	if e.Nodes != nil {
		w(e.Nodes)
	}
}

func (*Expression) _node() {}
