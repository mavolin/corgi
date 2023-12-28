package file

// ============================================================================
// Doctype
// ======================================================================================

// Doctype represents a doctype directive (`doctype(html)`).
type Doctype struct {
	Position
}

func (Doctype) _scopeItem() {}

// ============================================================================
// CorgiComment
// ======================================================================================

// HTMLComment represents a comment that is printed.
type HTMLComment struct {
	Comment string
	Position
}

// ============================================================================
// Element
// ======================================================================================

// Element represents a single HTML element.
type Element struct {
	Name       string
	Attributes []AttributeCollection
	// Void is true, if the element was manually marked as void.
	Void bool

	Body Body

	Position
}

func (Element) _scopeItem() {}

// ============================================================================
// Raw Element
// ======================================================================================

// RawElement represents the special !raw element, which includes all of its
// contents verbatim into the generated HTML.
type RawElement struct {
	Body BracketText
	Position
}

func (RawElement) _scopeItem() {}
