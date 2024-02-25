package ast

// ============================================================================
// Body
// ======================================================================================

// Body is a pointer to either a [Scope], [BracketText], or a
// [UnderscoreBlockShorthand].
type Body interface {
	Node
	_body()
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
	Nodes  []ScopeNode
	RBrace *Position
}

var _ Body = (*Scope)(nil)

func (s *Scope) Pos() Position { return s.LBrace }
func (s *Scope) End() Position {
	if s.RBrace != nil {
		return *s.RBrace
	} else if len(s.Nodes) > 0 {
		return s.Nodes[len(s.Nodes)-1].End()
	}
	return deltaPos(s.LBrace, 1)
}

func (*Scope) _node() {}
func (*Scope) _body() {}

// ===================================== Scope Node =====================================

// ScopeNode represents a node that can appear in a [Scope].
type ScopeNode interface {
	Node
	_scopeNode()
}

// ============================================================================
// BadNode
// ======================================================================================

type BadNode struct {
	// Line contains the entire bad line, excluding leading whitespace.
	Line     string
	Body     Body // may be nil
	Position Position
}

var _ ScopeNode = (*BadNode)(nil)

func (n *BadNode) Pos() Position { return n.Position }
func (n *BadNode) End() Position {
	if n.Body != nil {
		return n.Body.End()
	}
	return deltaPos(n.Position, len(n.Line))
}

func (*BadNode) _node()      {}
func (*BadNode) _scopeNode() {}

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
