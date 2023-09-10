package woof

/*
This file contains excerpts from the Go standard library package html/template,
licensed under the below license:

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

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Strings of content from a trusted source.
type (
	// HTMLText represents an HTML fragment consisting purely of escaped text
	// suitable to be placed both in the body of a tag (or the root of the
	// document), and as an attribute value, possibly later enclosed in quotes.
	//
	// This effectively makes HTMLText a mixture of HTMLBody and HTMLAttrVal.
	//
	// As such, it must not contain any elements and must escape ["<&].
	//
	// Corgi only considers HTMLText safe for "plain" attributes and won't,
	// for example, allow the usage of a HTMLText, as an href.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	HTMLText string
	// HTMLBody represents a known safe n HTML fragment to be placed in the
	// body of an element or the root of the document.
	//
	// It should not be used for HTML from a third-party, or HTML with
	// unclosed tags or comments and must escape at least [<&].
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	HTMLBody string

	// HTMLAttrVal represents a known safe HTML attribute value fragment that
	// can be placed between double-quotes to be used as part of an HTML
	// attribute.
	//
	// It must at least escape ["&].
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	HTMLAttrVal string

	// CSS encapsulates known safe content that matches any of:
	//   1. The CSS3 stylesheet production, such as `p { color: purple }`.
	//   2. The CSS3 rule production, such as `a[href=~"https:"].foo#bar`.
	//   3. CSS3 declaration productions, such as `color: red; margin: 2px`.
	//   4. The CSS3 value production, such as `rgba(0, 0, 255, 127)`.
	// See https://www.w3.org/TR/css3-syntax/#parsing and
	// https://web.archive.org/web/20090211114933/http://w3.org/TR/css3-syntax#style
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	CSS string

	// URL represents a known safe URL or URL fragment.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will only be HTML escaped before including in template output.
	URL string

	// Srcset represents a known safe srcset attribute value, safe to be
	// embedded between two double quotes and used as a srcset attribute or
	// part of a srcset attribute.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will only be HTML escaped before including in template output.
	Srcset string

	// JS encapsulates a known safe EcmaScript5 Expression, for example,
	// `(x + y * z())`.
	// Template authors are responsible for ensuring that typed expressions
	// do not break the intended precedence and that there is no
	// statement/expression ambiguity as when passing an expression like
	// "{ foo: bar() }\n['foo']()", which is both a valid Expression and a
	// valid Program with a very different meaning.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	//
	// Using JS to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	JS string
	// JSStr encapsulates a sequence of characters meant to be embedded
	// between quotes in a JavaScript expression.
	// The string must match a series of StringCharacters:
	//   StringCharacter :: SourceCharacter but not `\` or LineTerminator
	//                    | EscapeSequence
	// Note that LineContinuations are not allowed.
	// JSStr("foo\\nbar") is fine, but JSStr("foo\\\nbar") is not.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	JSStr string

	// JSAttrVal is a js attribute safe to be embedded in double quotes and used
	// as an attribute value.
	//
	// Use of this type presents a security risk:
	// the encapsulated content should come from a trusted source,
	// as it will be included verbatim in the template output.
	//
	// Using JS to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	JSAttrVal string
)

const UnsafeReplacement = "ZcorgiZ"

// ============================================================================
// CSS
// ======================================================================================

// isCSSNmchar reports whether rune is allowed anywhere in a CSS identifier.
func isCSSNmchar(r rune) bool {
	// Based on the CSS3 nmchar production but ignores multi-rune escape
	// sequences.
	// https://www.w3.org/TR/css3-syntax/#SUBTOK-nmchar
	return 'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		'0' <= r && r <= '9' ||
		r == '-' ||
		r == '_' ||
		// Non-ASCII cases below.
		0x80 <= r && r <= 0xd7ff ||
		0xe000 <= r && r <= 0xfffd ||
		0x10000 <= r && r <= 0x10ffff
}

// decodeCSS decodes CSS3 escapes given a sequence of stringchars.
// If there is no change, it returns the input, otherwise it returns a slice
// backed by a new array.
// https://www.w3.org/TR/css3-syntax/#SUBTOK-stringchar defines stringchar.
func decodeCSS(s []byte) []byte {
	i := bytes.IndexByte(s, '\\')
	if i == -1 {
		return s
	}
	// The UTF-8 sequence for a codepoint is never longer than 1 + the
	// number hex digits need to represent that codepoint, so len(s) is an
	// upper bound on the output length.
	b := make([]byte, 0, len(s))
	for len(s) != 0 {
		i := bytes.IndexByte(s, '\\')
		if i == -1 {
			i = len(s)
		}
		b, s = append(b, s[:i]...), s[i:]
		if len(s) < 2 {
			break
		}
		// https://www.w3.org/TR/css3-syntax/#SUBTOK-escape
		// escape ::= unicode | '\' [#x20-#x7E#x80-#xD7FF#xE000-#xFFFD#x10000-#x10FFFF]
		if isHex(s[1]) {
			// https://www.w3.org/TR/css3-syntax/#SUBTOK-unicode
			//   unicode ::= '\' [0-9a-fA-F]{1,6} wc?
			j := 2
			for j < len(s) && j < 7 && isHex(s[j]) {
				j++
			}
			r := hexDecode(s[1:j])
			if r > unicode.MaxRune {
				r, j = r/16, j-1
			}
			n := utf8.EncodeRune(b[len(b):cap(b)], r)
			// The optional space at the end allows a hex
			// sequence to be followed by a literal hex.
			// string(decodeCSS([]byte(`\A B`))) == "\nB"
			b, s = b[:len(b)+n], skipCSSSpace(s[j:])
		} else {
			// `\\` decodes to `\` and `\"` to `"`.
			_, n := utf8.DecodeRune(s[1:])
			b, s = append(b, s[1:1+n]...), s[1+n:]
		}
	}
	return b
}

// isHex reports whether the given character is a hex digit.
func isHex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

// hexDecode decodes a short hex digit sequence: "10" -> 16.
func hexDecode(s []byte) rune {
	n := '\x00'
	for _, c := range s {
		n <<= 4
		switch {
		case '0' <= c && c <= '9':
			n |= rune(c - '0')
		case 'a' <= c && c <= 'f':
			n |= rune(c-'a') + 10
		case 'A' <= c && c <= 'F':
			n |= rune(c-'A') + 10
		default:
			panic(fmt.Sprintf("Bad hex digit in %q", s))
		}
	}
	return n
}

// skipCSSSpace returns a suffix of c, skipping over a single space.
func skipCSSSpace(c []byte) []byte {
	if len(c) == 0 {
		return c
	}
	// wc ::= #x9 | #xA | #xC | #xD | #x20
	switch c[0] {
	case '\t', '\n', '\f', ' ':
		return c[1:]
	case '\r':
		// This differs from CSS3's wc production because it contains a
		// probable spec error whereby wc contains all the single byte
		// sequences in nl (newline) but not CRLF.
		if len(c) >= 2 && c[1] == '\n' {
			return c[2:]
		}
		return c[1:]
	}
	return c
}

// isCSSSpace reports whether b is a CSS space char as defined in wc.
func isCSSSpace(b byte) bool {
	switch b {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	}
	return false
}

var cssReplacementTable = []string{
	0:    `\0`,
	'\t': `\9`,
	'\n': `\a`,
	'\f': `\c`,
	'\r': `\d`,
	// Encode HTML specials as hex so the output can be embedded
	// in HTML attributes without further encoding.
	'"':  `\22`,
	'&':  `\26`,
	'\'': `\27`,
	'(':  `\28`,
	')':  `\29`,
	'+':  `\2b`,
	'/':  `\2f`,
	':':  `\3a`,
	';':  `\3b`,
	'<':  `\3c`,
	'>':  `\3e`,
	'\\': `\\`,
	'{':  `\7b`,
	'}':  `\7d`,
}

// EscapeCSSValue escapes HTML and CSS special characters using \<hex>+ escapes.
//
// The output is safe to use in HTML bodies and (possibly quote-wrapped)
// attributes without further HTML escaping.
func EscapeCSSValue(val any) (CSS, error) {
	if css, ok := val.(CSS); ok {
		return css, nil
	}

	s, err := stringify(val, escapeCSS)
	return CSS(s), err
}

var (
	expressionBytes = []byte("expression")
	mozBindingBytes = []byte("mozbinding")
)

// FilterCSSValue allows innocuous CSS values in the output including CSS
// quantities (10px or 25%), ID or class literals (#foo, .bar), keyword values
// (inherit, blue), and colors (#888).
// It filters out unsafe values, such as those that affect token boundaries,
// and anything that might execute scripts.
func FilterCSSValue(val any) (CSS, error) {
	if css, ok := val.(CSS); ok {
		return css, nil
	}

	s, err := stringify(val, func(s string) string {
		b, id := decodeCSS([]byte(s)), make([]byte, 0, 64)

		// CSS3 error handling is specified as honoring string boundaries per
		// https://www.w3.org/TR/css3-syntax/#error-handling :
		//     Malformed declarations. User agents must handle unexpected
		//     tokens encountered while parsing a declaration by reading until
		//     the end of the declaration, while observing the rules for
		//     matching pairs of (), [], {}, "", and '', and correctly handling
		//     escapes. For example, a malformed declaration may be missing a
		//     property, colon (:) or value.
		// So we need to make sure that values do not have mismatched bracket
		// or quote characters to prevent the browser from restarting parsing
		// inside a string that might embed JavaScript source.
		for i, c := range b {
			switch c {
			case 0, '"', '\'', '(', ')', '/', ';', '@', '[', '\\', ']', '`', '{', '}', '<', '>':
				return UnsafeReplacement
			case '-':
				// Disallow <!-- or -->.
				// -- should not appear in valid identifiers.
				if i != 0 && b[i-1] == '-' {
					return UnsafeReplacement
				}
			default:
				if c < utf8.RuneSelf && isCSSNmchar(rune(c)) {
					id = append(id, c)
				}
			}
		}
		id = bytes.ToLower(id)
		if bytes.Contains(id, expressionBytes) || bytes.Contains(id, mozBindingBytes) {
			return UnsafeReplacement
		}
		return escapeCSS(string(b))
	})
	return CSS(s), err
}

func escapeCSS(s string) string {
	var b strings.Builder
	r, w, written := rune(0), 0, 0 //nolint:wastedassign

Loop:
	for i := 0; i < len(s); i += w {
		// See comment in htmlEscaper.
		r, w = utf8.DecodeRuneInString(s[i:])
		var repl string
		switch {
		case int(r) < len(cssReplacementTable) && cssReplacementTable[r] != "":
			repl = cssReplacementTable[r]
		default:
			continue Loop
		}
		if written == 0 {
			b.Grow(len(s))
		}
		b.WriteString(s[written:i])
		b.WriteString(repl)
		written = i + w
		if repl != `\\` && (written == len(s) || isHex(s[written]) || isCSSSpace(s[written])) {
			b.WriteByte(' ')
		}
	}
	if written == 0 {
		return s
	}
	b.WriteString(s[written:])
	return b.String()
}

// ============================================================================
// HTML
// ======================================================================================

var (
	htmlEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`'`, "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&#34;", // "&#34;" is shorter than "&quot;".
	)
	htmlBodyToHTMLTextEscaper = strings.NewReplacer(
		`'`, "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
		`>`, "&gt;",
		`"`, "&#34;", // "&#34;" is shorter than "&quot;".
	)
	htmlAttrValToHTMLTextEscaper = strings.NewReplacer(
		`'`, "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
		`<`, "&lt;",
		`>`, "&gt;",
	)
)

// EscapeHTML replaces [&'<>"] with escape sequences.
//
// Usually, [EscapeHTMLBody] and [EscapeAttrVal] are more appropriate, as they
// have a smaller set of replacements, specific to their context.
//
// This does not mean however, that content escaped with EscapeHTML, is not
// valid in HTML bodies or attributes (the opposite is true).
// Instead, EscapeHTML simply escapes more characters than required by those
// contexts.
func EscapeHTML(val any) (HTMLText, error) {
	switch val := val.(type) {
	case HTMLText:
		return val, nil
	case HTMLBody:
		return HTMLText(htmlBodyToHTMLTextEscaper.Replace(string(val))), nil
	case HTMLAttrVal:
		return HTMLText(htmlAttrValToHTMLTextEscaper.Replace(string(val))), nil
	}

	s, err := stringify(val, htmlEscaper.Replace)
	return HTMLText(s), err
}

var (
	htmlBodyEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
	)
	htmlAttrValToHTMLBodyEscaper = strings.NewReplacer(`"`, "&#34;")
)

// EscapeHTMLBody replaces [&<] with escape sequences.
//
// It should only be used to escape a tag body's content.
//
// Since '>' is not escaped, previous start or end tags must be closed using a
// '>' to not cause unexpected side effects.
// This is the case for all escaped code generated by corgi and could only be a
// problem if user manually writes unclosed tags using unescaped assigns, which
// is discouraged by the corgi documentation.
func EscapeHTMLBody(val any) (HTMLBody, error) {
	switch val := val.(type) {
	case HTMLBody:
		return val, nil
	case HTMLText:
		return HTMLBody(val), nil
	case HTMLAttrVal:
		return HTMLBody(htmlAttrValToHTMLBodyEscaper.Replace(string(val))), nil
	}

	s, err := stringify(val, htmlBodyEscaper.Replace)
	return HTMLBody(s), err
}

var (
	htmlAttrValEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`"`, "&#34;",
	)
	htmlBodyToHTMLAttrValEscaper = strings.NewReplacer(`<`, "&lt;")
)

// EscapeHTMLAttrVal escapes s by replacing [&"] so that it can safely be
// placed between double  quotes as an attribute value.
//
// It does not perform further context specific escaping.
func EscapeHTMLAttrVal(val any) (HTMLAttrVal, error) {
	switch val := val.(type) {
	case HTMLAttrVal:
		return val, nil
	case HTMLText:
		return HTMLAttrVal(val), nil
	case HTMLBody:
		return HTMLAttrVal(htmlBodyToHTMLAttrValEscaper.Replace(string(val))), nil
	}

	s, err := stringify(val, htmlAttrValEscaper.Replace)
	return HTMLAttrVal(s), err
}

// ============================================================================
// URL
// ======================================================================================

func FilterURL(vals ...any) (URL, error) {
	u, safe, err := escapeURL(vals...)
	if err != nil {
		return "", err
	}

	if !safe {
		return "#" + UnsafeReplacement, nil
	}

	return u, nil
}

func EscapeURL(vals ...any) (URL, error) {
	u, _, err := escapeURL(vals...)
	return u, err
}

func escapeURL(vals ...any) (u URL, safe bool, err error) {
	if len(vals) == 0 {
		return "", true, nil
	}

	safe = true
	var protoEnd bool
	var inQuery bool

	var b strings.Builder
	for _, val := range vals {
		if u, ok := val.(URL); ok {
			inQuery2, protoEndPos := normalizeURL(&b, u)
			if !inQuery {
				inQuery = inQuery2
			}

			if protoEnd {
				continue
			}

			// if the protocol hasn't ended yet and u adds to it
			if protoEndPos != 0 {
				protoEnd = true
				continue
			}

			// u ends the proto, but the proto was written entirely by
			// non-URL types
			if protoEndPos == 0 {
				if u[protoEndPos] == ':' {
					safe = isSafeURLProtocol(b.String()[:(b.Len()-len(u))+protoEndPos])
				}
				protoEnd = true
			}
			continue
		}

		s, err := Stringify(val)
		if err != nil {
			return "", false, err
		}

		if inQuery {
			escapeURLQuery(&b, s)
			continue
		}

		inQuery2, protoEndPos := normalizeURL(&b, URL(s))
		if !inQuery {
			inQuery = inQuery2
		}

		if !protoEnd && protoEndPos >= 0 {
			if s[protoEndPos] == ':' {
				safe = isSafeURLProtocol(b.String()[:(b.Len()-len(s))+protoEndPos])
			}
			protoEnd = true
		}
	}

	return URL(b.String()), safe, nil
}

func isSafeURLProtocol(protocol string) bool {
	if !strings.EqualFold(protocol, "http") && !strings.EqualFold(protocol, "https") &&
		!strings.EqualFold(protocol, "mailto") && !strings.EqualFold(protocol, "tel") {
		return false
	}
	return true
}

func NormalizeURL(u URL) URL {
	var sb strings.Builder
	normalizeURL(&sb, u)
	return URL(sb.String())
}

func normalizeURL(b *strings.Builder, u URL) (inQuery bool, protoEndPos int) {
	protoEndPos = -1
	b.Grow(len(u) + 16)
	// The byte loop below assumes that all URLs use UTF-8 as the
	// content-encoding. This is similar to the URI to IRI encoding scheme
	// defined in section 3.1 of  RFC 3987, and behaves the same as the
	// EcmaScript builtin encodeURIComponent.
	// It should not cause any misencoding of URLs in pages with
	// Content-type: text/html;charset=UTF-8.
	var written int
	for i, n := 0, len(u); i < n; i++ {
		c := u[i]
		switch c {
		// Single quote and parens are sub-delims in RFC 3986, but we
		// escape them so the output can be embedded in single
		// quoted attributes and unquoted CSS url(...) constructs.
		// Single quotes are reserved in URLs, but are only used in
		// the obsolete "mark" rule in an appendix in RFC 3986
		// so can be safely encoded.
		case ':', '/':
			if protoEndPos <= 0 {
				protoEndPos = i
			}
			continue
		case '!', '#', '$', '&', '*', '+', ',', ';', '=', '@', '[', ']':
			continue
		case '?':
			inQuery = true
			continue
		// Unreserved according to RFC 3986 sec 2.3
		// "For consistency, percent-encoded octets in the ranges of
		// ALPHA (%41-%5A and %61-%7A), DIGIT (%30-%39), hyphen (%2D),
		// period (%2E), underscore (%5F), or tilde (%7E) should not be
		// created by URI producers
		case '-', '.', '_', '~':
			continue
		case '%':
			// When normalizing do not re-encode valid escapes.
			if i+2 < len(u) && isHex(u[i+1]) && isHex(u[i+2]) {
				continue
			}
		default:
			// Unreserved according to RFC 3986 sec 2.3
			if 'a' <= c && c <= 'z' {
				continue
			}
			if 'A' <= c && c <= 'Z' {
				continue
			}
			if '0' <= c && c <= '9' {
				continue
			}
		}
		b.WriteString(string(u[written:i]))
		fmt.Fprintf(b, "%%%02x", c)
		written = i + 1
	}
	b.WriteString(string(u[written:]))
	return inQuery, protoEndPos
}

func escapeURLQuery(b *strings.Builder, s string) {
	b.Grow(len(s) + 16)
	// The byte loop below assumes that all URLs use UTF-8 as the
	// content-encoding. This is similar to the URI to IRI encoding scheme
	// defined in section 3.1 of  RFC 3987, and behaves the same as the
	// EcmaScript builtin encodeURIComponent.
	// It should not cause any misencoding of URLs in pages with
	// Content-type: text/html;charset=UTF-8.
	var written int
	for i, n := 0, len(s); i < n; i++ {
		c := s[i]
		// Unreserved according to RFC 3986 sec 2.3
		if 'a' <= c && c <= 'z' {
			continue
		}
		if 'A' <= c && c <= 'Z' {
			continue
		}
		if '0' <= c && c <= '9' {
			continue
		}
		b.WriteString(s[written:i])
		fmt.Fprintf(b, "%%%02x", c)
		written = i + 1
	}
	b.WriteString(s[written:])
}

// ============================================================================
// Srcset
// ======================================================================================

func FilterSrcset(vals ...any) (Srcset, error) {
	var b strings.Builder

	for _, val := range vals {
		var trusted bool

		var s string

		switch val := val.(type) {
		case Srcset:
			trusted = true
			s = htmlAttrValEscaper.Replace(string(val))
		case URL:
			// Normalizing gets rid of all HTML whitespace
			// which separate the image URL from its metadata.
			u := NormalizeURL(val)
			// Additionally, commas separate one source from another.
			s = strings.ReplaceAll(string(u), ",", "%2c")
			trusted = true
		default:
			var err error
			s, err = Stringify(val)
			if err != nil {
				return "", err
			}
		}

		if trusted {
			b.WriteString(s)
			continue
		}

		written := 0
		for i := 0; i < len(s); i++ {
			if s[i] == ',' {
				filterSrcsetElement(s, written, i, &b)
				b.WriteString(",")
				written = i + 1
			}
		}
		filterSrcsetElement(s, written, len(s), &b)
	}

	return Srcset(b.String()), nil
}

func urlProto(s string) string {
	proto, _, ok := strings.Cut(s, ":")
	if !ok || strings.Contains(proto, "/") {
		return ""
	}

	return proto
}

func filterSrcsetElement(s string, left int, right int, b *strings.Builder) {
	start := left
	for start < right && isHTMLSpace(s[start]) {
		start++
	}
	end := right
	for i := start; i < right; i++ {
		if isHTMLSpace(s[i]) {
			end = i
			break
		}
	}
	url := s[start:end]
	proto := urlProto(url)
	if proto == "" || isSafeURLProtocol(proto) {
		// If image metadata is only spaces or alnums then
		// we don't need to URL normalize it.
		metadataOk := true
		for i := end; i < right; i++ {
			if !isHTMLSpaceOrASCIIAlnum(s[i]) {
				metadataOk = false
				break
			}
		}
		if metadataOk {
			b.WriteString(s[left:start])
			normalizeURL(b, URL(url))
			b.WriteString(s[end:right])
			return
		}
	}
	b.WriteString("#")
	b.WriteString(UnsafeReplacement)
}

// Derived from https://play.golang.org/p/Dhmj7FORT5
const htmlSpaceAndASCIIAlnumBytes = "\x00\x36\x00\x00\x01\x00\xff\x03\xfe\xff\xff\x07\xfe\xff\xff\x07"

// isHTMLSpace is true iff c is a whitespace character per
// https://infra.spec.whatwg.org/#ascii-whitespace
func isHTMLSpace(c byte) bool {
	return (c <= 0x20) && 0 != (htmlSpaceAndASCIIAlnumBytes[c>>3]&(1<<uint(c&0x7)))
}

func isHTMLSpaceOrASCIIAlnum(c byte) bool {
	return (c < 0x80) && 0 != (htmlSpaceAndASCIIAlnumBytes[c>>3]&(1<<uint(c&0x7)))
}

// ============================================================================
// JS
// ======================================================================================

func EscapeJSAttrVal(val any) (JSAttrVal, error) {
	attrVal, ok := val.(JSAttrVal)
	if ok {
		return attrVal, nil
	}

	s, err := JSify(val)
	if err != nil {
		return "", err
	}

	return JSAttrVal(htmlAttrValEscaper.Replace(string(s))), nil
}

// EscapeJSStr escapes the stringified version of val so that it fulfills
// the conditions outlined in the [JSStr] doc.
func EscapeJSStr(val any) (JSStr, error) {
	s, err := Stringify(val)
	if err != nil {
		return "", err
	}

	js, err := JSify(s)
	if err != nil {
		return "", err
	}

	return JSStr(js[1 : len(js)-1]), nil
}

// isJSIdentPart reports whether the given rune is a JS identifier part.
// It does not handle all the non-Latin letters, joiners, and combining marks,
// but it does handle every codepoint that can occur in a numeric literal or
// a keyword.
func isJSIdentPart(r rune) bool {
	switch {
	case r == '$':
		return true
	case '0' <= r && r <= '9':
		return true
	case 'A' <= r && r <= 'Z':
		return true
	case r == '_':
		return true
	case 'a' <= r && r <= 'z':
		return true
	}
	return false
}
