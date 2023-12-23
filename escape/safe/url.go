package safe

import "strings"

var (
	// URLSchemes is a list of trusted URL schemes.
	URLSchemes = []string{"http", "https", "mailto", "tel"}
	// ResourceURLSchemes is a list of trusted schemes for resource URLs.
	ResourceURLSchemes = []string{"https", "wss"}
	// DevelopmentResourceURLSchemes is a list of trusted schemes for resource URLs
	// that is more permissive than ResourceURLSchemes and allows http.
	// It is intended for use in development environments only and can be used
	// by calling [github.com/mavolin/corgi.DevelopmentMode()].
	//
	// Using this list instead does not affect static resource URLs, which are
	// never allowed to use http.
	DevelopmentResourceURLSchemes = []string{"http", "https", "ws", "wss"}
)

// IsSafeURLScheme returns whether the given scheme is a known safe URLAttr scheme
// as defined by the [URLSchemes] global variable.
func IsSafeURLScheme(probe string) bool {
	return isSafeURLScheme(probe, URLSchemes)
}

// IsSafeResourceURLScheme returns whether the given scheme is a known safe
// ResourceURLAttr scheme as defined by the [ResourceURLSchemes] global variable.
func IsSafeResourceURLScheme(probe string) bool {
	return isSafeURLScheme(probe, ResourceURLSchemes)
}

func isSafeURLScheme(probe string, safeSchemes []string) bool {
	for _, scheme := range safeSchemes {
		if strings.EqualFold(scheme, probe) {
			return true
		}
	}
	return false
}

// URLAttr represents a known safe URL attribute fragment, safe to be embedded
// between double quotes.
//
// As such, it must escape any raw double quotes.
type (
	// URLAttr represents a known safe URL attribute fragment, safe to be embedded
	// between double quotes.
	//
	// As such, it must escape any raw double quotes.
	//
	// A URLAttr cannot be used as part of a resource URL, however, it may
	// very well be used as part of a URLList.
	URLAttr struct{ val string }

	// ResourceURLAttr only differs semantically from URLAttr it fulfills a
	// higher security requirement.
	// ResourceURLAttrs are attributes that are used as part of a URL that
	// loads a sensitive resource, such as a script or stylesheet.
	//
	// ResourceURLAttrs escaped by package escape only allow https URLs,
	// deeming http URLs unsafe.
	// This can be changed for dynamically interpolated values during
	// development by calling [github.com/mavolin/corgi.DevelopmentMode()]
	// (Also see [DevelopmentResourceURLSchemes]).
	// Relative URLs are always allowed.
	//
	// It must escape double quotes.
	//
	// A ResourceURLAttr can be used as part of a regular URLAttr, but not
	// vice versa.
	ResourceURLAttr struct{ val string }

	// URLListAttr represents a space-separated list of URLs.
	//
	// If multiple URLListAttrs are used as part of the same attribute, corgi
	// will ensure that they are separated by at least one space.
	//
	// The same rules as for URLAttr apply.
	URLListAttr struct{ val string }
)

// TrustedURLAttr creates a new URLAttr from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an URLAttr.
func TrustedURLAttr(s string) URLAttr { return URLAttr{val: s} }

// TrustedResourceURLAttr creates a new ResourceURLAttr from the given trusted
// string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a ResourceURLAttr.
func TrustedResourceURLAttr(s string) ResourceURLAttr { return ResourceURLAttr{val: s} }

// TrustedURLListAttr creates a new URLListAttr from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an URLListAttr.
func TrustedURLListAttr(s string) URLListAttr { return URLListAttr{val: s} }

func (a URLAttr) Escaped() string         { return a.val }
func (a URLListAttr) Escaped() string     { return a.val }
func (a ResourceURLAttr) Escaped() string { return a.val }
