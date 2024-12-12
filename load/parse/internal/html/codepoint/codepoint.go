// Package codepoint provides functions to match runes against HTML spec code
// points.
package codepoint

import (
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func MatchesAny(p *parser.Parser, fs ...func(rune) bool) bool {
	return parser.MatchesRunePredicate(p, func(r rune) bool {
		for _, f := range fs {
			if f(r) {
				return true
			}
		}
		return false
	})
}

func LeadingSurrogate(r rune) bool { // https://infra.spec.whatwg.org/#leading-surrogate
	return r >= 0xd800 && r <= 0xdbff
}

func TrailingSurrogate(r rune) bool { // https://infra.spec.whatwg.org/#trailing-surrogate
	return r >= 0xdc00 && r <= 0xdfff
}

func Surrogate(r rune) bool { // https://infra.spec.whatwg.org/#surrogate
	return LeadingSurrogate(r) || TrailingSurrogate(r)
}

func ScalarValue(r rune) bool { // https://infra.spec.whatwg.org/#scalar-value
	return !Surrogate(r)
}

func Noncharacter(r rune) bool { // https://infra.spec.whatwg.org/#noncharacter
	return (r >= 0xfdd0 && r <= 0xfdef) || (r&0xfffe == 0xfffe && r >= 0xfffe && r <= 0x10ffff)
}

func ASCII(r rune) bool { // https://infra.spec.whatwg.org/#ascii
	return r >= 0 && r <= 0x7f
}

func ASCIITabOrNewline(r rune) bool { // https://infra.spec.whatwg.org/#ascii-tab-or-newline
	return r == 0x9 || r == 0xa || r == 0xd
}

func ASCIIWhitespace(r rune) bool { // https://infra.spec.whatwg.org/#ascii-whitespace
	return r == 0x9 || r == 0xa || r == 0xc || r == 0xd || r == 0x20
}

func C0Control(r rune) bool { // https://infra.spec.whatwg.org/#c0-control
	return r >= 0 && r <= 0x1f
}

func C0ControlOrSpace(r rune) bool { // https://infra.spec.whatwg.org/#c0-control-or-space
	return r >= 0 && r <= 0x20
}

func Control(r rune) bool {
	return C0Control(r) || (r >= 0x7f && r <= 0x9f)
}

func ASCIIDigit(r rune) bool { // https://infra.spec.whatwg.org/#ascii-digit
	return r >= '0' && r <= '9'
}

func ASCIIUpperHexDigit(r rune) bool { // https://infra.spec.whatwg.org/#ascii-upper-hex-digit
	return ASCIIDigit(r) || (r >= 'A' && r <= 'F')
}

func ASCIILowerHexDigit(r rune) bool { // https://infra.spec.whatwg.org/#ascii-lower-hex-digit
	return ASCIIDigit(r) || (r >= 'a' && r <= 'f')
}

func ASCIIHexDigit(r rune) bool { // https://infra.spec.whatwg.org/#ascii-hex-digit
	return ASCIIDigit(r) || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

func ASCIIUpperAlpha(r rune) bool { // https://infra.spec.whatwg.org/#ascii-upper-alpha
	return r >= 'A' && r <= 'Z'
}

func ASCIILowerAlpha(r rune) bool { // https://infra.spec.whatwg.org/#ascii-lower-alpha
	return r >= 'a' && r <= 'z'
}

func ASCIIAlpha(r rune) bool { // https://infra.spec.whatwg.org/#ascii-alpha
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func ASCIIAlphanumeric(r rune) bool { // https://infra.spec.whatwg.org/#ascii-alphanumeric
	return ASCIIAlpha(r) || ASCIIDigit(r)
}
