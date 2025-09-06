package ast

// ============================================================================
// Component Call
// ======================================================================================

type ComponentCall struct {
	Colon  *Position
	Header *ComponentCallHeader
	Body   ComponentCallBody
}

var (
	_ ScopeNode            = (*ComponentCall)(nil)
	_ ContentWriter        = (*ComponentCall)(nil)
	_ ElementWriter        = (*ComponentCall)(nil)
	_ AttributeWriter      = (*ComponentCall)(nil)
	_ AttributeInhibitor   = (*ComponentCall)(nil)
	_ AndPlaceholderWriter = (*ComponentCall)(nil)
	_ CodeNode             = (*ComponentCall)(nil)
	_ Highlighter          = (*ComponentCall)(nil)
)

func (c *ComponentCall) Start() Position {
	if c.Colon != nil {
		return *c.Colon
	}
	if c.Header != nil {
		if start := c.Header.Start(); start != NoPosition {
			return start
		}
	}
	if c.Body != nil {
		if start := c.Body.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (c *ComponentCall) End() Position {
	if c.Body != nil {
		if end := c.Body.End(); end != NoPosition {
			return end
		}
	}
	if c.Header != nil {
		if end := c.Header.End(); end != NoPosition {
			return end
		}
	}
	if c.Colon != nil {
		return deltaPos(*c.Colon, len(":"))
	}
	return NoPosition
}

func (c *ComponentCall) Walk(w func(Node)) {
	if c.Header != nil {
		w(c.Header)
	}
	if c.Body != nil {
		w(c.Body)
	}
}

func (c *ComponentCall) Highlight() (start, end Position) {
	if c.Colon != nil {
		if c.Header != nil && c.Header.Name != nil {
			if end = c.Header.Name.End(); end != NoPosition {
				return *c.Colon, end
			}
		}
		return *c.Colon, deltaPos(*c.Colon, len(":"))
	}
	return c.Start(), c.End()
}

func (*ComponentCall) _node()                 {}
func (*ComponentCall) _scopeNode()            {}
func (*ComponentCall) _codeNode()             {}
func (*ComponentCall) _contentWriter()        {}
func (*ComponentCall) _elementWriter()        {}
func (*ComponentCall) _attributeWriter()      {}
func (*ComponentCall) _attributeInhibitor()   {}
func (*ComponentCall) _andPlaceholderWriter() {}

// ============================================================================
// Component Call Header
// ======================================================================================

type ComponentCallHeader struct {
	Name          FullIdentifier
	TypeArguments *TypeArguments // optional
	Arguments     *Arguments     // optional
}

var _ Node = (*ComponentCallHeader)(nil)

func (h *ComponentCallHeader) Start() Position {
	if h.Name != nil {
		if start := h.Name.Start(); start != NoPosition {
			return start
		}
	}
	if h.TypeArguments != nil {
		if start := h.TypeArguments.Start(); start != NoPosition {
			return start
		}
	}
	if h.Arguments != nil {
		if start := h.Arguments.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (h *ComponentCallHeader) End() Position {
	if h.Arguments != nil {
		if end := h.Arguments.End(); end != NoPosition {
			return end
		}
	}
	if h.TypeArguments != nil {
		if end := h.TypeArguments.End(); end != NoPosition {
			return end
		}
	}
	if h.Name != nil {
		if end := h.Name.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (h *ComponentCallHeader) Walk(w func(Node)) {
	if h.Name != nil {
		w(h.Name)
	}
	if h.TypeArguments != nil {
		w(h.TypeArguments)
	}
	if h.Arguments != nil {
		w(h.Arguments)
	}
}

func (*ComponentCallHeader) _node() {}

// ============================================================================
// Component Call Body
// ======================================================================================

// ComponentCallBody is the body of a component, either [Scope], or
// [DefaultBlockShorthand].
type ComponentCallBody interface {
	Node
	_componentCallBody()
}

// if this is changed, change the comment above
var (
	_ ComponentCallBody = (*Scope)(nil)
	_ ComponentCallBody = (*DefaultBlockShorthand)(nil)
)

// ============================================================================
// Default Block Shorthand
// ======================================================================================

type DefaultBlockShorthand struct {
	Body     Body
	Position *Position
}

func (s *DefaultBlockShorthand) Highlight() (start, end Position) {
	if s.Position != nil {
		return *s.Position, deltaPos(*s.Position, len("_{"))
	}
	return s.Body.Highlight()
}

var (
	_ ComponentCallBody = (*DefaultBlockShorthand)(nil)
	_ BlockSetter       = (*DefaultBlockShorthand)(nil)
	_ Highlighter       = (*DefaultBlockShorthand)(nil)
)

func (s *DefaultBlockShorthand) Name() string {
	return ""
}

func (s *DefaultBlockShorthand) Start() Position {
	if s.Position != nil {
		return *s.Position
	}
	if s.Body != nil {
		if start := s.Body.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (s *DefaultBlockShorthand) End() Position {
	if s.Body != nil {
		if end := s.Body.End(); end != NoPosition {
			return end
		}
	}
	if s.Position != nil {
		return deltaPos(*s.Position, len("_"))
	}
	return NoPosition
}

func (s *DefaultBlockShorthand) Walk(w func(Node)) {
	if s.Body != nil {
		w(s.Body)
	}
}

func (*DefaultBlockShorthand) _node()              {}
func (*DefaultBlockShorthand) _with()              {}
func (*DefaultBlockShorthand) _componentCallBody() {}

// ============================================================================
// BlockSetter
// ======================================================================================

// BlockSetter is either a [With] or a [DefaultBlockShorthand].
type BlockSetter interface {
	Node
	_with()
	Name() string
}

// if this is changed, change the comment above
var (
	_ BlockSetter = (*With)(nil)
	_ BlockSetter = (*DefaultBlockShorthand)(nil)
)

// ============================================================================
// With
// ======================================================================================

type With struct {
	With       *Position
	Identifier *Identifier // optional for default block
	Body       Body
}

var (
	_ BlockSetter = (*With)(nil)
	_ ScopeNode   = (*With)(nil)
	_ Highlighter = (*With)(nil)
)

// Name returns the name of the block, "" for the default block.
func (w *With) Name() string {
	if w.Identifier != nil {
		return w.Identifier.Name
	}
	return ""
}

func (w *With) Start() Position {
	if w.With != nil {
		return *w.With
	}
	if w.Identifier != nil {
		if start := w.Identifier.Start(); start != NoPosition {
			return start
		}
	}
	if w.Body != nil {
		if start := w.Body.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (w *With) End() Position {
	if w.Body != nil {
		if end := w.Body.End(); end != NoPosition {
			return end
		}
	}
	if w.Identifier != nil {
		if end := w.Identifier.End(); end != NoPosition {
			return end
		}
	}
	if w.With != nil {
		return deltaPos(*w.With, len("with"))
	}
	return NoPosition
}

func (w *With) Walk(f func(Node)) {
	if w.Identifier != nil {
		f(w.Identifier)
	}
	if w.Body != nil {
		f(w.Body)
	}
}

func (w *With) Highlight() (start, end Position) {
	if w.With != nil {
		if w.Identifier != nil {
			if end := w.Identifier.End(); end != NoPosition {
				return *w.With, end
			}
		}
		return *w.With, deltaPos(*w.With, len("with"))
	}
	return w.Start(), w.End()
}

func (*With) _node()      {}
func (*With) _with()      {}
func (*With) _scopeNode() {}
