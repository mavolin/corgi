package file

// ============================================================================
// Assign
// ======================================================================================

type Assign struct {
	Expression Expression
	Position
}

var _ ScopeItem = Assign{}

func (Assign) _typeScopeItem() {}

// ============================================================================
// Inline Text
// ======================================================================================

type InlineText struct {
	Text TextLine
	Position
}

var _ ScopeItem = InlineText{}

func (InlineText) _typeScopeItem() {}

// ============================================================================
// ArrowBlock
// ======================================================================================

type ArrowBlock struct {
	Lines []TextLine
	Position
}

var _ ScopeItem = ArrowBlock{}

func (ArrowBlock) _typeScopeItem() {}

// ============================================================================
// TextItem
// ======================================================================================

type TextItem interface {
	_typeTextItem()
	Poser
}

type TextLine []TextItem

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

var _ TextItem = Text{}

func (Text) _typeTextItem() {}

// ============================================================================
// Interpolation
// ======================================================================================

type Interpolation interface {
	_typeInterpolation()
	_typeTextItem()
	Poser
}

// ================================ SimpleInterpolation =================================

type SimpleInterpolation struct {
	NoEscape bool
	Value    InterpolationValue

	Position
}

var (
	_ TextItem      = SimpleInterpolation{}
	_ Interpolation = SimpleInterpolation{}
)

func (SimpleInterpolation) _typeInterpolation() {}
func (SimpleInterpolation) _typeTextItem()      {}

// ================================ ElementInterpolation ================================

type ElementInterpolation struct {
	Element Element
	Value   InterpolationValue // nil for void elems

	Position
}

var (
	_ TextItem      = ElementInterpolation{}
	_ Interpolation = ElementInterpolation{}
)

func (ElementInterpolation) _typeTextItem()      {}
func (ElementInterpolation) _typeInterpolation() {}

// ================================= MixinCallInterpolation =================================

type MixinCallInterpolation struct {
	MixinCall MixinCall
	Value     InterpolationValue // may be nil

	Position
}

var (
	_ TextItem      = MixinCallInterpolation{}
	_ Interpolation = MixinCallInterpolation{}
)

func (MixinCallInterpolation) _typeTextItem()      {}
func (MixinCallInterpolation) _typeInterpolation() {}

// ============================================================================
// InterpolationValue
// ======================================================================================

type InterpolationValue interface {
	_typeInterpolationValue()
	Poser
}

// =============================== TextInterpolationValue ===============================

type TextInterpolationValue struct {
	LBracketPos Position
	Text        string
	RBracketPos Position
}

var _ InterpolationValue = TextInterpolationValue{}

func (TextInterpolationValue) _typeInterpolationValue() {}

func (interp TextInterpolationValue) Pos() Position { return interp.LBracketPos }

// ============================ ExpressionInterpolationValue ============================

type ExpressionInterpolationValue struct {
	LBracePos       Position
	FormatDirective string // a Sprintf formatting placeholder, w/o preceding '%'
	Expression      Expression
	RBracePos       Position
}

var _ InterpolationValue = ExpressionInterpolationValue{}

func (ExpressionInterpolationValue) _typeInterpolationValue() {}

func (interp ExpressionInterpolationValue) Pos() Position { return interp.LBracePos }
