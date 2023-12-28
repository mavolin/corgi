package file

// ============================================================================
// ArrowBlock
// ======================================================================================

type ArrowBlock struct {
	Lines []TextLine
	Position
}

var _ ScopeItem = ArrowBlock{}

func (ArrowBlock) _scopeItem() {}

// ============================================================================
// BracketText
// ======================================================================================

type BracketText struct {
	LBracket Position
	Lines    []TextLine
	RBracket Position
}

var _ Body = BracketText{}

func (t BracketText) Pos() Position {
	return t.LBracket
}

func (BracketText) _body() {}

// ============================================================================
// TextItem
// ======================================================================================

type (
	TextLine []TextItem

	// TextItem is either [Text] or [Interpolation].
	TextItem interface {
		_textItem()
		Poser
	}
)

func (l TextLine) Pos() Position {
	if len(l) == 0 {
		return InvalidPosition
	}

	return l[0].Pos()
}

// ============================================================================
// Text
// ======================================================================================

// Text is a string of text written as content of an element.
// It is not HTML-escaped yet.
type Text struct {
	Text string
	Position
}

func (Text) _textItem() {}

// ============================================================================
// Interpolation
// ======================================================================================

type Interpolation interface {
	_interpolation()
	TextItem
	Poser
}

// ================================= Bad Interpolation ==================================

type BadInterpolation struct {
	Position
}

func (BadInterpolation) _interpolation() {}
func (BadInterpolation) _textItem()      {}

// ============================== ExpressionInterpolation ===============================

type ExpressionInterpolation struct {
	// a sprintf placeholder, excluding the leading %
	FormatDirective string
	LBrace          Position
	Expression      Expression
	RBrace          Position

	Position
}

var _ TextItem = ExpressionInterpolation{}
var _ Interpolation = ExpressionInterpolation{}

func (ExpressionInterpolation) _textItem()      {}
func (ExpressionInterpolation) _interpolation() {}

// ================================ TextInterpolation =================================

type TextInterpolation struct {
	NoEscape bool
	Value    InterpolationValue

	Position
}

func (TextInterpolation) _interpolation() {}
func (TextInterpolation) _textItem()      {}

// ================================ ElementInterpolation ================================

type ElementInterpolation struct {
	Element Element            // has no body
	Value   InterpolationValue // may be nil, always nil for void elems

	Position
}

func (ElementInterpolation) _textItem()      {}
func (ElementInterpolation) _interpolation() {}

// ================================= ComponentCallInterpolation =================================

type ComponentCallInterpolation struct {
	ComponentCall ComponentCall
	Value         InterpolationValue // may be nil

	Position
}

func (ComponentCallInterpolation) _textItem()      {}
func (ComponentCallInterpolation) _interpolation() {}

// ============================================================================
// InterpolationValue
// ======================================================================================

type InterpolationValue struct {
	LBracket Position
	Text     TextLine
	RBracket Position
}

func (v InterpolationValue) Pos() Position { return v.LBracket }
