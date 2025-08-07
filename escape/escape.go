// Package escape provides functions to escape and filter untrusted data for
// inclusion in HTMLText documents.
//
// It's escapers and filters are mostly based off, or exactly the same as the
// ones provided by Go's stdlib package html/template.
//
// # Terminology
//
// In this package, you will find three seemingly similar types of functions:
//
// Escapers take untrusted data for a specific content type and replace unsafe
// parts with escape sequences, so that the data can be used in its content
// domain without unintended side effects.
// The meaning of the data is not changed.
//
// Filters take untrusted data for a specific content type and replace unsafe
// parts with a safe replacement or delete unsafe parts, changing or fully
// replacing the data to ensure safety.
// Filters won't return an error if they find unsafe parts, but will instead
// replace them with [safe.UnsafeReplacement].
//
// Normalizers take trusted data and escape sequences that are not yet escaped.
// It differs from escapers in that it trusts already present escape sequences
// and does not escape them again.
//
// Not every content type has a filter or normalizer.
//
// If you are in doubt which function to use, and your content type offers
// both an escaper and a filter, you should most likely use the filter.
//
// Note that you don't need to escape interpolated content yourself.
// Corgi will do that automatically for you.
package escape

import (
	"github.com/mavolin/corgi/v2/escape/safe"
	"github.com/mavolin/corgi/v2/internal/escapelite"
)

type unescaped = string

type (
	Func[T safe.Fragment]        func(any) (T, error)
	ContextFunc[T safe.Fragment] func(...any) (T, error)
)

// VoidElements is a set of all HTML void elements.
// https://developer.mozilla.org/en-US/docs/Glossary/Empty_element
var VoidElements = map[string]struct{}{
	"area":   {},
	"base":   {},
	"br":     {},
	"col":    {},
	"embed":  {},
	"hr":     {},
	"img":    {},
	"input":  {},
	"link":   {},
	"meta":   {},
	"param":  {},
	"source": {},
	"track":  {},
	"wbr":    {},
}

// HTML replaces [&<] with escape sequences.
//
// It should only be used to escape the content of an element's body.
func HTML(val unescaped) safe.HTML {
	return safe.TrustedHTML(escapelite.Content(val))
}
