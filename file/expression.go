package file

type (
	Expression interface {
		_expression()
		AttributeValue
		Poser
	}

	ForExpression interface {
		_forExpression()
		Poser
	}

	IfExpression struct {
		Statement *GoCode
		Condition Expression
	}
)

func (e IfExpression) Pos() Position {
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

	EqualSign Position // Position of the '=' or ':='
	Declares  bool     // true if the range expression declares new variables (':=')
	Ordered   bool     // true if the range expression is ordered ('= ordered range')

	// RangeExpression is the expression that is being iterated over.
	RangeExpression Expression

	Position
}

func (RangeExpression) _forExpression() {}

// ============================================================================
// Chain Expression
// ======================================================================================

// ChainExpression is an expression that checks if elements in the chain are
// not zero, and if indexes exist.
type ChainExpression struct {
	// Root is the root expression that is checked.
	Root RawGoCode
	// CheckRoot specifies whether to check if the root is non-zero.
	CheckRoot bool
	// Chain is a list of GoCode that yield the desired value.
	Chain []ChainExpressionItem

	// DerefCount is the number of leading *, used to dereference the chain
	// expression value
	DerefCount int

	// DefaultOperator is the position of the '~' operator.
	//
	// Only set, if Default is.
	DefaultOperator *Position

	// Default represents the optional default value.
	Default *GoCode

	Position
}

func (ChainExpression) _expression()     {}
func (ChainExpression) _forExpression()  {}
func (ChainExpression) _attributeValue() {}

// =============================== Chain Expression Item ================================

// ChainExpressionItem represents an expression that can be chained.
//
// It is either a IndexExpression, or a DotIdentExpression.
type ChainExpressionItem interface {
	_chainExpressionItem()
	Poser
}

// IndexExpression represents either a map or slice index expression.
type IndexExpression struct {
	LBracket Position
	Index    GoCode
	RBracket Position

	CheckIndex bool
	CheckValue bool
}

func (e IndexExpression) Pos() Position { return e.LBracket }

func (IndexExpression) _chainExpressionItem() {}

// DotIdentExpression represents a dot followed by a Go identifier.
type DotIdentExpression struct {
	Ident Ident
	Check bool

	Position // of the dot
}

func (DotIdentExpression) _chainExpressionItem() {}

// ParenExpression represents a function call or the paren part of a type cast.
type ParenExpression struct {
	LParen Position
	Args   []GoCode
	RParen Position

	// Check indicates whether to check the return value.
	Check bool
}

func (e ParenExpression) Pos() Position { return e.LParen }

func (ParenExpression) _chainExpressionItem() {}

// TypeAssertionExpression represents a type assertion.
type TypeAssertionExpression struct {
	LParen       Position
	PointerCount int
	Package      *Ident
	Type         Ident
	RParen       Position

	// Check indicates whether to check if the assertion was successful.
	Check bool

	Position // of the dot
}

func (TypeAssertionExpression) _chainExpressionItem() {}
