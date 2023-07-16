package typeinfer

import (
	"regexp"
	"strings"

	"github.com/mavolin/corgi/file"
)

// Infer attempts to infer the type expr would yield.
//
// Infer can detect int, float, bool, rune, string, and composite literals, as
// well as the above wrapped in ternary expressions
//
// If Infer returns an empty string, it could not identify the type.
func Infer(expr file.Expression) string {
	if len(expr.Expressions) == 0 {
		return ""
	}

	if t := inferLit(expr); t != "" {
		return t
	} else if t := inferTypeAssertion(expr); t != "" {
		return t
	} else if t := inferTernaryExpression(expr); t != "" {
		return t
	}

	return ""
}

func inferLit(expr file.Expression) string {
	if t := inferPrimitiveLit(expr); t != "" {
		return t
	} else if t := inferCompositeLit(expr); t != "" {
		return t
	}

	return ""
}

var (
	// https://go.dev/ref/spec#int_lit
	numLitRegexp = regexp.MustCompile(`(?i)^(?:0|[1-9](?:_?\d)*|0b(?:_?[01])+|0o?(?:_?[0-8])+|0x[_\da-f]+)`)
	// https://go.dev/ref/spec#float_lit
	floatLitRegexp = regexp.MustCompile(`(?i)^(?:` +
		`\d(?:_?\d)*\.(?:\d(?:_?\d)*)?(?:e[+-]?\d(?:_?\d)*)?|` +
		`\d(?:_?\d)*e[+-]?\d(?:_?\d)*|` +
		`\.\d(?:_?\d)*(?:e[+-]?\d(?:_?\d)*)?|` +
		`0x(?:(?:_?[\da-f])+(?:\.(?:[\da-f](?:_?[\da-f])*)?)?|\.[\da-f](?:_?[\da-f])*)p[+-]?\d(?:_?\d)*` +
		`)`)
)

func inferPrimitiveLit(expr file.Expression) string {
	if _, ok := expr.Expressions[0].(file.StringExpression); ok {
		return "string"
	}

	gexpr, ok := expr.Expressions[0].(file.GoExpression)
	if !ok {
		return ""
	}

	e := gexpr.Expression
	if len(e) == 0 {
		return ""
	}

	// using the first rune, we can narrow down the possible types, so we don't
	// need to run all regexps
	switch e[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		if numLitRegexp.MatchString(e) {
			return "int"
		} else if floatLitRegexp.MatchString(e) {
			return "float64"
		}
		return ""
	case 't', 'f':
		if strings.HasPrefix(e, "true") || strings.HasPrefix(e, "false") {
			return "bool"
		}
		return ""
	case '\'':
		return "rune"
	default:
		return ""
	}
}

var compositeLitRegexp = regexp.MustCompile(`(?i)^(?:` +
	`\[ *(?:\d(?:_?\d)+)? *] *[a-z0-9_]+(?: *\. *[a-z0-9_]+)?|` + // slice/array
	`map\[ *[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *] *[a-z0-9_]+(?: *\. *[a-z0-9_]+)?|` + // map
	`[a-z0-9_]+(?: *\. *[a-z0-9_]+)?` + // struct/named array/named map
	`)`)

func inferCompositeLit(expr file.Expression) string {
	gexpr, ok := expr.Expressions[0].(file.GoExpression)
	if !ok {
		return ""
	}

	e := gexpr.Expression
	t := compositeLitRegexp.FindString(e)

	e = strings.TrimLeft(e[len(t):], " \t")
	if len(e) == 0 || e[0] != '{' {
		return ""
	}

	return t
}

var typeAssertionRegexp = regexp.MustCompile(`(?i)\. *\(([^)]+)\)$`)

func inferTypeAssertion(expr file.Expression) string {
	gexpr, ok := expr.Expressions[len(expr.Expressions)-1].(file.GoExpression)
	if !ok {
		return ""
	}

	e := gexpr.Expression
	t := typeAssertionRegexp.FindStringSubmatch(e)

	if len(t) != 2 {
		return ""
	}

	return t[1]
}

func inferTernaryExpression(expr file.Expression) string {
	tern, ok := expr.Expressions[0].(file.TernaryExpression)
	if !ok {
		return ""
	}

	if t := Infer(tern.IfTrue); t != "" {
		return t
	}

	return Infer(tern.IfFalse)
}
