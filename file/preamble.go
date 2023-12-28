package file

// ============================================================================
// Package Directive
// ======================================================================================

type PackageDirective struct {
	// Name is the name of the package.
	Name Ident

	Position
}

// ============================================================================
// Import
// ======================================================================================

type (
	Import struct {
		LParen  *Position // nil if this is a single-line import
		Imports []ImportScopeItem
		RParen  *Position // nil if this is a single-line import

		Position
	}

	// ImportScopeItem is either an [ImportSpec], a [CorgiComment], or a
	// [BadImportSpec].
	ImportScopeItem interface {
		_importScopeItem()
		Poser
	}
)

func (Import) _scopeItem() {}

type (
	ImportSpec struct {
		// Alias is the alias of the import, if any.
		Alias *Ident
		Path  StaticString

		Position
	}

	BadImportSpec struct {
		Line string
		Position
	}
)

func (ImportSpec) _importScopeItem()    {}
func (BadImportSpec) _importScopeItem() {}

// ============================================================================
// State
// ======================================================================================

type (
	// State represents a state variable declaration.
	State struct {
		LParen *Position // nil if this is a single-line state
		Vars   []StateScopeItem
		RParen *Position // nil if this is a single-line state

		Position
	}

	// StateScopeItem is either an [StateVar], a [CorgiComment], or a
	// [BadStateVar].
	StateScopeItem interface {
		_stateScopeItem()
		Poser
	}
)

func (State) _scopeItem() {}

type (
	StateVar struct {
		Names        []Ident
		Type         *Type
		InferredType string // set by package typeinfer before linking

		Assign *Position // nil if this has no default value
		Values []GoCode  // nil if this has no default value
	}

	BadStateVar struct {
		Line string
		Position
	}
)

func (StateVar) _stateScopeItem()    {}
func (BadStateVar) _stateScopeItem() {}

func (v StateVar) Pos() Position {
	if len(v.Names) > 0 {
		return v.Names[0].Pos()
	}
	return InvalidPosition
}
