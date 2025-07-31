package ast

import (
	"regexp"
	"slices"
	"strings"

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
	switch {
	case d.Attr != nil:
		return *d.Attr
	case d.Prefix != nil:
		return d.Prefix.Start()
	case d.LParen != nil:
		return *d.LParen
	}
	for _, s := range d.Specs {
		if s != nil {
			return s.Start()
		}
	}
	if d.RParen != nil {
		return *d.RParen
	}
	return Position{}
}

func (d *AttributeDefinition) End() Position {
	switch {
	case d.RParen != nil:
		return deltaPos(*d.RParen, len(")"))
	case len(d.Specs) > 0:
		return d.Specs[len(d.Specs)-1].End()
	case d.LParen != nil:
		return deltaPos(*d.LParen, len("("))
	case d.Prefix != nil:
		return d.Prefix.End()
	case d.Attr != nil:
		return deltaPos(*d.Attr, len("attr"))
	}
	return Position{}
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
		return a.Selector.Start()
	} else if a.Ruleset != nil {
		return a.Ruleset.Start()
	}
	return Position{}
}

func (a *AttributeSpec) End() Position {
	if a.Ruleset != nil {
		return a.Ruleset.End()
	} else if a.Selector != nil {
		return a.Selector.End()
	}
	return Position{}
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
			return r.Start()
		}
	}
	if a.RBrace != nil {
		return *a.RBrace
	}
	return Position{}
}

func (a *AttributeRuleset) End() Position {
	if a.RBrace != nil {
		return deltaPos(*a.RBrace, len("}"))
	}
	for _, r := range slices.Backward(a.List) {
		if r != nil {
			return r.End()
		}
	}
	if a.LBrace != nil {
		return deltaPos(*a.LBrace, len("{"))
	}
	return Position{}
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
		return a.Selector.Start()
	} else if a.Type != nil {
		return a.Type.Start()
	}
	return Position{}
}

func (a *AttributeRule) End() Position {
	if a.Type != nil {
		return a.Type.End()
	} else if a.Selector != nil {
		return a.Selector.End()
	}
	return Position{}
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
	Matches(s string) bool
	_attributeSelector()
}

// ============================== Basic Attribute Matcher ===============================

type BasicAttributeSelector struct {
	Name     string
	Wildcard bool // optional
	Position *Position
}

var _ AttributeSelector = (*BasicAttributeSelector)(nil)

func (b *BasicAttributeSelector) Start() Position {
	if b.Position != nil {
		return *b.Position
	}
	return Position{}
}

func (b *BasicAttributeSelector) End() Position {
	if b.Position == nil {
		return Position{}
	}

	if b.Wildcard {
		return deltaPos(*b.Position, len(b.Name)+len("*"))
	}
	return deltaPos(*b.Position, len(b.Name))
}
func (b *BasicAttributeSelector) Walk(func(Node)) {}

func (b *BasicAttributeSelector) Matches(s string) bool {
	name := strings.ToLower(b.Name)
	s = strings.ToLower(s)
	if !b.Wildcard {
		return name == s
	}
	// at least one rune longer than b.Name
	return len(s) > len(name) && s[:len(name)] == b.Name
}

func (b *BasicAttributeSelector) _node()              {}
func (b *BasicAttributeSelector) _attributeSelector() {}

// ============================== Regexp Attribute Matcher ==============================

type RegexpAttributeSelector struct {
	Regexp   *Position
	LParen   *Position
	Raw      *StaticString
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
	case r.Raw != nil:
		return r.Raw.Start()
	case r.RParen != nil:
		return *r.RParen
	}
	return Position{}
}

func (r *RegexpAttributeSelector) End() Position {
	switch {
	case r.RParen != nil:
		return deltaPos(*r.RParen, len(")"))
	case r.Raw != nil:
		return r.Raw.End()
	case r.LParen != nil:
		return deltaPos(*r.LParen, len("("))
	case r.Regexp != nil:
		return deltaPos(*r.Regexp, len("regexp"))
	}
	return Position{}
}

func (r *RegexpAttributeSelector) Walk(w func(Node)) {
	if r.Raw != nil {
		w(r.Raw)
	}
}

func (r *RegexpAttributeSelector) Matches(s string) bool {
	s = strings.ToLower(s)
	return r.Compiled.MatchString(s)
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
	return Position{}
}

func (w *WildcardElementSelector) End() Position {
	if w.Asterisk != nil {
		return deltaPos(*w.Asterisk, len("*"))
	}
	return Position{}
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
			return e.Start()
		}
	}
	return Position{}
}

func (l *ListElementSelector) End() Position {
	for _, e := range slices.Backward(l.List) {
		if e != nil {
			return e.End()
		}
	}
	return Position{}
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
	return Position{}
}

func (a *AttributeTypeName) End() Position {
	if a.Position != nil {
		return deltaPos(*a.Position, len(a.Name))
	}
	return Position{}
}
func (a *AttributeTypeName) Walk(func(Node)) {}

func (*AttributeTypeName) _node() {}
