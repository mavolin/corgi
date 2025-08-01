package ast

type Interpolation interface {
	Node
	_interpolation()
}

// ============================================================================
// Bad TextInterpolation
// ======================================================================================

type BadInterpolation struct {
	From, Until Position
}

var (
	_ TextInterpolation   = (*BadInterpolation)(nil)
	_ StringInterpolation = (*BadInterpolation)(nil)
)

func (b *BadInterpolation) Start() Position { return b.From }
func (b *BadInterpolation) End() Position   { return b.Until }
func (b *BadInterpolation) Walk(func(Node)) {}

func (*BadInterpolation) _node()          {}
func (*BadInterpolation) _interpolation() {}
func (*BadInterpolation) _textNode()      {}
func (*BadInterpolation) _stringNode()    {}

// ============================================================================
// Escaped Hash
// ======================================================================================

type EscapedHash struct { // ##
	Hash *Position
}

var (
	_ TextInterpolation   = (*EscapedHash)(nil)
	_ StringInterpolation = (*EscapedHash)(nil)
	_ ContentWriter       = (*EscapedHash)(nil)
)

func (h *EscapedHash) Start() Position {
	if h.Hash != nil {
		return *h.Hash
	}
	return Position{}
}

func (h *EscapedHash) End() Position {
	if h.Hash != nil {
		return deltaPos(*h.Hash, len("##"))
	}
	return Position{}
}
func (h *EscapedHash) Walk(func(Node)) {}

func (*EscapedHash) _node()          {}
func (*EscapedHash) _interpolation() {}
func (*EscapedHash) _textNode()      {}
func (*EscapedHash) _stringNode()    {}
func (*EscapedHash) _contentWriter() {}

// ============================================================================
// Hash Space
// ======================================================================================

type HashSpace struct { // #_
	Hash *Position
}

var (
	_ TextInterpolation = (*HashSpace)(nil)
	_ ContentWriter     = (*HashSpace)(nil)
)

func (h *HashSpace) Start() Position {
	if h.Hash != nil {
		return *h.Hash
	}
	return Position{}
}

func (h *HashSpace) End() Position {
	if h.Hash != nil {
		return deltaPos(*h.Hash, len("#_"))
	}
	return Position{}
}
func (h *HashSpace) Walk(func(Node)) {}

func (*HashSpace) _node()          {}
func (*HashSpace) _interpolation() {}
func (*HashSpace) _textNode()      {}
func (*HashSpace) _contentWriter() {}

// ============================================================================
// Hash Right Bracket
// ======================================================================================

type EscapedRBracket struct { // #]
	Hash *Position
}

var (
	_ TextInterpolation = (*EscapedRBracket)(nil)
	_ ContentWriter     = (*EscapedRBracket)(nil)
)

func (h *EscapedRBracket) Start() Position {
	if h.Hash != nil {
		return *h.Hash
	}
	return Position{}
}

func (h *EscapedRBracket) End() Position {
	if h.Hash != nil {
		return deltaPos(*h.Hash, len("#]"))
	}
	return Position{}
}
func (h *EscapedRBracket) Walk(func(Node)) {}

func (*EscapedRBracket) _node()          {}
func (*EscapedRBracket) _interpolation() {}
func (*EscapedRBracket) _textNode()      {}
func (*EscapedRBracket) _contentWriter() {}

// ============================================================================
// Expression Interpolation
// ======================================================================================

type ExpressionInterpolation struct {
	Hash            *Position
	FormatDirective string // a sprintf placeholder, excluding the leading %
	LBrace          *Position
	Expression      *Expression
	RBrace          *Position
}

var (
	_ TextInterpolation   = (*ExpressionInterpolation)(nil)
	_ StringInterpolation = (*ExpressionInterpolation)(nil)
	_ ContentWriter       = (*ExpressionInterpolation)(nil)
)

func (interp *ExpressionInterpolation) Start() Position {
	switch {
	case interp.Hash != nil:
		return *interp.Hash
	case interp.LBrace != nil:
		return *interp.LBrace
	case interp.Expression != nil:
		return interp.Expression.Start()
	case interp.RBrace != nil:
		return *interp.RBrace
	}
	return Position{}
}

func (interp *ExpressionInterpolation) End() Position {
	switch {
	case interp.RBrace != nil:
		return deltaPos(*interp.RBrace, len("}"))
	case interp.Expression != nil:
		return interp.Expression.End()
	case interp.LBrace != nil:
		return deltaPos(*interp.LBrace, len("{"))
	case interp.Hash != nil:
		if interp.FormatDirective != "" {
			return deltaPos(*interp.Hash, len("#%")+len(interp.FormatDirective))
		}
		return deltaPos(*interp.Hash, len("#"))
	}
	return Position{}
}

func (interp *ExpressionInterpolation) Walk(w func(Node)) {
	if interp.Expression != nil {
		w(interp.Expression)
	}
}

func (*ExpressionInterpolation) _node()          {}
func (*ExpressionInterpolation) _interpolation() {}
func (*ExpressionInterpolation) _textNode()      {}
func (*ExpressionInterpolation) _stringNode()    {}
func (*ExpressionInterpolation) _contentWriter() {}

// ============================================================================
// Element TextInterpolation
// ======================================================================================

type ElementInterpolation struct {
	Hash    *Position
	Element *Element // Body is BracketText, if present
}

var _ TextInterpolation = (*ElementInterpolation)(nil)

func (interp *ElementInterpolation) Start() Position {
	if interp.Hash != nil {
		return *interp.Hash
	} else if interp.Element != nil {
		return interp.Element.Start()
	}
	return Position{}
}

func (interp *ElementInterpolation) End() Position {
	if interp.Element != nil {
		return interp.Element.End()
	} else if interp.Hash != nil {
		return deltaPos(*interp.Hash, len("#"))
	}
	return Position{}
}

func (interp *ElementInterpolation) Walk(w func(Node)) {
	if interp.Element != nil {
		w(interp.Element)
	}
}

func (*ElementInterpolation) _node()          {}
func (*ElementInterpolation) _interpolation() {}
func (*ElementInterpolation) _textNode()      {}

// ============================================================================
// Component Call TextInterpolation
// ======================================================================================

type ComponentCallInterpolation struct {
	Hash          *Position
	ComponentCall *ComponentCall // Body is implicit DefaultBlockShorthand, if present
}

var (
	_ TextInterpolation   = (*ComponentCallInterpolation)(nil)
	_ StringInterpolation = (*ComponentCallInterpolation)(nil)
)

func (interp *ComponentCallInterpolation) Start() Position {
	if interp.Hash != nil {
		return *interp.Hash
	} else if interp.ComponentCall != nil {
		return interp.ComponentCall.Start()
	}
	return Position{}
}

func (interp *ComponentCallInterpolation) End() Position {
	if interp.ComponentCall != nil {
		return interp.ComponentCall.End()
	} else if interp.Hash != nil {
		return deltaPos(*interp.Hash, len("#"))
	}
	return Position{}
}

func (interp *ComponentCallInterpolation) Walk(w func(Node)) {
	if interp.ComponentCall != nil {
		w(interp.ComponentCall)
	}
}

func (*ComponentCallInterpolation) _node()          {}
func (*ComponentCallInterpolation) _interpolation() {}
func (*ComponentCallInterpolation) _textNode()      {}
func (*ComponentCallInterpolation) _stringNode()    {}

// ============================================================================
// Character Reference
// ======================================================================================

type CharacterReference struct {
	Hash  *Position
	Name  string // w/o & and ;
	Chars string // the characters it represents
}

var (
	_ TextInterpolation   = (*CharacterReference)(nil)
	_ StringInterpolation = (*CharacterReference)(nil)
	_ ContentWriter       = (*CharacterReference)(nil)
)

func (c *CharacterReference) Start() Position {
	if c.Hash != nil {
		return *c.Hash
	}
	return Position{}
}

func (c *CharacterReference) End() Position {
	if c.Hash != nil {
		return deltaPos(*c.Hash, len("#")+len(c.Name)+len(";"))
	}
	return Position{}
}
func (c *CharacterReference) Walk(func(Node)) {}

func (*CharacterReference) _node()          {}
func (*CharacterReference) _interpolation() {}
func (*CharacterReference) _textNode()      {}
func (*CharacterReference) _stringNode()    {}
func (*CharacterReference) _contentWriter() {}
