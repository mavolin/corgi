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

import "github.com/mavolin/corgi/escape/safe"

/*
This package contains excerpts from the Go standard library package
html/template licensed under the below license:

Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

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

// IsVoidElement reports whether the element with the passed name is an HTML
// void element.
func IsVoidElement(name string) bool {
	_, ok := VoidElements[name]
	return ok
}
