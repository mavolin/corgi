package ast

import "strings"

// ============================================================================
// GoCode
// ======================================================================================

type (
	// GoCode represents a sequence of actual Go code and corgi extensions to
	// the Go language.
	GoCode struct {
		Expressions []GoCodeNode
	}

	// GoCodeNode is a pointer to either [RawGoCode], [String], or [BlockFunction].
	GoCodeNode interface {
		Node
		_goCodeNode()
	}
)

var (
	_ Expression     = (*GoCode)(nil)
	_ ForHeader      = (*GoCode)(nil)
	_ AttributeValue = (*GoCode)(nil)
)

func (c *GoCode) Pos() Position {
	if len(c.Expressions) > 0 {
		return c.Expressions[0].Pos()
	}
	return InvalidPosition
}
func (c *GoCode) End() Position {
	if len(c.Expressions) > 0 {
		return c.Expressions[len(c.Expressions)-1].End()
	}
	return InvalidPosition
}

func (*GoCode) _node()           {}
func (*GoCode) _expression()     {}
func (*GoCode) _forHeader()      {}
func (*GoCode) _attributeValue() {}

// ============================================================================
// Go Code
// ======================================================================================

// RawGoCode represents actual Go code, i.e. without any corgi language extensions.
type RawGoCode struct {
	Code     string
	Position Position
}

var _ GoCodeNode = (*RawGoCode)(nil)

func (c *RawGoCode) Pos() Position { return c.Position }
func (c *RawGoCode) End() Position { return deltaPos(c.Position, len(c.Code)) }

func (*RawGoCode) _node()       {}
func (*RawGoCode) _goCodeNode() {}

// ============================================================================
// Block Function
// ======================================================================================

// BlockFunction represents the "built-in" block existence check function.
type BlockFunction struct {
	LParen *Position
	Block  *Ident
	RParen *Position

	Position Position
}

var _ GoCodeNode = (*BlockFunction)(nil)

func (f *BlockFunction) Pos() Position { return f.Position }
func (f *BlockFunction) End() Position {
	if f.RParen != nil {
		return *f.RParen
	} else if f.Block != nil {
		return f.Block.End()
	} else if f.LParen != nil {
		return *f.LParen
	}
	return deltaPos(f.Position, len("block"))
}

func (*BlockFunction) _node()       {}
func (*BlockFunction) _goCodeNode() {}

// ============================================================================
// String
// ======================================================================================

// String represents a Go string literal extended to allow Character References, and
// Interpolation.
type String struct {
	Open     Position
	Quote    byte // either '"' or '`'
	Contents []StringContent
	Close    *Position
}

var _ GoCodeNode = (*String)(nil)

func (s *String) Pos() Position { return s.Open }
func (s *String) End() Position {
	if s.Close != nil {
		return *s.Close
	} else if len(s.Contents) > 0 {
		return s.Contents[len(s.Contents)-1].End()
	}
	return deltaPos(s.Open, len(`"`))
}

func (*String) _node()       {}
func (*String) _goCodeNode() {}

// ============================================================================
// String Node
// ======================================================================================

// StringContent is a pointer to either [StringText] or [StringInterpolation].
type StringContent interface {
	Node
	_stringContent()
}

// if this is changed, change the comment above
var (
	_ StringContent = (*StringText)(nil)
	_ StringContent = (*EscapedHash)(nil)
	_ StringContent = (*ExpressionInterpolation)(nil)
	_ StringContent = (*CharacterReference)(nil)
	_ StringContent = (*BadInterpolation)(nil)
)

// ==================================== String Text =====================================

type StringText struct {
	Text     string
	Position Position
}

var _ StringContent = (*StringText)(nil)

func (t StringText) Unescape() string {
	return strings.ReplaceAll(t.Text, "##", "#")
}

func (t *StringText) Pos() Position { return t.Position }
func (t *StringText) End() Position { return deltaPos(t.Position, len(t.Text)) }

func (*StringText) _node()          {}
func (*StringText) _stringContent() {}

// ================================ String Interpolation ================================

type StringInterpolation interface {
	StringContent
	_stringInterpolation()
}
