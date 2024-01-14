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
// If Infer returns the empty string, it could not identify the type.
func Infer(expr file.Expression) string {
	if expr == nil {
		return ""
	}

	switch expr := expr.(type) {
	case file.GoCode:
		return inferGoCode(expr)
	case file.ChainExpression:
		return inferChainExpression(expr)
	}

	return ""
}

func inferChainExpression(expr file.ChainExpression) string {
	if t := inferChainExpressionChain(expr); t != "" {
		return t
	}

	if expr.Default == nil {
		return ""
	}

	return inferGoCode(*expr.Default)
}

func inferChainExpressionChain(expr file.ChainExpression) string {
	if len(expr.Chain) == 0 {
		return ""
	}

	last := expr.Chain[len(expr.Chain)-1]
	ta, ok := last.(file.TypeAssertionExpression)
	if !ok {
		return ""
	}

	return ta.Type.String()
}

func inferGoCode(expr file.GoCode) string {
	if t := inferLit(expr); t != "" {
		return t
	} else if t := inferTypeAssertion(expr); t != "" {
		return t
	}

	return ""
}

func inferLit(expr file.GoCode) string {
	if t := inferPrimitiveLit(expr); t != "" {
		return t
	} else if t := inferCompositeLit(expr); t != "" {
		return t
	}

	return ""
}

var (
	// https://go.dev/ref/spec#int_lit
	numLitRegexp = regexp.MustCompile(`(?i)^[+-]?(?:0|[1-9](?:_?\d)*|0b(?:_?[01])+|0o?(?:_?[0-8])+|0x[_\da-f]+)`)
	// https://go.dev/ref/spec#float_lit
	floatLitRegexp = regexp.MustCompile(`(?i)^[+-]?(?:` +
		`\d(?:_?\d)*\.(?:\d(?:_?\d)*)?(?:e[+-]?\d(?:_?\d)*)?|` +
		`\d(?:_?\d)*e[+-]?\d(?:_?\d)*|` +
		`\.\d(?:_?\d)*(?:e[+-]?\d(?:_?\d)*)?|` +
		`0x(?:(?:_?[\da-f])+(?:\.(?:[\da-f](?:_?[\da-f])*)?)?|\.[\da-f](?:_?[\da-f])*)p[+-]?\d(?:_?\d)*` +
		`)`)
)

func inferPrimitiveLit(expr file.GoCode) string {
	if _, ok := expr.Expressions[0].(file.String); ok {
		return "string"
	}

	gexpr, ok := expr.Expressions[0].(file.RawGoCode)
	if !ok {
		return ""
	}

	c := gexpr.Code
	if len(c) == 0 {
		return ""
	}

	// using the first rune, we can narrow down the possible types, so we don't
	// need to run all regexps
	switch c[0] {
	case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		if numLitRegexp.MatchString(c) {
			return "int"
		} else if floatLitRegexp.MatchString(c) {
			return "float64"
		}
		return ""
	case 't', 'f':
		if strings.HasPrefix(c, "true") || strings.HasPrefix(c, "false") {
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

func inferCompositeLit(expr file.GoCode) string {
	rgc, ok := expr.Expressions[0].(file.RawGoCode)
	if !ok {
		return ""
	}

	c := rgc.Code
	t := compositeLitRegexp.FindString(c)

	c = strings.TrimLeft(c[len(t):], " \t")
	if len(c) == 0 || c[0] != '{' {
		return ""
	}

	return t
}

var typeAssertionRegexp = regexp.MustCompile(`(?i)\. *\(([^)]+)\)$`)

func inferTypeAssertion(expr file.GoCode) string {
	rgc, ok := expr.Expressions[len(expr.Expressions)-1].(file.RawGoCode)
	if !ok {
		return ""
	}

	c := rgc.Code
	t := typeAssertionRegexp.FindStringSubmatch(c)

	if len(t) != 2 {
		return ""
	}

	return t[1]
}
