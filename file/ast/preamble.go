package ast

// ============================================================================
// Package Directive
// ======================================================================================

type PackageDirective struct {
	Name     *Ident // package name
	Position Position
}

var _ Node = (*PackageDirective)(nil)

func (d *PackageDirective) Pos() Position { return d.Position }
func (d *PackageDirective) End() Position {
	if d.Name != nil {
		return d.Name.End()
	}
	return deltaPos(d.Position, len("package"))
}

func (*PackageDirective) _node() {}

// ============================================================================
// Import
// ======================================================================================

type Import struct {
	LParen  *Position // nil if this is a single-line import
	Imports []ImportNode
	RParen  *Position // nil if this is a single-line import

	Position Position
}

var _ ScopeNode = (*Import)(nil)

func (i *Import) Pos() Position { return i.Position }
func (i *Import) End() Position {
	if i.RParen != nil {
		return deltaPos(*i.RParen, 1)
	} else if len(i.Imports) > 0 {
		return i.Imports[len(i.Imports)-1].End()
	} else if i.LParen != nil {
		return deltaPos(*i.LParen, 1)
	}
	return deltaPos(i.Position, len("import"))
}

func (*Import) _node()      {}
func (*Import) _scopeNode() {}

// ============================================================================
// Import Scope Node
// ======================================================================================

// ImportNode is a pointer to either an [ImportSpec], a [DevComment], or a
// [BadImportSpec].
type ImportNode interface {
	Node
	_importNode()
}

// if this is changed, change the comment above
var (
	_ ImportNode = (*ImportSpec)(nil)
	_ ImportNode = (*DevComment)(nil)
	_ ImportNode = (*BadImportSpec)(nil)
)

// ==================================== Import Spec =====================================

type ImportSpec struct {
	// Alias is the alias of the import, if any.
	Alias *Ident
	Path  *StaticString
}

var _ ImportNode = (*ImportSpec)(nil)

func (s *ImportSpec) Pos() Position {
	if s.Alias != nil {
		return s.Alias.Pos()
	}
	return s.Path.Pos()
}
func (s *ImportSpec) End() Position {
	if s.Path != nil {
		return s.Path.End()
	} else if s.Alias != nil {
		return s.Alias.End()
	}
	return InvalidPosition
}

func (*ImportSpec) _node()       {}
func (*ImportSpec) _importNode() {}

// ================================== Bad Import Sepc ===================================

type BadImportSpec struct {
	Line     string
	Position Position
}

var _ ImportNode = (*BadImportSpec)(nil)

func (s *BadImportSpec) Pos() Position { return s.Position }
func (s *BadImportSpec) End() Position { return deltaPos(s.Position, len(s.Line)) }

func (*BadImportSpec) _node()       {}
func (*BadImportSpec) _importNode() {}

// ============================================================================
// State
// ======================================================================================

type State struct {
	LParen *Position // nil if this is a single-line state
	Vars   []StateNode
	RParen *Position // nil if this is a single-line state

	Position Position
}

var _ ScopeNode = (*State)(nil)

func (s *State) Pos() Position { return s.Position }
func (s *State) End() Position {
	if s.RParen != nil {
		return deltaPos(*s.RParen, 1)
	} else if len(s.Vars) > 0 {
		return s.Vars[len(s.Vars)-1].End()
	} else if s.LParen != nil {
		return deltaPos(*s.LParen, 1)
	}
	return deltaPos(s.Position, len("state"))
}

func (*State) _node()      {}
func (*State) _scopeNode() {}

// ================================== State Scope Node ==================================

// StateNode is a pointer to either a [StateVar], a [DevComment], or a
// [BadStateVar].
type StateNode interface {
	Node
	_stateNode()
}

// if this is changed, change the comment above
var (
	_ StateNode = (*StateVar)(nil)
	_ StateNode = (*DevComment)(nil)
	_ StateNode = (*BadStateVar)(nil)
)

// ==================================== State Var =======================================

type StateVar struct {
	Names []*Ident
	Type  *Type // nil if type is inferred

	Assign *Position // nil if this has no default value
	Values []*GoCode // empty if no default value
}

var _ StateNode = (*StateVar)(nil)

func (v *StateVar) Pos() Position {
	if len(v.Names) > 0 {
		return v.Names[0].Pos()
	}
	return InvalidPosition
}
func (v *StateVar) End() Position {
	if len(v.Values) > 0 {
		return v.Values[len(v.Values)-1].End()
	} else if v.Assign != nil {
		return deltaPos(*v.Assign, 1)
	} else if v.Type != nil {
		return v.Type.End()
	} else if len(v.Names) > 0 {
		return v.Names[len(v.Names)-1].End()
	}
	return InvalidPosition
}

func (*StateVar) _node()      {}
func (*StateVar) _stateNode() {}

// =================================== Bad State Var ====================================

type BadStateVar struct {
	Line     string
	Position Position
}

var _ StateNode = (*BadStateVar)(nil)

func (v *BadStateVar) Pos() Position { return v.Position }
func (v *BadStateVar) End() Position { return deltaPos(v.Position, len(v.Line)) }

func (*BadStateVar) _node()      {}
func (*BadStateVar) _stateNode() {}
