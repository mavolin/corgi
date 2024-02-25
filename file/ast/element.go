package ast

// ============================================================================
// Doctype
// ======================================================================================

// Doctype represents a doctype directive (`!doctype(html)`).
type Doctype struct {
	Position Position
}

var _ ScopeNode = (*Doctype)(nil)

func (d *Doctype) Pos() Position { return d.Position }
func (d *Doctype) End() Position { return deltaPos(d.Position, len("!doctype(html)")) }

func (*Doctype) _node()      {}
func (*Doctype) _scopeNode() {}

// ============================================================================
// DevComment
// ======================================================================================

// HTMLComment represents a comment that is printed.
type HTMLComment struct {
	Comment  string
	Position Position
}

var _ ScopeNode = (*HTMLComment)(nil)

func (c *HTMLComment) Pos() Position { return c.Position }
func (c *HTMLComment) End() Position { return deltaPos(c.Position, len("//-")+len(c.Comment)) }

func (*HTMLComment) _node()      {}
func (*HTMLComment) _scopeNode() {}

// ============================================================================
// Element
// ======================================================================================

// Element represents a single HTML element.
type Element struct {
	Name string
	// Void is true, if the element was manually marked as void.
	Void       bool
	Attributes []AttributeCollection
	Body       Body

	Position Position
}

var _ ScopeNode = (*Element)(nil)

func (e *Element) Pos() Position { return e.Position }
func (e *Element) End() Position {
	if e.Body != nil {
		return e.Body.End()
	} else if len(e.Attributes) > 0 {
		return e.Attributes[len(e.Attributes)-1].End()
	}

	if e.Void {
		return deltaPos(e.Position, len(e.Name)+len("/"))
	}
	return deltaPos(e.Position, len(e.Name))
}

func (*Element) _node()      {}
func (*Element) _scopeNode() {}

// ============================================================================
// Raw Element
// ======================================================================================

// RawElement represents the special !raw element, which includes all of its
// contents verbatim into the generated HTML.
type RawElement struct {
	Body     *BracketText // not nil
	Position Position
}

var _ ScopeNode = (*RawElement)(nil)

func (e *RawElement) Pos() Position { return e.Position }
func (e *RawElement) End() Position {
	if e.Body != nil {
		return e.Body.End()
	}
	return deltaPos(e.Position, len("!raw"))
}

func (*RawElement) _node()      {}
func (*RawElement) _scopeNode() {}
