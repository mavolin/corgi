package ast

// ============================================================================
// Content Writer
// ======================================================================================

// ContentWriter is a node that can produce content, i.e. elements or text:
// [ElementWriter], [Block], [CharacterReference], [Doctype],
// [CharacterEscape], [ExpressionInterpolation], [Text], [ComponentCall], or as
// a special case the [BlockSetterContentWriter] wrapper type.
type ContentWriter interface {
	Node
	_contentWriter()
}

// if changed, change the comment above
var (
	_ ContentWriter = (*Element)(nil)
	_ ContentWriter = (*Doctype)(nil)
	_ ContentWriter = (*Block)(nil)
	_ ContentWriter = (*Text)(nil)
	_ ContentWriter = (*CharacterReference)(nil)
	_ ContentWriter = (*CharacterEscape)(nil)
	_ ContentWriter = (*ExpressionInterpolation)(nil)
	_ ContentWriter = (*ComponentCall)(nil)
	_ ContentWriter = (*BlockSetterContentWriter)(nil)
)

// ============================== Component Content Writer ==============================

// BlockSetterContentWriter is a special case of a ContentWriter.
//
// It captures a [ComponentCall] with a block setter, that cannot write
// attributes.
type BlockSetterContentWriter struct {
	ComponentCall *ComponentCall
	// BlockSetter is the [BlockSetter] preventing the component call from
	// writing attributes.
	//
	// Refer to the block setter's block's CannotForwardAttributes field for
	// the reason why.
	BlockSetter *BlockSetter
}

var _ ContentWriter = (*BlockSetterContentWriter)(nil)

func (w *BlockSetterContentWriter) Start() Position   { return w.ComponentCall.Start() }
func (w *BlockSetterContentWriter) End() Position     { return w.ComponentCall.End() }
func (w *BlockSetterContentWriter) Walk(f func(Node)) { w.ComponentCall.Walk(f) }
func (w *BlockSetterContentWriter) _node()            {}
func (w *BlockSetterContentWriter) _contentWriter()   {}

// ============================================================================
// Element Writer
// ======================================================================================

// ElementWriter is a ContentWriter that can produce elements:
// [Element], [ComponentCall], or as a special case the
// [BlockSetterElementWriter] wrapper type.
type ElementWriter interface {
	Node
	_elementWriter()
}

// if changed, change the comment above
var (
	_ ElementWriter = (*Element)(nil)
	_ ElementWriter = (*ComponentCall)(nil)
	_ ElementWriter = (*BlockSetterElementWriter)(nil)
)

// ============================== Component Element Writer ==============================

// BlockSetterElementWriter is a special case of an ElementWriter.
//
// It captures a [ComponentCall] with a block setter, that cannot is not
// forwarded, i.e. it is placed inside an element.
type BlockSetterElementWriter struct {
	ComponentCall *ComponentCall
	// BlockSetter is the [BlockSetter] preventing the component call from
	// being forwarded.
	//
	// Refer to the block setter's block's NotForwarded field for
	// the reason why.
	BlockSetter BlockSetter
}

var _ ElementWriter = (*BlockSetterElementWriter)(nil)

func (w *BlockSetterElementWriter) Start() Position   { return w.ComponentCall.Start() }
func (w *BlockSetterElementWriter) End() Position     { return w.ComponentCall.End() }
func (w *BlockSetterElementWriter) Walk(f func(Node)) { w.ComponentCall.Walk(f) }
func (w *BlockSetterElementWriter) _node()            {}
func (w *BlockSetterElementWriter) _elementWriter()   {}

// ============================================================================
// Attribute Writer
// ======================================================================================

// AttributeWriter is a node that can produce attributes:
// [ClassShorthand], [IDShorthand], [NamedAttribute], [ComponentCall], or as a
// special case the [ComponentAttributeWriter] wrapper type.
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
// [AndPlaceholder], [ComponentCallAndPlaceholderWriter].
//
// This is either an [AndPlaceholder] directly, or a
// [ComponentCallAndPlaceholderWriter] describing a component call that
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
	_ AndPlaceholderWriter = (*ComponentCallAndPlaceholderWriter)(nil)
)

// ======================= Component Call And Placeholder Writer ========================

// ComponentCallAndPlaceholderWriter is a special case of an
// AndPlaceholderWriter.
//
// It captures a [ComponentCall] that receives the &-placeholder, and the
// AndPlaceholderWriter inside the call's body that writes the &-placeholder.
type ComponentCallAndPlaceholderWriter struct {
	ComponentCall *ComponentCall
	// BlockSetter is the [BlockSetter] that contains the writer.
	//
	// Might be nil, in which case the &-placeholder writer is directly passed
	// to the &-placeholder of the component that is being called.
	//
	// If set, the block setter is forwarded.
	BlockSetter BlockSetter
	// Writer is the AndPlaceholderWriter in the call's (not the component's!)
	// body that writes the &-placeholder.
	//
	// The &-placeholder is a forwarded &-placeholder.
	Writer AndPlaceholderWriter
}

var _ AndPlaceholderWriter = (*ComponentCallAndPlaceholderWriter)(nil)

func (w *ComponentCallAndPlaceholderWriter) Start() Position        { return w.ComponentCall.Start() }
func (w *ComponentCallAndPlaceholderWriter) End() Position          { return w.ComponentCall.End() }
func (w *ComponentCallAndPlaceholderWriter) Walk(f func(Node))      { w.ComponentCall.Walk(f) }
func (w *ComponentCallAndPlaceholderWriter) _node()                 {}
func (w *ComponentCallAndPlaceholderWriter) _andPlaceholderWriter() {}
