package ast

import "slices"

// Attribute is a pointer to either a [IDShorthand], a [ClassShorthand], or a
// [NamedAttribute].
type Attribute interface {
	Argument
	_attribute()
}

// if this is changed, change the comment above
var (
	_ Attribute = (*IDShorthand)(nil)
	_ Attribute = (*ClassShorthand)(nil)
	_ Attribute = (*NamedAttribute)(nil)
)

// ============================================================================
// ID Shorthand
// ======================================================================================

type IDShorthand struct {
	Hash *Position
	ID   Shorthand
}

var _ Attribute = (*IDShorthand)(nil)

func (s *IDShorthand) Start() Position {
	if s.Hash != nil {
		return *s.Hash
	}
	return Position{}
}

func (s *IDShorthand) End() Position {
	if len(s.ID) > 0 {
		return s.ID[len(s.ID)-1].End()
	} else if s.Hash != nil {
		return deltaPos(*s.Hash, len("#"))
	}
	return Position{}
}

func (s *IDShorthand) Walk(w func(Node)) {
	if s.ID != nil {
		w(s.ID)
	}
}

func (*IDShorthand) _node()      {}
func (*IDShorthand) _argument()  {}
func (*IDShorthand) _attribute() {}

// ============================================================================
// Class Shorthand
// ======================================================================================

type ClassShorthand struct {
	Dot   *Position
	Names []Shorthand
}

var _ Attribute = (*ClassShorthand)(nil)

func (s *ClassShorthand) Start() Position {
	if s.Dot != nil {
		return *s.Dot
	}
	for _, name := range s.Names {
		if len(name) > 0 {
			return name.Start()
		}
	}
	return Position{}
}

func (s *ClassShorthand) End() Position {
	for _, name := range slices.Backward(s.Names) {
		if len(name) > 0 {
			return name.End()
		}
	}
	if s.Dot != nil {
		return deltaPos(*s.Dot, len("."))
	}
	return Position{}
}

func (s *ClassShorthand) Walk(w func(Node)) {
	for _, name := range s.Names {
		if name != nil {
			w(name)
		}
	}
}

func (*ClassShorthand) _node()     {}
func (*ClassShorthand) _argument() {}
func (ClassShorthand) _attribute() {}

// ============================================================================
// Shorthand
// ======================================================================================

// Shorthand is the text of an ID or Class shorthand.
type Shorthand []ShorthandNode

var _ Node = (Shorthand)(nil)

func (s Shorthand) Start() Position {
	for _, node := range s {
		if node != nil {
			return node.Start()
		}
	}
	return Position{}
}

func (s Shorthand) End() Position {
	for _, node := range slices.Backward(s) {
		if node != nil {
			return node.End()
		}
	}
	return Position{}
}

func (s Shorthand) Walk(w func(Node)) {
	for _, node := range s {
		if node != nil {
			w(node)
		}
	}
}

func (Shorthand) _node() {}

// =================================== Shorthand Node ===================================

// A ShorthandNode is a pointer to either [ShorthandText], or
// [ExpressionInterpolation].
type ShorthandNode interface {
	Node
	_shorthandNode()
}

var (
	_ ShorthandNode = (*ShorthandText)(nil)
	_ ShorthandNode = (*ShorthandInterpolation)(nil)
)

// ============================================================================
// Shorthand Text
// ======================================================================================

type ShorthandText struct {
	Text     string
	Position *Position
}

var _ ShorthandNode = (*ShorthandText)(nil)

func (t *ShorthandText) Start() Position {
	if t.Position != nil {
		return *t.Position
	}
	return Position{}
}

func (t *ShorthandText) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return Position{}
}
func (*ShorthandText) Walk(func(Node)) {}

func (*ShorthandText) _node()          {}
func (*ShorthandText) _shorthandNode() {}

// ============================================================================
// Shorthand Interpolation
// ======================================================================================

type ShorthandInterpolation ExpressionInterpolation

var _ ShorthandNode = (*ShorthandInterpolation)(nil)

func (interp *ShorthandInterpolation) Start() Position {
	return (*ExpressionInterpolation)(interp).Start()
}
func (interp *ShorthandInterpolation) End() Position { return (*ExpressionInterpolation)(interp).End() }
func (interp *ShorthandInterpolation) Walk(w func(Node)) {
	w((*ExpressionInterpolation)(interp))
}

func (*ShorthandInterpolation) _node()          {}
func (*ShorthandInterpolation) _shorthandNode() {}

// ============================================================================
// Named Attribute
// ======================================================================================

type NamedAttribute struct {
	Name      *AttributeReference
	EqualSign *Position      // nil for boolean attributes
	Value     AttributeValue // nil for boolean attributes
}

var _ Attribute = (*NamedAttribute)(nil)

func (a *NamedAttribute) Start() Position {
	switch {
	case a.Name != nil:
		return a.Name.Start()
	case a.EqualSign != nil:
		return *a.EqualSign
	case a.Value != nil:
		return a.Value.Start()
	}
	return Position{}
}

func (a *NamedAttribute) End() Position {
	switch {
	case a.Value != nil:
		return a.Value.End()
	case a.EqualSign != nil:
		return deltaPos(*a.EqualSign, len("="))
	case a.Name != nil:
		return a.Name.End()
	}
	return Position{}
}

func (a *NamedAttribute) Walk(w func(Node)) {
	if a.Name != nil {
		w(a.Name)
	}
	if a.Value != nil {
		w(a.Value)
	}
}

func (*NamedAttribute) _node()      {}
func (*NamedAttribute) _argument()  {}
func (*NamedAttribute) _attribute() {}

// ============================================================================
// Attribute Value
// ======================================================================================

// AttributeValue is a pointer to either an [ExpressionAttributeValue], or
// [TypedAttributeValue].
type AttributeValue interface {
	Node
	_attributeValue()
}

// if this is changed, change the comment above
var (
	_ AttributeValue = (*ExpressionAttributeValue)(nil) // interface
	_ AttributeValue = (*TypedAttributeValue)(nil)
)

// ============================= Expression Attribute Value =============================

type ExpressionAttributeValue Expression

var _ AttributeValue = (*ExpressionAttributeValue)(nil)

func (v *ExpressionAttributeValue) Start() Position { return (*Expression)(v).Start() }
func (v *ExpressionAttributeValue) End() Position   { return (*Expression)(v).End() }
func (v *ExpressionAttributeValue) Walk(w func(Node)) {
	w((*Expression)(v))
}

func (*ExpressionAttributeValue) _node()           {}
func (*ExpressionAttributeValue) _attributeValue() {}

// =============================== Typed Attribute Value ================================

type TypedAttributeValue struct {
	Type   *AttributeType
	LParen *Position
	Value  AttributeValue
	RParen *Position
}

var _ AttributeValue = (*TypedAttributeValue)(nil)

func (v *TypedAttributeValue) Start() Position {
	switch {
	case v.Type != nil:
		return v.Type.Start()
	case v.LParen != nil:
		return *v.LParen
	case v.Value != nil:
		return v.Value.Start()
	}
	return Position{}
}

func (v *TypedAttributeValue) End() Position {
	switch {
	case v.RParen != nil:
		return deltaPos(*v.RParen, len(")"))
	case v.Value != nil:
		return v.Value.End()
	case v.LParen != nil:
		return deltaPos(*v.LParen, len("("))
	case v.Type != nil:
		return v.Type.End()
	}
	return Position{}
}

func (v *TypedAttributeValue) Walk(w func(Node)) {
	if v.Type != nil {
		w(v.Type)
	}
	if v.Value != nil {
		w(v.Value)
	}
}

func (*TypedAttributeValue) _node()           {}
func (*TypedAttributeValue) _attributeValue() {}

// ============================================================================
// Attribute Selector
// ======================================================================================

type AttributeName struct {
	Name     string
	Position *Position
}

var _ Node = (*AttributeName)(nil)

func (n *AttributeName) Start() Position {
	if n.Position != nil {
		return *n.Position
	}
	return Position{}
}

func (n *AttributeName) End() Position {
	if n.Position != nil {
		return deltaPos(*n.Position, len(n.Name))
	}
	return Position{}
}
func (*AttributeName) Walk(func(Node)) {}

func (*AttributeName) _node() {}

// ============================================================================
// Attribute Reference
// ======================================================================================

type AttributeReference struct {
	Package *Identifier
	Dot     *Position
	Name    *AttributeName
}

var _ Node = (*AttributeReference)(nil)

func (r *AttributeReference) Start() Position {
	if r.Package != nil {
		return r.Package.Start()
	} else if r.Name != nil {
		return r.Name.Start()
	}
	return Position{}
}

func (r *AttributeReference) End() Position {
	if r.Name != nil {
		return r.Name.End()
	} else if r.Package != nil {
		return r.Package.End()
	}
	return Position{}
}

func (r *AttributeReference) Walk(w func(Node)) {
	if r.Package != nil {
		w(r.Package)
	}
	if r.Name != nil {
		w(r.Name)
	}
}

func (*AttributeReference) _node() {}
