package golang

import "unicode"

func Newline(r rune) bool {
	return r == '\n'
}

func Unicode_Char(r rune) bool { //nolint:revive
	return !Newline(r)
}

func Unicode_Letter(r rune) bool { //nolint:revive
	return unicode.IsLetter(r)
}

func Unicode_Digit(r rune) bool { //nolint:revive
	return unicode.IsDigit(r)
}

func Letter(r rune) bool {
	return Unicode_Letter(r) || r == '_'
}

func Decimal_Digit(r rune) bool { //nolint:revive
	return r >= '0' && r <= '9'
}

func Binary_Digit(r rune) bool { //nolint:revive
	return r == '0' || r == '1'
}

func Octal_Digit(r rune) bool { //nolint:revive
	return r >= '0' && r <= '7'
}

func Hex_Digit(r rune) bool { //nolint:revive
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}
