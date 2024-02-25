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

var _ ScopeNode = (*And)(nil)

func (a *And) Pos() Position { return a.Position }
func (a *And) End() Position {
	if len(a.Attributes) >= 0 {
		return a.Attributes[len(a.Attributes)-1].End()
	}
	return deltaPos(a.Position, 1)
}

func (*And) _node()      {}
func (*And) _scopeNode() {}

// ============================================================================
// AttributeCollection
// ======================================================================================

// An AttributeCollection is a pointer to either a [IDShorthand], a
// [ClassShorthand], or an [AttributeList].
type AttributeCollection interface {
	Node
	_attributeCollection()
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

func (s *IDShorthand) Pos() Position { return s.Position }
func (s *IDShorthand) End() Position { return deltaPos(s.Position, len("#")+len(s.ID)) }

func (*IDShorthand) _node()                {}
func (*IDShorthand) _attributeCollection() {}

// ============================================================================
// Class Shorthand
// ======================================================================================

type ClassShorthand struct {
	Name     string
	Position Position
}

var _ AttributeCollection = (*ClassShorthand)(nil)

func (s *ClassShorthand) Pos() Position { return s.Position }
func (s *ClassShorthand) End() Position { return deltaPos(s.Position, len(".")+len(s.Name)) }

func (*ClassShorthand) _node()               {}
func (ClassShorthand) _attributeCollection() {}

// ============================================================================
// Attribute List
// ======================================================================================

type AttributeList struct {
	LParen     Position
	Attributes []Attribute
	RParen     *Position
}

var _ AttributeCollection = (*AttributeList)(nil)

func (l *AttributeList) Pos() Position { return l.LParen }
func (l *AttributeList) End() Position {
	if l.RParen != nil {
		return *l.RParen
	} else if len(l.Attributes) > 0 {
		return l.Attributes[len(l.Attributes)-1].End()
	}
	return deltaPos(l.LParen, 1)
}

func (*AttributeList) _node()                {}
func (*AttributeList) _attributeCollection() {}

// ============================================================================
// Attribute
// ======================================================================================

// Attribute is a pointer to either an [AndPlaceholder], or a [SimpleAttribute].
type Attribute interface {
	Node
	_attribute()
}

// if this is changed, change the comment above
var (
	_ Attribute = (*AndPlaceholder)(nil)
	_ Attribute = (*SimpleAttribute)(nil)
)

// ================================== And Placeholder ===================================

// AndPlaceholder is an attribute 'named' `&` that is used as a placeholder for
// the attributes attached to a Component call.
//
//	comp foo() {
//	  div { span(&&) [ foo ] }
//	}
//
// It may only be used inside a mixin definition.
type AndPlaceholder struct {
	Position Position
}

var _ Attribute = (*AndPlaceholder)(nil)

func (p *AndPlaceholder) Pos() Position { return p.Position }
func (p *AndPlaceholder) End() Position { return deltaPos(p.Position, len("&&")) }

func (*AndPlaceholder) _node()      {}
func (*AndPlaceholder) _attribute() {}

// ================================== Simple Attribute ==================================

type SimpleAttribute struct {
	Name   string
	Assign *Position      // nil for boolean attributes
	Value  AttributeValue // nil for boolean attributes

	Position Position
}

var _ Attribute = (*SimpleAttribute)(nil)

func (a *SimpleAttribute) Pos() Position { return a.Position }
func (a *SimpleAttribute) End() Position {
	if a.Value != nil {
		return a.Value.End()
	} else if a.Assign != nil {
		return deltaPos(*a.Assign, 1)
	}
	return deltaPos(a.Position, len(a.Name))
}

func (*SimpleAttribute) _node()      {}
func (*SimpleAttribute) _attribute() {}

// ============================================================================
// Attribute Value
// ======================================================================================

// AttributeValue is a pointer to either an [Expression], a
// [ComponentCallAttributeValue], or a [TypedAttributeValue].
type AttributeValue interface {
	Node
	_attributeValue()
}

// if this is changed, change the comment above
var (
	_ AttributeValue = (Expression)(nil) // interface
	_ AttributeValue = (*ComponentCallAttributeValue)(nil)
	_ AttributeValue = (*TypedAttributeValue)(nil)
)

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

func (v *TypedAttributeValue) Pos() Position { return v.Position }
func (v *TypedAttributeValue) End() Position {
	if v.RParen != nil {
		return deltaPos(*v.RParen, 1)
	} else if v.Value != nil {
		return v.Value.End()
	} else if v.LParen != nil {
		return deltaPos(*v.LParen, 1)
	}
	return deltaPos(v.Position, len(v.Type.String()))
}

func (*TypedAttributeValue) _node()           {}
func (*TypedAttributeValue) _attributeValue() {}

// ============================================================================
// Component Call Attribute Value
// ======================================================================================

type ComponentCallAttributeValue struct {
	ComponentCall *ComponentCall
	Value         *InterpolationValue

	Position Position
}

var _ AttributeValue = (*ComponentCallAttributeValue)(nil)

func (a *ComponentCallAttributeValue) Pos() Position { return a.Position }
func (a *ComponentCallAttributeValue) End() Position {
	if a.Value != nil {
		return a.Value.End()
	}
	return a.ComponentCall.End()
}

func (*ComponentCallAttributeValue) _node()           {}
func (*ComponentCallAttributeValue) _attributeValue() {}
