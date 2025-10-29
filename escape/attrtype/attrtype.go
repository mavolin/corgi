// Package attrtype provides utilities to determine the content type of an
// attribute.
package attrtype

// Type represents the type of an attribute.
type Type interface {
	_type()
	String() string
}

// All is the list of all predefined attribute types.
var All = [...]Type{
	Bool,
	Int, Float, String,
	Identifier,
	Datetime,
	URL, ResourceURL, Srcset,
	Text,
	CSS, JS,
	Unsafe, UnsafeBool,
	SpaceList{Int},
	SpaceList{Float},
	SpaceList{String},
	SpaceList{URL},
	SpaceList{ResourceURL},
	CommaList{Int},
	CommaList{Float},
	CommaList{String},
	CommaList{URL},
	CommaList{ResourceURL},
}

var (
	// Bool is a boolean attribute, specially treated by HTML through the
	// presence or absence of the attribute name.
	//
	// Permitted Types: bool
	Bool = standalone{"bool"}

	// Int is an integer attribute, formatted according to the [signed integer]
	// microsyntax.
	//
	// Permitted Types: all (u)ints
	//
	// [signed integer]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#signed-integers
	Int = primitive{"int"}
	// Float is a floating point attribute, formatted according to the
	// [floating-point numbers] microsyntax.
	//
	// Permitted Types: all (u)ints, all floats
	//
	// [floating-point numbers]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#floating-point-numbers
	Float = primitive{"float"}
	// String is a safe (see Unsafe for the definition of the antonym)
	// machine-readable string attribute.
	// It is distinct from [Text], which is meant for human consumption.
	//
	// Permitted Types: string, all (u)ints, all floats
	String = primitive{"string"}

	// Identifier is an attribute that refers to an identifier for HTML
	// elements, for example the value of the id and name attributes,
	// that are vulnerable to DOM clobbering attacks if not carefully chosen.
	//
	// Permitted Types: safe.Identifier, strings with a constant prefix
	Identifier = standalone{"identifier"}

	// Datetime is an attribute formatted according to one of the various
	// date, time, and duration microsyntaxes in the HTML specification.
	//
	// Type: html.DateTime, constant strings
	Datetime = standalone{"datetime"}

	// URL is a safe (see Unsafe for the definition of the antonym) URL
	// formated according to the [valid URL string] microsyntax.
	//
	// When used as part of a comma list, all commas are escaped.
	//
	// Permitted Types: safe.URL, safe.ResourceURL,
	// strings with a constant or safe scheme
	//
	// [valid URL string]: https://url.spec.whatwg.org/#valid-url-string
	URL = primitive{"url"}
	// ResourceURL is a URL that is safe to use in a context that loads a
	// resource (e.g. script, stylesheet, image, ...).
	// It is distinct from [URL], which is safe to use in an arbitrary URL
	// context, but does not necessarily fulfill the additional requirements
	// to be safe to use as a resource URL.
	//
	// Permitted Types: safe.ResourceURL, strings with a constant or safe scheme
	ResourceURL = primitive{"resourceURL"}
	// Srcset is a comma-separated list of [image candidate strings], formatted
	// according to the [srcset attribute syntax].
	//
	// Permitted Types: safe.URL, safe.ResourceURL,
	// strings with a constant or safe scheme
	//
	// [image candidate strings]: https://html.spec.whatwg.org/multipage/images.html#image-candidate-string
	// [srcset attribute syntax]: https://html.spec.whatwg.org/multipage/images.html#srcset-attribute
	Srcset = standalone{"srcset"}

	// Text is an attribute containing text consumed by humans.
	// Effectively, its only difference from String is that Text attributes
	// should be localized while String attributes should not.
	//
	// Permitted Types: string, all (u)ints, all floats, big.Int, big.Float
	Text = standalone{"text"}

	// CSS is an attribute that accepts CSS declarations.
	//
	// Permitted Types: safe.CSSValue, safe.CSSDeclarations, string constants,
	// all (u)ints, all floats, color.Color
	CSS = standalone{"css"}
	// JS is an attribute that accepts a JavaScript [FunctionBody] production.
	//
	// [FunctionBody]: https://tc39.es/ecma262/#prod-FunctionBody
	//
	// Permitted Types: safe.JSLiteral, safe.JS
	JS = standalone{"js"}

	// Unsafe is used for values that affect how embedded content and network
	// messages are formed, vetted, or interpreted; or which credentials
	// network messages carry.
	//
	// Permitted Types: unsafe.Unsafe, string constants
	Unsafe = standalone{"unsafe"}
	// UnsafeBool is the combination of [Unsafe] and [Bool].
	//
	// Permitted Types: unsafe.UnsafeBool, bool constants
	UnsafeBool = standalone{"unsafeBool"}
)

type (
	// CommaList is a list of values separated by commas, conforming to the
	// syntax of a [set of comma-separated tokens].
	//
	// [set of comma-separated tokens]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#set-of-comma-separated-tokens
	CommaList struct {
		Element CommaListElement
	}
	CommaListElement interface {
		Type
		_commaListElement()
	}

	// SpaceList is a list of values separated by whitespace, conforming to the
	// syntax of a [set of space-separated tokens].
	//
	// [set of space-separated tokens]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#set-of-space-separated-tokens
	SpaceList struct {
		Element SpaceListElement
	}
	SpaceListElement interface {
		Type
		_spaceListElement()
	}
)

type (
	primitive struct {
		name string
	}
	standalone struct {
		name string
	}
)

var (
	_ Type             = primitive{}
	_ Type             = standalone{}
	_ Type             = CommaList{}
	_ CommaListElement = primitive{}
	_ Type             = SpaceList{}
	_ SpaceListElement = primitive{}
)

func (p primitive) String() string   { return p.name }
func (primitive) _type()             {}
func (primitive) _spaceListElement() {}
func (primitive) _commaListElement() {}

func (s standalone) String() string { return s.name }
func (standalone) _type()           {}

func (l SpaceList) String() string { return "spaceList[" + l.Element.String() + "]" }
func (SpaceList) _type()           {}

func (l CommaList) String() string { return "commaList[" + l.Element.String() + "]" }
func (CommaList) _type()           {}
