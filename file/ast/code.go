package ast

import "slices"

// Code is a sequence of Go code with corgi language extensions.
//
// Note that there may be multiple successive [GoCode] nodes in a [Code]
// object, especially if the code was generated from a ParsedStatement.
// There are no guarantees that between versions, the code will be split
// in the same way.
type Code []CodeNode

var _ Node = Code(nil)

func (c Code) Start() Position {
	for _, n := range c {
		if n != nil {
			if start := n.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (c Code) End() Position {
	for _, n := range slices.Backward(c) {
		if n != nil {
			if end := n.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (c Code) Walk(w func(Node)) {
	for _, n := range c {
		if n != nil {
			w(n)
		}
	}
}

func (Code) _node() {}

type CodeNode interface {
	Node
	_codeNode()
}

// ============================================================================
// Go Code
// ======================================================================================

// GoCode is actual Go code, i.e., without any corgi language extensions.
type GoCode struct {
	Code     string
	Position *Position
}

var _ CodeNode = (*GoCode)(nil)

func (c *GoCode) Start() Position {
	if c.Position != nil {
		return *c.Position
	}
	return NoPosition
}

func (c *GoCode) End() Position {
	if c.Position != nil {
		return deltaPos(*c.Position, len(c.Code))
	}
	return NoPosition
}
func (c *GoCode) Walk(func(Node)) {}

func (*GoCode) _node()     {}
func (*GoCode) _codeNode() {}

// ============================================================================
// Block Function
// ======================================================================================

// BlockFunction is the "built-in" block existence check function.
type BlockFunction struct {
	Block     *Position
	LParen    *Position
	BlockName *Identifier // optional for the default block
	RParen    *Position
}

var _ CodeNode = (*BlockFunction)(nil)

func (f *BlockFunction) Name() string {
	if f.BlockName != nil {
		return f.BlockName.Name
	}
	return ""
}

func (f *BlockFunction) Start() Position {
	if f.Block != nil {
		return *f.Block
	}
	if f.LParen != nil {
		return *f.LParen
	}
	if f.BlockName != nil {
		if start := f.BlockName.Start(); start != NoPosition {
			return start
		}
	}
	if f.RParen != nil {
		return *f.RParen
	}
	return NoPosition
}

func (f *BlockFunction) End() Position {
	if f.RParen != nil {
		return deltaPos(*f.RParen, len(")"))
	}
	if f.BlockName != nil {
		if end := f.BlockName.End(); end != NoPosition {
			return end
		}
	}
	if f.LParen != nil {
		return deltaPos(*f.LParen, len("("))
	}
	if f.Block != nil {
		return deltaPos(*f.Block, len("block"))
	}
	return NoPosition
}

func (f *BlockFunction) Walk(w func(Node)) {
	if f.BlockName != nil {
		w(f.BlockName)
	}
}

func (*BlockFunction) _node()     {}
func (*BlockFunction) _codeNode() {}

// ============================================================================
// Ternary
// ======================================================================================

type Ternary struct {
	QuestionMark *Position
	LParen       *Position
	Condition    *Expression
	TrueVal      *Expression
	FalseVal     *Expression
	RParen       *Position
}

var _ CodeNode = (*Ternary)(nil)

func (t *Ternary) Start() Position {
	if t.QuestionMark != nil {
		return *t.QuestionMark
	}
	if t.LParen != nil {
		return *t.LParen
	}
	if t.Condition != nil {
		if start := t.Condition.Start(); start != NoPosition {
			return start
		}
	}
	if t.TrueVal != nil {
		if start := t.TrueVal.Start(); start != NoPosition {
			return start
		}
	}
	if t.FalseVal != nil {
		if start := t.FalseVal.Start(); start != NoPosition {
			return start
		}
	}
	if t.RParen != nil {
		return *t.RParen
	}
	return NoPosition
}

func (t *Ternary) End() Position {
	if t.RParen != nil {
		return deltaPos(*t.RParen, len(")"))
	}
	if t.FalseVal != nil {
		if end := t.FalseVal.End(); end != NoPosition {
			return end
		}
	}
	if t.TrueVal != nil {
		if end := t.TrueVal.End(); end != NoPosition {
			return end
		}
	}
	if t.Condition != nil {
		if end := t.Condition.End(); end != NoPosition {
			return end
		}
	}
	if t.LParen != nil {
		return deltaPos(*t.LParen, len("("))
	}
	if t.QuestionMark != nil {
		return deltaPos(*t.QuestionMark, len("?"))
	}
	return NoPosition
}

func (t *Ternary) Walk(w func(Node)) {
	if t.Condition != nil {
		w(t.Condition)
	}
	if t.TrueVal != nil {
		w(t.TrueVal)
	}
	if t.FalseVal != nil {
		w(t.FalseVal)
	}
}

func (*Ternary) _node()     {}
func (*Ternary) _codeNode() {}

// ============================================================================
// String
// ======================================================================================

// String is a Go string literal extended to allow Character References, and
// StringInterpolation.
type String struct {
	Open     *Position
	Quote    byte // either '"' or '`'
	Contents []StringNode
	Close    *Position
}

var _ CodeNode = (*String)(nil)

func (s *String) Start() Position {
	if s.Open != nil {
		return *s.Open
	}
	for _, n := range s.Contents {
		if n != nil {
			if start := n.Start(); start != NoPosition {
				return start
			}
		}
	}
	if s.Close != nil {
		return *s.Close
	}
	return NoPosition
}

func (s *String) End() Position {
	if s.Close != nil {
		return deltaPos(*s.Close, len("\""))
	}
	for _, n := range slices.Backward(s.Contents) {
		if n != nil {
			if end := n.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.Open != nil {
		return deltaPos(*s.Open, len("\""))
	}
	return NoPosition
}

func (s *String) Walk(w func(Node)) {
	for _, n := range s.Contents {
		if n != nil {
			w(n)
		}
	}
}

func (*String) _node()     {}
func (*String) _codeNode() {}

// ============================================================================
// String Node
// ======================================================================================

// StringNode is a pointer to either [StringText] or [StringInterpolation].
type StringNode interface {
	Node
	_stringNode()
}

// if this is changed, change the comment above
var (
	_ StringNode = (*StringText)(nil)
	_ StringNode = StringInterpolation(nil)
)

// ==================================== String Text =====================================

type StringText struct {
	Text     string
	Position *Position
}

var _ StringNode = (*StringText)(nil)

func (t *StringText) Start() Position {
	if t.Position != nil {
		return *t.Position
	}
	return NoPosition
}

func (t *StringText) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return NoPosition
}
func (t *StringText) Walk(func(Node)) {}

func (*StringText) _node()       {}
func (*StringText) _stringNode() {}

// ================================ String Interpolation ================================

// StringInterpolation is a pointer to either [BadInterpolation],
// an [CharacterEscape], an [ExpressionInterpolation], a [CharacterReference],
// or a [ComponentCallInterpolation].
type StringInterpolation interface {
	StringNode
	Interpolation
}
