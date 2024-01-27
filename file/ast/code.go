package ast

// ============================================================================
// Code
// ======================================================================================

type Code struct {
	Statements []*GoCode
	// Implicit indicates whether this code was implicitly detected as such,
	// i.e. it didn't use the '-' operator.
	//
	// This field has no relevance for global code and may be any value.
	Implicit bool
	Position Position
}

var _ ScopeItem = (*Code)(nil)

func (c *Code) Pos() Position { return c.Position }
func (*Code) _scopeItem()     {}

// ============================================================================
// Return
// ======================================================================================

type Return struct {
	Err      *GoCode // optional
	Position Position
}

var _ ScopeItem = (*Return)(nil)

func (r *Return) Pos() Position { return r.Position }
func (*Return) _scopeItem()     {}

// ============================================================================
// Break
// ======================================================================================

type Break struct {
	Label    *Ident // optional
	Position Position
}

var _ ScopeItem = (*Break)(nil)

func (b *Break) Pos() Position { return b.Position }
func (*Break) _scopeItem()     {}

// ============================================================================
// Continue
// ======================================================================================

type Continue struct {
	Label    *Ident // optional
	Position Position
}

var _ ScopeItem = (*Continue)(nil)

func (c *Continue) Pos() Position { return c.Position }
func (*Continue) _scopeItem()     {}
