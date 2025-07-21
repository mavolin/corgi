package ast

// ============================================================================
// Component Call
// ======================================================================================

type ComponentCall struct {
	Colon  *Position
	Header *ComponentCallHeader
	Body   Body
}

var (
	_ ScopeNode   = (*ComponentCall)(nil)
	_ Highlighter = (*ComponentCall)(nil)
)

func (c *ComponentCall) Start() Position {
	switch {
	case c.Colon != nil:
		return *c.Colon
	case c.Header != nil:
		return c.Header.Start()
	case c.Body != nil:
		return c.Body.Start()
	}
	return Position{}
}

func (c *ComponentCall) End() Position {
	switch {
	case c.Body != nil:
		return c.Body.End()
	case c.Header != nil:
		return c.Header.End()
	case c.Colon != nil:
		return deltaPos(*c.Colon, len(":"))
	}
	return Position{}
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
			return *c.Colon, c.Header.Name.End()
		}
		return *c.Colon, deltaPos(*c.Colon, len(":"))
	}
	return c.Start(), c.End()
}

func (*ComponentCall) _node()      {}
func (*ComponentCall) _scopeNode() {}

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
	switch {
	case h.Name != nil:
		return h.Name.Start()
	case h.TypeArguments != nil:
		return h.TypeArguments.Start()
	case h.Arguments != nil:
		return h.Arguments.Start()
	}
	return Position{}
}

func (h *ComponentCallHeader) End() Position {
	switch {
	case h.Arguments != nil:
		return h.Arguments.End()
	case h.TypeArguments != nil:
		return h.TypeArguments.End()
	case h.Name != nil:
		return h.Name.End()
	}
	return Position{}
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
// With
// ======================================================================================

type With struct {
	With       *Position
	Identifier *Identifier // optional for default block
	Body       Body
}

var (
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
	switch {
	case w.With != nil:
		return *w.With
	case w.Identifier != nil:
		return w.Identifier.Start()
	case w.Body != nil:
		return w.Body.Start()
	}
	return Position{}
}

func (w *With) End() Position {
	switch {
	case w.Body != nil:
		return w.Body.End()
	case w.Identifier != nil:
		return w.Identifier.End()
	case w.With != nil:
		return deltaPos(*w.With, len("with"))
	}
	return Position{}
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
			return *w.With, w.Identifier.End()
		}
		return *w.With, deltaPos(*w.With, len("with"))
	}
	return w.Start(), w.End()
}

func (*With) _node()      {}
func (*With) _scopeNode() {}
