package file

// ============================================================================
// Assign
// ======================================================================================

type Assign struct {
	Expression Expression
	NoEscape   bool
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

var _ TextItem = SimpleInterpolation{}
var _ Interpolation = SimpleInterpolation{}

func (SimpleInterpolation) _typeInterpolation() {}
func (SimpleInterpolation) _typeTextItem()      {}

// ================================ ElementInterpolation ================================

type ElementInterpolation struct {
	NoEscape bool
	Element  Element
	Value    InterpolationValue // nil for void elems

	Position
}

var _ TextItem = ElementInterpolation{}
var _ Interpolation = ElementInterpolation{}

func (ElementInterpolation) _typeTextItem()      {}
func (ElementInterpolation) _typeInterpolation() {}

// ================================= MixinCallInterpolation =================================

type MixinCallInterpolation struct {
	NoEscape  bool
	MixinCall MixinCall
	Value     InterpolationValue

	Position
}

var _ TextItem = MixinCallInterpolation{}
var _ Interpolation = MixinCallInterpolation{}

func (MixinCallInterpolation) _typeTextItem()      {}
func (MixinCallInterpolation) _typeInterpolation() {}

// ============================================================================
// InterpolationValue
// ======================================================================================

type InterpolationValue interface {
	_typeInterpolationValue()
}

// =============================== TextInterpolationValue ===============================

type TextInterpolationValue struct {
	LBracketPos Position
	Text        string
	RBracketPos Position
}

var _ InterpolationValue = TextInterpolationValue{}

func (TextInterpolationValue) _typeInterpolationValue() {}

// ============================ ExpressionInterpolationValue ============================

type ExpressionInterpolationValue struct {
	LBracePos  Position
	Expression Expression
	RBracePos  Position
}

var _ InterpolationValue = ExpressionInterpolationValue{}

func (ExpressionInterpolationValue) _typeInterpolationValue() {}
