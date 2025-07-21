package ast

import "slices"

// ============================================================================
// Arrow BlockName
// ======================================================================================

type ArrowBlock struct {
	Arrow *Position
	Lines TextBlock
}

var (
	_ ScopeNode   = (*ArrowBlock)(nil)
	_ Highlighter = (*ArrowBlock)(nil)
)

func (b *ArrowBlock) Start() Position {
	if b.Arrow != nil {
		return *b.Arrow
	}
	if len(b.Lines) > 0 {
		return b.Lines.Start()
	}
	return Position{}
}

func (b *ArrowBlock) End() Position {
	if len(b.Lines) > 0 {
		return b.Lines.End()
	} else if b.Arrow != nil {
		return deltaPos(*b.Arrow, len(">"))
	}
	return Position{}
}

func (b *ArrowBlock) Walk(w func(Node)) {
	if b.Lines != nil {
		w(b.Lines)
	}
}

func (b *ArrowBlock) Highlight() (start, end Position) {
	if len(b.Lines) != 1 && b.Arrow != nil {
		return *b.Arrow, deltaPos(*b.Arrow, len(">"))
	}
	return b.Start(), b.End()
}

func (*ArrowBlock) _node()      {}
func (*ArrowBlock) _scopeNode() {}

// ============================================================================
// Text Node
// ======================================================================================

type (
	TextBlock []TextLine
	TextLine  []TextNode

	// TextNode is a pointer to either a pointer to [Text] or [TextInterpolation].
	TextNode interface {
		Node
		_textNode()
	}
)

var (
	_ Node = (TextBlock)(nil)
	_ Node = (TextLine)(nil)

	// change the above comment if this changes
	_ TextNode = (*Text)(nil)
	_ TextNode = (TextInterpolation)(nil)
)

func (b TextBlock) Start() Position {
	for _, l := range b {
		if len(l) > 0 {
			return l.Start()
		}
	}
	return Position{}
}

func (b TextBlock) End() Position {
	for _, l := range slices.Backward(b) {
		if len(l) > 0 {
			return l.End()
		}
	}
	return Position{}
}

func (b TextBlock) Walk(w func(Node)) {
	for _, l := range b {
		if l != nil {
			w(l)
		}
	}
}

func (TextBlock) _node() {}

func (l TextLine) Start() Position {
	if len(l) == 0 {
		return Position{}
	}

	return l[0].Start()
}

func (l TextLine) End() Position {
	if len(l) > 0 {
		return l[len(l)-1].End()
	}
	return Position{}
}

func (l TextLine) Walk(w func(Node)) {
	for _, n := range l {
		if n != nil {
			w(n)
		}
	}
}

func (TextLine) _node() {}

// ============================================================================
// Text
// ======================================================================================

// Text is a string of text written as content of an element.
// It is not HTML-escaped yet.
type Text struct {
	Text     string
	Position *Position
}

var _ TextNode = (*Text)(nil)

func (t *Text) Start() Position {
	if t.Position != nil {
		return *t.Position
	}
	return Position{}
}

func (t *Text) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return Position{}
}
func (t *Text) Walk(func(Node)) {}

func (t *Text) _node()   {}
func (*Text) _textNode() {}

// ============================================================================
// Text Interpolation
// ======================================================================================

// TextInterpolation is a pointer to either [BadInterpolation],
// an [EscapedHash], a [HashSpace], a [EscapedRBracket], an [ElementInterpolation],
// a [ComponentCallInterpolation], or a [CharacterReference].
type TextInterpolation interface {
	TextNode
	Interpolation
}

// if this is changed, change the comment above
var (
	_ TextInterpolation = (*BadInterpolation)(nil)
	_ TextInterpolation = (*EscapedHash)(nil)
	_ TextInterpolation = (*HashSpace)(nil)
	_ TextInterpolation = (*EscapedRBracket)(nil)
	_ TextInterpolation = (*ExpressionInterpolation)(nil)
	_ TextInterpolation = (*ElementInterpolation)(nil)
	_ TextInterpolation = (*ComponentCallInterpolation)(nil)
	_ TextInterpolation = (*CharacterReference)(nil)
)
