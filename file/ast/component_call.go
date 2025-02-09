package ast

// ============================================================================
// Component Call
// ======================================================================================

type ComponentCall struct {
	Colon  *Position
	Header *ComponentCallHeader
	Body   Body
}

var _ ScopeNode = (*ComponentCall)(nil)

func (c *ComponentCall) Start() Position {
	if c.Colon != nil {
		return *c.Colon
	} else if c.Header != nil {
		return c.Header.Start()
	} else if c.Body != nil {
		return c.Body.Start()
	}
	return Position{}
}
func (c *ComponentCall) End() Position {
	if c.Body != nil {
		return c.Body.End()
	} else if c.Header != nil {
		return c.Header.End()
	} else if c.Colon != nil {
		return deltaPos(*c.Colon, len(":"))
	}
	return Position{}
}

func (*ComponentCall) _node()      {}
func (*ComponentCall) _scopeNode() {}

// ============================================================================
// Component Call Header
// ======================================================================================

type ComponentCallHeader struct {
	Name          FullIdent
	TypeArguments *TypeArguments // optional
	Arguments     *Arguments     // optional
}

var _ Node = (*ComponentCallHeader)(nil)

func (h *ComponentCallHeader) Start() Position {
	if h.Name != nil {
		return h.Name.Start()
	} else if h.TypeArguments != nil {
		return h.TypeArguments.Start()
	} else if h.Arguments != nil {
		return h.Arguments.Start()
	}
	return Position{}
}
func (h *ComponentCallHeader) End() Position {
	if h.Arguments != nil {
		return h.Arguments.End()
	} else if h.TypeArguments != nil {
		return h.TypeArguments.End()
	} else if h.Name != nil {
		return h.Name.End()
	}
	return Position{}
}

func (*ComponentCallHeader) _node() {}

// ============================================================================
// With
// ======================================================================================

type With struct {
	With *Position
	Name *Ident
	Body Body
}

var _ ScopeNode = (*With)(nil)

func (w *With) Start() Position {
	if w.With != nil {
		return *w.With
	} else if w.Name != nil {
		return w.Name.Start()
	} else if w.Body != nil {
		return w.Body.Start()
	}
	return Position{}
}
func (w *With) End() Position {
	if w.Body != nil {
		return w.Body.End()
	} else if w.Name != nil {
		return w.Name.End()
	} else if w.With != nil {
		return deltaPos(*w.With, len("with"))
	}
	return Position{}
}

func (*With) _node()      {}
func (*With) _scopeNode() {}
