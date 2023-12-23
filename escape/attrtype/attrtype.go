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
type Type uint8

const (
	Unknown Type = iota
	Plain
	CSS
	JS
	URL
	URLList
	ResourceURL
	Srcset
	// Unsafe is used  for values that affect how embedded content and network
	//messages are formed, vetted, or interpreted; or which credentials network
	// messages carry.
	Unsafe
	_invalid
)

func (t Type) IsValid() bool {
	return t < _invalid
}

func (t Type) String() string {
	switch t {
	case Unknown:
		return "<unknown>"
	case Plain:
		return "plain"
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
