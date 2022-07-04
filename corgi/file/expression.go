// Package file provides structs that represent the structure of a corgi file.
package file

// GoLiteral represents a Go literal.
//
// It is purely used for easier identification of the expected contents of a
// string.
type GoLiteral string

// GoIdent represents a Go identifier.
//
// It is purely used for easier identification of the expected contents of a
// string.
type GoIdent string

// Ident represents a corgi identifier.
//
// It is purely used for easier identification of the expected contents of a
// string.
type Ident string

// Expression represents an Expression, it is either a GoExpression, a
// TernaryExpression, or a NilCheckExpression.
type Expression interface {
	_typeExpression()
	_typeInlineElementValue()
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

func (GoExpression) _typeExpression()         {}
func (GoExpression) _typeInlineElementValue() {}

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

func (TernaryExpression) _typeExpression()         {}
func (TernaryExpression) _typeInlineElementValue() {}

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

func (NilCheckExpression) _typeExpression()         {}
func (NilCheckExpression) _typeInlineElementValue() {}

// ================================== Expression Expression ==================================

// ValueExpression represents an expression that can be chained.
//
// It is either a IndexExpression, or a FieldFuncExpression.
type ValueExpression interface {
	_typeValueExpression()
}

// IndexExpression represents indexing being performed on another value, e.g.
// foo[1] (1) or bar["fooz"] ("fooz").
type IndexExpression GoExpression

func (IndexExpression) _typeValueExpression() {}

// FieldFuncExpression represents access to a field or function of another
// value, .e.g. foo.Bar (Bar), foo.Baz() (Baz()) or foo.Fooz("arg") (Fooz("arg)).
type FieldFuncExpression GoExpression

func (FieldFuncExpression) _typeValueExpression() {}
