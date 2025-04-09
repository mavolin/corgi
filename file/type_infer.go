package file

import (
	"regexp"
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
)

// InferType attempts to infer the type expr would yield.
// It can detect int, float, bool, rune, string, composite literals
// and type assertions.
//
// If InferType returns the empty string, it could not identify the type.
func InferType(f *File, expr *ast.Expression) string {
	typ, _ := inferType(f, expr)
	return typ
}

func inferType(f *File, expr *ast.Expression) (typ string, sure bool) {
	if expr == nil {
		return "", false
	} else if len(expr.Code) == 0 {
		return "", false
	}

	switch n := expr.Code[0].(type) {
	case *ast.BlockFunction:
		return "bool", true
	case *ast.GoCode:
		return inferGoCodeType(f, n)
	case *ast.String:
		return "string", true
	case *ast.Ternary:
		return inferTernaryType(f, n)
	case *ast.ZeroCoalescing:
		return inferZeroCoalescingType(f, n)
	}

	if len(expr.Code) == 1 {
		return "", false
	}

	switch n := expr.Code[len(expr.Code)-1].(type) {
	case *ast.BlockFunction:
		return "bool", true
	case *ast.GoCode:
		return inferLastGoCodeType(n)
	case *ast.String:
		return "string", true
	case *ast.Ternary:
		return inferTernaryType(nil, n)
	default:
		return "", false
	}
}

func inferTernaryType(f *File, expr *ast.Ternary) (typ string, sure bool) {
	if expr == nil || (expr.TrueVal == nil && expr.FalseVal == nil) {
		return "", false
	}

	trueType, trueSure := inferType(f, expr.TrueVal)
	if trueSure {
		return trueType, trueSure
	}

	falseType, falseSure := inferType(f, expr.FalseVal)
	if falseSure {
		return falseType, falseSure
	}

	if trueType == falseType {
		return trueType, false
	}

	if expr.TrueVal == nil {
		return falseType, falseSure
	} else if expr.FalseVal == nil {
		return trueType, trueSure
	}
	return "", false
}

func inferZeroCoalescingType(f *File, expr *ast.ZeroCoalescing) (typ string, sure bool) {
	if expr == nil {
		return "", false
	}

	if len(expr.Chain) == 0 {
		if expr.Default != nil {
			return inferType(f, expr.Default)
		}
		return "", false
	}

	last := expr.Chain[len(expr.Chain)-1]
	ta, _ := last.(*ast.ZCTypeAssertionExpression)
	if ta != nil {
		return ta.Type.Full(), true
	}

	if expr.Default != nil {
		return InferType(f, expr.Default), false
	}
	return "", false
}

func inferGoCodeType(f *File, expr *ast.GoCode) (typ string, sure bool) {
	if expr == nil {
		return "", false
	}

	if t := inferStateVariableType(f, expr); t != "" {
		return t, true
	} else if t = inferLit(expr); t != "" {
		return t, false
	} else if t = inferMakeNewType(expr); t != "" {
		return t, true
	}

	return "", false
}

func inferLit(expr *ast.GoCode) string {
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

func inferPrimitiveLit(expr *ast.GoCode) string {
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

func inferCompositeLit(expr *ast.GoCode) string {
	c := expr.Code
	t := compositeLitRegexp.FindString(c)
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

func inferMakeNewType(expr *ast.GoCode) string {
	c := expr.Code
	t := makeRegexp.FindStringSubmatch(c)
	if len(t) != 2 {
		return ""
	}

	return t[1]
}

var stateRegexp = regexp.MustCompile(`^state[ \t]*\.\s*([a-zA-Z_][a-zA-Z0-9_]*)`)

func inferStateVariableType(f *File, expr *ast.GoCode) string {
	c := expr.Code
	t := stateRegexp.FindStringSubmatch(c)
	if len(t) != 2 {
		return ""
	}

	name := t[1]
	state := f.Package.StateByName(name)
	if state == nil {
		return ""
	}

	return state.ResolvedType()
}

func inferLastGoCodeType(expr *ast.GoCode) (typ string, sure bool) {
	if expr == nil {
		return "", false
	}

	if t := inferTypeAssertion(expr); t != "" {
		return t, true
	}

	return "", false
}

var typeAssertionRegexp = regexp.MustCompile(`(?i)\. *\(([^)]+)\)$`)

func inferTypeAssertion(expr *ast.GoCode) string {
	c := expr.Code
	t := typeAssertionRegexp.FindStringSubmatch(c)

	if len(t) != 2 {
		return ""
	}

	return t[1]
}
