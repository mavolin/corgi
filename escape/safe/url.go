package safe

type (
	// URL represents a known safe URL.
	// It must be complete, not just a part, and as such cannot be
	// embedded or concatenated to another partial URL.
	//
	// It is allowed to interpolate a URL in a URL list attribute.
	// If needed, the escaper will add a leading/trailing space to make sure
	// that the URL is not accidentally concatenated to another URL.
	//
	// A URL cannot be used as a resource URL.
	URL struct{ val string }

	// URLList is a list of URLs for use in URL list attributes.
	URLList []URL

	// ResourceURL only differs semantically from URL, fulfilling a
	// higher security requirement.
	// A ResourceURL is a URL that load a sensitive resource, such as a script
	// or stylesheet.
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
