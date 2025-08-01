package ast

// ============================================================================
// Content Writer
// ======================================================================================

// ContentWriter is a node that can produce content, i.e. elements or text:
// [ElementWriter], [CharacterReference], [ComponentCall], [Doctype],
// [EscapedHash], [EscapedRBracket], [ExpressionInterpolation],
// [HashSpace], [Text].
type ContentWriter interface {
	Node
	_contentWriter()
}

// if changed, change the comment above
var (
	_ ContentWriter = (ElementWriter)(nil)
	_ ContentWriter = (*CharacterReference)(nil)
	_ ContentWriter = (*ComponentCall)(nil)
	_ ContentWriter = (*Doctype)(nil)
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
type AndPlaceholderWriter interface {
	Node
	_andPlaceholderWriter()
}

// if changed, change the comment above
var (
	_ AndPlaceholderWriter = (*AndPlaceholder)(nil)
	_ AndPlaceholderWriter = (*ComponentCall)(nil)
)
