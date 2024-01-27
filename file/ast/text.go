package ast

// ============================================================================
// ArrowBlock
// ======================================================================================

type ArrowBlock struct {
	Lines    []TextLine
	Position Position
}

var _ ScopeItem = (*ArrowBlock)(nil)

func (b *ArrowBlock) Pos() Position { return b.Position }
func (*ArrowBlock) _scopeItem()     {}

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
func (*BracketText) _body()          {}

// ============================================================================
// TextItem
// ======================================================================================

type (
	TextLine []TextItem

	// TextItem is a pointer to either a pointer to [Text] or [Interpolation].
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
	Text     string
	Position Position
}

var _ TextItem = (*Text)(nil)

func (t *Text) Pos() Position { return t.Position }
func (*Text) _textItem()      {}

// ============================================================================
// Interpolation
// ======================================================================================

// Interpolation is a pointer to either [BadInterpolation],
// [ExpressionInterpolation], [TextInterpolation], [ElementInterpolation], or
// [ComponentCallInterpolation].
type Interpolation interface {
	_interpolation()
	TextItem
	Poser
}

// if this is changed, change the comment above
var (
	_ Interpolation = (*BadInterpolation)(nil)
	_ Interpolation = (*ExpressionInterpolation)(nil)
	_ Interpolation = (*TextInterpolation)(nil)
	_ Interpolation = (*ElementInterpolation)(nil)
	_ Interpolation = (*ComponentCallInterpolation)(nil)
)

// ================================= Bad Interpolation ==================================

type BadInterpolation struct {
	Position Position
}

var _ Interpolation = (*BadInterpolation)(nil)

func (b *BadInterpolation) Pos() Position { return b.Position }
func (*BadInterpolation) _interpolation() {}
func (*BadInterpolation) _textItem()      {}

// ============================== ExpressionInterpolation ===============================

type ExpressionInterpolation struct {
	// a sprintf placeholder, excluding the leading %
	FormatDirective string
	LBrace          *Position
	Expression      Expression
	RBrace          *Position

	Position Position
}

var _ TextItem = (*ExpressionInterpolation)(nil)

func (i *ExpressionInterpolation) Pos() Position { return i.Position }
func (*ExpressionInterpolation) _textItem()      {}
func (*ExpressionInterpolation) _interpolation() {}

// ================================ TextInterpolation =================================

type TextInterpolation struct {
	NoEscape bool
	Value    *InterpolationValue

	Position Position
}

var _ TextItem = (*TextInterpolation)(nil)

func (i *TextInterpolation) Pos() Position { return i.Position }
func (*TextInterpolation) _textItem()      {}
func (*TextInterpolation) _interpolation() {}

// ================================ ElementInterpolation ================================

type ElementInterpolation struct {
	Element *Element            // has no body
	Value   *InterpolationValue // may be nil, always nil for void elems

	Position Position
}

var _ TextItem = (*ElementInterpolation)(nil)

func (i *ElementInterpolation) Pos() Position { return i.Position }
func (*ElementInterpolation) _textItem()      {}
func (*ElementInterpolation) _interpolation() {}

// ================================= ComponentCallInterpolation =================================

type ComponentCallInterpolation struct {
	ComponentCall *ComponentCall
	Value         *InterpolationValue // may be nil

	Position Position
}

var _ TextItem = (*ComponentCallInterpolation)(nil)

func (i *ComponentCallInterpolation) Pos() Position { return i.Position }
func (*ComponentCallInterpolation) _textItem()      {}
func (*ComponentCallInterpolation) _interpolation() {}

// ============================================================================
// InterpolationValue
// ======================================================================================

type InterpolationValue struct {
	LBracket Position
	Text     TextLine
	RBracket *Position
}

func (v InterpolationValue) Pos() Position { return v.LBracket }
