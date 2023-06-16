package file

// ============================================================================
// Doctype
// ======================================================================================

// Doctype represents a doctype directive (`doctype html`)
type Doctype struct {
	Position
}

var _ ScopeItem = Doctype{}

func (Doctype) _typeScopeItem() {}

// ============================================================================
// CorgiComment
// ======================================================================================

// HTMLComment represents a comment.
type HTMLComment struct {
	Lines []HTMLCommentLine
	Position
}

type HTMLCommentLine struct {
	Comment string
	Position
}

var _ ScopeItem = HTMLComment{}

func (HTMLComment) _typeScopeItem() {}

// ============================================================================
// Element
// ======================================================================================

// Element represents a single HTML element.
type Element struct {
	Name       string
	Attributes []AttributeCollection
	Void       bool

	Body Scope

	Position
}

var _ ScopeItem = Element{}

func (Element) _typeScopeItem() {}

type DivShorthand struct {
	// Attributes are the attributes of the element.
	//
	// It contains at least one item.
	//
	// Attributes[0] will be either of type ClassShorthand or IDShorthand.
	Attributes []AttributeCollection

	Body Scope

	Position
}

var _ ScopeItem = DivShorthand{}

func (DivShorthand) _typeScopeItem() {}

// ============================================================================
// And
// ======================================================================================

// And represents an '&' expression.
type And struct {
	Attributes []AttributeCollection

	Position
}

func (And) _typeScopeItem() {}

// ============================================================================
// AttributeCollection
// ======================================================================================

type AttributeCollection interface {
	_typeAttributeCollection()
	Poser
}

// ==================================== IDShorthand =====================================

type IDShorthand struct {
	ID string
	Position
}

var _ AttributeCollection = IDShorthand{}

func (IDShorthand) _typeAttributeCollection() {}

// =================================== ClassShorthand ===================================

type ClassShorthand struct {
	Name string
	Position
}

var _ AttributeCollection = ClassShorthand{}

func (ClassShorthand) _typeAttributeCollection() {}

// =================================== AttributeList ====================================

type AttributeList struct {
	LParenPos  Position
	Attributes []Attribute
	RParenPos  Position
}

var _ AttributeCollection = AttributeList{}

func (AttributeList) _typeAttributeCollection() {}
func (l AttributeList) Pos() Position           { return l.LParenPos.Pos() }

// ============================================================================
// Attribute
// ======================================================================================

type Attribute interface {
	_typeAttribute()
	Poser
}

// ================================== SimpleAttribute ===================================

type SimpleAttribute struct {
	Name string

	NoEscape  bool
	AssignPos *Position // pos of '=' or '!='; nil for boolean attributes

	Value *Expression // nil for boolean attributes

	Position
}

var _ Attribute = SimpleAttribute{}

func (SimpleAttribute) _typeAttribute() {}

// =================================== AndPlaceholder ===================================

// AndPlaceholder is an attribute 'named' `&` that is used as a placeholder for
// the attributes attached to a mixin call.
//
//	mixin foo()
//		div: span(&) foo
//
// It may only be used inside a mixin definition.
type AndPlaceholder struct {
	Position
}

var _ Attribute = AndPlaceholder{}

func (AndPlaceholder) _typeAttribute() {}

// ================================ Mixin Call Attribute ================================

type MixinCallAttribute struct {
	Name string

	NoEscape  bool
	AssignPos Position

	MixinCall MixinCall
	Value     InterpolationValue

	Position
}

var _ Attribute = MixinCallAttribute{}

func (MixinCallAttribute) _typeAttribute() {}
