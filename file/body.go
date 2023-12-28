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

type Scope struct {
	LBrace Position
	Items  []ScopeItem
	RBrace Position
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
	Body Body // may be nil
	Position
}

func (BadItem) _scopeItem() {}

// ============================================================================
// CorgiComment
// ======================================================================================

// CorgiComment represents a comment that is not printed.
type CorgiComment struct {
	Comment string
	Position
}

func (CorgiComment) _scopeItem()       {}
func (CorgiComment) _importScopeItem() {}
