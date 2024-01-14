package file

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

	GoCodeItem interface {
		_goCodeItem()
		Poser
	}
)

func (GoCode) _expression()     {}
func (GoCode) _forExpression()  {}
func (GoCode) _attributeValue() {}

func (e GoCode) Pos() Position {
	if len(e.Expressions) > 0 {
		return e.Expressions[0].Pos()
	}
	return InvalidPosition
}

// ============================================================================
// Go Code
// ======================================================================================

// RawGoCode represents actual Go code, i.e. without any corgi language extensions.
type RawGoCode struct {
	Code string
	Position
}

func (RawGoCode) _goCodeItem() {}

// ============================================================================
// Block Function
// ======================================================================================

// BlockFunction represents the "built-in" block existence check function.
type BlockFunction struct {
	LParen Position
	Block  Ident
	RParen Position

	Position
}

func (BlockFunction) _goCodeItem() {}

// ============================================================================
// String
// ======================================================================================

type (
	// String represents a Go string literal extended to allow interpolation.
	String struct {
		Start    Position
		Quote    byte // either '"' or '`'
		Contents []StringItem
		End      Position
	}

	StringItem interface {
		_stringItem()
	}
)

func (String) _goCodeItem()    {}
func (s String) Pos() Position { return s.Start }

type (
	StringText struct {
		Text string
		Position
	}

	StringInterpolation struct {
		FormatDirective string // a Sprintf formatting placeholder, w/o preceding '%'
		LBrace          Position
		Expression      Expression
		RBrace          Position
		Position
	}
)

func (StringText) _stringItem()          {}
func (StringInterpolation) _stringItem() {}

func (t StringText) Unescape() string {
	return strings.ReplaceAll(t.Text, "##", "#")
}
