// Package file provides structs that represent the structure of a corgi file.
package file

// Expression represents a chain of [ExpressionItem] elements.
type Expression struct {
	Expressions []ExpressionItem
}

func (e Expression) _typeElementInterpolationValue() {}

// ================================== Expression Item ===================================

type ExpressionItem interface {
	_typeExpressionItem()
}

// ============================================================================
// Go Expression
// ======================================================================================

// GoExpression is a raw Go expression.
//
// GoExpressions should get written to the generated file as is, without
// any interpretation.
type GoExpression struct {
	Expression string

	Pos
}

func (GoExpression) _typeExpressionItem() {}

// ============================================================================
// String Expression
// ======================================================================================

// StringExpression represents a sequence of characters enclosed in double
// quotes or backticks.
type StringExpression struct {
	Contents []StringExpressionItem

	Pos
}

func (StringExpression) _typeExpressionItem() {}

// ============================== String Expression Items ===============================

type StringExpressionItem interface {
	_typeStringExpressionItem()
}

type RawStringExpression struct {
	Expression string
	Pos
}

func (RawStringExpression) _typeStringExpressionItem() {}

type InterpolationStringExpression struct {
	Expression Expression
	NoEscape   bool
	Pos
}

// ============================================================================
// Ternary Expression
// ======================================================================================

// TernaryExpression represents a ternary expression.
type TernaryExpression struct {
	// Condition is the Expression yielding the condition that is being
	// evaluated.
	Condition GoExpression

	// IfTrue is the Expression used if the condition is true.
	IfTrue Expression
	// IfFalse is the Expression used if the condition is false.
	IfFalse Expression

	Pos
}

func (TernaryExpression) _typeExpressionItem()            {}
func (TernaryExpression) _typeElementInterpolationValue() {}

// ============================================================================
// Nil Check Expression
// ======================================================================================

// NilCheckExpression represents a nil check expression.
type NilCheckExpression struct {
	// Root is the root expression that is checked.
	Root GoExpression
	// Chain is a list of GoExpression that yield the desired value.
	Chain []ValueExpression

	// Deref contains optional dereference operators (*) to be applied to the
	// resolved value in case the value is accessible.
	Deref string

	// Default represents the optional default value.
	Default *GoExpression

	Pos
}

func (NilCheckExpression) _typeExpressionItem()            {}
func (NilCheckExpression) _typeElementInterpolationValue() {}

// ================================== Value Expression ==================================

// ValueExpression represents an expression that can be chained.
//
// It is either a IndexExpression, or a FieldMethodExpression.
type ValueExpression interface {
	_typeValueExpression()
}

// IndexExpression represents indexing being performed on another value.
//
// Examples
//
//	base[1] => 1
//	base["fooz"] => "fooz"
type IndexExpression GoExpression

func (IndexExpression) _typeValueExpression() {}

// FieldMethodExpression represents access to a field or method of a value.
//
// Examples
//
//		base.Bar => Bar
//		base.Baz() => Baz
//	 base.Fooz("arg") => Fooz
type FieldMethodExpression GoExpression

func (FieldMethodExpression) _typeValueExpression() {}

// FuncCallExpression represents a call to a function or method.
// It contains the raw args of the call.
//
// Examples
//
//	base("Foo") => "Foo"
//	base(12, true,  "bar") => 12, true,  "bar"
type FuncCallExpression GoExpression

func (FuncCallExpression) _typeValueExpression() {}
