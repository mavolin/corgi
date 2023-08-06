// Package file provides structs that represent the structure of a corgi file.
package file

// Expression represents a chain of [ExpressionItem] elements.
type Expression struct {
	Expressions []ExpressionItem
}

func (e Expression) Pos() Position {
	if len(e.Expressions) > 0 {
		return e.Expressions[0].Pos()
	}
	return InvalidPosition
}

type ExpressionItem interface {
	_typeExpressionItem()
	Poser
}

// ============================================================================
// Go Expression
// ======================================================================================

// GoExpression is a raw Go expression.
type GoExpression struct {
	Expression string
	Position
}

func (GoExpression) _typeExpressionItem() {}

// ============================================================================
// RangeExpression
// ======================================================================================

// RangeExpression represents a range expression as used for for-loops.
//
// Hence, it is only present on for [For] items.
type RangeExpression struct {
	Var1, Var2 *GoIdent

	EqPos    Position // Position of the '=' or ':='
	Declares bool     // true if the range expression declares new variables (':=')
	Ordered  bool     // true if the range expression is ordered ('= ordered range')

	// RangeExpression is the expression that is being iterated over.
	RangeExpression Expression

	Position
}

func (RangeExpression) _typeExpressionItem() {}

// ============================================================================
// String Expression
// ======================================================================================

// StringExpression represents a sequence of characters enclosed in double
// quotes or backticks.
type StringExpression struct {
	Quote    byte // either " or `
	Contents []StringExpressionItem

	// Position is the position of the quote.
	Position
}

var _ ExpressionItem = StringExpression{}

func (StringExpression) _typeExpressionItem() {}

// ============================== String Expression Items ===============================

type StringExpressionItem interface {
	_typeStringExpressionItem()
}

type StringExpressionText struct {
	Text string
	Position
}

var _ StringExpressionItem = StringExpressionText{}

func (StringExpressionText) _typeStringExpressionItem() {}

type StringExpressionInterpolation struct {
	FormatDirective      string // a Sprintf formatting placeholder, w/o preceding '%'
	Expression           Expression
	LBracePos, RBracePos Position
	Position
}

var _ StringExpressionItem = StringExpressionInterpolation{}

func (StringExpressionInterpolation) _typeStringExpressionItem() {}

// ============================================================================
// Ternary Expression
// ======================================================================================

// TernaryExpression represents a ternary expression.
type TernaryExpression struct {
	// Condition is the Expression yielding the condition that is being
	// evaluated.
	Condition Expression // not a ChainExpression

	// IfTrue is the Expression used if the condition is true.
	IfTrue Expression // not a ChainExpression
	// IfFalse is the Expression used if the condition is false.
	IfFalse Expression // not a ChainExpression

	RParenPos Position
	// LParenPos is Position.Col+1
	Position
}

var _ ExpressionItem = TernaryExpression{}

func (TernaryExpression) _typeExpressionItem() {}

// ============================================================================
// Chain Expression
// ======================================================================================

// ChainExpression is an expression that checks if elements in the chain are
// not zero, and if indexes exist.
type ChainExpression struct {
	// Root is the root expression that is checked.
	Root GoExpression
	// CheckRoot specifies whether to check if the root is non-zero.
	CheckRoot bool
	// Chain is a list of GoExpression that yield the desired value.
	Chain []ChainExpressionItem

	// DerefCount is the number of leading *, used to dereference the chain
	// expression value
	DerefCount int

	// DefaultOpPos is the position of the '?!' operator.
	//
	// Only set, if Default is.
	DefaultOpPos *Position

	// Default represents the optional default value.
	Default *Expression // not a ChainExpression

	Position
}

var _ ExpressionItem = ChainExpression{}

func (ChainExpression) _typeExpressionItem() {}

// =============================== Chain Expression Item ================================

// ChainExpressionItem represents an expression that can be chained.
//
// It is either a IndexExpression, or a DotIdentExpression.
type ChainExpressionItem interface {
	_typeChainExpressionItem()
	Poser
}

// IndexExpression represents either a map or slice index expression.
type IndexExpression struct {
	LBracePos Position
	Index     Expression
	RBracePos Position

	CheckIndex bool
	CheckValue bool
}

func (e IndexExpression) Pos() Position { return e.LBracePos }

var _ ChainExpressionItem = IndexExpression{}

func (IndexExpression) _typeChainExpressionItem() {}

// DotIdentExpression represents a dot followed by a Go identifier.
type DotIdentExpression struct {
	Ident GoIdent
	Check bool

	Position // of the dot
}

var _ ChainExpressionItem = DotIdentExpression{}

func (DotIdentExpression) _typeChainExpressionItem() {}

// ParenExpression represents a function call or the paren part of a type cast.
type ParenExpression struct {
	LParenPos Position
	Args      []Expression
	RParenPos Position

	// Check indicates whether to check the return value.
	Check bool
}

func (e ParenExpression) Pos() Position { return e.LParenPos }

var _ ChainExpressionItem = ParenExpression{}

func (ParenExpression) _typeChainExpressionItem() {}

// TypeAssertionExpression represents a type assertion.
type TypeAssertionExpression struct {
	PointerCount int
	Package      *GoIdent
	Type         GoIdent
	RParenPos    Position

	// Check indicates whether to check if the assertion was successful.
	Check bool

	Position // of the dot
}

var _ ChainExpressionItem = TypeAssertionExpression{}

func (TypeAssertionExpression) _typeChainExpressionItem() {}
