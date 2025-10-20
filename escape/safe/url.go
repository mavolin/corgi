package safe

import _ "unsafe" // for go:linkname

type (
	// URL represents a known safe URL.
	//
	// A URL cannot be used as a resource URL.
	URL struct{ val string }

	// ResourceURL only differs semantically from URL, fulfilling a
	// higher security requirement.
	// A ResourceURL is a URL that load a sensitive resource, such as a script
	// or stylesheet.
	//
	// A ResourceURL escaped by package escape only allows "https" and "wss"
	// URLs, deeming "http" and "ws" unsafe.
	// Relative URLs are always allowed.
	//
	// A ResourceURL can be used as a regular URL, but not vice versa.
	ResourceURL struct{ val string }
)

// ConstantURL creates a new URL wrapper from the given string constant.
//
// Before using this function, read the documentation of [URL].
func ConstantURL(c constant) URL {
	return trustedURL(string(c))
}

// ConstantResourceURL creates a new ResourceURL wrapper from the given string
// constant.
//
// Before using this function, read the documentation of [ResourceURL].
func ConstantResourceURL(c constant) ResourceURL {
	return trustedResourceURL(string(c))
}

//go:linkname trustedURL
func trustedURL(s string) URL { return URL{val: s} }

//go:linkname trustedResourceURL
func trustedResourceURL(s string) ResourceURL { return ResourceURL{val: s} }

func (a URL) Get() string         { return a.val }
func (a ResourceURL) Get() string { return a.val }
