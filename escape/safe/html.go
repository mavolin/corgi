package safe

// HTML represents a known safe HTML document fragment that can safely be
// placed at the root of the document or in the body of another element.
//
// It must not contain unclosed tags or unclosed comments and must at least
// escape '<' and '&'.
type HTML struct{ val string }

// TrustedHTML creates a new HTML fragment from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an HTML.
func TrustedHTML(s string) HTML { return HTML{val: s} }

func (h HTML) Get() string { return h.val }
