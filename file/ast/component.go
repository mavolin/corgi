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
	switch {
	case c.Comp != nil:
		return *c.Comp
	case c.Header != nil:
		return c.Header.Start()
	}
	return Position{}
}

func (c *Component) End() Position {
	switch {
	case c.Body != nil:
		return c.Body.End()
	case c.Header != nil:
		return c.Header.End()
	case c.Comp != nil:
		return deltaPos(*c.Comp, len("comp"))
	}
	return Position{}
}

func (c *Component) Highlight() (start, end Position) {
	if c.Comp != nil {
		if c.Header != nil && c.Header.Name != nil {
			return *c.Comp, c.Header.Name.End()
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
	switch {
	case h.Name != nil:
		return h.Name.Start()
	case h.TypeParameters != nil:
		return h.TypeParameters.Start()
	case h.Parameters != nil:
		return h.Parameters.Start()
	}
	return Position{}
}

func (h *ComponentHeader) End() Position {
	switch {
	case h.Parameters != nil:
		return h.Parameters.End()
	case h.TypeParameters != nil:
		return h.TypeParameters.End()
	case h.Name != nil:
		return h.Name.End()
	}
	return Position{}
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
	for _, param := range slices.Backward(p.List) {
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
	return Position{}
}

func (e *Extend) End() Position {
	if e.ComponentCall != nil {
		return e.ComponentCall.End()
	}
	return Position{}
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
	return Position{}, Position{}
}

func (*Extend) _node()          {}
func (*Extend) _componentBody() {}

// ============================================================================
// Component Alias
// ======================================================================================

type ComponentAlias struct {
	EqualSign     *Position
	ComponentCall *ComponentCall
}

var _ ComponentBody = (*ComponentAlias)(nil)

func (a *ComponentAlias) Start() Position {
	switch {
	case a.EqualSign != nil:
		return *a.EqualSign
	case a.ComponentCall != nil:
		return a.ComponentCall.Start()
	}
	return Position{}
}

func (a *ComponentAlias) End() Position {
	switch {
	case a.ComponentCall != nil:
		return a.ComponentCall.End()
	case a.EqualSign != nil:
		return deltaPos(*a.EqualSign, len("="))
	}
	return Position{}
}

func (a *ComponentAlias) Walk(w func(Node)) {
	if a.ComponentCall != nil {
		w(a.ComponentCall)
	}
}

func (*ComponentAlias) _node()          {}
func (*ComponentAlias) _componentBody() {}

// ============================================================================
// Block
// ======================================================================================

type Block struct {
	Block      *Position
	Identifier *Identifier // optional for default block
	Default    Body        // may be nil
}

var (
	_ ScopeNode     = (*Block)(nil)
	_ Highlighter   = (*Block)(nil)
	_ ContentWriter = (*Block)(nil)
)

func (b *Block) Name() string {
	if b.Identifier != nil {
		return b.Identifier.Name
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

func (*Block) _node()          {}
func (*Block) _scopeNode()     {}
func (*Block) _contentWriter() {}
