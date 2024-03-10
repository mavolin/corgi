package ast

// ============================================================================
// DevComment
// ======================================================================================

// DevComment represents a comment that is not printed.
type DevComment struct {
	Comment  string
	Position Position
}

var (
	_ ScopeNode  = (*DevComment)(nil)
	_ ImportNode = (*DevComment)(nil)
	_ StateNode  = (*DevComment)(nil)
)

func (c *DevComment) Pos() Position { return c.Position }
func (c *DevComment) End() Position { return deltaPos(c.Position, len(c.Comment)) }

func (*DevComment) _node()       {}
func (*DevComment) _scopeNode()  {}
func (*DevComment) _importNode() {}
func (*DevComment) _stateNode()  {}
