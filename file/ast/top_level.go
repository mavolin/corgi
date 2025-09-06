package ast

import "slices"

// ============================================================================
// Top Level
// ======================================================================================

type TopLevel []TopLevelNode

var _ Node = (*TopLevel)(nil)

func (t TopLevel) Start() Position {
	for _, node := range t {
		if node != nil {
			if pos := node.Start(); pos != NoPosition {
				return pos
			}
		}
	}
	return NoPosition
}

func (t TopLevel) End() Position {
	for _, node := range slices.Backward(t) {
		if node != nil {
			if pos := node.End(); pos != NoPosition {
				return pos
			}
		}
	}
	return NoPosition
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
