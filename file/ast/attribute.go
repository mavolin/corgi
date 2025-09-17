package ast

import "slices"

// Attribute is a pointer to either an [AndPlaceholder], a [IDShorthand], a
// [ClassShorthand], or a [NamedAttribute].
type Attribute interface {
	Argument
	_attribute()
}

// if this is changed, change the comment above
var (
	_ Attribute = (*AndPlaceholder)(nil)
	_ Attribute = (*IDShorthand)(nil)
	_ Attribute = (*ClassShorthand)(nil)
	_ Attribute = (*NamedAttribute)(nil)
)

// ============================================================================
// And Placeholder
// ======================================================================================

// AndPlaceholder is an attribute 'named' `&` that is used as a placeholder for
// the attributes attached to a Component call.
//
//	comp foo() {
//	  div { span(&) [ foo ] }
//	}
type AndPlaceholder struct {
	And *Position
}

var (
	_ Attribute            = (*AndPlaceholder)(nil)
	_ AndPlaceholderWriter = (*AndPlaceholder)(nil)
)

func (p *AndPlaceholder) Start() Position {
	if p.And != nil {
		return *p.And
	}
	return NoPosition
}

func (p *AndPlaceholder) End() Position {
	if p.And != nil {
		return deltaPos(*p.And, len("&"))
	}
	return NoPosition
}

func (*AndPlaceholder) Walk(func(Node)) {}

func (*AndPlaceholder) _node()                 {}
func (*AndPlaceholder) _argument()             {}
func (*AndPlaceholder) _attribute()            {}
func (*AndPlaceholder) _andPlaceholderWriter() {}

// ============================================================================
// ID Shorthand
// ======================================================================================

type IDShorthand struct {
	Hash *Position
	ID   Shorthand
}

var (
	_ Attribute       = (*IDShorthand)(nil)
	_ AttributeWriter = (*IDShorthand)(nil)
)

func (s *IDShorthand) Start() Position {
	if s.Hash != nil {
		return *s.Hash
	}
	if len(s.ID) > 0 {
		if start := s.ID.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (s *IDShorthand) End() Position {
	if len(s.ID) > 0 {
		if end := s.ID.End(); end != NoPosition {
			return end
		}
	}
	if s.Hash != nil {
		return deltaPos(*s.Hash, len("#"))
	}
	return NoPosition
}

func (s *IDShorthand) Walk(w func(Node)) {
	if s.ID != nil {
		w(s.ID)
	}
}

func (*IDShorthand) _node()            {}
func (*IDShorthand) _argument()        {}
func (*IDShorthand) _attribute()       {}
func (*IDShorthand) _attributeWriter() {}

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
			if start := name.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (s *ClassShorthand) End() Position {
	for _, name := range slices.Backward(s.Names) {
		if len(name) > 0 {
			if end := name.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.Dot != nil {
		return deltaPos(*s.Dot, len("."))
	}
	return NoPosition
}

func (s *ClassShorthand) Walk(w func(Node)) {
	for _, name := range s.Names {
		if name != nil {
			w(name)
		}
	}
}

func (*ClassShorthand) _node()            {}
func (*ClassShorthand) _argument()        {}
func (*ClassShorthand) _attribute()       {}
func (*ClassShorthand) _attributeWriter() {}

// ============================================================================
// Shorthand
// ======================================================================================

// Shorthand is the text of an ID or Class shorthand.
type Shorthand []ShorthandNode

var _ Node = (Shorthand)(nil)

func (s Shorthand) Start() Position {
	for _, node := range s {
		if node != nil {
			if start := node.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (s Shorthand) End() Position {
	for _, node := range slices.Backward(s) {
		if node != nil {
			if end := node.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
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
	return NoPosition
}

func (t *ShorthandText) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return NoPosition
}
func (*ShorthandText) Walk(func(Node)) {}

func (*ShorthandText) _node()          {}
func (*ShorthandText) _shorthandNode() {}

// ============================================================================
// Shorthand Interpolation
// ======================================================================================

type ShorthandInterpolation ExpressionInterpolation

var _ ShorthandNode = (*ShorthandInterpolation)(nil)

func (inter *ShorthandInterpolation) Start() Position {
	return (*ExpressionInterpolation)(inter).Start()
}
func (inter *ShorthandInterpolation) End() Position { return (*ExpressionInterpolation)(inter).End() }
func (inter *ShorthandInterpolation) Walk(w func(Node)) {
	w((*ExpressionInterpolation)(inter))
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
	if a.Name != nil {
		if start := a.Name.Start(); start != NoPosition {
			return start
		}
	}
	if a.EqualSign != nil {
		return *a.EqualSign
	}
	if a.Value != nil {
		if start := a.Value.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (a *NamedAttribute) End() Position {
	if a.Value != nil {
		if end := a.Value.End(); end != NoPosition {
			return end
		}
	}
	if a.EqualSign != nil {
		return deltaPos(*a.EqualSign, len("="))
	}
	if a.Name != nil {
		if end := a.Name.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (a *NamedAttribute) Walk(w func(Node)) {
	if a.Name != nil {
		w(a.Name)
	}
	if a.Value != nil {
		w(a.Value)
	}
}

func (*NamedAttribute) _node()            {}
func (*NamedAttribute) _argument()        {}
func (*NamedAttribute) _attribute()       {}
func (*NamedAttribute) _attributeWriter() {}

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
	if v.Type != nil {
		if start := v.Type.Start(); start != NoPosition {
			return start
		}
	}
	if v.LParen != nil {
		return *v.LParen
	}
	if v.Value != nil {
		if start := v.Value.Start(); start != NoPosition {
			return start
		}
	}
	if v.RParen != nil {
		return *v.RParen
	}
	return NoPosition
}

func (v *TypedAttributeValue) End() Position {
	if v.RParen != nil {
		return deltaPos(*v.RParen, len(")"))
	}
	if v.Value != nil {
		if end := v.Value.End(); end != NoPosition {
			return end
		}
	}
	if v.LParen != nil {
		return deltaPos(*v.LParen, len("("))
	}
	if v.Type != nil {
		if end := v.Type.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
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
	Name          string
	CanonicalName string // ascii-lowercase version of Name
	Position      *Position
}

var _ Node = (*AttributeName)(nil)

func (n *AttributeName) Start() Position {
	if n.Position != nil {
		return *n.Position
	}
	return NoPosition
}

func (n *AttributeName) End() Position {
	if n.Position != nil {
		return deltaPos(*n.Position, len(n.Name))
	}
	return NoPosition
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
		if start := r.Package.Start(); start != NoPosition {
			return start
		}
	}
	if r.Dot != nil {
		return *r.Dot
	}
	if r.Name != nil {
		if start := r.Name.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (r *AttributeReference) End() Position {
	if r.Name != nil {
		if end := r.Name.End(); end != NoPosition {
			return end
		}
	}
	if r.Dot != nil {
		return deltaPos(*r.Dot, len("."))
	}
	if r.Package != nil {
		if end := r.Package.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
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
