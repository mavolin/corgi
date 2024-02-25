package ast

// Expression is either a [ChainExpression] or [GoCode].
type Expression interface {
	AttributeValue
	_expression()
}

// if this is changed, change the comment above
var (
	_ Expression = (*ChainExpression)(nil)
	_ Expression = (*GoCode)(nil)
)

// ============================================================================
// Chain Expression
// ======================================================================================

// ChainExpression is an expression that checks if elements in the chain are
// not zero, and if indexes exist.
type ChainExpression struct {
	Root       *RawGoCode
	CheckRoot  bool                  // check root for zero value
	Chain      []ChainExpressionNode // chain behind root
	DerefCount int                   // number of leading '*' pointer derefs

	DefaultOperator *Position // only set if we have a default
	Default         *GoCode   // default value

	Position Position
}

var (
	_ Expression     = (*ChainExpression)(nil)
	_ ForHeader      = (*ChainExpression)(nil)
	_ AttributeValue = (*ChainExpression)(nil)
)

func (c *ChainExpression) Pos() Position { return c.Position }
func (c *ChainExpression) End() Position {
	if c.Default != nil {
		return c.Default.End()
	} else if c.DefaultOperator != nil {
		return deltaPos(*c.DefaultOperator, 1)
	} else if len(c.Chain) > 0 {
		return c.Chain[len(c.Chain)-1].End()
	} else if c.Root != nil {
		if c.CheckRoot {
			return deltaPos(c.Root.End(), len("?"))
		}
		return c.Root.End()
	}
	return deltaPos(c.Position, c.DerefCount)
}

func (*ChainExpression) _node()           {}
func (*ChainExpression) _expression()     {}
func (*ChainExpression) _forHeader()      {}
func (*ChainExpression) _attributeValue() {}

// ============================================================================
// Chain Expression Node
// ======================================================================================

// ChainExpressionNode represents a node in a chain expression.
//
// It is either a [IndexExpression], [ParenExpression],
// [TypeAssertionExpression], or a [DotIdentExpression].
type ChainExpressionNode interface {
	Node
	_chainExpressionNode()
}

// if this is changed, change the comment above
var (
	_ ChainExpressionNode = (*IndexExpression)(nil)
	_ ChainExpressionNode = (*ParenExpression)(nil)
	_ ChainExpressionNode = (*TypeAssertionExpression)(nil)
	_ ChainExpressionNode = (*DotIdentExpression)(nil)
)

// ================================== IndexExpression ===================================

// IndexExpression represents either a map or slice index expression.
type IndexExpression struct {
	LBracket Position
	Index    *GoCode
	RBracket *Position

	CheckIndex bool
	CheckValue bool
}

var _ ChainExpressionNode = (*IndexExpression)(nil)

func (e *IndexExpression) Pos() Position { return e.LBracket }
func (e *IndexExpression) End() Position {
	if e.RBracket != nil {
		if e.CheckValue {
			return deltaPos(*e.RBracket, len("?"))
		}
		return *e.RBracket
	} else if e.Index != nil {
		if e.CheckIndex {
			return deltaPos(e.Index.End(), len("?"))
		}
		return e.Index.End()
	}
	return deltaPos(e.LBracket, 1)
}

func (*IndexExpression) _node()                {}
func (*IndexExpression) _chainExpressionNode() {}

// ================================ Dot Ident Expression ================================

// DotIdentExpression represents a dot followed by a Go identifier.
type DotIdentExpression struct {
	Ident *Ident
	Check bool

	Position // of the dot
}

var _ ChainExpressionNode = (*DotIdentExpression)(nil)

func (e *DotIdentExpression) Pos() Position { return e.Position }
func (e *DotIdentExpression) End() Position {
	if e.Ident != nil {
		if e.Check {
			return deltaPos(e.Ident.End(), len("?"))
		}
		return e.Ident.End()
	}
	return deltaPos(e.Position, len("."))
}
func (*DotIdentExpression) _node()                {}
func (*DotIdentExpression) _chainExpressionNode() {}

// ================================= Paren Expression ====================================

// ParenExpression represents a function call or the paren part of a type cast.
type ParenExpression struct {
	LParen Position
	Args   []*GoCode
	RParen *Position
	Check  bool
}

var _ ChainExpressionNode = (*ParenExpression)(nil)

func (e *ParenExpression) Pos() Position { return e.LParen }
func (e *ParenExpression) End() Position {
	if e.RParen != nil {
		if e.Check {
			return deltaPos(*e.RParen, len("?"))
		}
		return *e.RParen
	} else if len(e.Args) > 0 {
		return e.Args[len(e.Args)-1].End()
	}
	return deltaPos(e.LParen, 1)
}

func (*ParenExpression) _node()                {}
func (*ParenExpression) _chainExpressionNode() {}

// ================================ Type Assertion Expression ================================

// TypeAssertionExpression represents a type assertion.
type TypeAssertionExpression struct {
	LParen       *Position
	PointerCount int
	Package      *Ident // nil if no package
	Type         *Ident
	CheckType    bool
	RParen       *Position
	CheckValue   bool

	Position Position // of the dot
}

var _ ChainExpressionNode = (*TypeAssertionExpression)(nil)

func (e *TypeAssertionExpression) Pos() Position { return e.Position }
func (e *TypeAssertionExpression) End() Position {
	if e.RParen != nil {
		if e.CheckValue {
			return deltaPos(*e.RParen, len("?"))
		}
		return *e.RParen
	} else if e.Type != nil {
		if e.CheckType {
			return deltaPos(e.Type.End(), len("?"))
		}
		return e.Type.End()
	} else if e.LParen != nil {
		return deltaPos(*e.LParen, 1)
	}
	return deltaPos(e.Position, len("."))
}

func (*TypeAssertionExpression) _node()                {}
func (*TypeAssertionExpression) _chainExpressionNode() {}
