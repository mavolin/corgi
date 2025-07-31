package ast

import "slices"

// ============================================================================
// State Declaration
// ======================================================================================

type StateDeclaration struct {
	State  *Position
	LParen *Position // nil if this is a single-line state
	Specs  []*StateSpec
	RParen *Position // nil if this is a single-line state
}

var _ TopLevelNode = (*StateDeclaration)(nil)

func (s *StateDeclaration) Start() Position {
	if s.State != nil {
		return *s.State
	} else if s.LParen != nil {
		return *s.LParen
	}
	for _, spec := range s.Specs {
		if spec != nil {
			return spec.Start()
		}
	}
	if s.RParen != nil {
		return *s.RParen
	}
	return Position{}
}

func (s *StateDeclaration) End() Position {
	if s.RParen != nil {
		return deltaPos(*s.RParen, len(")"))
	}
	for _, spec := range slices.Backward(s.Specs) {
		if spec != nil {
			return spec.End()
		}
	}
	if s.LParen != nil {
		return deltaPos(*s.LParen, len("("))
	} else if s.State != nil {
		return deltaPos(*s.State, len("state"))
	}
	return Position{}
}

func (s *StateDeclaration) Walk(w func(Node)) {
	for _, spec := range s.Specs {
		if spec != nil {
			w(spec)
		}
	}
}

func (s *StateDeclaration) Highlight() (start, end Position) {
	start, end = s.Start(), s.End()
	if s.LParen == nil && start.Line >= end.Line-3 {
		return start, end
	}
	if s.State != nil {
		return *s.State, deltaPos(*s.State, len("state"))
	}
	return start, end
}

func (*StateDeclaration) _node()         {}
func (*StateDeclaration) _topLevelNode() {}

// ============================================================================
// StateDeclaration Var
// ======================================================================================

type StateSpec struct {
	Names []*Identifier
	Type  *Type // nil if type is inferred

	EqualSign *Position     // nil if this has no default value
	Values    []*Expression // empty if no default value
}

var _ Node = (*StateSpec)(nil)

func (v *StateSpec) Start() Position {
	for _, name := range v.Names {
		if name != nil {
			return name.Start()
		}
	}
	if v.Type != nil {
		return v.Type.Start()
	} else if v.EqualSign != nil {
		return *v.EqualSign
	}
	for _, value := range v.Values {
		if value != nil {
			return value.Start()
		}
	}
	return Position{}
}

func (v *StateSpec) End() Position {
	for _, value := range slices.Backward(v.Values) {
		if value != nil {
			return value.End()
		}
	}
	if v.EqualSign != nil {
		return deltaPos(*v.EqualSign, len("="))
	} else if v.Type != nil {
		return v.Type.End()
	}
	for _, name := range slices.Backward(v.Names) {
		if name != nil {
			return name.End()
		}
	}
	return Position{}
}

func (v *StateSpec) Walk(w func(Node)) {
	for _, name := range v.Names {
		if name != nil {
			w(name)
		}
	}
	if v.Type != nil {
		w(v.Type)
	}
	for _, value := range v.Values {
		if value != nil {
			w(value)
		}
	}
}

func (*StateSpec) _node() {}
