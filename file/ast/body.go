package ast

// ============================================================================
// Body
// ======================================================================================

// Body is a pointer to either a [Scope], [BracketText],
// [UnderscoreBlockShorthand], or an [Extend].
type Body interface {
	Node
	_body()
}

// if this is changed, change the comment above
var (
	_ Body = (*Scope)(nil)
	_ Body = (*BracketText)(nil)
	_ Body = (*UnderscoreBlockShorthand)(nil)
	_ Body = (*Extend)(nil)
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

// ============================================================================
// BracketText
// ======================================================================================

type BracketText struct {
	LBracket Position
	Lines    []TextLine
	RBracket *Position
}

var _ Body = (*BracketText)(nil)

func (t *BracketText) Pos() Position { return t.LBracket }
func (t *BracketText) End() Position {
	if t.RBracket != nil {
		return *t.RBracket
	} else if len(t.Lines) > 0 {
		return t.Lines[len(t.Lines)-1].End()
	}
	return deltaPos(t.LBracket, 1)
}

func (*BracketText) _node() {}
func (*BracketText) _body() {}

// ============================================================================
// Underscore Block Shorthand
// ======================================================================================

type UnderscoreBlockShorthand struct {
	Body     Body
	Position Position
}

var _ Body = (*UnderscoreBlockShorthand)(nil)

func (s *UnderscoreBlockShorthand) Pos() Position { return s.Position }
func (s *UnderscoreBlockShorthand) End() Position {
	if s.Body != nil {
		return s.Body.End()
	}
	return deltaPos(s.Position, len("_"))
}

func (*UnderscoreBlockShorthand) _node() {}
func (*UnderscoreBlockShorthand) _body() {}

// ============================================================================
// Extend
// ======================================================================================

type Extend struct {
	ComponentCall *ComponentCall
}

var _ Body = (*Extend)(nil)

func (e *Extend) Pos() Position { return e.ComponentCall.Pos() }
func (e *Extend) End() Position { return e.ComponentCall.End() }

func (*Extend) _node() {}
func (*Extend) _body() {}

// ============================================================================
// Scope Node
// ======================================================================================

// ScopeNode represents a node that can appear in a [Scope].
type ScopeNode interface {
	Node
	_scopeNode()
}

// ====================================== Bad Node ======================================

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
