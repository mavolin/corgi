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

var (
	_ TopLevelNode = (*ElementDefinition)(nil)
	_ Highlighter  = (*ElementDefinition)(nil)
)

func (d *ElementDefinition) Start() Position {
	if d.Elem != nil {
		return *d.Elem
	}
	if d.Prefix != nil {
		if start := d.Prefix.Start(); start != NoPosition {
			return start
		}
	}
	if d.LParen != nil {
		return *d.LParen
	}
	for _, spec := range d.Specs {
		if spec != nil {
			if start := spec.Start(); start != NoPosition {
				return start
			}
		}
	}
	if d.RParen != nil {
		return *d.RParen
	}
	return NoPosition
}

func (d *ElementDefinition) End() Position {
	if d.RParen != nil {
		return deltaPos(*d.RParen, len(")"))
	}
	for _, spec := range slices.Backward(d.Specs) {
		if spec != nil {
			if end := spec.End(); end != NoPosition {
				return end
			}
		}
	}
	if d.LParen != nil {
		return deltaPos(*d.LParen, len("("))
	}
	if d.Prefix != nil {
		if end := d.Prefix.End(); end != NoPosition {
			return end
		}
	}
	if d.Elem != nil {
		return deltaPos(*d.Elem, len("element"))
	}
	return NoPosition
}

func (d *ElementDefinition) Walk(w func(Node)) {
	if d.Prefix != nil {
		w(d.Prefix)
	}
	for _, spec := range d.Specs {
		if spec != nil {
			w(spec)
		}
	}
}

func (d *ElementDefinition) Highlight() (start, end Position) {
	start, end = d.Start(), d.End()
	if d.LParen == nil && start.Line >= end.Line-3 {
		return start, end
	}
	if d.Elem != nil {
		return *d.Elem, deltaPos(*d.Elem, len("elem"))
	}
	return start, end
}

func (*ElementDefinition) _node()         {}
func (*ElementDefinition) _topLevelNode() {}

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
		if start := a.Name.Start(); start != NoPosition {
			return start
		}
	}
	if a.Type != nil {
		if start := a.Type.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (a *ElementSpec) End() Position {
	if a.Type != nil {
		if end := a.Type.End(); end != NoPosition {
			return end
		}
	}
	if a.Name != nil {
		if end := a.Name.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (a *ElementSpec) Walk(w func(Node)) {
	if a.Name != nil {
		w(a.Name)
	}
	if a.Type != nil {
		w(a.Type)
	}
}

func (*ElementSpec) _node() {}

// ============================================================================
// Element Type
// ======================================================================================

// ElementType is either a [BasicElementType] or an [AliasElementType].
type ElementType interface {
	Node
	_elementType()
}

// if this is changed, change the comment above
var (
	_ ElementType = (*BasicElementType)(nil)
	_ ElementType = (*AliasElementType)(nil)
)

// ============================================================================
// Basic Element Type
// ======================================================================================

type BasicElementType struct {
	Type *ElementTypeName
}

var _ ElementType = (*BasicElementType)(nil)

func (t *BasicElementType) Start() Position {
	if t.Type != nil {
		if start := t.Type.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (t *BasicElementType) End() Position {
	if t.Type != nil {
		if end := t.Type.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (t *BasicElementType) Walk(w func(Node)) {
	if t.Type != nil {
		w(t.Type)
	}
}

func (t *BasicElementType) _node()        {}
func (t *BasicElementType) _elementType() {}

// ============================================================================
// Alias Element Type
// ======================================================================================

type AliasElementType struct {
	EqualSign *Position
	Name      *ElementReference
}

var _ ElementType = (*AliasElementType)(nil)

func (t *AliasElementType) Start() Position {
	if t.EqualSign != nil {
		return *t.EqualSign
	}
	if t.Name != nil {
		if start := t.Name.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (t *AliasElementType) End() Position {
	if t.Name != nil {
		if end := t.Name.End(); end != NoPosition {
			return end
		}
	}
	if t.EqualSign != nil {
		return deltaPos(*t.EqualSign, len("="))
	}
	return NoPosition
}

func (t *AliasElementType) Walk(w func(Node)) {
	if t.Name != nil {
		w(t.Name)
	}
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
	return NoPosition
}

func (a *ElementTypeName) End() Position {
	if a.Position != nil {
		return deltaPos(*a.Position, len([]rune(a.Name)))
	}
	return NoPosition
}
func (a *ElementTypeName) Walk(func(Node)) {}

func (*ElementTypeName) _node() {}
