package analyze

import (
	"regexp"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/switches"
)

// InferType attempts to infer the type expr would yield.
// It can detect int, float, bool, rune, string, composite literals
// and type assertions.
//
// If InferType returns the empty string, it could not identify the type.
//
// The “exact” return value indicates if this prediction is certain, i.e. that
// the expression will exactly yield the type returned.
// If “exact” is false, InferType encountered an untyped literal:
// While the expression can be cast to the type returned, in the context
// surrounding the expression it might also be used to yield a different, more
// concrete type.
func InferType(f *file.File, expr *ast.Expression) (typ file.Type, exact bool) {
	if expr == nil || len(expr.Nodes) == 0 {
		return "", false
	}

	switches.CodeNode(expr.Nodes[0],
		func(*ast.BlockFunction) { typ, exact = "bool", true },
		func(*ast.ComponentCall) { typ, exact = "string", true },
		func(gc *ast.GoCode) { typ, exact = inferGoCodeType(gc) },
		func(*ast.InterpretedString) { typ, exact = "string", false },
		func(*ast.RawString) { typ, exact = "string", false },
		func(n *ast.Ternary) {
			if len(expr.Nodes) == 1 {
				typ, exact = inferTernaryType(f, n)
			}
		},
		func(n *ast.ZeroCoalescing) { typ, exact = inferZeroCoalescingType(f, n) })
	if typ != "" && exact || len(expr.Nodes) == 1 {
		return typ, exact
	}

	switches.CodeNode(expr.Nodes[len(expr.Nodes)-1],
		func(*ast.BlockFunction) { typ, exact = "bool", true },
		func(*ast.ComponentCall) { typ, exact = "string", true },
		func(gc *ast.GoCode) { typ, exact = inferLastGoCodeType(gc) },
		func(*ast.InterpretedString) { typ, exact = "string", false },
		func(*ast.RawString) { typ, exact = "string", false },
		func(n *ast.Ternary) { typ, exact = inferTernaryType(f, n) },
		func(*ast.ZeroCoalescing) {})
	return typ, exact
}

func inferTernaryType(f *file.File, expr *ast.Ternary) (typ file.Type, exact bool) {
	if expr == nil || (expr.TrueVal == nil && expr.FalseVal == nil) {
		return "", false
	}

	trueType, trueExact := InferType(f, expr.TrueVal)
	if trueExact {
		return trueType, trueExact
	}

	falseType, falseExact := InferType(f, expr.FalseVal)
	if falseExact {
		return falseType, true
	}

	switch {
	case trueType == falseType:
		return trueType, false
	case expr.TrueVal == nil:
		return falseType, falseExact
	case expr.FalseVal == nil:
		return trueType, trueExact
	}
	return "", false
}

func inferZeroCoalescingType(f *file.File, expr *ast.ZeroCoalescing) (typ file.Type, exact bool) {
	if expr == nil {
		return "", false
	}

	if len(expr.Chain) == 0 {
		if expr.Default != nil {
			return InferType(f, expr.Default)
		}
		return "", false
	}

	last := expr.Chain[len(expr.Chain)-1]
	switches.ZeroCoalescingNode(last,
		func(*ast.ZCIndexExpression) {},
		func(*ast.ZCParenExpression) {},
		func(*ast.ZCSelectorExpression) {},
		func(e *ast.ZCTypeAssertionExpression) { typ, exact = file.Type(e.Type.Full()), true },
	)
	if typ != "" {
		return typ, exact
	}

	if expr.Default != nil {
		return InferType(f, expr.Default)
	}
	return "", false
}

func inferGoCodeType(expr *ast.GoCode) (typ file.Type, exact bool) {
	if expr == nil {
		return "", false
	}

	if t := inferBooleanType(expr); t != "" {
		return t, true
	} else if t = inferLit(expr); t != "" {
		return t, false
	} else if t = inferMakeNewType(expr); t != "" {
		return t, true
	}

	return "", false
}

func inferBooleanType(expr *ast.GoCode) file.Type {
	if expr == nil {
		return ""
	}

	c := expr.Code
	switch {
	case c == "true" || c == "false":
		return "bool"
	case strings.HasPrefix(c, "true ") || strings.HasPrefix(c, "false "):
		return "bool"
	case strings.HasPrefix(c, "!"):
		return "bool"
	}

	var parenCount int
	for i, r := range c {
		switch r {
		case '(', '[', '{':
			parenCount++
		case ')', ']', '}':
			parenCount--
		case '!', '=', '<', '>':
			if parenCount == 0 && i+1 < len(c) && c[i+1] == '=' {
				return "bool"
			}
		case '&':
			if parenCount == 0 && i+1 < len(c) && c[i+1] == '&' {
				return "bool"
			}
		case '|':
			if parenCount == 0 && i+1 < len(c) && c[i+1] == '|' {
				return "bool"
			}
		}
	}

	return ""
}

func inferLit(expr *ast.GoCode) file.Type {
	if t := inferPrimitiveLit(expr); t != "" {
		return t
	} else if t = inferCompositeLit(expr); t != "" {
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

func inferPrimitiveLit(expr *ast.GoCode) file.Type {
	c := expr.Code
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
	`\[ *(?:\d(?:_?\d)+)? *] *[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *(?:\[\s+[^[\]]+])?|` + // slice/array
	`map\[ *[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *] *[a-z0-9_]+(?: *\. *[a-z0-9_]+)?|` + // map
	`[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *(?:\[\s+[^[\]]+])?` + // named type
	`)`)

func inferCompositeLit(expr *ast.GoCode) file.Type {
	c := expr.Code
	t := file.Type(compositeLitRegexp.FindString(c))
	if t == "struct" || t == "interface" {
		return ""
	}

	c = strings.TrimLeft(c[len(t):], " \t")
	if len(c) == 0 || c[0] != '{' {
		return ""
	}

	return t
}

var makeRegexp = regexp.MustCompile(`(?i)^(?:make|new)\s*\(\s*(?:` +
	`\[ *(?:\d(?:_?\d)+)? *] *[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *(?:\[\s*[^[\]]+])?|` + // slice/array
	`map\[ *[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *] *[a-z0-9_]+(?: *\. *[a-z0-9_]+)?|` + // map
	`[a-z0-9_]+(?: *\. *[a-z0-9_]+)? *(?:\[\s*[^[\]]+])?` + // named type
	`)`)

func inferMakeNewType(expr *ast.GoCode) file.Type {
	c := expr.Code
	t := makeRegexp.FindStringSubmatch(c)
	if len(t) != 2 {
		return ""
	}

	return file.Type(t[1])
}

func inferLastGoCodeType(expr *ast.GoCode) (typ file.Type, exact bool) {
	if expr == nil {
		return "", false
	}

	if t := inferTypeAssertion(expr); t != "" {
		return t, true
	}

	return "", false
}

var typeAssertionRegexp = regexp.MustCompile(`(?i)\. *\(([^)]+)\)$`)

func inferTypeAssertion(expr *ast.GoCode) file.Type {
	c := expr.Code
	t := typeAssertionRegexp.FindStringSubmatch(c)

	if len(t) != 2 {
		return ""
	}

	return file.Type(t[1])
}
