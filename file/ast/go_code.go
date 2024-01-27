package ast

import "strings"

// ============================================================================
// GoCode
// ======================================================================================

type (
	// GoCode represents a sequence of actual Go code and corgi extensions to
	// the Go language.
	GoCode struct {
		Expressions []GoCodeItem
	}

	// GoCodeItem is a pointer to either [RawGoCode], [String], or [BlockFunction].
	GoCodeItem interface {
		_goCodeItem()
		Poser
	}
)

var (
	_ Expression     = (*GoCode)(nil)
	_ ForExpression  = (*GoCode)(nil)
	_ AttributeValue = (*GoCode)(nil)
)

func (e *GoCode) Pos() Position {
	if len(e.Expressions) > 0 {
		return e.Expressions[0].Pos()
	}
	return InvalidPosition
}

func (*GoCode) _expression()     {}
func (*GoCode) _forExpression()  {}
func (*GoCode) _attributeValue() {}

// ============================================================================
// Go Code
// ======================================================================================

// RawGoCode represents actual Go code, i.e. without any corgi language extensions.
type RawGoCode struct {
	Code     string
	Position Position
}

var _ GoCodeItem = (*RawGoCode)(nil)

func (c *RawGoCode) Pos() Position { return c.Position }
func (*RawGoCode) _goCodeItem()    {}

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

var _ GoCodeItem = (*BlockFunction)(nil)

func (f *BlockFunction) Pos() Position { return f.Position }
func (*BlockFunction) _goCodeItem()    {}

// ============================================================================
// String
// ======================================================================================

// String represents a Go string literal extended to allow interpolation.
type String struct {
	Start    Position
	Quote    byte // either '"' or '`'
	Contents []StringItem
	End      *Position
}

// if this is changed, change the comment above
var (
	_ StringItem = (*StringText)(nil)
	_ StringItem = (*StringInterpolation)(nil)
)

var _ GoCodeItem = (*String)(nil)

// ============================================================================
// String Item
// ======================================================================================

// StringItem is a pointer to either [StringText] or [StringInterpolation].
type StringItem interface {
	_stringItem()
	Poser
}

// ==================================== String Text =====================================

func (s *String) Pos() Position { return s.Start }
func (*String) _goCodeItem()    {}

type StringText struct {
	Text     string
	Position Position
}

var _ StringItem = (*StringText)(nil)

func (t StringText) Unescape() string {
	return strings.ReplaceAll(t.Text, "##", "#")
}

func (t *StringText) Pos() Position { return t.Position }
func (*StringText) _stringItem()    {}

// ================================ String Interpolation ================================

type StringInterpolation struct {
	FormatDirective string // a Sprintf formatting placeholder, w/o preceding '%'
	LBrace          *Position
	Expression      Expression
	RBrace          *Position
	Position        Position // of the hash
}

var _ StringItem = (*StringInterpolation)(nil)

func (i *StringInterpolation) Pos() Position { return i.Position }
func (*StringInterpolation) _stringItem()    {}
