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
	}
	if s.LParen != nil {
		return *s.LParen
	}
	for _, spec := range s.Specs {
		if spec != nil {
			if start := spec.Start(); start != NoPosition {
				return start
			}
		}
	}
	if s.RParen != nil {
		return *s.RParen
	}
	return NoPosition
}

func (s *StateDeclaration) End() Position {
	if s.RParen != nil {
		return deltaPos(*s.RParen, len(")"))
	}
	for _, spec := range slices.Backward(s.Specs) {
		if spec != nil {
			if end := spec.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.LParen != nil {
		return deltaPos(*s.LParen, len("("))
	}
	if s.State != nil {
		return deltaPos(*s.State, len("state"))
	}
	return NoPosition
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
			if start := name.Start(); start != NoPosition {
				return start
			}
		}
	}
	if v.Type != nil {
		if start := v.Type.Start(); start != NoPosition {
			return start
		}
	}
	if v.EqualSign != nil {
		return *v.EqualSign
	}
	for _, value := range v.Values {
		if value != nil {
			if start := value.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (v *StateSpec) End() Position {
	for _, value := range slices.Backward(v.Values) {
		if value != nil {
			if end := value.End(); end != NoPosition {
				return end
			}
		}
	}
	if v.EqualSign != nil {
		return deltaPos(*v.EqualSign, len("="))
	}
	if v.Type != nil {
		if end := v.Type.End(); end != NoPosition {
			return end
		}
	}
	for _, name := range slices.Backward(v.Names) {
		if name != nil {
			if end := name.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
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
