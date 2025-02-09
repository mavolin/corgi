package ast

import (
	"slices"

	"github.com/mavolin/corgi/v2/escape/elemtype"
)

type ElementDefinition struct {
	Elem   *Position
	Prefix *ElementName
	LParen *Position // nil if not a list
	Specs  []*ElementSpec
	RParen *Position // nil if not a list
}

var _ ScopeNode = (*ElementDefinition)(nil)

func (d *ElementDefinition) Start() Position {
	if d.Elem != nil {
		return *d.Elem
	} else if d.Prefix != nil {
		return d.Prefix.Start()
	} else if d.LParen != nil {
		return *d.LParen
	}
	for _, spec := range d.Specs {
		if spec != nil {
			return spec.Start()
		}
	}
	if d.RParen != nil {
		return *d.RParen
	}
	return Position{}
}
func (d *ElementDefinition) End() Position {
	if d.RParen != nil {
		return deltaPos(*d.RParen, len(")"))
	}
	for _, spec := range slices.Backward(d.Specs) {
		if spec != nil {
			return spec.End()
		}
	}
	if d.LParen != nil {
		return deltaPos(*d.LParen, len("("))
	} else if d.Prefix != nil {
		return d.Prefix.End()
	} else if d.Elem != nil {
		return deltaPos(*d.Elem, len("elem"))
	}
	return Position{}
}

func (*ElementDefinition) _node()      {}
func (*ElementDefinition) _scopeNode() {}

// ============================================================================
// Element Spec
// ======================================================================================

type ElementSpec struct {
	Name *ElementName
	Type ElementType
}

var _ Node = (*ElementSpec)(nil)

func (a *ElementSpec) Start() Position {
	if a.Name != nil {
		return a.Name.Start()
	} else if a.Type != nil {
		return a.Type.Start()
	}
	return Position{}
}
func (a *ElementSpec) End() Position {
	if a.Type != nil {
		return a.Type.End()
	} else if a.Name != nil {
		return a.Name.End()
	}
	return Position{}
}

func (*ElementSpec) _node() {}

// ============================================================================
// Element Type
// ======================================================================================

type ElementType interface {
	Node
	_elementType()
}

// ================================= Basic Element Type =================================

type BasicElementType struct {
	Type *ElementTypeName
}

var _ ElementType = (*BasicElementType)(nil)

func (t *BasicElementType) Start() Position {
	if t.Type != nil {
		return t.Type.Start()
	}
	return Position{}
}
func (t *BasicElementType) End() Position {
	if t.Type != nil {
		return t.Type.End()
	}
	return Position{}
}
func (t *BasicElementType) _node()        {}
func (t *BasicElementType) _elementType() {}

// ================================= Alias Element Type =================================

type AliasElementType struct {
	EqualSign *Position
	Name      *ElementName
}

var _ ElementType = (*AliasElementType)(nil)

func (t *AliasElementType) Start() Position {
	if t.EqualSign != nil {
		return *t.EqualSign
	} else if t.Name != nil {
		return t.Name.Start()
	}
	return Position{}
}
func (t *AliasElementType) End() Position {
	if t.Name != nil {
		return t.Name.End()
	} else if t.EqualSign != nil {
		return deltaPos(*t.EqualSign, len("="))
	}
	return Position{}
}

func (t *AliasElementType) _node()        {}
func (t *AliasElementType) _elementType() {}

// ============================================================================
// Element Type Name
// ======================================================================================

type ElementTypeName struct {
	Name     string
	Type     elemtype.Type
	Position *Position
}

var _ Node = (*ElementTypeName)(nil)

func (a *ElementTypeName) Start() Position {
	if a.Position != nil {
		return *a.Position
	}
	return Position{}
}
func (a *ElementTypeName) End() Position {
	if a.Position != nil {
		return deltaPos(*a.Position, len(a.Name))
	}
	return Position{}
}
func (*ElementTypeName) _node() {}
