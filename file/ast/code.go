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
			return n.Start()
		}
	}
	return Position{}
}

func (c Code) End() Position {
	for _, n := range slices.Backward(c) {
		if n != nil {
			return n.End()
		}
	}
	return Position{}
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
	return Position{}
}

func (c *GoCode) End() Position {
	if c.Position != nil {
		return deltaPos(*c.Position, len(c.Code))
	}
	return Position{}
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
	BlockName *Ident
	RParen    *Position
}

var _ CodeNode = (*BlockFunction)(nil)

func (f *BlockFunction) Start() Position {
	switch {
	case f.Block != nil:
		return *f.Block
	case f.LParen != nil:
		return *f.LParen
	case f.BlockName != nil:
		return f.BlockName.Start()
	case f.RParen != nil:
		return *f.RParen
	}
	return Position{}
}

func (f *BlockFunction) End() Position {
	switch {
	case f.RParen != nil:
		return deltaPos(*f.RParen, len(")"))
	case f.BlockName != nil:
		return f.BlockName.End()
	case f.LParen != nil:
		return deltaPos(*f.LParen, len("("))
	case f.Block != nil:
		return deltaPos(*f.Block, len("block"))
	}
	return Position{}
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
	switch {
	case t.QuestionMark != nil:
		return *t.QuestionMark
	case t.LParen != nil:
		return *t.LParen
	case t.Condition != nil:
		return t.Condition.Start()
	case t.TrueVal != nil:
		return t.TrueVal.Start()
	case t.FalseVal != nil:
		return t.FalseVal.Start()
	case t.RParen != nil:
		return *t.RParen
	}
	return Position{}
}

func (t *Ternary) End() Position {
	switch {
	case t.RParen != nil:
		return deltaPos(*t.RParen, len(")"))
	case t.FalseVal != nil:
		return t.FalseVal.End()
	case t.TrueVal != nil:
		return t.TrueVal.End()
	case t.Condition != nil:
		return t.Condition.End()
	case t.LParen != nil:
		return deltaPos(*t.LParen, len("("))
	case t.QuestionMark != nil:
		return deltaPos(*t.QuestionMark, len("?"))
	}
	return Position{}
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
			return n.Start()
		}
	}
	if s.Close != nil {
		return *s.Close
	}
	return Position{}
}

func (s *String) End() Position {
	if s.Close != nil {
		return deltaPos(*s.Close, len(`"`))
	}
	for _, n := range slices.Backward(s.Contents) {
		if n != nil {
			return n.End()
		}
	}
	if s.Open != nil {
		return deltaPos(*s.Open, len(`"`))
	}
	return Position{}
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
	return Position{}
}

func (t *StringText) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return Position{}
}
func (t *StringText) Walk(func(Node)) {}

func (*StringText) _node()       {}
func (*StringText) _stringNode() {}

// ================================ String Interpolation ================================

// StringInterpolation is a pointer to either [BadInterpolation],
// an [EscapedHash], an [ExpressionInterpolation], a [CharacterReference],
// or a [ComponentCallInterpolation].
type StringInterpolation interface {
	StringNode
	Interpolation
}
