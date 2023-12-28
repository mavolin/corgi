package file

import "github.com/mavolin/corgi/escape/attrtype"

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

// ============================================================================
// ID Shorthand
// ======================================================================================

type IDShorthand struct {
	ID string
	Position
}

func (IDShorthand) _attributeCollection() {}

// ============================================================================
// Class Shorthand
// ======================================================================================

type ClassShorthand struct {
	Name string
	Position
}

var _ AttributeCollection = ClassShorthand{}

func (ClassShorthand) _attributeCollection() {}

// ============================================================================
// Attribute List
// ======================================================================================

type (
	AttributeList struct {
		LParen     Position
		Attributes []Attribute
		RParen     Position
	}

	Attribute interface {
		_attribute()
		Poser
	}
)

func (AttributeList) _attributeCollection() {}
func (l AttributeList) Pos() Position       { return l.LParen.Pos() }

// ============================================================================
// And Placeholder
// ======================================================================================

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

// ============================================================================
// Simple Attribute
// ======================================================================================

type (
	SimpleAttribute struct {
		Name   string
		Assign *Position // nil for boolean attributes
		Value  AttributeValue

		Position
	}

	// AttributeValue is either an [Expression], a
	// [ComponentCallAttributeValue], or a [TypedAttributeValue].
	AttributeValue interface {
		_attributeValue()
		Poser
	}
)

func (SimpleAttribute) _attribute() {}

// ============================================================================
// Typed Attribute Value
// ======================================================================================

type TypedAttributeValue struct {
	Type   attrtype.Type
	LParen Position
	Value  AttributeValue
	RParen Position

	Position
}

func (TypedAttributeValue) _attributeValue() {}

// ============================================================================
// Component Call Attribute Value
// ======================================================================================

type ComponentCallAttribute struct {
	ComponentCall ComponentCall
	Value         InterpolationValue

	Position
}

func (ComponentCallAttribute) _attributeValue() {}
