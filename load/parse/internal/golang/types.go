package golang

import (
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
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
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
			}, nil
		}

		t := parser.Try(p, TypeLit())
		if t != nil {
			return t, nil
		}

		namedType := parser.Try(p, NamedType())
		if namedType != nil {
			return &ast.Type{
				Type:   p.AST.Raw[startIndex:p.Index()],
				Parsed: namedType,
				From:   namedType.Start(),
				Until:  namedType.End(),
			}, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message:  "missing type",
			Primary:  quickanno.Expected(p, p.Pos(), "a type"),
			Examples: []diagnostic.Example{{Example: "`int`, `[]string`, or `woof.Bark`"}},
		}
	}
}

func TypeName() parser.Func[ast.FullIdent] { // https://go.dev/ref/spec#TypeName
	return FullIdent()
}

func TypeArgs() parser.Func[*ast.TypeArguments] { // https://go.dev/ref/spec#TypeArgs
	return func(p *parser.Parser) (*ast.TypeArguments, *diagnostic.Diagnostic) {
		l, err := parser.TryErr(p, list.BracketList("type arguments", Type()))
		if err != nil {
			return nil, err
		}
		return &ast.TypeArguments{
			LBracket: l.Open,
			Types:    l.Elems,
			RBracket: l.Close,
		}, nil
	}
}

func TypeLit() parser.Func[*ast.Type] { // https://go.dev/ref/spec#TypeLit
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := parser.TryInOrder(p, ArrayType(), StructType(), PointerType(), FunctionType(),
			InterfaceType(), SliceType(), MapType(), ChannelType())
		if t == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing type",
				Primary:  quickanno.Expected(p, p.Pos(), "a type literal"),
				Examples: []diagnostic.Example{{Example: "`[]string` or `func(string) bool`"}},
			}
		}

		return t, nil
	}
}

// ============================================================================
// Array types
// ======================================================================================

func ArrayType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#ArrayType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		var t ast.Type
		t.From = p.Pos()

		startIndex := p.Index()
		if !parser.TryRune(p, '[') {
			return nil, &diagnostic.Diagnostic{
				Message: "missing array",
				Primary: quickanno.Expected(p, p.Pos(), "an opening bracket"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		l := parser.Try(p, ArrayLength())
		if l == "" {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing array",
				Primary:  quickanno.Expected(p, p.Pos(), "an array length"),
				Examples: []diagnostic.Example{{Example: "`[5]string`"}},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		if !parser.TryRune(p, ']') {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "array length: missing closing bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a closing bracket"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, t.Start(), "for the opening bracket here"),
				},
			})
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		parser.Must(p, ElementType())

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return &t, nil
	}
}

// ArrayLength parses a simpler superset of the array length defined in the
// Go spec.
func ArrayLength() parser.Func[string] { // https://go.dev/ref/spec#ArrayType
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
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
		// Before consuming any runes, we make a clone of the state.
		// We capture the ArrayLength like we normally (heuristically) would.
		// When done, we inspect the captured string and trim right-outer
		// whitespace.
		// we then restore the clone, and capture the token without the
		// whitespace
		var bracketCount int
		state := p.CloneState()
		s := parser.TokenWhile(p, func() bool {
			if parser.MatchesToken(p, "[") {
				bracketCount++
				return true
			} else if parser.MatchesToken(p, "]") {
				bracketCount--
				return bracketCount >= 0
			}
			return !parser.MatchesToken(p, ";")
		})
		if s == "" {
			return "", &diagnostic.Diagnostic{
				Message: "missing array length",
				Primary: quickanno.Expected(p, p.Pos(), "a number, constant, or expression"),
			}
		}
	loop:
		for i := len(s) - 1; i >= 0; i-- {
			for _, r := range whitespace.Runes {
				if rune(s[i]) == r {
					s = s[:i]
					continue loop
				}
			}
			break
		}
		p.RestoreState(state)
		parser.TryToken(p, s)

		return s, nil
	}
}

func ElementType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#ElementType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := parser.Try(p, Type())
		if t == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing element type",
				Primary:  quickanno.Expected(p, p.Pos(), "a type"),
				Examples: []diagnostic.Example{{Example: "`[]string`, `map[string]int`, or `chan string`"}},
			}
		}
		return t, nil
	}
}

// ============================================================================
// Struct types
// ======================================================================================

// StructType parses a simpler superset of the struct type defined in the Go
// spec.
func StructType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#StructType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		var t ast.Type
		t.From = p.Pos()
		startIndex := p.Index()

		if !parser.TryToken(p, "struct") {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing struct type",
				Primary:  quickanno.Expected(p, p.Pos(), "a struct type"),
				Examples: []diagnostic.Example{{Example: "`struct{ Foo string }`"}},
			}
		}

		hasWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		lBracePos := parser.TryRuneAt(p, '{')
		if lBracePos == nil {
			if !hasWS {
				return nil, &diagnostic.Diagnostic{
					Message:  "struct type: missing opening brace",
					Primary:  quickanno.Expected(p, p.Pos(), "an opening curly brace"),
					Examples: []diagnostic.Example{{Example: "`struct { Foo string }`"}},
				}
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "struct type: missing opening brace",
				Primary:  quickanno.Expected(p, p.Pos(), "an opening curly brace"),
				Examples: []diagnostic.Example{{Example: "`struct { Foo() }`"}},
			})
		}

		// see array length on why this is necessary
		var braceCount int
		state := p.CloneState()
		s := parser.TokenWhile(p, func() bool {
			if parser.MatchesToken(p, "{") {
				braceCount++
				return true
			} else if parser.MatchesToken(p, "}") {
				braceCount--
				return braceCount >= 0
			}
			return true
		})
	loop:
		for i := len(s) - 1; i >= 0; i-- {
			for _, r := range whitespace.Runes {
				if rune(s[i]) == r {
					s = s[:i]
					continue loop
				}
			}
			break
		}
		p.RestoreState(state)
		parser.TryToken(p, s)

		parser.TrySkip(p, comment.OrAnyWhitespace())

		if !parser.TryRune(p, '}') {
			err := &diagnostic.Diagnostic{
				Message: "struct type: missing closing brace",
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
		return &t, nil
	}
}

// ============================================================================
// Pointer types
// ======================================================================================

func PointerType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#PointerType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := &ast.Type{From: p.Pos()}
		startIndex := p.Index()

		if !parser.TryRune(p, '*') {
			return nil, &diagnostic.Diagnostic{
				Message: "missing pointer",
				Primary: quickanno.Expected(p, p.Pos(), "an asterisk"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		parser.Must(p, BaseType(t.Start()))

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return t, nil
	}
}

func BaseType(asteriskPos ast.Position) parser.Func[*ast.Type] { // https://go.dev/ref/spec#BaseType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := parser.Try(p, Type())
		if t == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing base type",
				Primary: quickanno.Expected(p, p.Pos(), "a type"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, asteriskPos, "for the pointer asterisk here"),
				},
				Examples: []diagnostic.Example{{Example: "`int`"}},
			}
		}
		return t, nil
	}
}

// ============================================================================
// Function types
// ======================================================================================

// FunctionType parses a simpler superset of the function type defined in the Go
// spec.
func FunctionType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#FunctionType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := &ast.Type{From: p.Pos()}
		startIndex := p.Index()

		if !parser.TryToken(p, "func") {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing function type",
				Primary:  quickanno.Expected(p, p.Pos(), "the keyword `func`"),
				Examples: []diagnostic.Example{{Example: "`func(a, b int) string`"}},
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		parser.Must(p, Signature())

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return t, nil
	}
}

// Signature parses a simpler superset of the function signature defined in the
// Go spec.
func Signature() parser.Func[string] { // https://go.dev/ref/spec#Signature
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		_, err := parser.TryErr(p, Parameters())
		if err != nil {
			return "", err
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		parser.Try(p, Result())
		return p.AST.Raw[startIndex:p.Index()], nil
	}
}

// Result parses a simpler superset of the result defined in the Go spec.
func Result() parser.Func[string] { // https://go.dev/ref/spec#Result
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		startIndex := p.Index()

		params := parser.Try(p, Parameters())
		if params != "" {
			return p.AST.Raw[startIndex:p.Index()], nil
		}

		t := parser.Try(p, Type())
		if t != nil {
			return p.AST.Raw[startIndex:p.Index()], nil
		}

		return "", &diagnostic.Diagnostic{
			Message: "function signature: missing result",
			Primary: quickanno.Expected(p, p.Pos(), "a type or a parameter list"),
		}
	}
}

// Parameters parses a simpler superset of the parameters defined in the Go
// spec.
// In particular, it allows variadic parameters everywhere.
func Parameters() parser.Func[string] { // https://go.dev/ref/spec#Parameters
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		l := parser.Try(p, list.ParenList("parameters", NamedParameterDecl()))
		if l != nil {
			return p.AST.Raw[startIndex:p.Index()], nil
		}

		_, err := parser.TryErr(p, list.ParenList("parameters", UnnamedParameterDecl()))
		if err != nil {
			return "", err
		}
		return p.AST.Raw[startIndex:p.Index()], nil
	}
}

func NamedParameterDecl() parser.Func[string] { // https://go.dev/ref/spec#ParameterDecl
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		_, err := parser.TryErr(p, list.CommaList("parameter declaration", "parameter declarations", Identifier()))
		if err != nil {
			return "", err
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		if parser.TryOptionalToken(p, "...", nil) {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			parser.Must(p, Type())
		} else {
			parser.Try(p, Type())
		}

		return p.AST.Raw[startIndex:p.Index()], nil
	}
}

func UnnamedParameterDecl() parser.Func[string] { // https://go.dev/ref/spec#ParameterDecl
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		startPos := p.Pos()

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		parser.TryToken(p, "...")
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		t := parser.Try(p, Type())
		if t == nil {
			return "", &diagnostic.Diagnostic{
				Message: "missing parameter declaration",
				Primary: quickanno.Expected(p, startPos, "an optional list of identifiers followed by a type"),
			}
		}

		return p.AST.Raw[startIndex:p.Index()], nil
	}
}

// ============================================================================
// Interface types
// ======================================================================================

func InterfaceType() parser.Func[*ast.Type] {
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := &ast.Type{From: p.Pos()}
		startIndex := p.Index()

		if !parser.TryToken(p, "interface") {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing struct type",
				Primary:  quickanno.Expected(p, p.Pos(), "a interface type"),
				Examples: []diagnostic.Example{{Example: "`struct{ Foo string }`"}},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		lBracePos := parser.TryRuneAt(p, '{')
		if lBracePos == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "interface type: missing opening brace",
				Primary:  quickanno.Expected(p, p.Pos(), "an opening curly brace"),
				Examples: []diagnostic.Example{{Example: "`interface { Foo() }`"}},
			})
		}

		// see array length on why this is necessary
		var braceCount int
		state := p.CloneState()
		s := parser.TokenWhile(p, func() bool {
			if parser.MatchesToken(p, "{") {
				braceCount++
				return true
			} else if parser.MatchesToken(p, "}") {
				braceCount--
				return braceCount >= 0
			}
			return true
		})
	loop:
		for i := len(s) - 1; i >= 0; i-- {
			for _, r := range whitespace.Runes {
				if rune(s[i]) == r {
					s = s[:i]
					continue loop
				}
			}
			break
		}
		p.RestoreState(state)
		parser.TryToken(p, s)

		parser.TrySkip(p, comment.OrAnyWhitespace())

		if !parser.TryRune(p, '}') {
			err := &diagnostic.Diagnostic{
				Message: "interface type: missing closing brace",
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
		return t, nil
	}
}

// ============================================================================
// Slice types
// ======================================================================================

func SliceType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#SliceType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		t := &ast.Type{From: p.Pos()}

		if !parser.TryRune(p, '[') {
			return nil, &diagnostic.Diagnostic{
				Message: "missing slice",
				Primary: quickanno.Expected(p, p.Pos(), "an opening bracket"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		if !parser.TryRune(p, ']') {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "slice: missing closing bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a closing bracket"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, t.Start(), "for the opening bracket here"),
				},
			})
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		parser.Must(p, ElementType())

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return t, nil
	}
}

// ============================================================================
// Map types
// ======================================================================================

func MapType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#MapType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		t := &ast.Type{From: p.Pos()}

		if !parser.TryToken(p, "map") {
			return nil, &diagnostic.Diagnostic{
				Message: "missing map",
				Primary: quickanno.Expected(p, p.Pos(), "the keyword `map`"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		if !parser.TryRune(p, '[') {
			// check if only the [ is missing, or if the key along with the
			// closing bracket has been omitted entirely
			ok := parser.Matches(p, func(p *parser.Parser) (struct{}, *diagnostic.Diagnostic) {
				parser.TrySkip(p, comment.OrAnyWhitespace())
				parser.Try(p, KeyType())
				parser.TrySkip(p, comment.OrHorizontalWhitespace())

				if !parser.TryRune(p, ']') {
					return struct{}{}, &diagnostic.Diagnostic{}
				}
				return struct{}{}, nil
			})

			if ok {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "map: missing opening bracket",
					Primary: quickanno.Expected(p, p.Pos(), "an opening bracket"),
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message:  "map: missing key",
					Primary:  quickanno.Expected(p, p.Pos(), "a key type enclosed in `[` and `]`"),
					Examples: []diagnostic.Example{{Example: "`map[string]int`"}},
				})
				goto elemType
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		parser.Must(p, KeyType())

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

	elemType:
		parser.TrySkip(p, comment.OrAnyWhitespace())
		parser.Must(p, ElementType())
		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return t, nil
	}
}

func KeyType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#KeyType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		t := parser.Try(p, Type())
		if t == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing key type",
				Primary:  quickanno.Expected(p, p.Pos(), "a type"),
				Examples: []diagnostic.Example{{Example: "`map[string]int`"}},
			}
		}
		return t, nil
	}
}

// ============================================================================
// Channel types
// ======================================================================================

func ChannelType() parser.Func[*ast.Type] { // https://go.dev/ref/spec#ChannelType
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		startIndex := p.Index()
		t := &ast.Type{From: p.Pos()}

		if parser.TryToken(p, "<-") {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			if !parser.TryToken(p, "chan") {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "channel type: missing `chan` keyword",
					Primary: quickanno.Expected(p, p.Pos(), "the keyword `chan`"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, p.Pos(), "considering, you already placed a `<-` here"),
					},
				})
			}
		} else {
			if !parser.TryToken(p, "chan") {
				return nil, &diagnostic.Diagnostic{
					Message: "missing channel type",
					Primary: quickanno.Expected(p, p.Pos(), "a channel type"),
				}
			}

			parser.TrySkip(p, comment.OrAnyWhitespace())
			parser.TryToken(p, "<-")
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		parser.Must(p, ElementType())

		t.Type = p.AST.Raw[startIndex:p.Index()]
		t.Until = p.Pos()
		return t, nil
	}
}

// ============================================================================
// Parsed Types
// ======================================================================================

func NamedType() parser.Func[*ast.NamedType] { // essentially the first of https://go.dev/ref/spec#Type
	return func(p *parser.Parser) (*ast.NamedType, *diagnostic.Diagnostic) {
		var t ast.NamedType

		t.Name = parser.Try(p, TypeName())
		if t.Name == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing named type",
				Primary:  quickanno.Expected(p, p.Pos(), "a type name"),
				Examples: []diagnostic.Example{{Example: "`int` or `strings.Builder`"}},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		t.TypeArgs = parser.Try(p, TypeArgs())

		return &t, nil
	}
}

// ============================================================================
// Type parameter declarations
// ======================================================================================

func TypeParameters() parser.Func[*ast.TypeParameters] {
	return func(p *parser.Parser) (*ast.TypeParameters, *diagnostic.Diagnostic) {
		l, err := parser.TryErr(p, list.BracketList("type parameters", TypeParameterDecl()))
		if err != nil {
			return nil, err
		}

		return &ast.TypeParameters{
			LBracket: l.Open,
			Params:   l.Elems,
			RBracket: l.Close,
		}, nil
	}
}

func TypeParameterDecl() parser.Func[*ast.TypeParameter] {
	return func(p *parser.Parser) (*ast.TypeParameter, *diagnostic.Diagnostic) {
		var tp ast.TypeParameter

		tp.Names = parser.Try(p, list.CommaList("type parameter name", "type parameter names", Identifier()))
		if tp.Names == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing type parameter",
				Primary: quickanno.Expected(p, p.Pos(), "a type parameter"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		tp.Type = parser.Must(p, Type())

		return &tp, nil
	}
}
