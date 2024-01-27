package ast

type (
	// Expression is either a [ChainExpression] or [GoCode].
	Expression interface {
		_expression()
		AttributeValue
		Poser
	}

	// ForExpression is either a [RangeExpression], [GoCode], or a [ChainExpression].
	ForExpression interface {
		_forExpression()
		Poser
	}

	IfExpression struct {
		Statement *GoCode
		Condition Expression
	}
)

// if this is changed, change the comment above
var (
	_ Expression = (*ChainExpression)(nil)
	_ Expression = (*GoCode)(nil)

	_ ForExpression = (*RangeExpression)(nil)
	_ ForExpression = (*GoCode)(nil)
	_ ForExpression = (*ChainExpression)(nil)
)

func (e *IfExpression) Pos() Position {
	if e.Statement != nil {
		return e.Statement.Pos()
	}

	return e.Condition.Pos()
}

// ============================================================================
// RangeExpression
// ======================================================================================

// RangeExpression represents a range expression as used for for-loops.
//
// Hence, it is only present on for [For] items.
type RangeExpression struct {
	Var1, Var2 *Ident

	EqualSign *Position // Position of the '=' or ':='
	Declares  bool      // true if the range expression declares new variables (':=')
	Ordered   bool      // true if the range expression is ordered ('= ordered range')

	// RangeExpression is the expression that is being iterated over.
	RangeExpression Expression

	Position Position
}

var _ ForExpression = (*RangeExpression)(nil)

func (r *RangeExpression) Pos() Position { return r.Position }
func (*RangeExpression) _forExpression() {}

// ============================================================================
// Chain Expression
// ======================================================================================

// ChainExpression is an expression that checks if elements in the chain are
// not zero, and if indexes exist.
type ChainExpression struct {
	Root       *RawGoCode
	CheckRoot  bool                  // check root for zero value
	Chain      []ChainExpressionItem // chain behind root
	DerefCount int                   // number of leading '*' pointer derefs

	DefaultOperator *Position // only set if we have a default
	Default         *GoCode   // default value

	Position Position
}

var (
	_ Expression     = (*ChainExpression)(nil)
	_ ForExpression  = (*ChainExpression)(nil)
	_ AttributeValue = (*ChainExpression)(nil)
)

func (c *ChainExpression) Pos() Position  { return c.Position }
func (*ChainExpression) _expression()     {}
func (*ChainExpression) _forExpression()  {}
func (*ChainExpression) _attributeValue() {}

// ============================================================================
// Chain Expression Item
// ======================================================================================

// ChainExpressionItem represents an expression that can be chained.
//
// It is either a [IndexExpression], [ParenExpression],
// [TypeAssertionExpression], or a [DotIdentExpression].
type ChainExpressionItem interface {
	_chainExpressionItem()
	Poser
}

// if this is changed, change the comment above
var (
	_ ChainExpressionItem = (*IndexExpression)(nil)
	_ ChainExpressionItem = (*ParenExpression)(nil)
	_ ChainExpressionItem = (*TypeAssertionExpression)(nil)
	_ ChainExpressionItem = (*DotIdentExpression)(nil)
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

var _ ChainExpressionItem = (*IndexExpression)(nil)

func (e *IndexExpression) Pos() Position       { return e.LBracket }
func (*IndexExpression) _chainExpressionItem() {}

// ================================ Dot Ident Expression ================================

// DotIdentExpression represents a dot followed by a Go identifier.
type DotIdentExpression struct {
	Ident *Ident
	Check bool

	Position // of the dot
}

var _ ChainExpressionItem = (*DotIdentExpression)(nil)

func (e *DotIdentExpression) Pos() Position       { return e.Position }
func (*DotIdentExpression) _chainExpressionItem() {}

// ================================= Paren Expression ====================================

// ParenExpression represents a function call or the paren part of a type cast.
type ParenExpression struct {
	LParen Position
	Args   []*GoCode
	RParen *Position
	Check  bool
}

var _ ChainExpressionItem = (*ParenExpression)(nil)

func (e *ParenExpression) Pos() Position       { return e.LParen }
func (*ParenExpression) _chainExpressionItem() {}

// ================================ Type Assertion Expression ================================

// TypeAssertionExpression represents a type assertion.
type TypeAssertionExpression struct {
	LParen       *Position
	PointerCount int
	Package      *Ident // nil if no package
	Type         *Ident
	RParen       *Position
	Check        bool

	Position Position // of the dot
}

var _ ChainExpressionItem = (*TypeAssertionExpression)(nil)

func (e *TypeAssertionExpression) Pos() Position       { return e.Position }
func (*TypeAssertionExpression) _chainExpressionItem() {}
