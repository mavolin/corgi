package file

// ============================================================================
// Body
// ======================================================================================

// Body is either [Scope] or [BracketText].
type Body interface {
	_body()
	Poser
}

// ============================================================================
// Scope
// ======================================================================================

// A Scope represents a level of indentation.
// Every Component available inside a scope is also available in its child scopes.
type Scope struct {
	// Items contains the items in this scope.
	Items          []ScopeItem
	LBrace, RBrace Position
}

func (s Scope) Pos() Position { return s.LBrace }
func (Scope) _body()          {}

// ============================================================================
// ScopeItem
// ======================================================================================

// ScopeItem represents an item in a [Scope].
type ScopeItem interface {
	_scopeItem()
	Poser
}

// ============================================================================
// BadItem
// ======================================================================================

type BadItem struct {
	// Line contains the entire bad line, excluding leading whitespace.
	Line string
	// Body contains the body of the item, if it has any.
	Body Body
	Position
}

func (BadItem) _scopeItem()       {}
func (BadItem) _importScopeItem() {}

// ============================================================================
// CorgiComment
// ======================================================================================

// CorgiComment represents a comment that is not printed.
type CorgiComment struct {
	Text string
	Position
}

func (CorgiComment) _scopeItem()       {}
func (CorgiComment) _importScopeItem() {}
