package escapelite

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"unicode/utf8"
)

// CSSFloat formats f as a string suitable for use as a CSS floating point
// production.
//
// See:
// https://www.w3.org/TR/css-values-4/#number
// https://www.w3.org/TR/css-syntax-3/#number-token-diagram
func CSSFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// CSSInt formats n as a string suitable for use as a CSS integer production.
//
// See: https://www.w3.org/TR/css-values-4/#integer
func CSSInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

// CSSUint formats n as a string suitable for use as a CSS integer production.
//
// See: https://www.w3.org/TR/css-values-4/#integer
func CSSUint(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// CSSColor formats the given [color.Color] into a CSS hex color string.
//
// See: https://www.w3.org/TR/css-color-4/#typedef-hex-color
func CSSColor(c color.Color) string {
	nrgba := color.NRGBAModel.Convert(c).(color.NRGBA) //nolint:errcheck
	if nrgba.A == 255 {
		// Opaque color
		return fmt.Sprintf("#%02x%02x%02x", nrgba.R, nrgba.G, nrgba.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", nrgba.R, nrgba.G, nrgba.B, nrgba.A)
}

// CSSString formats s as a CSS [string literal], escaping characters
// as necessary.
//
// See:
// https://www.w3.org/TR/css-syntax-3/#string-token-diagram
// https://www.w3.org/TR/css-syntax-3/#consume-string-token
func CSSString(s string) string {
	var b strings.Builder
	b.Grow(len(`"`) + len(s) + len(`"`))
	b.WriteByte('"')

	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '<': // so we comply with safe.CSSValue's contract
			writeHexEscape(&b, '<') // \00003C
		case '\n', '\r', '\f':
			writeHexEscape(&b, r)
		case 0: // NUL -> U+FFFD
			writeHexEscape(&b, '\uFFFD')
		default:
			// Escape other C0 controls and DEL
			if r < 0x20 || r == 0x7F {
				writeHexEscape(&b, r)
			} else {
				// Write as-is (UTF-8)
				b.WriteRune(r)
			}
		}
	}

	b.WriteByte('"')
	return b.String()
}

// writeHexEscape writes a 6-digit CSS hex escape for rune r: \XXXXXX
// Using 6 digits avoids needing a terminating space even if the next
// character is a hex digit.
func writeHexEscape(b *strings.Builder, r rune) {
	// Ensure r is a valid Unicode scalar value
	if !utf8.ValidRune(r) {
		r = '\uFFFD'
	}
	fmt.Fprintf(b, `\%06X`, r)
}
