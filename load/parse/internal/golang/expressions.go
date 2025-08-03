package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// https://go.dev/ref/spec#Expressions

// ============================================================================
// Qualified identifiers
// ======================================================================================

func QualifiedIdent() parser.Func[*ast.QualifiedIdentifier] { // https://go.dev/ref/spec#QualifiedIdent
	return func(p *parser.Parser) *ast.QualifiedIdentifier {
		var ident ast.QualifiedIdentifier

		ident.Package = parser.Try(p, PackageName())
		if ident.Package == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		ident.Dot = parser.TryRuneAt(p, '.')
		if ident.Dot == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "qualified identifier: missing dot and name in package",
				Primary: quickanno.Expected(p, p.Pos(), "a dot"),
				Explanation: "A qualified identifier consists of a package name, " +
					"and the name of a symbol in that package separated by a dot. " +
					"You are missing the dot and the name of the symbol.",
				Examples: []diagnostic.Example{{Example: "`" + ident.Package.Name + ".Woof`"}},
			})
			return &ident
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		ident.Name = parser.Try(p, Identifier())
		if ident.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "qualified identifier: missing name in package",
				Primary: quickanno.Expected(p, *ident.Dot, "an identifier"),
			})
		}

		return &ident
	}
}

// ============================================================================
// Operators
// ======================================================================================

func AddOp() parser.Func[string] {
	return func(p *parser.Parser) string {
		r := parser.TryAnyRune(p, '+', '-', '|', '^')
		if r == 0 {
			return ""
		}

		return string(r)
	}
}

func MulOp() parser.Func[string] {
	return func(p *parser.Parser) string {
		if op := parser.TryAnyToken(p, "<<", ">>", "&^"); op != "" {
			return op
		}

		r := parser.TryAnyRune(p, '*', '/', '%', '&')
		if r == 0 {
			return ""
		}

		return string(r)
	}
}
