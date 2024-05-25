// Package attrtype provides utilities to determine the content type of an
// attribute.
package attrtype

// Func is a function that returns the type for a given attribute on a given
// element.
//
// If the Func cannot determine the type, it should return Unknown.
type Func func(element, attr string) Type

// Combine combines multiple Funcs into a single Func.
func Combine(fs ...Func) Func {
	return func(element, attr string) Type {
		for _, f := range fs {
			t := f(element, attr)
			if t.IsValid() {
				return t
			}
		}
		return Unknown
	}
}

// Type represents the type of attribute.
//
// Never use the numeric values of a Type directly, but only the provided
// constants.
//
// A Type is a single of the types, or a bitmask of TextList and at least one
// of Comma, Semicolon, or Space as the delimiters.
type Type uint8

const (
	Unknown Type = iota
	// Unsafe is used for values that affect how embedded content and network
	// messages are formed, vetted, or interpreted; or which credentials
	// network messages carry.
	Unsafe
	UnsafeBool
	Bool
	Text        // plain text with only HTML escapes
	CSS         // CSS code with HTML escapes
	JS          // JS code with HTML escapes
	URL         // a single URL
	URLList     // space separated of URLs
	ResourceURL // a URL loading a resource; stricter security requirements
	Srcset      // a srcset-like attribute
	invalid
)

func (t Type) IsValid() bool {
	return t > Unknown && t < invalid
}

// String returns the string representation of t.
//
// For a text list, it returns the string "text list" followed by all the valid
// delimiters for that list in square brackets.
func (t Type) String() string {
	switch t {
	case Unknown:
		return "<unknown>"
	case Bool:
		return "bool"
	case Text:
		return "text"
	case CSS:
		return "css"
	case JS:
		return "js"
	case URL:
		return "url"
	case URLList:
		return "url"
	case ResourceURL:
		return "resource url"
	case Srcset:
		return "srcset"
	default:
		return "<invalid>"
	}
}
