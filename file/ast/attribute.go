package ast

import (
	"github.com/mavolin/corgi/escape/attrtype"
)

// ============================================================================
// And
// ======================================================================================

// And represents an '&' expression.
type And struct {
	Attributes []AttributeCollection
	Position   Position
}

var _ ScopeItem = (*And)(nil)

func (a *And) Pos() Position { return a.Position }
func (*And) _scopeItem()     {}

// ============================================================================
// AttributeCollection
// ======================================================================================

// An AttributeCollection is a pointer to either a [IDShorthand], a
// [ClassShorthand], or an [AttributeList].
type AttributeCollection interface {
	_attributeCollection()
	Poser
}

// if this is changed, change the comment above
var (
	_ AttributeCollection = (*IDShorthand)(nil)
	_ AttributeCollection = (*ClassShorthand)(nil)
	_ AttributeCollection = (*AttributeList)(nil)
)

// ============================================================================
// ID Shorthand
// ======================================================================================

type IDShorthand struct {
	ID       string
	Position Position
}

var _ AttributeCollection = (*IDShorthand)(nil)

func (s *IDShorthand) Pos() Position       { return s.Position }
func (*IDShorthand) _attributeCollection() {}

// ============================================================================
// Class Shorthand
// ======================================================================================

type ClassShorthand struct {
	Name     string
	Position Position
}

var _ AttributeCollection = (*ClassShorthand)(nil)

func (s *ClassShorthand) Pos() Position      { return s.Position }
func (ClassShorthand) _attributeCollection() {}

// ============================================================================
// Attribute List
// ======================================================================================

type (
	AttributeList struct {
		LParen     Position
		Attributes []Attribute
		RParen     *Position
	}

	// Attribute is a pointer to either an [AndPlaceholder], or a [SimpleAttribute].
	Attribute interface {
		_attribute()
		Poser
	}
)

// if this is changed, change the comment above
var (
	_ Attribute = (*AndPlaceholder)(nil)
	_ Attribute = (*SimpleAttribute)(nil)
)

var _ AttributeCollection = (*AttributeList)(nil)

func (l *AttributeList) Pos() Position       { return l.LParen }
func (*AttributeList) _attributeCollection() {}

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
	Position Position
}

var _ Attribute = (*AndPlaceholder)(nil)

func (p *AndPlaceholder) Pos() Position { return p.Position }
func (*AndPlaceholder) _attribute()     {}

// ============================================================================
// Simple Attribute
// ======================================================================================

type (
	SimpleAttribute struct {
		Name   string
		Assign *Position      // nil for boolean attributes
		Value  AttributeValue // nil for boolean attributes

		Position Position
	}

	// AttributeValue is a pointer to either an [Expression], a
	// [ComponentCallAttribute], or a [TypedAttributeValue].
	AttributeValue interface {
		_attributeValue()
		Poser
	}
)

// if this is changed, change the comment above
var (
	_ AttributeValue = (Expression)(nil) // interface
	_ AttributeValue = (*ComponentCallAttribute)(nil)
	_ AttributeValue = (*TypedAttributeValue)(nil)
)

var _ Attribute = (*SimpleAttribute)(nil)

func (a *SimpleAttribute) Pos() Position { return a.Position }
func (*SimpleAttribute) _attribute()     {}

// ============================================================================
// Typed Attribute Value
// ======================================================================================

type TypedAttributeValue struct {
	Type   attrtype.Type
	LParen *Position
	Value  AttributeValue
	RParen *Position

	Position Position
}

var _ AttributeValue = (*TypedAttributeValue)(nil)

func (v *TypedAttributeValue) Pos() Position  { return v.Position }
func (*TypedAttributeValue) _attributeValue() {}

// ============================================================================
// Component Call Attribute Value
// ======================================================================================

type ComponentCallAttribute struct {
	ComponentCall *ComponentCall
	Value         *InterpolationValue

	Position Position
}

var _ AttributeValue = (*ComponentCallAttribute)(nil)

func (a *ComponentCallAttribute) Pos() Position  { return a.Position }
func (*ComponentCallAttribute) _attributeValue() {}
