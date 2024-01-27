package ast

// ============================================================================
// Body
// ======================================================================================

// Body is a pointer to either a [Scope], [BracketText], or a
// [UnderscoreBlockShorthand].
type Body interface {
	_body()
	Poser
}

// if this is changed, change the comment above
var (
	_ Body = (*Scope)(nil)
	_ Body = (*BracketText)(nil)
	_ Body = (*UnderscoreBlockShorthand)(nil)
)

// ============================================================================
// Scope
// ======================================================================================

type Scope struct {
	LBrace Position
	Items  []ScopeItem
	RBrace *Position
}

var _ Body = (*Scope)(nil)

func (s *Scope) Pos() Position { return s.LBrace }
func (*Scope) _body()          {}

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
	Line     string
	Body     Body // may be nil
	Position Position
}

var _ ScopeItem = (*BadItem)(nil)

func (itm *BadItem) Pos() Position { return itm.Position }
func (*BadItem) _scopeItem()       {}

// ============================================================================
// DevComment
// ======================================================================================

// DevComment represents a comment that is not printed.
type DevComment struct {
	Comment  string
	Position Position
}

var (
	_ ScopeItem       = (*DevComment)(nil)
	_ ImportScopeItem = (*DevComment)(nil)
	_ StateScopeItem  = (*DevComment)(nil)
)

func (c *DevComment) Pos() Position   { return c.Position }
func (*DevComment) _scopeItem()       {}
func (*DevComment) _importScopeItem() {}
func (*DevComment) _stateScopeItem()  {}
