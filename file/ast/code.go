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

var _ ScopeNode = (*Code)(nil)

func (c *Code) Pos() Position { return c.Position }
func (c *Code) End() Position {
	if len(c.Statements) >= 0 {
		return c.Statements[len(c.Statements)-1].End()
	}

	if c.Implicit {
		return InvalidPosition
	}
	return deltaPos(c.Position, len("-"))
}

func (*Code) _node()      {}
func (*Code) _scopeNode() {}

// ============================================================================
// Return
// ======================================================================================

type Return struct {
	Err      *GoCode // optional
	Position Position
}

var _ ScopeNode = (*Return)(nil)

func (r *Return) Pos() Position { return r.Position }
func (r *Return) End() Position {
	if r.Err != nil {
		return r.Err.End()
	}
	return r.Position
}

func (*Return) _node()      {}
func (*Return) _scopeNode() {}

// ============================================================================
// Break
// ======================================================================================

type Break struct {
	Label    *Ident // optional
	Position Position
}

var _ ScopeNode = (*Break)(nil)

func (b *Break) Pos() Position { return b.Position }
func (b *Break) End() Position {
	if b.Label != nil {
		return b.Label.End()
	}
	return b.Position
}

func (*Break) _node()      {}
func (*Break) _scopeNode() {}

// ============================================================================
// Continue
// ======================================================================================

type Continue struct {
	Label    *Ident // optional
	Position Position
}

var _ ScopeNode = (*Continue)(nil)

func (c *Continue) Pos() Position { return c.Position }
func (c *Continue) End() Position {
	if c.Label != nil {
		return c.Label.End()
	}
	return c.Position
}

func (*Continue) _node()      {}
func (*Continue) _scopeNode() {}
