package ast

// ============================================================================
// Doctype
// ======================================================================================

// Doctype is a doctype directive (`!doctype(html)`).
type Doctype struct {
	Doctype *Position
	LParen  *Position
	HTML    *Position
	RParen  *Position
}

var (
	_ ScopeNode          = (*Doctype)(nil)
	_ ContentWriter      = (*Doctype)(nil)
	_ ElementWriter      = (*Doctype)(nil)
	_ AttributeInhibitor = (*Doctype)(nil)
)

func (d *Doctype) Start() Position {
	if d.Doctype != nil {
		return *d.Doctype
	}
	if d.LParen != nil {
		return *d.LParen
	}
	if d.HTML != nil {
		return *d.HTML
	}
	if d.RParen != nil {
		return *d.RParen
	}
	return NoPosition
}

func (d *Doctype) End() Position {
	if d.RParen != nil {
		return deltaPos(*d.RParen, len(")"))
	}
	if d.HTML != nil {
		return deltaPos(*d.HTML, len("html"))
	}
	if d.LParen != nil {
		return deltaPos(*d.LParen, len("("))
	}
	if d.Doctype != nil {
		return deltaPos(*d.Doctype, len("!doctype"))
	}
	return NoPosition
}

func (d *Doctype) Walk(func(Node)) {}

func (*Doctype) _node()               {}
func (*Doctype) _elementWriter()      {}
func (*Doctype) _contentWriter()      {}
func (*Doctype) _attributeInhibitor() {}
func (*Doctype) _scopeNode()          {}

// ============================================================================
// Element
// ======================================================================================

// Element is a HTML element.
type Element struct {
	Header *ElementHeader
	Body   Body
}

var (
	_ ScopeNode          = (*Element)(nil)
	_ ContentWriter      = (*Element)(nil)
	_ ElementWriter      = (*Element)(nil)
	_ ContainingElement  = (*Element)(nil)
	_ AttributeInhibitor = (*Element)(nil)
	_ Highlighter        = (*Element)(nil)
)

func (e *Element) Start() Position {
	if e.Header != nil {
		if start := e.Header.Start(); start != NoPosition {
			return start
		}
	}
	if e.Body != nil {
		if start := e.Body.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (e *Element) End() Position {
	if e.Body != nil {
		if end := e.Body.End(); end != NoPosition {
			return end
		}
	}
	if e.Header != nil {
		if end := e.Header.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (e *Element) Walk(w func(Node)) {
	if e.Header != nil {
		w(e.Header)
	}
	if e.Body != nil {
		w(e.Body)
	}
}

func (e *Element) Highlight() (start, end Position) {
	if e.Header != nil && e.Header.Name != nil {
		start, end = e.Header.Name.Start(), e.Header.Name.End()
		if start != NoPosition && end != NoPosition {
			return start, end
		}
	}
	return e.Start(), e.End()
}

func (*Element) _node()               {}
func (*Element) _scopeNode()          {}
func (*Element) _contentWriter()      {}
func (*Element) _elementWriter()      {}
func (*Element) _containingElement()  {}
func (*Element) _attributeInhibitor() {}

// ============================================================================
// Element Header
// ======================================================================================

type ElementHeader struct {
	Name       *ElementReference
	Attributes *Arguments
}

var _ Node = (*ElementHeader)(nil)

func (h *ElementHeader) Start() Position {
	if h.Name != nil {
		if start := h.Name.Start(); start != NoPosition {
			return start
		}
	}
	if h.Attributes != nil {
		if start := h.Attributes.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (h *ElementHeader) End() Position {
	if h.Attributes != nil {
		if end := h.Attributes.End(); end != NoPosition {
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

func (h *ElementHeader) Walk(w func(Node)) {
	if h.Name != nil {
		w(h.Name)
	}
	if h.Attributes != nil {
		w(h.Attributes)
	}
}

func (*ElementHeader) _node() {}

// ============================================================================
// Element Name
// ======================================================================================

type ElementName struct {
	Name     string
	Position *Position
}

var _ Node = (*ElementName)(nil)

func (n *ElementName) Start() Position {
	if n.Position != nil {
		return *n.Position
	}
	return NoPosition
}

func (n *ElementName) End() Position {
	if n.Position != nil {
		return deltaPos(*n.Position, len(n.Name))
	}
	return NoPosition
}
func (n *ElementName) Walk(func(Node)) {}

func (*ElementName) _node() {}

// ============================================================================
// Element Reference
// ======================================================================================

type ElementReference struct {
	Package *Identifier
	Dot     *Position
	Name    *ElementName
}

var _ Node = (*ElementReference)(nil)

func (r *ElementReference) Start() Position {
	if r.Package != nil {
		if start := r.Package.Start(); start != NoPosition {
			return start
		}
	}
	if r.Name != nil {
		if start := r.Name.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (r *ElementReference) End() Position {
	if r.Name != nil {
		if end := r.Name.End(); end != NoPosition {
			return end
		}
	}
	if r.Package != nil {
		if end := r.Package.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (r *ElementReference) Walk(w func(Node)) {
	if r.Package != nil {
		w(r.Package)
	}
	if r.Name != nil {
		w(r.Name)
	}
}

func (*ElementReference) _node() {}

// ============================================================================
// Raw Element
// ======================================================================================

// RawElement is the special !raw element, which includes all of its
// contents verbatim in the generated HTML.
type RawElement struct {
	Raw  *Position
	Body *BracketText // should only contain text nodes, not nil
}

var (
	_ ScopeNode          = (*RawElement)(nil)
	_ Highlighter        = (*RawElement)(nil)
	_ ContentWriter      = (*RawElement)(nil)
	_ AttributeInhibitor = (*RawElement)(nil)
)

func (e *RawElement) Start() Position {
	if e.Raw != nil {
		return *e.Raw
	}
	if e.Body != nil {
		if start := e.Body.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (e *RawElement) End() Position {
	if e.Body != nil {
		if end := e.Body.End(); end != NoPosition {
			return end
		}
	}
	if e.Raw != nil {
		return deltaPos(*e.Raw, len("!raw"))
	}
	return NoPosition
}

func (e *RawElement) Walk(w func(Node)) {
	if e.Body != nil {
		w(e.Body)
	}
}

func (e *RawElement) Highlight() (start, end Position) {
	if e.Raw != nil {
		return *e.Raw, deltaPos(*e.Raw, len("!raw"))
	}
	return e.Start(), e.End()
}

func (*RawElement) _node()               {}
func (*RawElement) _scopeNode()          {}
func (*RawElement) _contentWriter()      {}
func (*RawElement) _attributeInhibitor() {}

// ============================================================================
// And
// ======================================================================================

type And struct {
	And        *Position
	Attributes *Arguments
}

var _ ScopeNode = (*And)(nil)

func (a *And) Start() Position {
	if a.And != nil {
		return *a.And
	}
	if a.Attributes != nil {
		if start := a.Attributes.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (a *And) End() Position {
	if a.Attributes != nil {
		if end := a.Attributes.End(); end != NoPosition {
			return end
		}
	}
	if a.And != nil {
		return deltaPos(*a.And, len("&"))
	}
	return NoPosition
}

func (a *And) Walk(w func(Node)) {
	if a.Attributes != nil {
		w(a.Attributes)
	}
}

func (*And) _node()      {}
func (*And) _scopeNode() {}
