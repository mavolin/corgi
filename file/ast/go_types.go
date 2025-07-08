package ast

import "slices"

// Types representing their Go counterparts.

// ============================================================================
// Type
// ======================================================================================

// Type represents a Go type.
//
// As we currently use heuristics to capture types, this type captures a
// superset of all valid Go types.
//
// If a future version of the parser is able to parse types more accurately,
// the set of accepted, but invalid types may decrease.
// As such, users should not rely on type objects for invalid types to be
// placed into the ComponentAST by future versions of the parser.
//
// In fact, invalid types are specifically exempt from compatibility
// guarantees, and you should not expect a future version of corgi to continue
// to parse them.
type Type struct {
	// Type is the textual representation of the type, as we found it in the
	// source.
	Type   string
	Parsed ParsedType // may be nil; see doc of ParsedType

	From  Position
	Until Position
}

var (
	_ Node     = (*Type)(nil)
	_ TypeTerm = (*Type)(nil)
)

func (t *Type) Start() Position { return t.From }
func (t *Type) End() Position   { return t.Until }
func (t *Type) Walk(w func(Node)) {
	if t.Parsed != nil {
		w(t.Parsed)
	}
}

func (*Type) _node()     {}
func (*Type) _typeTerm() {}

// ============================================================================
// Parsed Type
// ======================================================================================

// ParsedType represents a type that we actually have an ComponentAST representation
// for.
// This is usually for the subset of types that we need to properly identify
// later on.
//
// Parsing all Go Types would add a huge bloat of ComponentAST nodes, which we currently
// have no need for, and as such would only add to our maintenance burden.
// Solving the problem this way, allows us to incrementally expand the list of
// types as we need, without introducing breaking changes.
//
// Currently, the only types parsed are [NamedType], however, that list may
// be extended in the future.
//
// Hence, a ParsedType field will only be set, if the type is one of the types
// listed above.
type ParsedType interface {
	Node
	_type()
}

// ==================================== Named Type =====================================

type NamedType struct {
	Name     FullIdent
	TypeArgs *TypeArguments // nil if no type args
}

var _ ParsedType = (*NamedType)(nil)

func (t *NamedType) Start() Position {
	if t.Name != nil {
		return t.Name.Start()
	} else if t.TypeArgs != nil {
		return t.TypeArgs.Start()
	}
	return Position{}
}

func (t *NamedType) End() Position {
	if t.TypeArgs != nil {
		return t.TypeArgs.End()
	} else if t.Name != nil {
		return t.Name.End()
	}
	return Position{}
}

func (t *NamedType) Walk(w func(Node)) {
	if t.Name != nil {
		w(t.Name)
	}
	if t.TypeArgs != nil {
		w(t.TypeArgs)
	}
}

func (*NamedType) _node() {}
func (*NamedType) _type() {}

// =================================== Attribute Type ===================================

type AttributeType struct {
	Quote *Position
	Name  *AttributeTypeName

	LBracket  *Position      // optional
	Attribute *AttributeName // optional, only needed for unsafe attributes
	RBracket  *Position      // optional
}

var _ ParsedType = (*AttributeType)(nil)

func (t *AttributeType) Start() Position {
	if t.Quote != nil {
		return *t.Quote
	} else if t.Name != nil {
		return t.Name.Start()
	}
	return Position{}
}

func (t *AttributeType) End() Position {
	switch {
	case t.RBracket != nil:
		return deltaPos(*t.RBracket, len("]"))
	case t.Attribute != nil:
		return t.Attribute.End()
	case t.LBracket != nil:
		return deltaPos(*t.LBracket, len("["))
	case t.Name != nil:
		return t.Name.End()
	case t.Quote != nil:
		return deltaPos(*t.Quote, len("'"))
	}
	return Position{}
}

func (t *AttributeType) Walk(w func(Node)) {
	if t.Name != nil {
		w(t.Name)
	}
	if t.Attribute != nil {
		w(t.Attribute)
	}
}

func (*AttributeType) _node() {}
func (*AttributeType) _type() {}

// ============================================================================
// Type Constraint
// ======================================================================================

type TypeConstraint struct {
	Terms []TypeTerm
}

var _ Node = (*TypeConstraint)(nil)

func (c *TypeConstraint) Start() Position {
	for _, term := range c.Terms {
		if term != nil {
			return term.Start()
		}
	}
	return Position{}
}

func (c *TypeConstraint) End() Position {
	for _, term := range slices.Backward(c.Terms) {
		if term != nil {
			return term.End()
		}
	}
	return Position{}
}

func (c *TypeConstraint) Walk(w func(Node)) {
	for _, term := range c.Terms {
		if term != nil {
			w(term)
		}
	}
}

func (*TypeConstraint) _node() {}

// ============================================================================
// Type Term
// ======================================================================================

// A TypeTerm is a pointer to either a [Type], or an [UnderlyingParam].
type TypeTerm interface {
	Node
	_typeTerm()
}

// ================================== Underlying Type ===================================

type UnderlyingParam struct {
	Tilde *Position
	Type  *Type
}

var _ TypeTerm = (*UnderlyingParam)(nil)

func (p *UnderlyingParam) Start() Position {
	if p.Type != nil {
		return p.Type.Start()
	} else if p.Tilde != nil {
		return *p.Tilde
	}
	return Position{}
}

func (p *UnderlyingParam) End() Position {
	if p.Type != nil {
		return p.Type.End()
	} else if p.Tilde != nil {
		return deltaPos(*p.Tilde, len("~"))
	}
	return Position{}
}

func (p *UnderlyingParam) Walk(w func(Node)) {
	if p.Type != nil {
		w(p.Type)
	}
}

func (*UnderlyingParam) _node()     {}
func (*UnderlyingParam) _typeTerm() {}

// ============================================================================
// Type Args
// ======================================================================================

type TypeArguments struct {
	LBracket *Position
	Types    []*Type
	RBracket *Position
}

var _ Node = (*TypeArguments)(nil)

func (a *TypeArguments) Start() Position {
	switch {
	case a.LBracket != nil:
		return *a.LBracket
	case len(a.Types) > 0:
		return a.Types[0].Start()
	case a.RBracket != nil:
		return deltaPos(*a.RBracket, len("]"))
	}
	return Position{}
}

func (a *TypeArguments) End() Position {
	switch {
	case a.RBracket != nil:
		return deltaPos(*a.RBracket, len("]"))
	case len(a.Types) > 0:
		return a.Types[len(a.Types)-1].End()
	case a.LBracket != nil:
		return deltaPos(*a.LBracket, len("["))
	}
	return Position{}
}

func (a *TypeArguments) Walk(w func(Node)) {
	for _, arg := range a.Types {
		if arg != nil {
			w(arg)
		}
	}
}

func (*TypeArguments) _node() {}

// ============================================================================
// Type Parameters
// ======================================================================================

type TypeParameters struct {
	LBracket *Position
	Params   []*TypeParameter
	RBracket *Position
}

var _ Node = (*TypeParameters)(nil)

func (p *TypeParameters) Start() Position {
	if p.LBracket != nil {
		return *p.LBracket
	}
	for _, param := range p.Params {
		if param != nil {
			return param.Start()
		}
	}
	if p.RBracket != nil {
		return deltaPos(*p.RBracket, len("]"))
	}
	return Position{}
}

func (p *TypeParameters) End() Position {
	if p.RBracket != nil {
		return deltaPos(*p.RBracket, len("["))
	}
	for _, param := range slices.Backward(p.Params) {
		if param != nil {
			return param.End()
		}
	}
	if p.LBracket != nil {
		return deltaPos(*p.LBracket, len("["))
	}
	return Position{}
}

func (p *TypeParameters) Walk(w func(Node)) {
	for _, param := range p.Params {
		if param != nil {
			w(param)
		}
	}
}

func (*TypeParameters) _node() {}

// ============================================================================
// Type Param
// ======================================================================================

type TypeParameter struct {
	Names []*Ident
	Type  TypeTerm
}

var _ Node = (*TypeParameter)(nil)

func (p *TypeParameter) Start() Position {
	for _, name := range p.Names {
		if name != nil {
			return name.Start()
		}
	}
	if p.Type != nil {
		return p.Type.Start()
	}
	return Position{}
}

func (p *TypeParameter) End() Position {
	if p.Type != nil {
		return p.Type.End()
	}
	for _, name := range slices.Backward(p.Names) {
		if name != nil {
			return name.End()
		}
	}
	return Position{}
}

func (p *TypeParameter) Walk(w func(Node)) {
	for _, name := range p.Names {
		if name != nil {
			w(name)
		}
	}
	if p.Type != nil {
		w(p.Type)
	}
}

func (*TypeParameter) _node() {}
