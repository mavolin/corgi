package ast

import (
	"regexp"
	"slices"

	"github.com/mavolin/corgi/v2/escape/attrtype"
)

type AttributeDefinition struct {
	Attr   *Position
	Prefix *AttributeName // nil if no prefix
	LParen *Position      // nil if not a list
	Specs  []*AttributeSpec
	RParen *Position // nil if not a list
}

var (
	_ TopLevelNode = (*AttributeDefinition)(nil)
	_ Highlighter  = (*AttributeDefinition)(nil)
)

func (d *AttributeDefinition) Start() Position {
	if d.Attr != nil {
		return *d.Attr
	}
	if d.Prefix != nil {
		if start := d.Prefix.Start(); start != NoPosition {
			return start
		}
	}
	if d.LParen != nil {
		return *d.LParen
	}
	for _, s := range d.Specs {
		if s != nil {
			if start := s.Start(); start != NoPosition {
				return start
			}
		}
	}
	if d.RParen != nil {
		return *d.RParen
	}
	return NoPosition
}

func (d *AttributeDefinition) End() Position {
	if d.RParen != nil {
		return deltaPos(*d.RParen, len(")"))
	}
	for _, s := range slices.Backward(d.Specs) {
		if s != nil {
			if end := s.End(); end != NoPosition {
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
	if d.Attr != nil {
		return deltaPos(*d.Attr, len("attr"))
	}
	return NoPosition
}

func (d *AttributeDefinition) Highlight() (start, end Position) {
	start, end = d.Start(), d.End()
	if d.LParen == nil && start.Line >= end.Line-3 {
		return start, end
	}
	if d.Attr != nil {
		return *d.Attr, deltaPos(*d.Attr, len("attr"))
	}
	return start, end
}

func (d *AttributeDefinition) Walk(w func(Node)) {
	if d.Prefix != nil {
		w(d.Prefix)
	}
	for _, s := range d.Specs {
		if s != nil {
			w(s)
		}
	}
}

func (*AttributeDefinition) _node()         {}
func (*AttributeDefinition) _topLevelNode() {}

// ============================================================================
// Attribute Rule
// ======================================================================================

type AttributeSpec struct {
	Selector AttributeSelector
	Ruleset  *AttributeRuleset
}

var _ Node = (*AttributeSpec)(nil)

func (a *AttributeSpec) Start() Position {
	if a.Selector != nil {
		if start := a.Selector.Start(); start != NoPosition {
			return start
		}
	} else if a.Ruleset != nil {
		if start := a.Ruleset.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (a *AttributeSpec) End() Position {
	if a.Ruleset != nil {
		if end := a.Ruleset.End(); end != NoPosition {
			return end
		}
	} else if a.Selector != nil {
		if end := a.Selector.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (a *AttributeSpec) Walk(w func(Node)) {
	if a.Selector != nil {
		w(a.Selector)
	}
	if a.Ruleset != nil {
		w(a.Ruleset)
	}
}

func (*AttributeSpec) _node() {}

// ============================================================================
// Attribute Ruleset
// ======================================================================================

type AttributeRuleset struct {
	LBrace *Position
	List   []*AttributeRule
	RBrace *Position
}

var _ Node = (*AttributeRuleset)(nil)

func (a *AttributeRuleset) Start() Position {
	if a.LBrace != nil {
		return *a.LBrace
	}
	for _, r := range a.List {
		if r != nil {
			if start := r.Start(); start != NoPosition {
				return start
			}
		}
	}
	if a.RBrace != nil {
		return *a.RBrace
	}
	return NoPosition
}

func (a *AttributeRuleset) End() Position {
	if a.RBrace != nil {
		return deltaPos(*a.RBrace, len("}"))
	}
	for _, r := range slices.Backward(a.List) {
		if r != nil {
			if end := r.End(); end != NoPosition {
				return end
			}
		}
	}
	if a.LBrace != nil {
		return deltaPos(*a.LBrace, len("{"))
	}
	return NoPosition
}

func (a *AttributeRuleset) Walk(w func(Node)) {
	for _, r := range a.List {
		if r != nil {
			w(r)
		}
	}
}

func (*AttributeRuleset) _node() {}

// ============================================================================
// Attribute Ruleset Rule
// ======================================================================================

type AttributeRule struct {
	Selector ElementSelector
	Type     *AttributeTypeName
}

var _ Node = (*AttributeRule)(nil)

func (a *AttributeRule) Start() Position {
	if a.Selector != nil {
		if start := a.Selector.Start(); start != NoPosition {
			return start
		}
	} else if a.Type != nil {
		if start := a.Type.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (a *AttributeRule) End() Position {
	if a.Type != nil {
		if end := a.Type.End(); end != NoPosition {
			return end
		}
	} else if a.Selector != nil {
		if end := a.Selector.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (a *AttributeRule) Walk(w func(Node)) {
	if a.Selector != nil {
		w(a.Selector)
	}
	if a.Type != nil {
		w(a.Type)
	}
}

func (*AttributeRule) _node() {}

// ============================================================================
// Attribute Selector
// ======================================================================================

type AttributeSelector interface {
	Node
	// Matches returns true if the given attribute name, in canonical form,
	// matches this selector.
	Matches(canonicalName string) bool
	_attributeSelector()
}

// ============================== Basic Attribute Matcher ===============================

type BasicAttributeSelector struct {
	Name          string
	CanonicalName string // ascii-lowercase version of Name
	Wildcard      bool   // optional
	Position      *Position
}

var _ AttributeSelector = (*BasicAttributeSelector)(nil)

func (b *BasicAttributeSelector) Start() Position {
	if b.Position != nil {
		return *b.Position
	}
	return NoPosition
}

func (b *BasicAttributeSelector) End() Position {
	if b.Position == nil {
		return NoPosition
	}

	if b.Wildcard {
		return deltaPos(*b.Position, len([]rune(b.Name))+len("*"))
	}
	return deltaPos(*b.Position, len([]rune(b.Name)))
}
func (b *BasicAttributeSelector) Walk(func(Node)) {}

func (b *BasicAttributeSelector) Matches(canonicalName string) bool {
	if !b.Wildcard {
		return canonicalName == b.CanonicalName
	}
	// at least one rune longer than b.Name
	return len(canonicalName) > len(b.CanonicalName) && canonicalName[:len(b.CanonicalName)] == b.CanonicalName
}

func (b *BasicAttributeSelector) _node()              {}
func (b *BasicAttributeSelector) _attributeSelector() {}

// ============================== Regexp Attribute Matcher ==============================

type RegexpAttributeSelector struct {
	Regexp *Position
	LParen *Position
	Raw    *StaticString
	// Compiled is the compiled version of Raw, with start and end anchors
	// added to ensure full-string matches.
	Compiled *regexp.Regexp
	RParen   *Position
}

var _ AttributeSelector = (*RegexpAttributeSelector)(nil)

func (r *RegexpAttributeSelector) Start() Position {
	switch {
	case r.Regexp != nil:
		return *r.Regexp
	case r.LParen != nil:
		return *r.LParen
	}
	if r.Raw != nil {
		if start := r.Raw.Start(); start != NoPosition {
			return start
		}
	}
	if r.RParen != nil {
		return *r.RParen
	}
	return NoPosition
}

func (r *RegexpAttributeSelector) End() Position {
	if r.RParen != nil {
		return deltaPos(*r.RParen, len(")"))
	}
	if r.Raw != nil {
		if end := r.Raw.End(); end != NoPosition {
			return end
		}
	}
	switch {
	case r.LParen != nil:
		return deltaPos(*r.LParen, len("("))
	case r.Regexp != nil:
		return deltaPos(*r.Regexp, len("regexp"))
	}
	return NoPosition
}

func (r *RegexpAttributeSelector) Walk(w func(Node)) {
	if r.Raw != nil {
		w(r.Raw)
	}
}

func (r *RegexpAttributeSelector) Matches(canonicalName string) bool {
	return r.Compiled.MatchString(canonicalName)
}

func (r *RegexpAttributeSelector) _node()              {}
func (r *RegexpAttributeSelector) _attributeSelector() {}

// ============================================================================
// Element Selector
// ======================================================================================

type ElementSelector interface {
	Node
	_elementSelector()
}

// ============================= Wildcard Element Selector ==============================

type WildcardElementSelector struct {
	Asterisk *Position
}

var _ ElementSelector = (*WildcardElementSelector)(nil)

func (w *WildcardElementSelector) Start() Position {
	if w.Asterisk != nil {
		return *w.Asterisk
	}
	return NoPosition
}

func (w *WildcardElementSelector) End() Position {
	if w.Asterisk != nil {
		return deltaPos(*w.Asterisk, len("*"))
	}
	return NoPosition
}

func (w *WildcardElementSelector) Walk(func(Node)) {}

func (w *WildcardElementSelector) _node()            {}
func (w *WildcardElementSelector) _elementSelector() {}

// =============================== Element List Selector ================================

type ListElementSelector struct {
	List []*ElementReference
}

var _ ElementSelector = (*ListElementSelector)(nil)

func (l *ListElementSelector) Start() Position {
	for _, e := range l.List {
		if e != nil {
			if start := e.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (l *ListElementSelector) End() Position {
	for _, e := range slices.Backward(l.List) {
		if e != nil {
			if end := e.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (l *ListElementSelector) Walk(w func(Node)) {
	for _, e := range l.List {
		if e != nil {
			w(e)
		}
	}
}

func (l *ListElementSelector) _node()            {}
func (l *ListElementSelector) _elementSelector() {}

// ============================================================================
// Attribute Type Name
// ======================================================================================

type AttributeTypeName struct {
	Name     string
	Type     attrtype.Type
	Position *Position
}

var _ Node = (*AttributeTypeName)(nil)

func (a *AttributeTypeName) Start() Position {
	if a.Position != nil {
		return *a.Position
	}
	return NoPosition
}

func (a *AttributeTypeName) End() Position {
	if a.Position != nil {
		return deltaPos(*a.Position, len(a.Name))
	}
	return NoPosition
}
func (a *AttributeTypeName) Walk(func(Node)) {}

func (*AttributeTypeName) _node() {}
