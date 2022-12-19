package file

// ============================================================================
// Scope
// ======================================================================================

// A Scope represents a level of indentation.
// Every mixin available inside a scope is also available in its child scopes.
type Scope []ScopeItem

// ScopeItem represents an item in a scope.
type ScopeItem interface {
	_typeScopeItem()
	Position() (line, col int)
}

// ============================================================================
// Block
// ======================================================================================

type BlockType uint8

const (
	BlockTypeBlock BlockType = iota + 1
	BlockTypeAppend
	BlockTypePrepend
)

// Block represents a block with content.
// It is used for blocks from extendable templates as well as blocks in
// MixinCalls.
type Block struct {
	// Name is the name of the block.
	Name Ident

	// Type is the type of block.
	Type BlockType
	Body Scope

	Pos
}

func (Block) _typeScopeItem() {}

// ============================================================================
// Comment
// ======================================================================================

// Comment represents a comment.
type Comment struct {
	Comment string
	// Printed indicates whether the comment shall be included in the HTML
	// output.
	Printed bool

	Pos
}

func (Comment) _typeScopeItem() {}

// ============================================================================
// Element
// ======================================================================================

// Element represents a single HTML element.
type Element struct {
	// Name is the name of the element.
	Name string

	// Classes is a list of Classes directly added to the element.
	Classes []Class
	// Attributes is a list of Attributes, excluding 'class', that were
	// directly added to the element.
	Attributes []Attribute

	Body Scope

	Pos
}

func (Element) _typeScopeItem() {}

// ===================================== Attribute ======================================

type Attribute interface {
	_typeAttribute()
}

type AttributeLiteral struct {
	// Name is the name of the attribute.
	Name string
	// Value is the expression that yields the value of the attribute.
	Value string

	Pos
}

func (AttributeLiteral) _typeAttribute() {}

type AttributeExpression struct {
	// Name is the name of the attribute.
	Name string
	// Value is the expression that yields the value of the attribute.
	Value Expression
	// NoEscape indicates whether the value of the attribute should be
	// rendered without escaping.
	NoEscape bool

	Pos
}

func (AttributeExpression) _typeAttribute() {}

// ======================================= Class ========================================

type Class interface {
	_typeClass()
}

type ClassLiteral struct {
	Name string
	Pos
}

func (ClassLiteral) _typeClass() {}

type ClassExpression struct {
	// Name is the name of the class.
	Name Expression
	// NoEscape indicates whether the class name should be rendered without
	// escaping.
	NoEscape bool

	Pos
}

func (ClassExpression) _typeClass() {}

// ============================================================================
// Include
// ======================================================================================

type Include struct {
	// Path is the path to the file to include.
	Path string

	// Include is the included file.
	// It is populated by the linter.
	Include IncludeFile

	Pos
}

func (Include) _typeScopeItem() {}

// IncludeFile is the type used to represent an included file.
//
// Its concrete type is either a CorgiInclude or a OtherInclude.
type IncludeFile interface {
	_typeIncludeFile()
}

// =================================== Corgi Include ====================================

type CorgiInclude struct {
	File File
}

func (CorgiInclude) _typeIncludeFile() {}

// ==================================== Raw Include =====================================

// OtherInclude is an included file other than a Corgi file.
type OtherInclude struct {
	Contents string
}

func (OtherInclude) _typeIncludeFile() {}

// ============================================================================
// Code
// ======================================================================================

// Code represents a line or block of code.
type Code struct {
	Code string
	Pos
}

func (Code) _typeScopeItem() {}

// ============================================================================
// If
// ======================================================================================

// If represents an 'if' statement.
type If struct {
	// Condition is the condition of the if statement.
	Condition Expression

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Scope

	// ElseIfs are the else if statements, if this If has any.
	ElseIfs []ElseIf
	// Else is the scope of the Else statement, if this If has one.
	Else *Else

	Pos
}

func (If) _typeScopeItem() {}

// ElseIf represents an 'else if' statement.
type ElseIf struct {
	// Condition is the condition of the else if statement.
	Condition Expression

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Scope

	Pos
}

type Else struct {
	Then Scope
	Pos
}

// ============================================================================
// IfBlock
// ======================================================================================

// IfBlock represents an 'if block' directive.
type IfBlock struct {
	// Name is the name of the block, whose existence is checked.
	Name Ident

	// Then is the scope of the code that is executed if the block exists.
	Then Scope
	// ElseIfs are the else if statements, if this IfBlock has any.
	ElseIfs []ElseIfBlock
	// Else is the scope of the code that is executed if the block does not
	// exist.
	Else *Else

	Pos
}

func (IfBlock) _typeScopeItem() {}

type ElseIfBlock struct {
	// Name is the name of the block, whose existence is checked.
	Name Ident
	// Then is the scope of the code that is executed if the block exists.
	Then Scope

	Pos
}

// ============================================================================
// Switch
// ======================================================================================

// Switch represents a 'switch' statement.
type Switch struct {
	// Comparator is the expression that is compared against.
	//
	// It may be empty, in which case the cases will contain boolean
	// expressions.
	Comparator Expression

	// Cases are the cases of the Switch.
	Cases []Case
	// Default is the default case, if there is one.
	Default *DefaultCase

	Pos
}

func (Switch) _typeScopeItem() {}

type Case struct {
	// Expression is the expression written behind 'case'.
	Expression GoExpression
	// Then is the scope of the code that is executed if the condition
	// evaluates to true.
	Then Scope

	Pos
}

type DefaultCase struct {
	Then Scope
	Pos
}

// ============================================================================
// For
// ======================================================================================

// For represents a for loop.
type For struct {
	// Expression is the expression written in the head of the for, or nil if
	// this is an infinite loop.
	Expression Expression
	Body       Scope

	Pos
}

func (For) _typeScopeItem() {}

// ============================================================================
// &
// ======================================================================================

// And represents an '&' expression.
type And struct {
	// Classes is a list of Classes.
	Classes []Class
	// Attributes is a list of attributes, excluding 'class'.
	Attributes []Attribute

	Pos
}

func (And) _typeScopeItem() {}

// ============================================================================
// Contents
// ======================================================================================

// Text is a string of text written as content of an element.
// It is not HTML-escaped yet.
type Text struct {
	Text string
	Pos
}

func (Text) _typeScopeItem()                 {}
func (Text) _typeElementInterpolationValue() {}

// ============================================================================
// ExpressionInterpolation
// ======================================================================================

// ExpressionInterpolation is an Expression that is interpolated into the
// content of an element.
//
// It is generated through hash-interpolation or through lex.CodeAssigns.
type ExpressionInterpolation struct {
	// Expression is the expression that is interpolated.
	Expression Expression

	NoEscape bool

	Pos
}

func (ExpressionInterpolation) _typeScopeItem() {}

// ============================================================================
// ElementInterpolation
// ======================================================================================

type ElementInterpolation struct {
	// Name is the name of the element.
	Name string

	// Classes is a list of expressions that yield the classes of the element.
	Classes []Class
	// Attributes is a list of the attributes of the element, excluding 'class'.
	Attributes []Attribute

	NoEscape bool

	// Value is the value of the element.
	Value ElementInterpolationValue

	Pos
}

func (ElementInterpolation) _typeScopeItem() {}

// ============================ Element Interpolation Value =============================

// ElementInterpolationValue represents types that can be used as the value of an
// ElementInterpolation.
//
// Its concrete type is either Expression or Text.
type ElementInterpolationValue interface {
	_typeElementInterpolationValue()
}

// ============================================================================
// TextInterpolation
// ======================================================================================

type TextInterpolation struct {
	// Text is the interpolated text.
	Text string

	NoEscape bool

	Pos
}

func (TextInterpolation) _typeScopeItem() {}

// ============================================================================
// Mixin
// ======================================================================================

// Mixin represents the definition of a mixin.
type (
	Mixin struct {
		// Name is the name of the mixin.
		Name Ident

		// Params is a list of the parameters of the mixin.
		Params []MixinParam

		// Body is the scope of the mixin.
		Body Scope

		Pos
	}

	// MixinParam represents a parameter of a mixin.
	MixinParam struct {
		// Name is the name of the parameter.
		Name Ident

		// Default is the optional default value of the parameter.
		Default *GoExpression

		// Type is the name of the type of the parameter.
		Type GoIdent

		Pos
	}
)

func (Mixin) _typeScopeItem() {}

type (
	// MixinCall represents the call to a mixin.
	MixinCall struct {
		// Namespace is the namespace of the mixin.
		Namespace Ident
		// Name is the name of the mixin.
		Name Ident

		// Mixin is a pointer to the called mixin.
		//
		// It is set by the linker.
		Mixin *Mixin
		// MixinSource is the resource source that provides the Mixin.
		MixinSource string
		// MixinFile is the file that provides the mixin.
		MixinFile string

		// Args is a list of the arguments of given to the mixin.
		Args []MixinArg

		// Body is the body of the mixin call.
		//
		// It will only consist of If, IfBlock, Switch, And, and Block items.
		Body Scope

		Pos // for linking
	}

	// MixinArg represents a single argument given to a mixin.
	MixinArg struct {
		// Name is the name of the argument.
		Name Ident
		// Value is the expression that yields the value of the argument.
		Value Expression
		// NoEscape indicates whether the value of the argument should be
		// not be escaped.
		NoEscape bool

		Pos
	}
)

func (MixinCall) _typeScopeItem() {}

// ============================================================================
// Filter
// ======================================================================================

type Filter struct {
	Name string
	Args []string

	Body Text

	Pos
}

func (Filter) _typeScopeItem() {}
