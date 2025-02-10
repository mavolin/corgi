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

var _ ScopeNode = (*Doctype)(nil)

func (d *Doctype) Start() Position {
	if d.Doctype != nil {
		return *d.Doctype
	} else if d.HTML != nil {
		return *d.HTML
	} else if d.LParen != nil {
		return *d.LParen
	}
	return Position{}
}
func (d *Doctype) End() Position {
	if d.RParen != nil {
		return *d.RParen
	} else if d.HTML != nil {
		return *d.HTML
	} else if d.LParen != nil {
		return *d.LParen
	} else if d.Doctype != nil {
		return deltaPos(*d.Doctype, len("!doctype"))
	}
	return Position{}
}
func (d *Doctype) Walk(func(Node)) {}

func (*Doctype) _node()      {}
func (*Doctype) _scopeNode() {}

// ============================================================================
// Element
// ======================================================================================

// Element is a HTML element.
type Element struct {
	Header *ElementHeader
	Body   Body
}

var _ ScopeNode = (*Element)(nil)

func (e *Element) Start() Position {
	if e.Header != nil {
		return e.Header.Start()
	} else if e.Body != nil {
		return e.Body.Start()
	}
	return Position{}
}
func (e *Element) End() Position {
	if e.Body != nil {
		return e.Body.End()
	} else if e.Header != nil {
		return e.Header.End()
	}
	return Position{}
}
func (e *Element) Walk(w func(Node)) {
	if e.Header != nil {
		w(e.Header)
	}
	if e.Body != nil {
		w(e.Body)
	}
}

func (*Element) _node()      {}
func (*Element) _scopeNode() {}

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
		return h.Name.Start()
	} else if h.Attributes != nil {
		return h.Attributes.Start()
	}
	return Position{}
}
func (h *ElementHeader) End() Position {
	if h.Attributes != nil {
		return h.Attributes.End()
	} else if h.Name != nil {
		return h.Name.End()
	}
	return Position{}
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
	return Position{}
}
func (n *ElementName) End() Position {
	if n.Position != nil {
		return deltaPos(*n.Position, len(n.Name))
	}
	return Position{}
}
func (n *ElementName) Walk(func(Node)) {}

func (*ElementName) _node() {}

// ============================================================================
// Element Reference
// ======================================================================================

type ElementReference struct {
	Package *Ident
	Dot     *Position
	Name    *ElementName
}

var _ Node = (*ElementReference)(nil)

func (r *ElementReference) Start() Position {
	if r.Package != nil {
		return r.Package.Start()
	} else if r.Name != nil {
		return r.Name.Start()
	}
	return Position{}
}
func (r *ElementReference) End() Position {
	if r.Name != nil {
		return r.Name.End()
	} else if r.Package != nil {
		return r.Package.End()
	}
	return Position{}
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
// contents verbatim into the generated HTML.
type RawElement struct {
	Raw  *Position
	Body *BracketText // not nil
}

var _ ScopeNode = (*RawElement)(nil)

func (e *RawElement) Start() Position {
	if e.Raw != nil {
		return *e.Raw
	}
	return e.Body.Start()
}
func (e *RawElement) End() Position {
	if e.Body != nil {
		return e.Body.End()
	} else if e.Raw != nil {
		return deltaPos(*e.Raw, len("!raw"))
	}
	return Position{}
}
func (e *RawElement) Walk(w func(Node)) {
	if e.Body != nil {
		w(e.Body)
	}
}

func (*RawElement) _node()      {}
func (*RawElement) _scopeNode() {}

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
	} else if a.Attributes != nil {
		return a.Attributes.Start()
	}
	return Position{}
}
func (a *And) End() Position {
	if a.Attributes != nil {
		return a.Attributes.End()
	} else if a.And != nil {
		return deltaPos(*a.And, len("&"))
	}
	return Position{}
}
func (a *And) Walk(w func(Node)) {
	if a.Attributes != nil {
		w(a.Attributes)
	}
}

func (*And) _node()      {}
func (*And) _scopeNode() {}
