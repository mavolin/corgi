package elemtype

// Func is a function that returns the type for a given element.
//
// If the Func cannot determine the type, it should return Unknown.
type Func func(element string) Type

// Combine combines multiple Funcs into a single Func.
func Combine(fs ...Func) Func {
	return func(element string) Type {
		for _, f := range fs {
			t := f(element)
			if t.IsValid() {
				return t
			}
		}
		return Unknown
	}
}

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
	// An HTML element is the most common element type and can hold other HTML
	// elements.
	HTML
	// A Text element can only contain text, but no child elements.
	// Ampersand escapes are allowed in text elements.
	Text
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

// IsValid returns whether t is a valid Type.
func (t Type) IsValid() bool {
	return t > Unknown && t < invalid
}
