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
	_expressionItem()
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

func (GoExpression) _expressionItem() {}

// ============================================================================
// RangeExpression
// ======================================================================================

// RangeExpression represents a range expression as used for for-loops.
//
// Hence, it is only present on for [For] items.
type RangeExpression struct {
	Var1, Var2 *GoIdent

	EqualSign Position // Position of the '=' or ':='
	Declares  bool     // true if the range expression declares new variables (':=')
	Ordered   bool     // true if the range expression is ordered ('= ordered range')

	// RangeExpression is the expression that is being iterated over.
	RangeExpression Expression

	Position
}

func (RangeExpression) _expressionItem() {}

// ============================================================================
// String Expression
// ======================================================================================

// StringExpression represents a sequence of characters enclosed in double
// quotes or backticks.
type StringExpression struct {
	Quote    byte // either '"' or '`'
	Contents []StringExpressionItem

	// Position is the position of the quote.
	Position
}

func (StringExpression) _expressionItem() {}

// ============================== String Expression Items ===============================

type StringExpressionItem interface {
	_stringExpressionItem()
}

type StringExpressionText struct {
	Text string
	Position
}

func (StringExpressionText) _stringExpressionItem() {}

type StringExpressionInterpolation struct {
	FormatDirective string // a Sprintf formatting placeholder, w/o preceding '%'
	Expression      Expression
	LBrace, RBrace  Position
	Position
}

func (StringExpressionInterpolation) _stringExpressionItem() {}

// ============================================================================
// Ternary Expression
// ======================================================================================

// TernaryFunction represents a ternary expression.
type TernaryFunction struct {
	Condition Expression // not a ChainExpression

	IfTrue  Expression // not a ChainExpression
	IfFalse Expression // not a ChainExpression

	LParen Position
	RParen Position
	Position
}

func (TernaryFunction) _expressionItem() {}

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

	// DefaultOperator is the position of the '~' operator.
	//
	// Only set, if Default is.
	DefaultOperator *Position

	// Default represents the optional default value.
	Default *Expression // not a ChainExpression

	Position
}

func (ChainExpression) _expressionItem() {}

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

func (IndexExpression) _typeChainExpressionItem() {}

// DotIdentExpression represents a dot followed by a Go identifier.
type DotIdentExpression struct {
	Ident GoIdent
	Check bool

	Position // of the dot
}

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

func (TypeAssertionExpression) _typeChainExpressionItem() {}
