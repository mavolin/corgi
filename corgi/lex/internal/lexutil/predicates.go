package lexutil

// IsElementName matches all runes that can be part of an element's name.
func IsElementName(r rune) bool {
	// c.f. https://html.spec.whatwg.org/multipage/syntax.html#syntax-tag-name
	return IsASCIIAlphanumeric(r)
}

// IsAttributeName matches all runes that formally can be part of an
// attribute's name.
func IsAttributeName(r rune) bool {
	// c.f. https://html.spec.whatwg.org/multipage/syntax.html#syntax-attributes
	return !IsControl(r) &&
		r != ' ' && r != '"' && r != '\'' && r != '>' && r != '`' && r != '=' &&
		!IsNoncharacter(r)

}

func IsNoncharacter(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#noncharacter
	return (r >= '\uFDD0' && r <= '\uFDEF') || r == '\uFFFE' || r == '\uFFFF' ||
		r == '\U0001FFFE' || r == '\U0001FFFF' || r == '\U0002FFFE' ||
		r == '\U0002FFFF' || r == '\U0003FFFE' || r == '\U0003FFFF' ||
		r == '\U0004FFFE' || r == '\U0004FFFF' || r == '\U0005FFFE' ||
		r == '\U0005FFFF' || r == '\U0006FFFE' || r == '\U0006FFFF' ||
		r == '\U0007FFFE' || r == '\U0007FFFF' || r == '\U0008FFFE' ||
		r == '\U0008FFFF' || r == '\U0009FFFE' || r == '\U0009FFFF' ||
		r == '\U000AFFFE' || r == '\U000AFFFF' || r == '\U000BFFFE' ||
		r == '\U000BFFFF' || r == '\U000CFFFE' || r == '\U000CFFFF' ||
		r == '\U000DFFFE' || r == '\U000DFFFF' || r == '\U000EFFFE' ||
		r == '\U000EFFFF' || r == '\U000FFFFE' || r == '\U000FFFFF' ||
		r == '\U0010FFFE' || r == '\U0010FFFF'
}

func IsC0Control(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#c0-control
	return r >= '\u0000' && r <= '\u001F'
}

func IsControl(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#control
	return IsC0Control(r) || (r >= '\u007F' && r <= '\u009F')
}

func IsASCIIDigit(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#ascii-digit
	return r >= '0' && r <= '9'
}

func IsASCIIUpperAlpha(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#ascii-upper-alpha
	return r >= 'A' && r <= 'Z'
}

func IsASCIILowerAlpha(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#ascii-lower-alpha
	return r >= 'a' && r <= 'z'
}

func IsASCIIAlpha(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#ascii-alpha
	return IsASCIIUpperAlpha(r) || IsASCIILowerAlpha(r)
}

func IsASCIIAlphanumeric(r rune) bool {
	// c.f. https://infra.spec.whatwg.org/#ascii-alphanumeric
	return IsASCIIAlpha(r) || IsASCIIDigit(r)
}
