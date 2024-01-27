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

var _ ScopeItem = (*Component)(nil)

func (c *Component) Pos() Position { return c.Position }
func (*Component) _scopeItem()     {}

// ==================================== Type Param =====================================

type TypeParam struct {
	Names []Ident
	Type  Type
}

func (p TypeParam) Pos() Position {
	if len(p.Names) > 0 {
		return p.Names[0].Pos()
	}
	return InvalidPosition
}

// ==================================== Component Param =====================================

// ComponentParam represents a parameter of a Component.
type ComponentParam struct {
	Name    *Ident
	Type    *Type // nil if inferred from default, set if Default is nil
	Colon   *Position
	Default *GoCode // optional, set if Type is nil

	Position Position
}

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

var _ ScopeItem = (*ComponentCall)(nil)

func (c *ComponentCall) Pos() Position { return c.Position }
func (*ComponentCall) _scopeItem()     {}

// =================================== Component Call Arg ===================================

type ComponentArg struct {
	Name  *Ident
	Colon *Position
	Value *Expression

	Position Position
}

// ============================================================================
// Block
// ======================================================================================

type Block struct {
	Name *Ident
	Body Body // may be nil

	Position Position
}

var _ ScopeItem = (*Block)(nil)

func (b *Block) Pos() Position { return b.Position }
func (*Block) _scopeItem()     {}

// ============================================================================
// Underscore Block Shorthand
// ======================================================================================

type UnderscoreBlockShorthand struct {
	Body     Body
	Position Position
}

var _ Body = (*UnderscoreBlockShorthand)(nil)

func (u *UnderscoreBlockShorthand) Pos() Position { return u.Position }
func (*UnderscoreBlockShorthand) _body()          {}
