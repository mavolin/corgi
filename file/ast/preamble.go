package ast

// ============================================================================
// Package Directive
// ======================================================================================

type PackageDirective struct {
	Name     *Ident // package name
	Position Position
}

// ============================================================================
// Import
// ======================================================================================

type Import struct {
	LParen  *Position // nil if this is a single-line import
	Imports []ImportScopeItem
	RParen  *Position // nil if this is a single-line import

	Position Position
}

var _ ScopeItem = (*Import)(nil)

func (i *Import) Pos() Position { return i.Position }
func (*Import) _scopeItem()     {}

// ============================================================================
// Import Scope Item
// ======================================================================================

// ImportScopeItem is a pointer to either an [ImportSpec], a [DevComment], or a
// [BadImportSpec].
type ImportScopeItem interface {
	_importScopeItem()
	Poser
}

// if this is changed, change the comment above
var (
	_ ImportScopeItem = (*ImportSpec)(nil)
	_ ImportScopeItem = (*DevComment)(nil)
	_ ImportScopeItem = (*BadImportSpec)(nil)
)

// ==================================== Import Spec =====================================

type ImportSpec struct {
	// Alias is the alias of the import, if any.
	Alias *Ident
	Path  *StaticString

	Position Position
}

var _ ImportScopeItem = (*ImportSpec)(nil)

func (s *ImportSpec) Pos() Position   { return s.Position }
func (*ImportSpec) _importScopeItem() {}

// ================================== Bad Import Sepc ===================================

type BadImportSpec struct {
	Line     string
	Position Position
}

var _ ImportScopeItem = (*BadImportSpec)(nil)

func (s *BadImportSpec) Pos() Position   { return s.Position }
func (*BadImportSpec) _importScopeItem() {}

// ============================================================================
// State
// ======================================================================================

type State struct {
	LParen *Position // nil if this is a single-line state
	Vars   []StateScopeItem
	RParen *Position // nil if this is a single-line state

	Position Position
}

var _ ScopeItem = (*State)(nil)

func (s *State) Pos() Position { return s.Position }
func (*State) _scopeItem()     {}

// ================================== State Scope Item ==================================

// StateScopeItem is a pointer to either a [StateVar], a [DevComment], or a
// [BadStateVar].
type StateScopeItem interface {
	_stateScopeItem()
	Poser
}

// if this is changed, change the comment above
var (
	_ StateScopeItem = (*StateVar)(nil)
	_ StateScopeItem = (*DevComment)(nil)
	_ StateScopeItem = (*BadStateVar)(nil)
)

// ==================================== State Var =======================================

type StateVar struct {
	Names []*Ident
	Type  *Type // nil if type is inferred

	Assign *Position // nil if this has no default value
	Values []*GoCode // empty if no default value
}

var _ StateScopeItem = (*StateVar)(nil)

func (v *StateVar) Pos() Position {
	if len(v.Names) > 0 {
		return v.Names[0].Pos()
	}
	return InvalidPosition
}
func (*StateVar) _stateScopeItem() {}

// =================================== Bad State Var ====================================

type BadStateVar struct {
	Line     string
	Position Position
}

var _ StateScopeItem = (*BadStateVar)(nil)

func (v *BadStateVar) Pos() Position  { return v.Position }
func (*BadStateVar) _stateScopeItem() {}
