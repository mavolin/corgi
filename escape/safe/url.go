package safe

import "strings"

var (
	// URLSchemes is a list of trusted URL schemes.
	URLSchemes = []string{"http", "https", "ws", "wss", "mailto", "tel"}
	// ResourceURLSchemes is a list of trusted schemes for resource URLs.
	// It must be a subset of URLSchemes.
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

// IsSafeURLScheme returns whether the given scheme is a known safe URL scheme
// as defined by the [URLSchemes] global variable.
func IsSafeURLScheme(probe string) bool {
	return isSafeURLScheme(probe, URLSchemes)
}

// IsSafeResourceURLScheme returns whether the given scheme is a known safe
// ResourceURL scheme as defined by the [ResourceURLSchemes] global variable.
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

// URL represents a known safe URL attribute fragment, safe to be embedded
// between double quotes.
//
// As such, it must escape any raw double quotes.
type (
	// URL represents a known safe URL.
	// It must be complete, not just a part, and as such cannot be
	// embedded or concatenated to another partial URL.
	//
	// It is allowed to interpolate a URL in a URL list attribute.
	// If needed, the escaper will add a leading/trailing space to make sure
	// that the URL is not accidentally concatenated to another URL.
	//
	// A URL cannot be used as part of a resource URL, however, it may
	// very well be used as part of a URLList.
	URL struct{ val string }

	// URLList is a list of URLs for use in URL list attributes.
	URLList []URL

	// ResourceURL only differs semantically from URL, fulfilling a
	// higher security requirement.
	// ResourceURL are URLs that load a sensitive resource, such as a script or
	// stylesheet.
	//
	// A ResourceURL escaped by package escape only allows "https" and "wss"
	// URLs, deeming "http" and "ws" unsafe.
	// This can be changed for dynamically interpolated values during
	// development by calling [github.com/mavolin/corgi.DevelopmentMode()]
	// (Also see [DevelopmentResourceURLSchemes]).
	// Relative URLs are always allowed.
	//
	// A ResourceURL can be used as a regular URL, but not vice versa.
	ResourceURL struct{ val string }
)

// TrustedURL creates a new URL from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a URL.
func TrustedURL(s string) URL { return URL{val: s} }

// TrustedResourceURL creates a new ResourceURL from the given trusted
// string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a ResourceURL.
func TrustedResourceURL(s string) ResourceURL { return ResourceURL{val: s} }

func (a URL) Get() string         { return a.val }
func (a ResourceURL) Get() string { return a.val }
