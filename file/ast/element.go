package ast

// ============================================================================
// Doctype
// ======================================================================================

// Doctype represents a doctype directive (`!doctype(html)`).
type Doctype struct {
	Position Position
}

var _ ScopeItem = (*Doctype)(nil)

func (d *Doctype) Pos() Position { return d.Position }
func (*Doctype) _scopeItem()     {}

// ============================================================================
// DevComment
// ======================================================================================

// HTMLComment represents a comment that is printed.
type HTMLComment struct {
	Comment  string
	Position Position
}

var _ ScopeItem = (*HTMLComment)(nil)

func (c *HTMLComment) Pos() Position { return c.Position }
func (*HTMLComment) _scopeItem()     {}

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

	Position Position
}

var _ ScopeItem = (*Element)(nil)

func (e *Element) Pos() Position { return e.Position }
func (*Element) _scopeItem()     {}

// ============================================================================
// Raw Element
// ======================================================================================

// RawElement represents the special !raw element, which includes all of its
// contents verbatim into the generated HTML.
type RawElement struct {
	Body     *BracketText // not nil
	Position Position
}

var _ ScopeItem = (*RawElement)(nil)

func (e *RawElement) Pos() Position { return e.Position }
func (*RawElement) _scopeItem()     {}
