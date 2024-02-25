package ast

// ============================================================================
// Component
// ======================================================================================

type Component struct {
	Name *Ident

	LBracket   *Position // nil if no type params
	TypeParams []*TypeParam
	RBracket   *Position // nil if no type params

	LParen *Position
	Params []*ComponentParam
	RParen *Position

	Body Body

	Position Position
}

var _ ScopeNode = (*Component)(nil)

func (c *Component) Pos() Position { return c.Position }
func (c *Component) End() Position {
	if c.Body != nil {
		return c.Body.End()
	} else if c.RParen != nil {
		return deltaPos(*c.RParen, 1)
	} else if len(c.Params) > 0 {
		return c.Params[len(c.Params)-1].End()
	} else if c.LParen != nil {
		return deltaPos(*c.LParen, 1)
	} else if c.RBracket != nil {
		return deltaPos(*c.RBracket, 1)
	} else if len(c.TypeParams) > 0 {
		return c.TypeParams[len(c.TypeParams)-1].End()
	} else if c.LBracket != nil {
		return deltaPos(*c.LBracket, 1)
	} else if c.Name != nil {
		return c.Name.End()
	}
	return deltaPos(c.Position, len("comp"))
}

func (*Component) _node()      {}
func (*Component) _scopeNode() {}

// ==================================== Type Param =====================================

type TypeParam struct {
	Names []*Ident
	Type  *Type
}

var _ Node = (*TypeParam)(nil)

func (p *TypeParam) Pos() Position {
	if len(p.Names) > 0 {
		return p.Names[0].Pos()
	}
	return InvalidPosition
}
func (p *TypeParam) End() Position {
	if p.Type != nil {
		return p.Type.End()
	}
	if len(p.Names) > 0 {
		return p.Names[len(p.Names)-1].End()
	}
	return InvalidPosition
}

func (*TypeParam) _node() {}

// ==================================== Component Param =====================================

// ComponentParam represents a parameter of a Component.
type ComponentParam struct {
	Name    *Ident
	Type    *Type // nil if inferred from default, set if Default is nil
	Colon   *Position
	Default *GoCode // optional, set if Type is nil

	Position Position
}

var _ Node = (*ComponentParam)(nil)

func (p *ComponentParam) Pos() Position { return p.Position }
func (p *ComponentParam) End() Position {
	if p.Default != nil {
		return p.Default.End()
	} else if p.Colon != nil {
		return deltaPos(*p.Colon, 1)
	} else if p.Type != nil {
		return p.Type.End()
	} else if p.Name != nil {
		return p.Name.End()
	}
	return InvalidPosition
}

func (*ComponentParam) _node() {}

// ============================================================================
// Component Call
// ======================================================================================

type ComponentCall struct {
	Namespace *Ident // may be nil
	Name      *Ident

	LBracket *Position // nil if no type params
	TypeArgs []*Type
	RBracket *Position // nil if no type params

	LParen *Position
	Args   []*ComponentArg
	RParen *Position

	Body Body

	Position Position
}

var _ ScopeNode = (*ComponentCall)(nil)

func (c *ComponentCall) Pos() Position { return c.Position }
func (c *ComponentCall) End() Position {
	if c.Body != nil {
		return c.Body.End()
	} else if c.RParen != nil {
		return deltaPos(*c.RParen, 1)
	} else if len(c.Args) > 0 {
		return c.Args[len(c.Args)-1].End()
	} else if c.LParen != nil {
		return deltaPos(*c.LParen, 1)
	} else if c.RBracket != nil {
		return deltaPos(*c.RBracket, 1)
	} else if len(c.TypeArgs) > 0 {
		return c.TypeArgs[len(c.TypeArgs)-1].End()
	} else if c.LBracket != nil {
		return deltaPos(*c.LBracket, 1)
	} else if c.Name != nil {
		return c.Name.End()
	}
	return deltaPos(c.Position, len("+"))
}

func (*ComponentCall) _node()      {}
func (*ComponentCall) _scopeNode() {}

// =================================== Component Call Arg ===================================

type ComponentArg struct {
	Name  *Ident
	Colon *Position
	Value Expression

	Position Position
}

var _ Node = (*ComponentArg)(nil)

func (a *ComponentArg) Pos() Position { return a.Position }
func (a *ComponentArg) End() Position {
	if a.Value != nil {
		return a.Value.End()
	} else if a.Colon != nil {
		return deltaPos(*a.Colon, 1)
	} else if a.Name != nil {
		return a.Name.End()
	}
	return InvalidPosition
}

func (*ComponentArg) _node() {}

// ============================================================================
// Block
// ======================================================================================

type Block struct {
	Name *Ident
	Body Body // may be nil

	Position Position
}

var _ ScopeNode = (*Block)(nil)

func (b *Block) Pos() Position { return b.Position }
func (b *Block) End() Position {
	if b.Body != nil {
		return b.Body.End()
	} else if b.Name != nil {
		return b.Name.End()
	}
	return deltaPos(b.Position, len("block"))
}

func (*Block) _node()      {}
func (*Block) _scopeNode() {}

// ============================================================================
// Underscore Block Shorthand
// ======================================================================================

type UnderscoreBlockShorthand struct {
	Body     Body
	Position Position
}

var _ Body = (*UnderscoreBlockShorthand)(nil)

func (s *UnderscoreBlockShorthand) Pos() Position { return s.Position }
func (s *UnderscoreBlockShorthand) End() Position {
	if s.Body != nil {
		return s.Body.End()
	}
	return deltaPos(s.Position, len("_"))
}

func (*UnderscoreBlockShorthand) _node() {}
func (*UnderscoreBlockShorthand) _body() {}
