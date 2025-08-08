package ast

// ============================================================================
// Content Writer
// ======================================================================================

// ContentWriter is a node that can produce content, i.e. elements or text:
// [ElementWriter], [Block], [CharacterReference], [ComponentCall], [Doctype],
// [EscapedHash], [EscapedRBracket], [ExpressionInterpolation],
// [HashSpace], [Text].
//
// Additionally, a [BlockSetter] may also be used as a ContentWriter, not
// because it produces content, but because it holds the ability to inhibit
// the writing of attributes, if it's associated block cannot forward
// attributes.
type ContentWriter interface {
	Node
	_contentWriter()
}

// if changed, change the comment above
var (
	_ ContentWriter = (ElementWriter)(nil)
	_ ContentWriter = (*Doctype)(nil)
	_ ContentWriter = (*Block)(nil)
	_ ContentWriter = (BlockSetter)(nil)
	_ ContentWriter = (*ComponentCall)(nil)
	_ ContentWriter = (*Text)(nil)
	_ ContentWriter = (*CharacterReference)(nil)
	_ ContentWriter = (*EscapedHash)(nil)
	_ ContentWriter = (*EscapedRBracket)(nil)
	_ ContentWriter = (*ExpressionInterpolation)(nil)
	_ ContentWriter = (*HashSpace)(nil)
)

// ============================================================================
// Element Writer
// ======================================================================================

// ElementWriter is a ContentWriter that can produce elements:
// [Element], [ComponentCall].
type ElementWriter interface {
	ContentWriter
	_elementWriter()
}

// if changed, change the comment above
var (
	_ ElementWriter = (*Element)(nil)
	_ ElementWriter = (*ComponentCall)(nil)
)

// ============================================================================
// Attribute Writer
// ======================================================================================

// AttributeWriter is a node that can produce attributes:
// [ClassShorthand], [ComponentCall], [IDShorthand], [NamedAttribute].
type AttributeWriter interface {
	Node
	_attributeWriter()
}

// if changed, change the comment above
var (
	_ AttributeWriter = (*ClassShorthand)(nil)
	_ AttributeWriter = (*ComponentCall)(nil)
	_ AttributeWriter = (*IDShorthand)(nil)
	_ AttributeWriter = (*NamedAttribute)(nil)
)

// ============================================================================
// And Placeholder Writer
// ======================================================================================

// AndPlaceholderWriter is a node that can produce attributes if the
// &-placeholder is set:
// [AndPlaceholder], [ComponentCall].
//
// This is either an [AndPlaceholder] directly, or a [ComponentCall] that
// forwards the &-placeholder it is called with to the element containing the
// component call again:
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
