package ast

// ============================================================================
// ArrowBlock
// ======================================================================================

type ArrowBlock struct {
	Lines    []TextLine
	Position Position
}

var _ ScopeNode = (*ArrowBlock)(nil)

func (b *ArrowBlock) Pos() Position { return b.Position }
func (b *ArrowBlock) End() Position {
	if len(b.Lines) > 0 {
		return b.Lines[len(b.Lines)-1].End()
	}
	return deltaPos(b.Position, len(">"))
}

func (*ArrowBlock) _node()      {}
func (*ArrowBlock) _scopeNode() {}

// ============================================================================
// TextNode
// ======================================================================================

type (
	TextLine []TextNode

	// TextNode is a pointer to either a pointer to [Text] or [Interpolation].
	TextNode interface {
		Node
		_textNode()
	}
)

var (
	_ Node = (TextLine)(nil)

	// change above comment if this changes
	_ TextNode = (*Text)(nil)
	_ TextNode = (Interpolation)(nil)
)

func (l TextLine) Pos() Position {
	if len(l) == 0 {
		return InvalidPosition
	}

	return l[0].Pos()
}
func (l TextLine) End() Position {
	if len(l) > 0 {
		return l[len(l)-1].End()
	}
	return InvalidPosition
}

func (TextLine) _node() {}

// ============================================================================
// Text
// ======================================================================================

// Text is a string of text written as content of an element.
// It is not HTML-escaped yet.
type Text struct {
	Text     string
	Position Position
}

var _ TextNode = (*Text)(nil)

func (t *Text) Pos() Position { return t.Position }
func (t *Text) End() Position { return deltaPos(t.Position, len(t.Text)) }

func (t *Text) _node()   {}
func (*Text) _textNode() {}

// ============================================================================
// Interpolation
// ======================================================================================

// Interpolation is a pointer to either [BadInterpolation],
// [ExpressionInterpolation], [TextInterpolation], [ElementInterpolation],
// [ComponentCallInterpolation], or [CharacterReference].
type Interpolation interface {
	TextNode
	_interpolation()
}

// if this is changed, change the comment above
var (
	_ Interpolation = (*BadInterpolation)(nil)
	_ Interpolation = (*EscapedHash)(nil)
	_ Interpolation = (*HashSpace)(nil)
	_ Interpolation = (*ExpressionInterpolation)(nil)
	_ Interpolation = (*ElementInterpolation)(nil)
	_ Interpolation = (*ComponentCallInterpolation)(nil)
	_ Interpolation = (*CharacterReference)(nil)
)

// ================================= Bad Interpolation ==================================

type BadInterpolation struct {
	Position Position
}

var (
	_ Interpolation       = (*BadInterpolation)(nil)
	_ StringInterpolation = (*BadInterpolation)(nil)
)

func (b *BadInterpolation) Pos() Position { return b.Position }
func (b *BadInterpolation) End() Position { return deltaPos(b.Position, len("#")) }

func (*BadInterpolation) _node()                {}
func (*BadInterpolation) _interpolation()       {}
func (*BadInterpolation) _textNode()            {}
func (*BadInterpolation) _stringInterpolation() {}
func (*BadInterpolation) _stringContent()       {}

// ===================================== EscapedHash =====================================

type EscapedHash struct { // ##
	Position
}

var (
	_ Interpolation       = (*EscapedHash)(nil)
	_ StringInterpolation = (*EscapedHash)(nil)
)

func (h *EscapedHash) Pos() Position { return h.Position }
func (h *EscapedHash) End() Position { return deltaPos(h.Position, len("##")) }

func (*EscapedHash) _node()                {}
func (*EscapedHash) _textNode()            {}
func (*EscapedHash) _interpolation()       {}
func (*EscapedHash) _stringContent()       {}
func (*EscapedHash) _stringInterpolation() {}

// ===================================== HashSpace =====================================

type HashSpace struct { // #_
	Position
}

var _ Interpolation = (*HashSpace)(nil)

func (h *HashSpace) Pos() Position { return h.Position }
func (h *HashSpace) End() Position { return deltaPos(h.Position, len("#_")) }

func (*HashSpace) _node()          {}
func (*HashSpace) _textNode()      {}
func (*HashSpace) _interpolation() {}

// ============================== ExpressionInterpolation ===============================

type ExpressionInterpolation struct {
	// a sprintf placeholder, excluding the leading %
	FormatDirective string
	LBrace          *Position
	Expression      Expression
	RBrace          *Position

	Position Position
}

var (
	_ Interpolation       = (*ExpressionInterpolation)(nil)
	_ StringInterpolation = (*ExpressionInterpolation)(nil)
)

func (interp *ExpressionInterpolation) Pos() Position { return interp.Position }
func (interp *ExpressionInterpolation) End() Position {
	if interp.RBrace != nil {
		return *interp.RBrace
	} else if interp.Expression != nil {
		return interp.Expression.End()
	} else if interp.LBrace != nil {
		return deltaPos(*interp.LBrace, 1)
	}
	return deltaPos(interp.Position, len("#")+len(interp.FormatDirective))
}

func (*ExpressionInterpolation) _node()                {}
func (*ExpressionInterpolation) _textNode()            {}
func (*ExpressionInterpolation) _interpolation()       {}
func (*ExpressionInterpolation) _stringContent()       {}
func (*ExpressionInterpolation) _stringInterpolation() {}

// ================================ ElementInterpolation ================================

type ElementInterpolation struct {
	Element *Element            // has no body
	Value   *InterpolationValue // may be nil, always nil for void elems

	Position Position
}

var _ Interpolation = (*ElementInterpolation)(nil)

func (interp *ElementInterpolation) Pos() Position { return interp.Position }
func (interp *ElementInterpolation) End() Position {
	if interp.Value != nil {
		return interp.Value.End()
	} else if interp.Element != nil {
		return interp.Element.End()
	}
	return deltaPos(interp.Position, len("#"))
}

func (*ElementInterpolation) _node()          {}
func (*ElementInterpolation) _textNode()      {}
func (*ElementInterpolation) _interpolation() {}

// ================================= ComponentCallInterpolation =================================

type ComponentCallInterpolation struct {
	ComponentCall *ComponentCall
	Value         *InterpolationValue // may be nil

	Position Position
}

var _ Interpolation = (*ComponentCallInterpolation)(nil)

func (interp *ComponentCallInterpolation) Pos() Position { return interp.Position }
func (interp *ComponentCallInterpolation) End() Position {
	if interp.Value != nil {
		return interp.Value.End()
	} else if interp.ComponentCall != nil {
		return interp.ComponentCall.End()
	}
	return deltaPos(interp.Position, len("#"))
}

func (*ComponentCallInterpolation) _node()          {}
func (*ComponentCallInterpolation) _textNode()      {}
func (*ComponentCallInterpolation) _interpolation() {}

// ================================= CharacterReference =================================

type CharacterReference struct {
	Name string // w/o & and ;
	Position
}

var (
	_ Interpolation       = (*CharacterReference)(nil)
	_ StringInterpolation = (*CharacterReference)(nil)
)

func (c *CharacterReference) Pos() Position { return c.Position }
func (c *CharacterReference) End() Position {
	return deltaPos(c.Position, len("#")+len(c.Name)+len(";"))
}

func (*CharacterReference) _node()                {}
func (*CharacterReference) _textNode()            {}
func (*CharacterReference) _interpolation()       {}
func (*CharacterReference) _stringContent()       {}
func (*CharacterReference) _stringInterpolation() {}

// ============================================================================
// InterpolationValue
// ======================================================================================

type InterpolationValue struct {
	LBracket Position
	Text     TextLine
	RBracket *Position
}

var _ Node = (*InterpolationValue)(nil)

func (v *InterpolationValue) Pos() Position { return v.LBracket }
func (v *InterpolationValue) End() Position {
	if v.RBracket != nil {
		return *v.RBracket
	} else if len(v.Text) > 0 {
		return v.Text[len(v.Text)-1].End()
	}
	return deltaPos(v.LBracket, 1)
}

func (*InterpolationValue) _node() {}
