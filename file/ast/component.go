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
	if c.Comp != nil {
		return *c.Comp
	} else if c.Header != nil {
		return c.Header.Start()
	} else if c.Colon != nil {
		return *c.Colon
	} else if c.Extend != nil {
		return c.Extend.Start()
	}
	return Position{}
}
func (c *Component) End() Position {
	if c.Body != nil {
		return c.Body.End()
	} else if c.Extend != nil {
		return c.Extend.End()
	} else if c.Colon != nil {
		return deltaPos(*c.Colon, len(":"))
	} else if c.Header != nil {
		return c.Header.End()
	} else if c.Comp != nil {
		deltaPos(*c.Comp, len("comp"))
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
	Params     *ComponentParameters
}

var _ Node = (*ComponentHeader)(nil)

func (h *ComponentHeader) Start() Position {
	if h.Name != nil {
		return h.Name.Start()
	} else if h.TypeParams != nil {
		return h.TypeParams.Start()
	} else if h.Params != nil {
		return h.Params.Start()
	}
	return Position{}
}
func (h *ComponentHeader) End() Position {
	if h.Params != nil {
		return h.Params.End()
	} else if h.TypeParams != nil {
		return h.TypeParams.End()
	} else if h.Name != nil {
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
	if h.Params != nil {
		w(h.Params)
	}
}

func (*ComponentHeader) _node() {}

// ============================================================================
// Component Params
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
	if p.Name != nil {
		return p.Name.Start()
	} else if p.Type != nil {
		return p.Type.Start()
	} else if p.Colon != nil {
		return *p.Colon
	} else if p.Default != nil {
		return p.Default.Start()
	}
	return Position{}
}
func (p *ComponentParameter) End() Position {
	if p.Default != nil {
		return p.Default.End()
	} else if p.Colon != nil {
		return deltaPos(*p.Colon, len(":"))
	} else if p.Type != nil {
		return p.Type.End()
	} else if p.Name != nil {
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
	if a.Alias != nil {
		return *a.Alias
	} else if a.Header != nil {
		return a.Header.Start()
	} else if a.ComponentCall != nil {
		return a.ComponentCall.Start()
	}
	return Position{}
}
func (a *Alias) End() Position {
	if a.ComponentCall != nil {
		return a.ComponentCall.End()
	} else if a.Header != nil {
		return a.Header.End()
	} else if a.Alias != nil {
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
// General
// ======================================================================================

type Block struct {
	Block   *Position
	Name    *Ident
	Default Body // may be nil
}

var _ ScopeNode = (*Block)(nil)

func (b *Block) Start() Position {
	if b.Block != nil {
		return *b.Block
	} else if b.Name != nil {
		return b.Name.Start()
	} else if b.Default != nil {
		return b.Default.Start()
	}
	return Position{}
}
func (b *Block) End() Position {
	if b.Default != nil {
		return b.Default.End()
	} else if b.Name != nil {
		return b.Name.End()
	} else if b.Block != nil {
		return deltaPos(*b.Block, len("block"))
	}
	return Position{}
}
func (b *Block) Walk(w func(Node)) {
	if b.Name != nil {
		w(b.Name)
	}
	if b.Default != nil {
		w(b.Default)
	}
}

func (*Block) _node()      {}
func (*Block) _scopeNode() {}
