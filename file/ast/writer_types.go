package ast

// ============================================================================
// Content Writer
// ======================================================================================

// ContentWriter is a node that can produce content, i.e. elements or text:
// [ElementWriter], [RawElement], [Block], [CharacterReference], [Doctype],
// [CharacterEscape], [ExpressionInterpolation], [Text], [ComponentCall].
type ContentWriter interface {
	Node
	AttributeInhibitor
	_contentWriter()
}

// if changed, change the comment above
var (
	_ ContentWriter = (*Element)(nil)
	_ ContentWriter = (*Doctype)(nil)
	_ ContentWriter = (*RawElement)(nil)
	_ ContentWriter = (*Block)(nil)
	_ ContentWriter = (*Text)(nil)
	_ ContentWriter = (*CharacterReference)(nil)
	_ ContentWriter = (*CharacterEscape)(nil)
	_ ContentWriter = (*ExpressionInterpolation)(nil)
	_ ContentWriter = (*ComponentCall)(nil)
)

// ============================================================================
// Attribute Inhibitor
// ======================================================================================

// AttributeInhibitor is a node that prevents adding to the attribute of an
// element.
// It includes all content writers:
// [Element], [Doctype], [RawElement], [Block], [Text], [CharacterReference],
// [CharacterEscape], [ExpressionInterpolation], [ComponentCall] or a
// [BlockSetterAttributeInhibitor].
type AttributeInhibitor interface {
	Node
	_attributeInhibitor()
}

// if changed, change the comment above
var (
	_ AttributeInhibitor = ContentWriter(nil)
	_ AttributeInhibitor = (*BlockSetterAttributeInhibitor)(nil)
)

// ========================== Block Setter Attribute Inhibitor ==========================

// BlockSetterAttributeInhibitor is a block setter for a block that does not
// allow its setter to forward attributes.
type BlockSetterAttributeInhibitor struct {
	ComponentCall *ComponentCall
	// BlockSetter is the [BlockSetter] preventing the component call from
	// writing attributes.
	//
	// Refer to the block setter's block's CannotForwardAttributes field for
	// the reason why.
	BlockSetter BlockSetter
}

var _ AttributeInhibitor = (*BlockSetterAttributeInhibitor)(nil)

func (w *BlockSetterAttributeInhibitor) Start() Position    { return w.ComponentCall.Start() }
func (w *BlockSetterAttributeInhibitor) End() Position      { return w.ComponentCall.End() }
func (w *BlockSetterAttributeInhibitor) Walk(f func(Node))  { w.ComponentCall.Walk(f) }
func (*BlockSetterAttributeInhibitor) _node()               {}
func (*BlockSetterAttributeInhibitor) _attributeInhibitor() {}

// ============================================================================
// Element Writer
// ======================================================================================

// ElementWriter is a ContentWriter that can produce elements:
// [Element], [Doctype], [ComponentCall].
type ElementWriter interface {
	Node
	_elementWriter()
}

// if changed, change the comment above
var (
	_ ElementWriter = (*Element)(nil)
	_ ElementWriter = (*Doctype)(nil)
	_ ElementWriter = (*ComponentCall)(nil)
)

// ============================================================================
// Containing Element
// ======================================================================================

// ContainingElement describes an element or something producing an element
// containing a node:
// [Element], [BlockSetterContainingElement], or
// [AndPlaceholderContainingElement].
type ContainingElement interface {
	Node
	_containingElement()
}

// if changed, change the comment above
var (
	_ ContainingElement = (*Element)(nil)
	_ ContainingElement = (*BlockSetterContainingElement)(nil)
	_ ContainingElement = (*AndPlaceholderContainingElement)(nil)
)

// ========================== Block Setter Containing Element ===========================

// BlockSetterContainingElement is a block setter for a block where at least
// one instance of that block is contained in an element.
type BlockSetterContainingElement struct {
	ComponentCall *ComponentCall
	BlockSetter   BlockSetter
}

var _ ContainingElement = (*BlockSetterContainingElement)(nil)

func (w *BlockSetterContainingElement) Start() Position   { return w.BlockSetter.Start() }
func (w *BlockSetterContainingElement) End() Position     { return w.BlockSetter.End() }
func (w *BlockSetterContainingElement) Walk(f func(Node)) { w.BlockSetter.Walk(f) }
func (*BlockSetterContainingElement) _node()              {}
func (*BlockSetterContainingElement) _containingElement() {}

// ========================= And Placeholder Containing Element =========================

// AndPlaceholderContainingElement is a component call that accepts attributes.
type AndPlaceholderContainingElement ComponentCall

var _ ContainingElement = (*AndPlaceholderContainingElement)(nil)

func (w *AndPlaceholderContainingElement) Start() Position   { return (*ComponentCall)(w).Start() }
func (w *AndPlaceholderContainingElement) End() Position     { return (*ComponentCall)(w).End() }
func (w *AndPlaceholderContainingElement) Walk(f func(Node)) { (*ComponentCall)(w).Walk(f) }
func (*AndPlaceholderContainingElement) _node()              {}
func (*AndPlaceholderContainingElement) _containingElement() {}

// ============================================================================
// Attribute Writer
// ======================================================================================

// AttributeWriter is a node that can produce attributes:
// [ClassShorthand], [IDShorthand], [NamedAttribute], [ComponentCall].
type AttributeWriter interface {
	Node
	_attributeWriter()
}

// if changed, change the comment above
var (
	_ AttributeWriter = (*ClassShorthand)(nil)
	_ AttributeWriter = (*IDShorthand)(nil)
	_ AttributeWriter = (*NamedAttribute)(nil)
	_ AttributeWriter = (*ComponentCall)(nil)
)

// ============================================================================
// And Placeholder Writer
// ======================================================================================

// AndPlaceholderWriter is a node that can produce attributes if the
// &-placeholder is set:
// [AndPlaceholder], [ComponentCall].
//
// This is either an [AndPlaceholder] directly, or a [ComponentCall] that
// accepts and forwards the &-placeholder.
//
//	comp Foo() {
//	  &(&)
//	}
//	:Foo(&)
type AndPlaceholderWriter interface {
	Node
	_andPlaceholderWriter()
}

// if changed, change the comment above
var (
	_ AndPlaceholderWriter = (*AndPlaceholder)(nil)
	_ AndPlaceholderWriter = (*ComponentCall)(nil)
)
