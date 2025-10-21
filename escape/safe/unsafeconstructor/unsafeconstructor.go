// Package unsafeconstructor provides functions to create safe types from
// plain trusted strings.
//
// Usage of these types poses a security risk and should be done with great
// care, as the encapsulated content will be included verbatim in the template.
// Manually creating instances of these types may undermine the security
// guarantees provided by corgi.
// Always ensure that content fully complies with the requirements put forth in
// the documentation of the respective type.
// Content should always come from a trusted source and never from a
// third party like an end-user, or any source that the developer does not have
// full control over.
package unsafeconstructor

import (
	_ "unsafe" // for go:linkname

	"github.com/mavolin/corgi/v2/escape/safe"
	"github.com/mavolin/corgi/v2/internal/escapelite"
)

// TrustedCSSValue creates a new CSSValue wrapper from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for CSSValue.
//
//go:linkname TrustedCSSValue github.com/mavolin/corgi/v2/escape/safe.trustedCSSValue
func TrustedCSSValue(string) safe.CSSValue

// TrustedCSSString creates a css string literal wrapped in a CSSValue.
//
// This function is in unsafeconstrutor, because further contextual escaping/
// validation would be needed to safely assert that the string can be used as
// any part of a CSS value.
// A common example would be a url(...) value, where if the string contains
// a malicious URL and used in a url(...) context, it could lead to an XSS
// vulnerability.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a CSSValue.
func TrustedCSSString(s string) safe.CSSValue {
	return TrustedCSSValue(escapelite.CSSString(s))
}

// TrustedCSSDeclarations creates a new CSSDeclarations wrapper from the given
// trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for CSSDeclarations.
//
//go:linkname TrustedCSSDeclarations github.com/mavolin/corgi/v2/escape/safe.trustedCSSDeclarations
func TrustedCSSDeclarations(string) safe.CSSDeclarations

// TrustedStylesheet creates a new Stylesheet from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Stylesheet.
//
//go:linkname TrustedStylesheet github.com/mavolin/corgi/v2/escape/safe.trustedStylesheet
func TrustedStylesheet(string) safe.Stylesheet

// TrustedHTML creates a new HTML from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an HTML.
//
//go:linkname TrustedHTML github.com/mavolin/corgi/v2/escape/safe.trustedHTML
func TrustedHTML(string) safe.HTML

// TrustedIdentifier creates a new Identifier from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an Identifier.
//
//go:linkname TrustedIdentifier github.com/mavolin/corgi/v2/escape/safe.trustedIdentifier
func TrustedIdentifier(string) safe.Identifier

// TrustedJSLiteral creates a new JSLiteral from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a JS.
//
//go:linkname TrustedJSLiteral github.com/mavolin/corgi/v2/escape/safe.trustedJSLiteral
func TrustedJSLiteral(string) safe.JSLiteral

// TrustedScript creates a new Script from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Script.
//
//go:linkname TrustedScript github.com/mavolin/corgi/v2/escape/safe.trustedScript
func TrustedScript(string) safe.Script

// TrustedSrcset creates a new Srcset from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Srcset.
//
//go:linkname TrustedSrcset github.com/mavolin/corgi/v2/escape/safe.trustedSrcset
func TrustedSrcset(string) safe.Srcset

// TrustedUnsafe creates a new Unsafe attribute from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an Unsafe
// attribute.
//
//go:linkname TrustedUnsafe github.com/mavolin/corgi/v2/escape/safe.trustedUnsafe
func TrustedUnsafe(attr, val string) safe.Unsafe

// TrustedUnsafeBool creates a new UnsafeBool attribute from the given
// trusted boolean.
//
// Only use this function if you have read the package documentation and are
// sure that the passed boolean satisfies the requirements for an UnsafeBool
// attribute.
//
//go:linkname TrustedUnsafeBool github.com/mavolin/corgi/v2/escape/safe.trustedUnsafeBool
func TrustedUnsafeBool(attr string, val bool) safe.UnsafeBool

// TrustedURL creates a new URL from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a URL.
//
//go:linkname TrustedURL github.com/mavolin/corgi/v2/escape/safe.trustedURL
func TrustedURL(string) safe.URL

// TrustedResourceURL creates a new ResourceURL from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a ResourceURL.
//
//go:linkname TrustedResourceURL github.com/mavolin/corgi/v2/escape/safe.trustedResourceURL
func TrustedResourceURL(string) safe.ResourceURL
