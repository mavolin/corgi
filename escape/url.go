package escape

import (
	"fmt"
	"strings"

	"github.com/mavolin/corgi/escape/safe"
)

func URL(vals ...any) (safe.URLAttr, error) {
	s, err := escapeURL(vals, urlTypeURL, false)
	return safe.TrustedURLAttr(s), err
}

func FilterURL(vals ...any) (safe.URLAttr, error) {
	s, err := escapeURL(vals, urlTypeURL, true)
	return safe.TrustedURLAttr(s), err
}

func URLList(vals ...any) (safe.URLListAttr, error) {
	s, err := escapeURL(vals, urlTypeURLList, false)
	return safe.TrustedURLListAttr(s), err
}

func FilterURLList(vals ...any) (safe.URLListAttr, error) {
	s, err := escapeURL(vals, urlTypeURLList, true)
	return safe.TrustedURLListAttr(s), err
}

func FilterResourceURL(vals ...any) (safe.ResourceURLAttr, error) {
	s, err := escapeURL(vals, urlTypeResourceURL, true)
	return safe.TrustedResourceURLAttr(s), err
}

func NormalizeURL(u safe.URLAttr) (safe.URLAttr, error) {
	var sb strings.Builder
	normalizeURL(&sb, u)
	return URL(sb.String())
}

type urlType uint8

const (
	urlTypeURL urlType = iota + 1
	urlTypeResourceURL
	urlTypeURLList
)

func escapeURL(vals []any, t urlType) (u string, safe bool, err error) {
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

func normalizeURL(b *strings.Builder, u string) (inQuery bool, protoEndPos int) {
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
