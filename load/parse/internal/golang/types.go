package golang

import (
	"slices"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// https://go.dev/ref/spec#Types

func Type() parser.Func[*ast.Type] { // https://go.dev/ref/spec#Type
	return func(p *parser.Parser) *ast.Type {
		startIndex := p.Index()
		starPos := p.Pos()

		if parser.TryRune(p, '(') {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			t := parser.TryOptional(p, Type(), comment.OrHorizontalWhitespace())
			if t == nil {
				err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), ')')
				if err != nil {
					err.Message = "unexpected runes after opening parenthesis"
					p.CaptureError(err)
				}
			}

			if !parser.TryRune(p, ')') {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "missing closing parenthesis",
					Primary: quickanno.Expected(p, p.Pos(), "a closing parenthesis"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, starPos, "for the opening parenthesis here"),
					},
				})
			}

			return &ast.Type{
				Type:  p.AST.Raw[startIndex:p.Index()],
				From:  starPos,
				Until: p.Pos(),
			}
		}

		t := parser.Try(p, TypeLit())
		if t != nil {
			return t
		}

		namedType := parser.Try(p, NamedType())
		if namedType != nil {
			return &ast.Type{
				Type:   p.AST.Raw[startIndex:p.Index()],
				Parsed: namedType,
				From:   namedType.Start(),
				Until:  namedType.End(),
			}
		}

		return nil
	}
}

func TypeName() parser.Func[ast.FullIdentifier] { // https://go.dev/ref/spec#TypeName
	return FullIdent()
}

func TypeArgs() parser.Func[*ast.TypeArguments] { // https://go.dev/ref/spec#TypeArgs
	return func(p *parser.Parser) *ast.TypeArguments {
		l := parser.Try(p, list.BracketList("type argument", "type arguments", Type()))
		if l == nil {
			return nil
		}
		return &ast.TypeArguments{
			LBracket: l.Open,
			Types:    l.Elems,
			RBracket: l.Close,
		}
	}
}

func TypeLit() parser.Func[*ast.Type] { // https://go.dev/ref/spec#TypeLit
	return func(p *parser.Parser) *ast.Type {
		t := parser.TryInOrder(p, ArrayType(), StructType(), PointerType(), FunctionType(),
			InterfaceType(), SliceType(), MapType(), ChannelType())
		if t == nil {
			return nil
		}
		return t
	}
}

// ============================================================================
// Array types
// ======================================================================================

func ArrayType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#ArrayType
	return func(p *parser.Parser) *ast.Type {
		startIndex := p.Index()
		from := p.Pos()

		if !parser.TryRune(p, '[') {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		if !parser.Try(p, ArrayLength()) {
			return nil
		}

		var t ast.Type
		t.From = from

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		if !parser.TryRune(p, ']') {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "type: array: length: missing closing bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a closing bracket"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, t.Start(), "for the opening bracket here"),
				},
			})
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		elementType := parser.Try(p, ElementType())
		if elementType == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: array: missing element type",
				Primary:  quickanno.Expected(p, p.Pos(), "an element type"),
				Examples: []diagnostic.Example{{Example: "`[5]string`"}},
			})
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

// ArrayLength parses a simpler superset of the array length defined in the
// Go spec.
func ArrayLength() parser.Func[bool] { // https://go.dev/ref/spec#ArrayType
	return func(p *parser.Parser) bool {
		// This is quite possibly the worst code in the entirety of the parser,
		// so let me explain what's going on here:
		// Since we're not properly parsing the array length, but are rather
		// heuristically consuming everything up until a ']', it is possible
		// that the last runes we consume are whitespace.
		// This in itself isn't so bad, if we continue parsing afterward,
		// however, if the closing ']' is missing, the ArrayType() parser will
		// abort and exit with a type with trailing whitespace.
		// This is obviously not desired.
		//
		// Therefore, we need a little hack.
		// We parse the input string ourselves, and remember the last
		// non-whitespace rune we encounter.
		// We can then call parser.NextRune until we reach that rune.
		start := p.Index()
		var bracketCount int
		end := len(p.AST.Raw)
		for i, r := range p.AST.Raw[p.Index():] {
			if r == '[' { //nolint:gocritic
				bracketCount++
				end = p.Index() + i + len("[")
			} else if r == ']' {
				bracketCount--
				if bracketCount < 0 {
					end = p.Index() + i
					break
				}
				end = p.Index() + i + len("]")
			} else if r == ';' {
				break
			} else if !slices.Contains(whitespace.Runes, r) {
				end = p.Index() + i + utf8.RuneLen(r)
			}
		}
		if start == end {
			return false
		}
		for p.Index() < end {
			parser.NextRune(p)
		}

		return true
	}
}

func ElementType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#ElementType
	return func(p *parser.Parser) *ast.Type {
		t := parser.Try(p, Type())
		if t == nil {
			return nil
		}
		return t
	}
}

// ============================================================================
// Struct types
// ======================================================================================

// StructType parses a simpler superset of the struct type defined in the Go
// spec.
func StructType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#StructType
	return func(p *parser.Parser) *ast.Type {
		startIndex := p.Index()
		from := p.Pos()

		if !parser.TryToken(p, "struct") {
			return nil
		}

		hasWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		lBracePos := parser.TryRuneAt(p, '{')
		if lBracePos == nil {
			if !hasWS {
				return nil
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: struct: missing opening brace",
				Primary:  quickanno.Expected(p, p.Pos(), "an opening curly brace"),
				Examples: []diagnostic.Example{{Example: "`struct { Woof string }`"}},
			})
			return &ast.Type{
				Type:  p.AST.Raw[startIndex:p.Index()],
				From:  from,
				Until: p.Pos(),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var t ast.Type
		t.From = from

		// see array length on why this is necessary
		var braceCount int
		end := len(p.AST.Raw)
		for i, r := range p.AST.Raw[p.Index():] {
			if r == '{' { //nolint:gocritic
				braceCount++
				end = p.Index() + i + len("{")
			} else if r == '}' {
				braceCount--
				if braceCount < 0 {
					end = p.Index() + i
					break
				}
				end = p.Index() + i + len("}")
			} else if !slices.Contains(whitespace.Runes, r) {
				end = p.Index() + i + utf8.RuneLen(r)
			}
		}
		for p.Index() < end {
			parser.NextRune(p)
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		if !parser.TryRune(p, '}') {
			err := &diagnostic.Diagnostic{
				Message: "type: struct: missing closing brace",
				Primary: quickanno.Expected(p, p.Pos(), "a closing brace"),
			}
			if lBracePos != nil {
				err.Secondary = append(err.Secondary,
					anno.Position(p.File, *lBracePos, "for the opening brace here"))
			}
			p.CaptureError(err)
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

// ============================================================================
// Pointer types
// ======================================================================================

func PointerType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#PointerType
	return func(p *parser.Parser) *ast.Type {
		pos := p.Pos()
		startIndex := p.Index()

		if !parser.TryRune(p, '*') {
			return nil
		}

		var t ast.Type
		t.From = pos

		parser.TrySkip(p, comment.OrAnyWhitespace())

		baseType := parser.Try(p, BaseType())
		if baseType == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: pointer: missing base type",
				Primary:  quickanno.Expected(p, p.Pos(), "a base type"),
				Examples: []diagnostic.Example{{Example: "`*string`"}},
			})
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

func BaseType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#BaseType
	return func(p *parser.Parser) *ast.Type {
		t := parser.Try(p, Type())
		if t == nil {
			return nil
		}
		return t
	}
}

// ============================================================================
// Function types
// ======================================================================================

// FunctionType parses a simpler superset of the function type defined in the Go
// spec.
func FunctionType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#FunctionType
	return func(p *parser.Parser) *ast.Type {
		pos := p.Pos()
		startIndex := p.Index()

		if !parser.TryToken(p, "func") {
			return nil
		}

		var t ast.Type
		t.From = pos

		parser.TrySkip(p, comment.OrAnyWhitespace())
		signature := parser.Try(p, Signature())
		if !signature {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: function: missing signature",
				Primary:  quickanno.Expected(p, p.Pos(), "a function signature"),
				Examples: []diagnostic.Example{{Example: "`func(a, b int) string`"}},
			})
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

// Signature parses a simpler superset of the function signature defined in the
// Go spec.
func Signature() parser.Func[bool] { // https://go.dev/ref/spec#Signature
	return func(p *parser.Parser) bool {
		if !parser.Try(p, Parameters()) {
			return false
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		parser.Try(p, Result())
		return true
	}
}

// Result parses a simpler superset of the result defined in the Go spec.
func Result() parser.Func[bool] { // https://go.dev/ref/spec#Result
	return func(p *parser.Parser) bool {
		return parser.Try(p, Parameters()) || parser.Try(p, Type()) != nil
	}
}

// Parameters parses a simpler superset of the parameters defined in the Go
// spec.
// In particular, it allows variadic parameters everywhere.
func Parameters() parser.Func[bool] { // https://go.dev/ref/spec#Parameters
	return func(p *parser.Parser) bool {
		l := parser.Try(p, list.ParenList("parameter", "parameters", NamedParameterDecl()))
		if l != nil {
			return true
		}

		return parser.Try(p, list.ParenList("parameter", "parameters", UnnamedParameterDecl())) != nil
	}
}

func NamedParameterDecl() parser.Func[bool] { // https://go.dev/ref/spec#ParameterDecl
	return func(p *parser.Parser) bool {
		l := parser.Try(p, list.CommaList("parameter declaration", "parameter declarations", Identifier()))
		if l == nil {
			return false
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		if parser.TryOptionalToken(p, "...", nil) {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			typ := parser.Try(p, Type())
			if typ == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "parameter: missing type",
					Primary: quickanno.Expected(p, p.Pos(), "a type"),
				})
			}
		} else {
			parser.Try(p, Type())
		}

		return true
	}
}

func UnnamedParameterDecl() parser.Func[bool] { // https://go.dev/ref/spec#ParameterDecl
	return func(p *parser.Parser) bool {
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		parser.TryToken(p, "...")
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		return parser.Try(p, Type()) != nil
	}
}

// ============================================================================
// Interface types
// ======================================================================================

func InterfaceType() parser.Func[*ast.Type] {
	return func(p *parser.Parser) *ast.Type {
		from := p.Pos()
		startIndex := p.Index()

		if !parser.TryToken(p, "interface") {
			return nil
		}

		hasWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		lBracePos := parser.TryRuneAt(p, '{')
		if lBracePos == nil {
			if !hasWS {
				return nil
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: interface: missing opening brace",
				Primary:  quickanno.Expected(p, p.Pos(), "an opening curly brace"),
				Examples: []diagnostic.Example{{Example: "`interface { Foo() }`"}},
			})
			return &ast.Type{
				Type:  p.AST.Raw[startIndex:p.Index()],
				From:  from,
				Until: p.Pos(),
			}
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		var t ast.Type
		t.From = from

		// see array length on why this is necessary
		var braceCount int
		end := len(p.AST.Raw)
		for i, r := range p.AST.Raw[p.Index():] {
			if r == '{' { //nolint:gocritic
				braceCount++
				end = p.Index() + i + len("{")
			} else if r == '}' {
				braceCount--
				if braceCount < 0 {
					end = p.Index() + i
					break
				}
				end = p.Index() + i + len("}")
			} else if !slices.Contains(whitespace.Runes, r) {
				end = p.Index() + i + utf8.RuneLen(r)
			}
		}
		for p.Index() < end {
			parser.NextRune(p)
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		if !parser.TryRune(p, '}') {
			err := &diagnostic.Diagnostic{
				Message: "type: interface: missing closing brace",
				Primary: quickanno.Expected(p, p.Pos(), "a closing brace"),
			}
			if lBracePos != nil {
				err.Secondary = append(err.Secondary,
					anno.Position(p.File, *lBracePos, "for the opening brace here"))
			}
			p.CaptureError(err)
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

// ============================================================================
// Slice types
// ======================================================================================

func SliceType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#SliceType
	return func(p *parser.Parser) *ast.Type {
		pos := p.Pos()
		startIndex := p.Index()

		if !parser.TryRune(p, '[') {
			return nil
		}

		var t ast.Type
		t.From = pos

		parser.TrySkip(p, comment.OrAnyWhitespace())
		if !parser.TryRune(p, ']') {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "type: slice: missing closing bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a closing bracket"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, t.Start(), "for the opening bracket here"),
				},
			})
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		elemType := parser.Try(p, ElementType())
		if elemType == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: slice: missing element type",
				Primary:  quickanno.Expected(p, p.Pos(), "a type"),
				Examples: []diagnostic.Example{{Example: "`[]string`"}},
			})
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

// ============================================================================
// Map types
// ======================================================================================

func MapType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#MapType
	return func(p *parser.Parser) *ast.Type {
		pos := p.Pos()
		startIndex := p.Index()

		if !parser.TryToken(p, "map") {
			return nil
		}

		var t ast.Type
		t.From = pos

		parser.TrySkip(p, comment.OrAnyWhitespace())

		if parser.TryRune(p, '[') {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			keyType := parser.Try(p, KeyType())
			if keyType == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message:  "map: missing key type",
					Primary:  quickanno.Expected(p, p.Pos(), "a key type"),
					Examples: []diagnostic.Example{{Example: "`map[string]int`"}},
				})
			}

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if !parser.TryRune(p, ']') {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "map: missing closing bracket",
					Primary: quickanno.Expected(p, p.Pos(), "a closing bracket"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, t.Start(), "for the opening bracket here"),
					},
				})
			}
		} else {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "map: missing key",
				Primary:  quickanno.Expected(p, p.Pos(), "a key type enclosed in `[` and `]`"),
				Examples: []diagnostic.Example{{Example: "`map[string]int`"}},
			})
			t.Type = p.AST.Raw[startIndex:p.Index()]
			t.Until = p.Pos()
			return &t
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		elementType := parser.Try(p, ElementType())
		if elementType == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "map: missing element type",
				Primary:  quickanno.Expected(p, p.Pos(), "an element type"),
				Examples: []diagnostic.Example{{Example: "`map[string]int`"}},
			})
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

func KeyType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#KeyType
	return func(p *parser.Parser) *ast.Type {
		t := parser.Try(p, Type())
		if t == nil {
			return nil
		}
		return t
	}
}

// ============================================================================
// Channel types
// ======================================================================================

func ChannelType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#ChannelType
	return func(p *parser.Parser) *ast.Type {
		pos := p.Pos()
		startIndex := p.Index()

		if parser.TryToken(p, "<-") {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			if !parser.TryToken(p, "chan") {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "type: channel: missing `chan` keyword",
					Primary: quickanno.Expected(p, p.Pos(), "the keyword `chan`"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, p.Pos(), "considering, you already placed a `<-` here"),
					},
				})
			}
		} else {
			if !parser.TryToken(p, "chan") {
				return nil
			}

			parser.TrySkip(p, comment.OrAnyWhitespace())
			parser.TryToken(p, "<-")
		}

		var t ast.Type
		t.From = pos

		parser.TrySkip(p, comment.OrAnyWhitespace())
		elemType := parser.Try(p, ElementType())
		if elemType == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type: channel: missing element type",
				Primary:  quickanno.Expected(p, p.Pos(), "an element type"),
				Examples: []diagnostic.Example{{Example: "`chan string` or `chan<- int`"}},
			})
		}

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t
	}
}

// ============================================================================
// Parsed Types
// ======================================================================================

func NamedType() parser.Func[*ast.NamedType] { // essentially the first of https://go.dev/ref/spec#Type
	return func(p *parser.Parser) *ast.NamedType {
		name := parser.Try(p, TypeName())
		if name == nil {
			return nil
		}

		var t ast.NamedType
		t.Name = name

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		t.TypeArgs = parser.Try(p, TypeArgs())

		return &t
	}
}

// ============================================================================
// Type parameter declarations
// ======================================================================================

func TypeParameters() parser.Func[*ast.TypeParameters] {
	return func(p *parser.Parser) *ast.TypeParameters {
		l := parser.Try(p, list.BracketList("type parameter", "type parameters", TypeParameterDecl()))
		if l == nil {
			return nil
		}

		return &ast.TypeParameters{
			LBracket: l.Open,
			Params:   l.Elems,
			RBracket: l.Close,
		}
	}
}

func TypeParameterDecl() parser.Func[*ast.TypeParameter] {
	return func(p *parser.Parser) *ast.TypeParameter {
		names := parser.Try(p, list.CommaList("type parameter name", "type parameter names", Identifier()))
		if names == nil {
			return nil
		}

		var tp ast.TypeParameter
		tp.Names = names

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		pos := p.Pos()
		if t := parser.Try(p, Type()); t != nil {
			tp.Type = t // typed nil
		} else {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "type parameter: missing type",
				Primary:  quickanno.Expected(p, pos, "a type"),
				Examples: []diagnostic.Example{{Example: "`T any`"}},
			})
		}

		return &tp
	}
}
