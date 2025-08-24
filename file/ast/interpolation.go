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
// Character Escape
// ======================================================================================

// CharacterEscape is for escaping a special character in text or string.
//
// In strings, that is '#'.
// In text, that is '#', ' ', and ']'.
type CharacterEscape struct {
	Hash   *Position
	Symbol rune // one of '#', '_', ']'
	Rune   rune // the actual rune it represents, '#', ' ', or ']'
}

var (
	_ TextInterpolation   = (*CharacterEscape)(nil)
	_ StringInterpolation = (*CharacterEscape)(nil)
	_ ContentWriter       = (*CharacterEscape)(nil)
)

func (ce *CharacterEscape) Start() Position {
	if ce.Hash != nil {
		return *ce.Hash
	}
	return Position{}
}

func (ce *CharacterEscape) End() Position {
	if ce.Hash != nil {
		return deltaPos(*ce.Hash, len("#_"))
	}
	return Position{}
}

func (ce *CharacterEscape) Walk(func(Node)) {}

func (*CharacterEscape) _node()          {}
func (*CharacterEscape) _interpolation() {}
func (*CharacterEscape) _textNode()      {}
func (*CharacterEscape) _stringNode()    {}
func (*CharacterEscape) _contentWriter() {}

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

func (ei *ExpressionInterpolation) Start() Position {
	switch {
	case ei.Hash != nil:
		return *ei.Hash
	case ei.LBrace != nil:
		return *ei.LBrace
	case ei.Expression != nil:
		return ei.Expression.Start()
	case ei.RBrace != nil:
		return *ei.RBrace
	}
	return Position{}
}

func (ei *ExpressionInterpolation) End() Position {
	switch {
	case ei.RBrace != nil:
		return deltaPos(*ei.RBrace, len("}"))
	case ei.Expression != nil:
		return ei.Expression.End()
	case ei.LBrace != nil:
		return deltaPos(*ei.LBrace, len("{"))
	case ei.Hash != nil:
		return deltaPos(*ei.Hash, len("#"))
	}
	return Position{}
}

func (ei *ExpressionInterpolation) Walk(w func(Node)) {
	if ei.Expression != nil {
		w(ei.Expression)
	}
}

func (*ExpressionInterpolation) _node()          {}
func (*ExpressionInterpolation) _interpolation() {}
func (*ExpressionInterpolation) _textNode()      {}
func (*ExpressionInterpolation) _stringNode()    {}
func (*ExpressionInterpolation) _contentWriter() {}

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

func (r *CharacterReference) Start() Position {
	if r.Hash != nil {
		return *r.Hash
	}
	return Position{}
}

func (r *CharacterReference) End() Position {
	if r.Hash != nil {
		return deltaPos(*r.Hash, len("#")+len(r.Name)+len(";"))
	}
	return Position{}
}
func (r *CharacterReference) Walk(func(Node)) {}

func (*CharacterReference) _node()          {}
func (*CharacterReference) _interpolation() {}
func (*CharacterReference) _textNode()      {}
func (*CharacterReference) _stringNode()    {}
func (*CharacterReference) _contentWriter() {}

// ============================================================================
// Mode Switch
// ======================================================================================

type ModeSwitch struct {
	Hash *Position
	Node ScopeNode
}

func (s *ModeSwitch) Start() Position {
	if s.Hash != nil {
		return *s.Hash
	} else if s.Node != nil {
		return s.Node.Start()
	}
	return Position{}
}

func (s *ModeSwitch) End() Position {
	if s.Node != nil {
		return s.Node.End()
	} else if s.Hash != nil {
		return deltaPos(*s.Hash, len("#"))
	}
	return Position{}
}

func (s *ModeSwitch) Walk(f func(Node)) {
	f(s.Node)
}

func (s *ModeSwitch) _interpolation() {}
func (s *ModeSwitch) _node()          {}
func (s *ModeSwitch) _textNode()      {}

var _ TextInterpolation = (*ModeSwitch)(nil)

// ============================================================================
// Component Call Interpolation
// ======================================================================================

// ComponentCallInterpolation is a string interpolation.
// It has no body.
// Within text, the equivalent is you should use a mode switch.
type ComponentCallInterpolation struct {
	Hash          *Position
	ComponentCall *ComponentCall // Body is implicit DefaultBlockShorthand, if present
}

var _ StringInterpolation = (*ComponentCallInterpolation)(nil)

func (cci *ComponentCallInterpolation) Start() Position {
	if cci.Hash != nil {
		return *cci.Hash
	} else if cci.ComponentCall != nil {
		return cci.ComponentCall.Start()
	}
	return Position{}
}

func (cci *ComponentCallInterpolation) End() Position {
	if cci.ComponentCall != nil {
		return cci.ComponentCall.End()
	} else if cci.Hash != nil {
		return deltaPos(*cci.Hash, len("#"))
	}
	return Position{}
}

func (cci *ComponentCallInterpolation) Walk(w func(Node)) {
	if cci.ComponentCall != nil {
		w(cci.ComponentCall)
	}
}

func (*ComponentCallInterpolation) _node()          {}
func (*ComponentCallInterpolation) _interpolation() {}
func (*ComponentCallInterpolation) _stringNode()    {}
