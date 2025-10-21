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
	"strings"

	"github.com/mavolin/corgi/v2/escape/safe"
	"github.com/mavolin/corgi/v2/escape/safe/unsafeconstructor"
	"github.com/mavolin/corgi/v2/internal/escapelite"
)

var (
	// URLSchemes is a list of trusted URL schemes.
	URLSchemes = []string{"http", "https", "ws", "wss", "mailto", "tel"}
	// ResourceURLSchemes is a list of trusted schemes for resource URLs.
	// It must be a subset of URLSchemes.
	ResourceURLSchemes = []string{"https", "wss"}
)

// IsSafeURLScheme returns whether the given scheme is a known safe URL scheme
// as defined by the [URLSchemes] global variable.
func IsSafeURLScheme(probe string) bool {
	return isSafeURLScheme(probe, URLSchemes)
}

// IsSafeResourceURLScheme returns whether the given scheme is a known safe
// ResourceURL scheme as defined by the [ResourceURLSchemes] global variable.
func IsSafeResourceURLScheme(probe string) bool {
	return isSafeURLScheme(probe, ResourceURLSchemes)
}

func isSafeURLScheme(probe string, safeSchemes []string) bool {
	for _, scheme := range safeSchemes {
		if strings.EqualFold(scheme, probe) {
			return true
		}
	}
	return false
}

type unescaped = string

// HTML replaces [&<] with escape sequences.
func HTML(val unescaped) safe.HTML {
	return unsafeconstructor.TrustedHTML(escapelite.Content(val))
}
