package file

// ============================================================================
// And
// ======================================================================================

// And represents an '&' expression.
type And struct {
	Attributes []AttributeCollection

	Position
}

func (And) _scopeItem() {}

// ============================================================================
// AttributeCollection
// ======================================================================================

type AttributeCollection interface {
	_attributeCollection()
	Poser
}

// ==================================== IDShorthand =====================================

type IDShorthand struct {
	ID string
	Position
}

func (IDShorthand) _attributeCollection() {}

// =================================== ClassShorthand ===================================

type ClassShorthand struct {
	Name string
	Position
}

var _ AttributeCollection = ClassShorthand{}

func (ClassShorthand) _attributeCollection() {}

// =================================== AttributeList ====================================

type AttributeList struct {
	LParenPos  Position
	Attributes []Attribute
	RParenPos  Position
}

func (AttributeList) _attributeCollection() {}
func (l AttributeList) Pos() Position       { return l.LParenPos.Pos() }

// ============================================================================
// Attribute
// ======================================================================================

type Attribute interface {
	_attribute()
	Poser
}

// ================================== SimpleAttribute ===================================

type SimpleAttribute struct {
	Name   string
	Assign *Position   // nil for boolean attributes
	Value  *Expression // nil for boolean attributes

	Position
}

func (SimpleAttribute) _attribute() {}

// =================================== AndPlaceholder ===================================

// AndPlaceholder is an attribute 'named' `&` that is used as a placeholder for
// the attributes attached to a Component call.
//
//		comp foo() {
//		  div { span(&&) [ foo ] }
//	 }
//
// It may only be used inside a mixin definition.
type AndPlaceholder struct {
	Position
}

func (AndPlaceholder) _attribute() {}

// ================================ Component Attribute =================================

type ComponentAttribute struct {
	Name string

	AssignPos Position

	ComponentCall ComponentCall
	Value         InterpolationValue

	Position
}

func (ComponentAttribute) _attribute() {}
