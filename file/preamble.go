package file

// ============================================================================
// PackageDirective
// ======================================================================================

type PackageDirective struct {
	// Name is the name of the package.
	Name GoIdent

	Position
}

// ============================================================================
// Import
// ======================================================================================

// Import represents a single import.
type Import struct {
	LParen  *Position // nil if this is a single-line import
	Imports []ImportScopeItem
	RParen  *Position // nil if this is a single-line import

	Position
}

// ImportScopeItem is either an [ImportSpec], a [CorgiComment], or a [BadItem].
type ImportScopeItem interface {
	_importScopeItem()
}

type ImportSpec struct {
	// Alias is the alias of the import, if any.
	Alias *GoIdent
	Path  String

	Position
}

func (ImportSpec) _importScopeItem() {}

// ============================================================================
// State
// ======================================================================================

// State represents a state variable declaration.

type State struct {
	LParen *Position // nil if this is a single-line state
	Vars   []StateVar
	RParen *Position // nil if this is a single-line state

	Position
}

func (State) _scopeItem() {}

type StateVar struct {
	// Name is the name of the state variable.
	Name GoIdent
	Type GoType

	Assign *Position  // nil if this has no default value
	Value  Expression // GoExpression or StringExpression
}
