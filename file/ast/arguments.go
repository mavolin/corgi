package ast

import "slices"

type Arguments struct {
	LParen *Position
	List   []Argument
	RParen *Position
}

var _ Node = (*Arguments)(nil)

func (a *Arguments) Start() Position {
	if a.LParen != nil {
		return *a.LParen
	}
	for _, arg := range a.List {
		if arg != nil {
			if start := arg.Start(); start != NoPosition {
				return start
			}
		}
	}
	if a.RParen != nil {
		return *a.RParen
	}
	return NoPosition
}

func (a *Arguments) End() Position {
	if a.RParen != nil {
		return deltaPos(*a.RParen, len(")"))
	}
	for _, arg := range slices.Backward(a.List) {
		if arg != nil {
			if end := arg.End(); end != NoPosition {
				return end
			}
		}
	}
	if a.LParen != nil {
		return deltaPos(*a.LParen, len("("))
	}
	return NoPosition
}

func (a *Arguments) Walk(w func(Node)) {
	for _, arg := range a.List {
		if arg != nil {
			w(arg)
		}
	}
}

func (*Arguments) _node() {}

// ============================================================================
// Argument
// ======================================================================================

// Argument is either a [ComponentArgument] or an [Attribute].
type Argument interface {
	Node
	_argument()
}

// if this is changed, change the comment above
var (
	_ Argument = (*ComponentArgument)(nil)
	_ Argument = (Attribute)(nil)
)

// ============================================================================
// Component Argument
// ======================================================================================

type ComponentArgument struct {
	Name  *Identifier
	Colon *Position
	Value *Expression
}

var _ Node = (*ComponentArgument)(nil)

func (a *ComponentArgument) Start() Position {
	if a.Name != nil {
		if a.Name.Start() != NoPosition {
			return a.Name.Start()
		}
	}
	if a.Colon != nil {
		return *a.Colon
	}
	if a.Value != nil {
		if a.Value.Start() != NoPosition {
			return a.Value.Start()
		}
	}
	return NoPosition
}

func (a *ComponentArgument) End() Position {
	if a.Value != nil {
		if a.Value.End() != NoPosition {
			return a.Value.End()
		}
	}
	if a.Colon != nil {
		return deltaPos(*a.Colon, len(":"))
	}
	if a.Name != nil {
		if a.Name.End() != NoPosition {
			return a.Name.End()
		}
	}
	return NoPosition
}

func (a *ComponentArgument) Walk(w func(Node)) {
	if a.Name != nil {
		w(a.Name)
	}
	if a.Value != nil {
		w(a.Value)
	}
}

func (*ComponentArgument) _node()     {}
func (*ComponentArgument) _argument() {}
