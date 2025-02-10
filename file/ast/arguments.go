package ast

import "slices"

type Arguments struct {
	LParen *Position
	Args   []Argument
	RParen *Position
}

var _ Node = (*Arguments)(nil)

func (a *Arguments) Start() Position {
	if a.LParen != nil {
		return *a.LParen
	}
	for _, arg := range a.Args {
		if arg != nil {
			return arg.Start()
		}
	}
	if a.RParen != nil {
		return *a.RParen
	}
	return Position{}
}

func (a *Arguments) End() Position {
	if a.RParen != nil {
		return deltaPos(*a.RParen, len(")"))
	}
	for _, arg := range slices.Backward(a.Args) {
		if arg != nil {
			return arg.End()
		}
	}
	if a.LParen != nil {
		return deltaPos(*a.LParen, len("("))
	}
	return Position{}
}

func (a *Arguments) Walk(w func(Node)) {
	for _, arg := range a.Args {
		if arg != nil {
			w(arg)
		}
	}
}

func (*Arguments) _node() {}

// ============================================================================
// Argument
// ======================================================================================

// Argument is either a [ComponentArgument], or an [Attribute].
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
	Name  *Ident
	Colon *Position
	Value *Expression
}

var _ Node = (*ComponentArgument)(nil)

func (a *ComponentArgument) Start() Position {
	if a.Name != nil {
		return a.Name.Start()
	} else if a.Colon != nil {
		return *a.Colon
	} else if a.Value != nil {
		return a.Value.Start()
	}
	return Position{}
}
func (a *ComponentArgument) End() Position {
	if a.Value != nil {
		return a.Value.End()
	} else if a.Colon != nil {
		return deltaPos(*a.Colon, len(":"))
	} else if a.Name != nil {
		return a.Name.End()
	}
	return Position{}
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
