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

func (*BadInterpolation) _node()                  {}
func (*BadInterpolation) _interpolation()         {}
func (*BadInterpolation) _textNode()              {}
func (*BadInterpolation) _interpretedStringNode() {}
func (*BadInterpolation) _rawStringNode()         {}

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
	_ TextInterpolation           = (*CharacterEscape)(nil)
	_ StringInterpolation         = (*CharacterEscape)(nil)
	_ ConstantStringInterpolation = (*CharacterEscape)(nil)
	_ ContentWriter               = (*CharacterEscape)(nil)
	_ AttributeInhibitor          = (*CharacterEscape)(nil)
)

func (ce *CharacterEscape) Start() Position {
	if ce.Hash != nil {
		return *ce.Hash
	}
	return NoPosition
}

func (ce *CharacterEscape) End() Position {
	if ce.Hash != nil {
		return deltaPos(*ce.Hash, len("#_"))
	}
	return NoPosition
}

func (ce *CharacterEscape) Walk(func(Node)) {}

// ConstantValue returns the replacement for the character escape, which is the
// character it represents.
//
// It always returns true.
func (ce *CharacterEscape) ConstantValue() (string, bool) {
	return string(ce.Rune), true
}

func (*CharacterEscape) _node()                  {}
func (*CharacterEscape) _interpolation()         {}
func (*CharacterEscape) _textNode()              {}
func (*CharacterEscape) _interpretedStringNode() {}
func (*CharacterEscape) _rawStringNode()         {}
func (*CharacterEscape) _contentWriter()         {}
func (*CharacterEscape) _attributeInhibitor()    {}

// ============================================================================
// Expression Interpolation
// ======================================================================================

type ExpressionInterpolation struct {
	Hash       *Position
	LBrace     *Position
	Expression *Expression
	RBrace     *Position
}

var (
	_ TextInterpolation   = (*ExpressionInterpolation)(nil)
	_ StringInterpolation = (*ExpressionInterpolation)(nil)
	_ ContentWriter       = (*ExpressionInterpolation)(nil)
	_ AttributeInhibitor  = (*ExpressionInterpolation)(nil)
)

func (ei *ExpressionInterpolation) Start() Position {
	if ei.Hash != nil {
		return *ei.Hash
	}
	if ei.LBrace != nil {
		return *ei.LBrace
	}
	if ei.Expression != nil {
		if start := ei.Expression.Start(); start != NoPosition {
			return start
		}
	}
	if ei.RBrace != nil {
		return *ei.RBrace
	}
	return NoPosition
}

func (ei *ExpressionInterpolation) End() Position {
	if ei.RBrace != nil {
		return deltaPos(*ei.RBrace, len("}"))
	}
	if ei.Expression != nil {
		if end := ei.Expression.End(); end != NoPosition {
			return end
		}
	}
	if ei.LBrace != nil {
		return deltaPos(*ei.LBrace, len("{"))
	}
	if ei.Hash != nil {
		return deltaPos(*ei.Hash, len("#"))
	}
	return NoPosition
}

func (ei *ExpressionInterpolation) Walk(w func(Node)) {
	if ei.Expression != nil {
		w(ei.Expression)
	}
}

func (*ExpressionInterpolation) _node()                  {}
func (*ExpressionInterpolation) _interpolation()         {}
func (*ExpressionInterpolation) _textNode()              {}
func (*ExpressionInterpolation) _interpretedStringNode() {}
func (*ExpressionInterpolation) _rawStringNode()         {}
func (*ExpressionInterpolation) _contentWriter()         {}
func (*ExpressionInterpolation) _attributeInhibitor()    {}

// ============================================================================
// Character Reference
// ======================================================================================

type CharacterReference struct {
	Hash  *Position
	Name  string // w/o & and ;
	Chars string // the characters it represents
}

var (
	_ TextInterpolation           = (*CharacterReference)(nil)
	_ StringInterpolation         = (*CharacterReference)(nil)
	_ ConstantStringInterpolation = (*CharacterReference)(nil)
	_ ContentWriter               = (*CharacterReference)(nil)
	_ AttributeInhibitor          = (*CharacterReference)(nil)
)

func (r *CharacterReference) Start() Position {
	if r.Hash != nil {
		return *r.Hash
	}
	return NoPosition
}

func (r *CharacterReference) End() Position {
	if r.Hash != nil {
		return deltaPos(*r.Hash, len("#")+len([]rune(r.Name))+len(";"))
	}
	return NoPosition
}
func (r *CharacterReference) Walk(func(Node)) {}

// ConstantValue returns the replacement for the character reference, which is
// the characters it represents.
//
// If the character reference is unknown, it returns an empty string and false.
func (r *CharacterReference) ConstantValue() (string, bool) {
	return r.Chars, r.Chars != ""
}

func (*CharacterReference) _node()                  {}
func (*CharacterReference) _interpolation()         {}
func (*CharacterReference) _textNode()              {}
func (*CharacterReference) _interpretedStringNode() {}
func (*CharacterReference) _rawStringNode()         {}
func (*CharacterReference) _contentWriter()         {}
func (*CharacterReference) _attributeInhibitor()    {}

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
	}
	if s.Node != nil {
		if start := s.Node.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (s *ModeSwitch) End() Position {
	if s.Node != nil {
		if end := s.Node.End(); end != NoPosition {
			return end
		}
	}
	if s.Hash != nil {
		return deltaPos(*s.Hash, len("#"))
	}
	return NoPosition
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

// ComponentCallInterpolation allows using component calls inside strings.
// Within text, the same can be achieved with a [ModeSwitch].
type ComponentCallInterpolation struct {
	Hash          *Position
	ComponentCall *ComponentCall // Body is implicit DefaultBlockShorthand, if present
}

var _ StringInterpolation = (*ComponentCallInterpolation)(nil)

func (cci *ComponentCallInterpolation) Start() Position {
	if cci.Hash != nil {
		return *cci.Hash
	}
	if cci.ComponentCall != nil {
		if start := cci.ComponentCall.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (cci *ComponentCallInterpolation) End() Position {
	if cci.ComponentCall != nil {
		if end := cci.ComponentCall.End(); end != NoPosition {
			return end
		}
	}
	if cci.Hash != nil {
		return deltaPos(*cci.Hash, len("#"))
	}
	return NoPosition
}

func (cci *ComponentCallInterpolation) Walk(w func(Node)) {
	if cci.ComponentCall != nil {
		w(cci.ComponentCall)
	}
}

func (*ComponentCallInterpolation) _node()                  {}
func (*ComponentCallInterpolation) _interpolation()         {}
func (*ComponentCallInterpolation) _interpretedStringNode() {}
func (*ComponentCallInterpolation) _rawStringNode()         {}
