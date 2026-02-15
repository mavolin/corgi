package safe

import (
	"fmt"
	"regexp"

	_ "unsafe" // for go:linkname
)

// Identifier represents a known safe identifier to be used as an identifier
// for HTML elements, for example the value of the id and name attributes.
//
// Special care must be taken to ensure the value does not shadow existing
// DOM properties, as this can lead to [DOM clobbering attacks].
//
// [DOM clobbering attacks]: https://en.wikipedia.org/wiki/DOM_clobbering
type Identifier struct{ val string }

var identifierPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// ConstantIdentifier creates a new Identifier from the given string constant.
//
// For the sake of simplicity, the identifier must start with an ASCII letter
// and contain only ASCII letters, digits, hyphens, and underscores.
// ConstantIdentifier panics if this is not the case.
//
// Before using this function, read the documentation of [Identifier].
func ConstantIdentifier(c constant) Identifier {
	if !identifierPattern.MatchString(string(c)) {
		panic(fmt.Sprintf("invalid identifier %q", c))
	}
	return trustedIdentifier(string(c))
}

var prefixedIdentifierValuePattern = regexp.MustCompile(`^[A-Za-z0-9_-]*$`)

// PrefixedIdentifier creates a new Identifier by joining prefix and value.
//
// prefix must start with an ASCII letter and contain only ASCII letters,
// digits, hyphens, and underscores.
// Likewise, value must also contain only ASCII letters, digits, hyphens, and
// underscores.
// Value may be empty.
//
// Before using this function, read the documentation of [Identifier].
func PrefixedIdentifier(prefix constant, value string) (Identifier, error) {
	if !identifierPattern.MatchString(string(prefix)) {
		return Identifier{}, fmt.Errorf("invalid identifier prefix %q", prefix)
	} else if !prefixedIdentifierValuePattern.MatchString(value) {
		return Identifier{}, fmt.Errorf("invalid identifier value %q", value)
	}
	return trustedIdentifier(string(prefix) + value), nil
}

// MustPrefixedIdentifier is like PrefixedIdentifier but panics on error.
//
// Before using this function, read the documentation of [PrefixedIdentifier].
func MustPrefixedIdentifier(prefix constant, value string) Identifier {
	id, err := PrefixedIdentifier(prefix, value)
	if err != nil {
		panic(err)
	}
	return id
}

//go:linkname trustedIdentifier
func trustedIdentifier(s string) Identifier { return Identifier{val: s} }

func (id Identifier) Get() string { return id.val }
