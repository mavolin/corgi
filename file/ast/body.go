package ast

import "slices"

// ============================================================================
// Body
// ======================================================================================

// A Body is a group of nodes.
// It is a pointer to either a [Scope], [BracketText], or
// [UnderscoreBlockShorthand].
//
// The Go spec calls this a "block", but obviously that name is already taken.
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

// Scope is a sequence of [ScopeNode] nodes enclosed in braces.
//
// They are equivalent to blocks in the Go spec; whereas corgi has multiple
// types of blocks, the Scope type is closer to Go's block, than the others.
type Scope struct {
	LBrace *Position
	Nodes  []ScopeNode
	RBrace *Position
}

var _ Body = (*Scope)(nil)

func (s *Scope) Start() Position {
	if s.LBrace != nil {
		return *s.LBrace
	}
	for _, node := range s.Nodes {
		if node != nil {
			return node.Start()
		}
	}
	if s.RBrace != nil {
		return *s.RBrace
	}
	return Position{}
}
func (s *Scope) End() Position {
	if s.RBrace != nil {
		return deltaPos(*s.RBrace, len("}"))
	}
	for _, node := range slices.Backward(s.Nodes) {
		if node != nil {
			return node.End()
		}
	}
	if s.LBrace != nil {
		return deltaPos(*s.LBrace, len("{"))
	}
	return Position{}
}

func (*Scope) _node() {}
func (*Scope) _body() {}

// ============================================================================
// Scope Node
// ======================================================================================

// ScopeNode is a node that can appear in a [Scope].
//
// These are equivalent to statements in the Go spec.
type ScopeNode interface {
	Node
	_scopeNode()
}

type BadScopeNode struct {
	From, Until Position
}

var _ ScopeNode = (*BadScopeNode)(nil)

func (b *BadScopeNode) Start() Position { return b.From }
func (b *BadScopeNode) End() Position   { return b.Until }

func (*BadScopeNode) _node()      {}
func (*BadScopeNode) _scopeNode() {}

// ============================================================================
// BracketText
// ======================================================================================

type BracketText struct {
	LBracket *Position
	Lines    TextBlock
	RBracket *Position
}

var _ Body = (*BracketText)(nil)

func (t *BracketText) Start() Position {
	if t.LBracket != nil {
		return *t.LBracket
	} else if len(t.Lines) > 0 {
		return t.Lines.Start()
	} else if t.RBracket != nil {
		return *t.RBracket
	}
	return Position{}
}
func (t *BracketText) End() Position {
	if t.RBracket != nil {
		return deltaPos(*t.RBracket, len("]"))
	} else if len(t.Lines) > 0 {
		return t.Lines.End()
	} else if t.LBracket != nil {
		return deltaPos(*t.LBracket, len("["))
	}
	return Position{}
}

func (*BracketText) _node() {}
func (*BracketText) _body() {}

// ============================================================================
// Underscore General Shorthand
// ======================================================================================

type UnderscoreBlockShorthand struct {
	// Implicit, if set to true, indicates that this shorthand has no leading
	// underscore.
	// As of writing, this is only true for interpolation.
	Implicit bool
	Body     Body
	Position *Position
}

var _ Body = (*UnderscoreBlockShorthand)(nil)

func (s *UnderscoreBlockShorthand) Start() Position {
	if s.Position != nil {
		return *s.Position
	} else if s.Body != nil {
		return s.Body.Start()
	}
	return Position{}
}
func (s *UnderscoreBlockShorthand) End() Position {
	if s.Body != nil {
		return s.Body.End()
	} else if s.Position != nil {
		if s.Implicit {
			return *s.Position
		}
		return deltaPos(*s.Position, len("_"))
	}
	return Position{}
}

func (*UnderscoreBlockShorthand) _node() {}
func (*UnderscoreBlockShorthand) _body() {}
