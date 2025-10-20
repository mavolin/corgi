package safe

import (
	"strings"
	_ "unsafe" // for go:linkname
)

// HTML represents a known safe HTML document fragment that can safely be
// placed at the root of the document or in the body of another element.
//
// It must not contain unclosed tags or unclosed comments and must at least
// escape '<' and '&'.
type HTML struct{ val string }

// ConstantHTML creates a new HTML wrapper from the given string constant.
//
// Before using this function, read the documentation of [HTML].
func ConstantHTML(c constant) HTML {
	if strings.ContainsRune(string(c), '<') {
		panic("HTML must not contain literal '<'")
	}
	return trustedHTML(string(c))
}

//go:linkname trustedHTML
func trustedHTML(s string) HTML { return HTML{val: s} }

func (h HTML) Get() string { return h.val }
