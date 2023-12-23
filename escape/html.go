package escape

import (
	"strings"

	"github.com/mavolin/corgi/escape/safe"
)

var (
	htmlEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
	)
	plainAttrToHTMLEscaper = strings.NewReplacer(`<`, "&lt;")
)

// HTML replaces [&<] with escape sequences.
//
// It should only be used to escape the content of an element's body.
func HTML(val any) (safe.HTML, error) {
	switch val := val.(type) {
	case safe.HTML:
		return val, nil
	case safe.PlainAttr:
		return safe.TrustedHTML(plainAttrToHTMLEscaper.Replace(val.Escaped())), nil
	}

	s, err := stringify(val, htmlEscaper.Replace)
	return safe.TrustedHTML(s), err
}

var (
	plainAttrEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`"`, "&#34;",
	)
	htmlToPlainAttrEscaper = strings.NewReplacer(`<`, "&lt;")
)

// PlainAttr escapes s by replacing [&"] so that it can safely be placed between
// double quotes as an attribute value.
//
// It does not perform further context specific escaping.
func PlainAttr(val any) (safe.PlainAttr, error) {
	switch val := val.(type) {
	case safe.PlainAttr:
		return val, nil
	case safe.HTML:
		return safe.TrustedPlainAttr(htmlToPlainAttrEscaper.Replace(val.Escaped())), nil
	}

	s, err := stringify(val, plainAttrEscaper.Replace)
	return safe.TrustedPlainAttr(s), err
}
