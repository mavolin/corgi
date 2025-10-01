// Package attrtype provides utilities to determine the content type of an
// attribute.
package attrtype

// Type represents the type of an attribute.
//
// Never use the numeric values of a Type directly, but only the provided
// constants.
type Type uint8

const (
	Unknown Type = iota
	// Unsafe is used for values that affect how embedded content and network
	// messages are formed, vetted, or interpreted; or which credentials
	// network messages carry.
	Unsafe
	UnsafeBool
	Bool
	// Text is an attribute containing text consumed by humans.
	// Effectively, its only difference from Innocuous is that Text attributes
	// should be localized while Innocuous attributes should not.
	Text
	Innocuous   // text attribute containing text consumed by machines
	CSS         // CSS code
	JS          // JS code
	URL         // a single URL
	URLList     // space separated list of URLs
	ResourceURL // a URL loading a resource; stricter security requirements
	Srcset      // a srcset-like attribute
	invalid
)

// All is the list of all valid Types.
var All = func() [invalid - 1]Type {
	var ts [invalid - 1]Type
	for t := Unknown + 1; t < invalid; t++ {
		ts[int(t)-1] = t
	}
	return ts
}()

func (t Type) IsValid() bool {
	return t > Unknown && t < invalid
}

// String returns the string representation of t.
func (t Type) String() string {
	switch t {
	case Unknown:
		return "<unknown>"
	case Unsafe:
		return "unsafe"
	case UnsafeBool:
		return "unsafe bool"
	case Bool:
		return "bool"
	case Text:
		return "text"
	case Innocuous:
		return "innocuous"
	case CSS:
		return "css"
	case JS:
		return "js"
	case URL:
		return "url"
	case URLList:
		return "url list"
	case ResourceURL:
		return "resource url"
	case Srcset:
		return "srcset"
	case invalid:
		fallthrough
	default:
		return "<invalid>"
	}
}
