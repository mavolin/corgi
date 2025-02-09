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

var _ ScopeNode = (*AttributeDefinition)(nil)

func (a *AttributeDefinition) Start() Position {
	if a.Attr != nil {
		return *a.Attr
	} else if a.Prefix != nil {
		return a.Prefix.Start()
	} else if a.LParen != nil {
		return *a.LParen
	}
	for _, s := range a.Specs {
		if s != nil {
			return s.Start()
		}
	}
	if a.RParen != nil {
		return *a.RParen
	}
	return Position{}
}
func (a *AttributeDefinition) End() Position {
	if a.RParen != nil {
		return deltaPos(*a.RParen, len(")"))
	} else if len(a.Specs) > 0 {
		return a.Specs[len(a.Specs)-1].End()
	} else if a.LParen != nil {
		return deltaPos(*a.LParen, len("("))
	} else if a.Prefix != nil {
		return a.Prefix.End()
	} else if a.Attr != nil {
		return deltaPos(*a.Attr, len("attr"))
	}
	return Position{}
}

func (*AttributeDefinition) _node()      {}
func (*AttributeDefinition) _scopeNode() {}

// ============================================================================
// Attribute Rule
// ======================================================================================

type AttributeSpec struct {
	Name    AttributeSelector
	Ruleset *AttributeRuleset
}

var _ Node = (*AttributeSpec)(nil)

func (a *AttributeSpec) Start() Position {
	if a.Name != nil {
		return a.Name.Start()
	} else if a.Ruleset != nil {
		return a.Ruleset.Start()
	}
	return Position{}
}
func (a *AttributeSpec) End() Position {
	if a.Ruleset != nil {
		return a.Ruleset.End()
	} else if a.Name != nil {
		return a.Name.End()
	}
	return Position{}
}

func (*AttributeSpec) _node() {}

// ============================================================================
// Attribute Ruleset
// ======================================================================================

type AttributeRuleset struct {
	LBrace *Position
	Rules  []*AttributeRule
	RBrace *Position
}

var _ Node = (*AttributeRuleset)(nil)

func (a *AttributeRuleset) Start() Position {
	if a.LBrace != nil {
		return *a.LBrace
	}
	for _, r := range a.Rules {
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
	for _, r := range slices.Backward(a.Rules) {
		if r != nil {
			return r.End()
		}
	}
	if a.LBrace != nil {
		return deltaPos(*a.LBrace, len("{"))
	}
	return Position{}
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

func (*AttributeRule) _node() {}

// ============================================================================
// Name Matcher
// ======================================================================================

type AttributeSelector interface {
	Node
	Matches(s string) bool
	_nameMatcher()
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

func (b *BasicAttributeSelector) Matches(s string) bool {
	if !b.Wildcard {
		return b.Name == s
	}
	// at least one rune longer than b.Name
	return len(s) > len(b.Name) && s[:len(b.Name)] == b.Name
}

func (b *BasicAttributeSelector) _node()        {}
func (b *BasicAttributeSelector) _nameMatcher() {}

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
	if r.Regexp != nil {
		return *r.Regexp
	} else if r.LParen != nil {
		return *r.LParen
	} else if r.Raw != nil {
		return r.Raw.Start()
	} else if r.RParen != nil {
		return *r.RParen
	}
	return Position{}
}
func (r *RegexpAttributeSelector) End() Position {
	if r.RParen != nil {
		return deltaPos(*r.RParen, len(")"))
	} else if r.Raw != nil {
		return r.Raw.End()
	} else if r.LParen != nil {
		return deltaPos(*r.LParen, len("("))
	} else if r.Regexp != nil {
		return deltaPos(*r.Regexp, len("regexp"))
	}
	return Position{}
}

func (r *RegexpAttributeSelector) Matches(s string) bool {
	return r.Compiled.MatchString(s)
}

func (r *RegexpAttributeSelector) _node()        {}
func (r *RegexpAttributeSelector) _nameMatcher() {}

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
func (w *WildcardElementSelector) _node()            {}
func (w *WildcardElementSelector) _elementSelector() {}

// =============================== Element List Selector ================================

type (
	ListElementSelector struct {
		Elements []*ListElementSelectorItem
	}
	ListElementSelectorItem struct {
		Name     string
		Position *Position
	}
)

var (
	_ ElementSelector = (*ListElementSelector)(nil)
	_ Node            = (*ListElementSelectorItem)(nil)
)

func (l *ListElementSelector) Start() Position {
	for _, e := range l.Elements {
		if e != nil {
			return e.Start()
		}
	}
	return Position{}
}
func (l *ListElementSelector) End() Position {
	for _, e := range slices.Backward(l.Elements) {
		if e != nil {
			return e.End()
		}
	}
	return Position{}
}

func (l *ListElementSelector) Matches(s string) bool {
	for _, e := range l.Elements {
		if e.Name == s {
			return true
		}
	}
	return false
}

func (l *ListElementSelector) _node()            {}
func (l *ListElementSelector) _elementSelector() {}

func (l *ListElementSelectorItem) Start() Position {
	if l.Position != nil {
		return *l.Position
	}
	return Position{}
}
func (l *ListElementSelectorItem) End() Position {
	if l.Position != nil {
		return deltaPos(*l.Position, len(l.Name))
	}
	return Position{}
}
func (l *ListElementSelectorItem) _node() {}

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
func (*AttributeTypeName) _node() {}
