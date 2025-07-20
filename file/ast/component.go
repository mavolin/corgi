package ast

import "slices"

// ============================================================================
// Component
// ======================================================================================

type Component struct {
	Comp   *Position
	Header *ComponentHeader
	Colon  *Position            // nil if no extend
	Extend *ComponentCallHeader // optional
	Body   Body
}

var _ ScopeNode = (*Component)(nil)

func (c *Component) Start() Position {
	switch {
	case c.Comp != nil:
		return *c.Comp
	case c.Header != nil:
		return c.Header.Start()
	case c.Colon != nil:
		return *c.Colon
	case c.Extend != nil:
		return c.Extend.Start()
	}
	return Position{}
}

func (c *Component) End() Position {
	switch {
	case c.Body != nil:
		return c.Body.End()
	case c.Extend != nil:
		return c.Extend.End()
	case c.Colon != nil:
		return deltaPos(*c.Colon, len(":"))
	case c.Header != nil:
		return c.Header.End()
	case c.Comp != nil:
		return deltaPos(*c.Comp, len("comp"))
	}
	return Position{}
}

func (c *Component) Walk(w func(Node)) {
	if c.Header != nil {
		w(c.Header)
	}
	if c.Extend != nil {
		w(c.Extend)
	}
	if c.Body != nil {
		w(c.Body)
	}
}

func (*Component) _node()      {}
func (*Component) _scopeNode() {}

// ============================================================================
// Component Header
// ======================================================================================

type ComponentHeader struct {
	Name       *Ident
	TypeParams *TypeParameters // optional
	Parameters *ComponentParameters
}

var _ Node = (*ComponentHeader)(nil)

func (h *ComponentHeader) Start() Position {
	switch {
	case h.Name != nil:
		return h.Name.Start()
	case h.TypeParams != nil:
		return h.TypeParams.Start()
	case h.Parameters != nil:
		return h.Parameters.Start()
	}
	return Position{}
}

func (h *ComponentHeader) End() Position {
	switch {
	case h.Parameters != nil:
		return h.Parameters.End()
	case h.TypeParams != nil:
		return h.TypeParams.End()
	case h.Name != nil:
		return h.Name.End()
	}
	return Position{}
}

func (h *ComponentHeader) Walk(w func(Node)) {
	if h.Name != nil {
		w(h.Name)
	}
	if h.TypeParams != nil {
		w(h.TypeParams)
	}
	if h.Parameters != nil {
		w(h.Parameters)
	}
}

func (*ComponentHeader) _node() {}

// ============================================================================
// Component Parameters
// ======================================================================================

type ComponentParameters struct {
	LParen *Position
	Params []*ComponentParameter
	RParen *Position
}

var _ Node = (*ComponentParameters)(nil)

func (p *ComponentParameters) Start() Position {
	if p.LParen != nil {
		return *p.LParen
	}
	for _, param := range p.Params {
		if param != nil {
			return param.Start()
		}
	}
	if p.RParen != nil {
		return *p.RParen
	}
	return Position{}
}

func (p *ComponentParameters) End() Position {
	if p.RParen != nil {
		return deltaPos(*p.RParen, len(")"))
	}
	for _, param := range slices.Backward(p.Params) {
		if param != nil {
			return param.End()
		}
	}
	if p.LParen != nil {
		return deltaPos(*p.LParen, len("("))
	}
	return Position{}
}

func (p *ComponentParameters) Walk(w func(Node)) {
	for _, param := range p.Params {
		if param != nil {
			w(param)
		}
	}
}

func (*ComponentParameters) _node() {}

// ============================================================================
// Component Param
// ======================================================================================

// ComponentParameter is a parameter of a Component.
type ComponentParameter struct {
	Name    *Ident
	Type    *Type       // nil if inferred from default, set if Default is nil
	Colon   *Position   // optional, set if Default
	Default *Expression // optional, set if Type is nil
}

var _ Node = (*ComponentParameter)(nil)

func (p *ComponentParameter) Start() Position {
	switch {
	case p.Name != nil:
		return p.Name.Start()
	case p.Type != nil:
		return p.Type.Start()
	case p.Colon != nil:
		return *p.Colon
	case p.Default != nil:
		return p.Default.Start()
	}
	return Position{}
}

func (p *ComponentParameter) End() Position {
	switch {
	case p.Default != nil:
		return p.Default.End()
	case p.Colon != nil:
		return deltaPos(*p.Colon, len(":"))
	case p.Type != nil:
		return p.Type.End()
	case p.Name != nil:
		return p.Name.End()
	}
	return Position{}
}

func (p *ComponentParameter) Walk(w func(Node)) {
	if p.Name != nil {
		w(p.Name)
	}
	if p.Type != nil {
		w(p.Type)
	}
	if p.Default != nil {
		w(p.Default)
	}
}

func (*ComponentParameter) _node() {}

// ============================================================================
// Header
// ======================================================================================

type Alias struct {
	Alias         *Position
	Header        *ComponentHeader
	ComponentCall *ComponentCall
}

var _ ScopeNode = (*Alias)(nil)

func (a *Alias) Start() Position {
	switch {
	case a.Alias != nil:
		return *a.Alias
	case a.Header != nil:
		return a.Header.Start()
	case a.ComponentCall != nil:
		return a.ComponentCall.Start()
	}
	return Position{}
}

func (a *Alias) End() Position {
	switch {
	case a.ComponentCall != nil:
		return a.ComponentCall.End()
	case a.Header != nil:
		return a.Header.End()
	case a.Alias != nil:
		return deltaPos(*a.Alias, len("alias"))
	}
	return Position{}
}

func (a *Alias) Walk(w func(Node)) {
	if a.Header != nil {
		w(a.Header)
	}
	if a.ComponentCall != nil {
		w(a.ComponentCall)
	}
}

func (*Alias) _node()      {}
func (*Alias) _scopeNode() {}

// ============================================================================
// Block
// ======================================================================================

type Block struct {
	Block      *Position
	Identifier *Ident // optional for default block
	Default    Body   // may be nil
}

var _ ScopeNode = (*Block)(nil)

func (b *Block) Name() string {
	if b.Identifier != nil {
		return b.Identifier.Ident
	}
	return ""
}

func (b *Block) Start() Position {
	switch {
	case b.Block != nil:
		return *b.Block
	case b.Identifier != nil:
		return b.Identifier.Start()
	case b.Default != nil:
		return b.Default.Start()
	}
	return Position{}
}

func (b *Block) End() Position {
	switch {
	case b.Default != nil:
		return b.Default.End()
	case b.Identifier != nil:
		return b.Identifier.End()
	case b.Block != nil:
		return deltaPos(*b.Block, len("block"))
	}
	return Position{}
}

func (b *Block) Walk(w func(Node)) {
	if b.Identifier != nil {
		w(b.Identifier)
	}
	if b.Default != nil {
		w(b.Default)
	}
}

func (*Block) _node()      {}
func (*Block) _scopeNode() {}
