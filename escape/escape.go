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
// replace them with [Replacement].
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

// Replacement is the replacement value used by filters to replace unsafe
// content or parts.
const Replacement = escapelite.Replacement

type unescaped = string

// HTML replaces [&<] with escape sequences.
//
// It should only be used to escape the content of an element's body.
func HTML(val unescaped) safe.HTML {
	return safe.TrustedHTML(escapelite.Content(val))
}

// CSSValue escapes CSS special characters using \<hex>+ escapes.
func CSSValue(s unescaped) safe.CSSValue {
	return safe.TrustedCSSValue(escapelite.CSSValue(s))
}

// FilterCSSValue allows innocuous CSS values in the output including CSS
// quantities (10px or 25%), ID or class literals (#foo, .bar), keyword values
// (inherit, blue), and colors (#888).
// It filters out unsafe values, such as those that affect token boundaries,
// and anything that might execute scripts.
func FilterCSSValue(s unescaped) safe.CSSValue {
	return safe.TrustedCSSValue(escapelite.FilterCSSValue(s))
}

// JSLiteral converts the passed value to a JavaScript literal.
func JSLiteral(val any) (safe.JSLiteral, error) {
	esc, err := escapelite.JS(val)
	if err != nil {
		return safe.JSLiteral{}, err
	}
	return safe.TrustedJSLiteral(esc), err
}
