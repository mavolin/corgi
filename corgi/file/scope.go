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
	//
	// This field is optional for blocks used in a mixin call.
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

// Comment represents a rendered HTML comment.
type Comment struct {
	Comment string
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

	// SelfClosing indicates whether this Element should use a '/' to
	// self-close.
	SelfClosing bool

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
}

func (AttributeExpression) _typeAttribute() {}

// ======================================= Class ========================================

type Class interface {
	_typeClass()
}

type ClassLiteral struct {
	Name string
}

func (ClassLiteral) _typeClass() {}

type ClassExpression struct {
	// Name is the name of the class.
	Name Expression
	// NoEscape indicates whether the class name should be rendered without
	// escaping.
	NoEscape bool
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
	Include IncludeValue

	Pos
}

func (Include) _typeScopeItem() {}

// IncludeValue is the type used to represent an included file.
//
// Its concrete type is either a CorgiInclude or a RawInclude.
type IncludeValue interface {
	_typeIncludeValue()
}

// =================================== Corgi Include ====================================

type CorgiInclude struct {
	File File
}

func (CorgiInclude) _typeIncludeValue() {}

// ==================================== Raw Include =====================================

type RawInclude struct {
	Text string
}

func (RawInclude) _typeIncludeValue() {}

// ============================================================================
// Code
// ======================================================================================

// Code represents a line or block of code.
type Code struct {
	Code string
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
}

func (If) _typeScopeItem() {}

// ElseIf represents an 'else if' statement.
type ElseIf struct {
	// Condition is the condition of the else if statement.
	Condition Expression

	// Then is scope of the code that is executed if the condition evaluates
	// to true.
	Then Scope
}

type Else struct {
	Then Scope
}

// ============================================================================
// IfBlock
// ======================================================================================

// IfBlock represents an 'if block' directive.
type IfBlock struct {
	// Name is the name of the block, whose existence is checked.
	//
	// It may be empty for mixins with a single block.
	Name Ident

	// Then is the scope of the code that is executed if the block exists.
	Then Scope
	// Else is the scope of the code that is executed if the block does not
	// exist.
	Else *Else
}

func (IfBlock) _typeScopeItem() {}

// ============================================================================
// Switch
// ======================================================================================

// Switch represents a 'switch' statement.
type Switch struct {
	// Comparator is the expression that is compared against.
	Comparator Expression

	// Cases are the cases of the Switch.
	Cases []Case
	// Default is the default case, if there is one.
	Default *DefaultCase
}

func (Switch) _typeScopeItem() {}

type Case struct {
	// Expression is the expression written behind 'case'.
	Expression GoExpression
	// Then is the scope of the code that is executed if the condition
	// evaluates to true.
	Then Scope
}

type DefaultCase struct {
	Then Scope
}

// ============================================================================
// For
// ======================================================================================

// For represents a for loop.
type For struct {
	// VarOne is the first variable of the range.
	VarOne GoIdent
	// VarTwo is the optional second variable of the range.
	VarTwo GoIdent

	// Range is the expression that is used to iterate over.
	Range Expression

	Body Scope

	Pos
}

func (For) _typeScopeItem() {}

// ============================================================================
// While
// ======================================================================================

// While represents a while loop.
type While struct {
	// Condition is the condition of the while loop.
	Condition GoExpression

	Body Scope

	Pos
}

func (While) _typeScopeItem() {}

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
// Text
// ======================================================================================

// Text is a string of text written as content of an element.
// It is not HTML-escaped yet.
type Text struct {
	Text string
}

func (Text) _typeScopeItem()          {}
func (Text) _typeInlineElementValue() {}

// ============================================================================
// Interpolation
// ======================================================================================

// Interpolation is an Expression that is interpolated into the content of an
// element.
//
// It is generated through hash-interpolation or through lex.CodeAssigns.
type Interpolation struct {
	// Expression is the expression that is interpolated.
	Expression Expression

	NoEscape bool
}

func (Interpolation) _typeScopeItem() {}

// ============================================================================
// InlineElement
// ======================================================================================

type InlineElement struct {
	// Name is the name of the element.
	Name string

	// Classes is a list of expressions that yield the classes of the element.
	Classes []Class
	// Attributes is a list of the attributes of the element, excluding 'class'.
	Attributes []Attribute

	SelfClosing bool

	NoEscape bool

	// Value is the value of the element.
	Value InlineElementValue
}

func (InlineElement) _typeScopeItem() {}

// ================================= InlineElementValue =================================

// InlineElementValue represents types that can be used as the value of an
// InlineElement.
//
// Its concrete type is either Expression or Text.
type InlineElementValue interface {
	_typeInlineElementValue()
}

// ============================================================================
// InlineText
// ======================================================================================

type InlineText struct {
	// Text is the interpolated text.
	Text string

	NoEscape bool
}

func (InlineText) _typeScopeItem() {}

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
		Mixin *Mixin
		// MixinSource is the resource source that provides the Mixin.
		MixinSource string
		// MixinFile is the file that provides the mixin.
		MixinFile string

		// Args is a list of the arguments of given to the mixin.
		Args []MixinArg

		// Body is the body of the mixin.
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

		Pos // for linking
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
