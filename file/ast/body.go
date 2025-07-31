package ast

import "slices"

// ============================================================================
// Body
// ======================================================================================

// A Body is a group of nodes.
// It is a pointer to either a [Scope], or [BracketText].
//
// The Go spec calls this a "block", but that name is already taken.
type Body interface {
	Node
	Highlighter
	_body()
	_componentBody()
}

// if this is changed, change the comment above
var (
	_ Body = (*Scope)(nil)
	_ Body = (*BracketText)(nil)
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

var (
	_ Body          = (*Scope)(nil)
	_ ComponentBody = (*Scope)(nil)
	_ Highlighter   = (*Scope)(nil)
)

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

func (s *Scope) Highlight() (start, end Position) {
	start, end = s.Start(), s.End()
	if start.Line == end.Line {
		return start, end
	}
	if s.LBrace != nil {
		return *s.LBrace, deltaPos(*s.LBrace, len("{"))
	}
	return start, end
}

func (s *Scope) Walk(w func(Node)) {
	for _, node := range s.Nodes {
		if node != nil {
			w(node)
		}
	}
}

func (*Scope) _node()              {}
func (*Scope) _body()              {}
func (*Scope) _componentBody()     {}
func (*Scope) _componentCallBody() {}

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

// ============================================================================
// Bracket Text
// ======================================================================================

type BracketText struct {
	LBracket *Position
	Lines    TextBlock
	RBracket *Position
}

var (
	_ Body        = (*BracketText)(nil)
	_ Highlighter = (*BracketText)(nil)
)

func (t *BracketText) Start() Position {
	switch {
	case t.LBracket != nil:
		return *t.LBracket
	case len(t.Lines) > 0:
		return t.Lines.Start()
	case t.RBracket != nil:
		return *t.RBracket
	}
	return Position{}
}

func (t *BracketText) End() Position {
	switch {
	case t.RBracket != nil:
		return deltaPos(*t.RBracket, len("]"))
	case len(t.Lines) > 0:
		return t.Lines.End()
	case t.LBracket != nil:
		return deltaPos(*t.LBracket, len("["))
	}
	return Position{}
}

func (t *BracketText) Highlight() (start, end Position) {
	start, end = t.Start(), t.End()
	if start.Line == end.Line {
		return start, end
	}
	if t.LBracket != nil {
		return *t.LBracket, deltaPos(*t.LBracket, len("["))
	}
	return start, end
}

func (t *BracketText) Walk(w func(Node)) {
	if t.Lines != nil {
		w(t.Lines)
	}
}

func (*BracketText) _node()          {}
func (*BracketText) _body()          {}
func (*BracketText) _componentBody() {}
