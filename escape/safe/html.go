package safe

type (
	// HTML represents a known safe HTML document fragment that can safely be
	// placed at the root of the document or in the body of another element.
	//
	// It must not contain unclosed tags or comments and must escape at least
	// [<&].
	HTML struct{ val string }

	// PlainAttr represents a known safe HTML attribute fragment that can
	// safely be placed between double quotes and used as a plain attribute.
	//
	// It must at least escape ["&].
	//
	// A plain attribute may not be used as part of an attribute with specific
	// escaping requirements, such as href or style.
	PlainAttr struct{ val string }
)

// TrustedHTML creates a new HTML fragment from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an HTML.
func TrustedHTML(s string) HTML { return HTML{val: s} }

// TrustedPlainAttr creates a new PlainAttr fragment from the given trusted
// string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a PlainAttr.
func TrustedPlainAttr(s string) PlainAttr { return PlainAttr{val: s} }

func (h HTML) Escaped() string      { return h.val }
func (a PlainAttr) Escaped() string { return a.val }
