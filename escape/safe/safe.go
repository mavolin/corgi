// Package safe provides types that hold data that is trusted to be used inside
// HTML documents.
//
// If you want to generate safe HTML from untrusted sources, use the parent
// package escape instead.
//
// Usage of these types poses a security risk and should be done with great
// care, as the encapsulated content will be included verbatim in the template.
// Manually creating instances of these types may undermine the security
// guarantees provided by corgi.
// Always ensure that content fully complies with the requirements put forth in
// the documentation of the respective type.
// Content should always come from a trusted source and never from a
// third party like an end-user, or any source that the developer does not have
// full control over.
//
// To create instances of these types, use the Trusted* functions.
// The [Concat] helper can be used to concatenate multiple safe fragments into
// one.
//
// If a content type can appear both in the body/root of a document and as an
// attribute value, it is represented by two different types with different
// requirements.
//
// Since corgi writes all attributes using double quotes, attribute types
// needn't escape single quotes.
//
// Characters that must be escaped are specified in regular expression
// character classes like [<&], which requires the characters '<' and '&' to be
// escaped, but not the character '[' or ']'.
//
// This package follows some of the design proposals from here:
// https://github.com/golang/go/issues/27926
package safe

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

import "strings"

// UnsafeReplacement is the string to be used as replacement for an unsafe
// fragment.
const UnsafeReplacement = "ZcorgiZ"

type (
	// Fragment is an interface fulfilled by all safe types.
	Fragment interface {
		bodyFragment | attrFragment
		Escaped() string
	}

	BodyFragment interface {
		bodyFragment
		Escaped() string
	}
	bodyFragment interface {
		CSSValue | HTML | JS
	}

	AttrFragment interface {
		attrFragment
		Escaped() string
	}
	attrFragment interface {
		CSSValueAttr | PlainAttr | JSAttr | SrcsetAttr | UnsafeAttr |
			URLAttr | ResourceURLAttr | URLListAttr
	}
)

// Concat concatenates multiple fragments of the same type into one.
func Concat[T Fragment](fs ...T) T {
	var n int
	for _, f := range fs {
		n += len(f.Escaped())
	}

	var b strings.Builder
	b.Grow(n)

	for _, f := range fs {
		b.WriteString(f.Escaped())
	}

	return T{val: b.String()}
}

// DevelopmentMode enables development mode for the remainder of the program.
//
// Usually, you would not call this function directly, but use
// [github.com/mavolin/corgi.DevelopmentMode] instead.
func DevelopmentMode() {
	developmentMode(nil)
}

// if you change this, please remember to update the documentation of
// corgi.DevelopmentMode
func developmentMode(f func()) {
	if f != nil {
		oldResourceURLSchemes := ResourceURLSchemes
		defer func() {
			ResourceURLSchemes = oldResourceURLSchemes
		}()
	}

	ResourceURLSchemes = DevelopmentResourceURLSchemes
}
