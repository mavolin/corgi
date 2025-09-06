package ast

import "slices"

// ============================================================================
// Component
// ======================================================================================

type Component struct {
	Comp   *Position
	Header *ComponentHeader
	Body   ComponentBody
}

var (
	_ TopLevelNode = (*Component)(nil)
	_ Highlighter  = (*Component)(nil)
)

func (c *Component) Start() Position {
	if c.Comp != nil {
		return *c.Comp
	}
	if c.Header != nil {
		if start := c.Header.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (c *Component) End() Position {
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
	if c.Comp != nil {
		return deltaPos(*c.Comp, len("comp"))
	}
	return NoPosition
}

func (c *Component) Highlight() (start, end Position) {
	if c.Comp != nil {
		if c.Header != nil && c.Header.Name != nil {
			if end = c.Header.Name.End(); end != NoPosition {
				return *c.Comp, end
			}
		}
		return *c.Comp, deltaPos(*c.Comp, len("comp"))
	}
	return c.Start(), c.End()
}

func (c *Component) Walk(w func(Node)) {
	if c.Header != nil {
		w(c.Header)
	}
	if c.Body != nil {
		w(c.Body)
	}
}

func (*Component) _node()         {}
func (*Component) _topLevelNode() {}

// ============================================================================
// Component Header
// ======================================================================================

type ComponentHeader struct {
	Name           *Identifier
	TypeParameters *TypeParameters // optional
	Parameters     *ComponentParameters
}

var _ Node = (*ComponentHeader)(nil)

func (h *ComponentHeader) Start() Position {
	if h.Name != nil {
		if start := h.Name.Start(); start != NoPosition {
			return start
		}
	}
	if h.TypeParameters != nil {
		if start := h.TypeParameters.Start(); start != NoPosition {
			return start
		}
	}
	if h.Parameters != nil {
		if start := h.Parameters.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (h *ComponentHeader) End() Position {
	if h.Parameters != nil {
		if end := h.Parameters.End(); end != NoPosition {
			return end
		}
	}
	if h.TypeParameters != nil {
		if end := h.TypeParameters.End(); end != NoPosition {
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

func (h *ComponentHeader) Walk(w func(Node)) {
	if h.Name != nil {
		w(h.Name)
	}
	if h.TypeParameters != nil {
		w(h.TypeParameters)
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
	List   []*ComponentParameter
	RParen *Position
}

var _ Node = (*ComponentParameters)(nil)

func (p *ComponentParameters) Start() Position {
	if p.LParen != nil {
		return *p.LParen
	}
	for _, param := range p.List {
		if param != nil {
			if start := param.Start(); start != NoPosition {
				return start
			}
		}
	}
	if p.RParen != nil {
		return *p.RParen
	}
	return NoPosition
}

func (p *ComponentParameters) End() Position {
	if p.RParen != nil {
		return deltaPos(*p.RParen, len(")"))
	}
	for _, param := range slices.Backward(p.List) {
		if param != nil {
			if end := param.End(); end != NoPosition {
				return end
			}
		}
	}
	if p.LParen != nil {
		return deltaPos(*p.LParen, len("("))
	}
	return NoPosition
}

func (p *ComponentParameters) Walk(w func(Node)) {
	for _, param := range p.List {
		if param != nil {
			w(param)
		}
	}
}

func (*ComponentParameters) _node() {}

// ============================================================================
// Component Parameter
// ======================================================================================

// ComponentParameter is a parameter of a Component.
type ComponentParameter struct {
	Name    *Identifier
	Type    *Type       // nil if inferred from default, set if Default is nil
	Colon   *Position   // optional, set if Default
	Default *Expression // optional, set if Type is nil
}

var _ Node = (*ComponentParameter)(nil)

func (p *ComponentParameter) Start() Position {
	if p.Name != nil {
		if start := p.Name.Start(); start != NoPosition {
			return start
		}
	}
	if p.Type != nil {
		if start := p.Type.Start(); start != NoPosition {
			return start
		}
	}
	if p.Colon != nil {
		return *p.Colon
	}
	if p.Default != nil {
		if start := p.Default.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (p *ComponentParameter) End() Position {
	if p.Default != nil {
		if end := p.Default.End(); end != NoPosition {
			return end
		}
	}
	if p.Colon != nil {
		return deltaPos(*p.Colon, len(":"))
	}
	if p.Type != nil {
		if end := p.Type.End(); end != NoPosition {
			return end
		}
	}
	if p.Name != nil {
		if end := p.Name.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
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
// Component Body
// ======================================================================================

// ComponentBody is the body of a component, either [Body] or [Extend].
type ComponentBody interface {
	Node
	_componentBody()
}

// if this is changed, change the comment above
var (
	_ ComponentBody = (Body)(nil)
	_ ComponentBody = (*Extend)(nil)
)

// ============================================================================
// Extend
// ======================================================================================

type Extend struct {
	ComponentCall *ComponentCall
}

var (
	_ ComponentBody = (*Extend)(nil)
	_ Highlighter   = (*Extend)(nil)
)

func (e *Extend) Start() Position {
	if e.ComponentCall != nil {
		return e.ComponentCall.Start()
	}
	return NoPosition
}

func (e *Extend) End() Position {
	if e.ComponentCall != nil {
		return e.ComponentCall.End()
	}
	return NoPosition
}

func (e *Extend) Walk(w func(Node)) {
	if e.ComponentCall != nil {
		w(e.ComponentCall)
	}
}

func (e *Extend) Highlight() (start, end Position) {
	if e.ComponentCall != nil {
		return e.ComponentCall.Highlight()
	}
	return NoPosition, NoPosition
}

func (*Extend) _node()          {}
func (*Extend) _componentBody() {}

// ============================================================================
// Block
// ======================================================================================

type Block struct {
	Block      *Position
	Identifier *Identifier // optional for default block
	Default    Body        // may be nil
}

var (
	_ ScopeNode          = (*Block)(nil)
	_ Highlighter        = (*Block)(nil)
	_ ContentWriter      = (*Block)(nil)
	_ AttributeInhibitor = (*Block)(nil)
)

func (b *Block) Name() string {
	if b.Identifier != nil {
		return b.Identifier.Name
	}
	return ""
}

func (b *Block) Start() Position {
	if b.Block != nil {
		return *b.Block
	}
	if b.Identifier != nil {
		if start := b.Identifier.Start(); start != NoPosition {
			return start
		}
	}
	if b.Default != nil {
		if start := b.Default.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (b *Block) End() Position {
	if b.Default != nil {
		if end := b.Default.End(); end != NoPosition {
			return end
		}
	}
	if b.Identifier != nil {
		if end := b.Identifier.End(); end != NoPosition {
			return end
		}
	}
	if b.Block != nil {
		return deltaPos(*b.Block, len("block"))
	}
	return NoPosition
}

func (b *Block) Highlight() (start, end Position) {
	if b.Block != nil {
		if b.Identifier != nil {
			return *b.Block, b.Identifier.End()
		}
		return *b.Block, deltaPos(*b.Block, len("block"))
	}
	return b.Start(), b.End()
}

func (b *Block) Walk(w func(Node)) {
	if b.Identifier != nil {
		w(b.Identifier)
	}
	if b.Default != nil {
		w(b.Default)
	}
}

func (*Block) _node()               {}
func (*Block) _scopeNode()          {}
func (*Block) _contentWriter()      {}
func (*Block) _attributeInhibitor() {}
