package ast

// ============================================================================
// Top Level
// ======================================================================================

type TopLevel []TopLevelNode

var _ Node = (*TopLevel)(nil)

func (t TopLevel) Start() Position {
	if len(t) == 0 {
		return Position{}
	}
	return t[0].Start()
}

func (t TopLevel) End() Position {
	if len(t) == 0 {
		return Position{}
	}
	return t[len(t)-1].End()
}

func (t TopLevel) Walk(w func(Node)) {
	for _, node := range t {
		if node != nil {
			w(node)
		}
	}
}

func (TopLevel) _node() {}

// ============================================================================
// Top Level Node
// ======================================================================================

// TopLevelNode is a [Component], [StateDeclaration], [AttributeDefinition],
// [ElementDefinition], or [Statement].
type TopLevelNode interface {
	Node
	_topLevelNode()
}
