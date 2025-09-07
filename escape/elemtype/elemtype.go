package elemtype

// Type represents the content type of an element, roughly equivalent to the
// HTML spec's content model.
//
// Never use the numeric values of a Type directly, but only the provided
// constants.
type Type uint8

const (
	Unknown Type = iota
	// A Void element has no body and no closing tag.
	Void
	// A Nothing element is akin to a void element: It has a closing tag, but
	// must have an empty body.
	Nothing
	// A Text element can only contain text, but no child elements.
	// Ampersand escapes are allowed in text elements.
	Text
	// A Normal element is the most common element type and can hold text or
	// other elements.
	Normal
	// A CSS element can only contain CSS code as text.
	// Ampersand escapes are unavailable in CSS elements and are ignored.
	// As such, special care must be taken to prevent premature end of
	// the element, by ensuring that it does not contain the case-insensitive
	// closing tag.
	CSS
	// A JS element can only contain JS code as text.
	// Ampersand escapes are unavailable in JS elements and are ignored.
	// As such, special care must be taken to prevent premature end of
	// the element, by ensuring that it does not contain the case-insensitive
	// closing tag.
	JS
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

// IsValid returns whether t is a valid Type.
func (t Type) IsValid() bool {
	return t > Unknown && t < invalid
}

func (t Type) String() string {
	switch t {
	case Unknown:
		return "<unknown>"
	case Void:
		return "void"
	case Nothing:
		return "nothing"
	case Normal:
		return "normal"
	case Text:
		return "text"
	case CSS:
		return "css"
	case JS:
		return "js"
	case invalid:
		fallthrough
	default:
		return "<invalid>"
	}
}
