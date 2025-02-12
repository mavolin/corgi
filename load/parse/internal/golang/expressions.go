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

func QualifiedIdent() parser.Func[*ast.QualifiedIdent] { // https://go.dev/ref/spec#QualifiedIdent
	return func(p *parser.Parser) (*ast.QualifiedIdent, *diagnostic.Diagnostic) {
		var ident ast.QualifiedIdent

		ident.Package = parser.Try(p, PackageName())
		if ident.Package == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing qualified identifier",
				Primary:  quickanno.Expected(p, p.Pos(), "an identifier"),
				Examples: []diagnostic.Example{{Example: "`woof.Bark`"}},
			}
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
				Examples: []diagnostic.Example{{Example: "`" + ident.Package.Ident + ".Woof`"}},
			})
			return &ident, nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		ident.Name = parser.Try(p, Identifier())
		if ident.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "qualified identifier: missing name in package",
				Primary: quickanno.Expected(p, *ident.Dot, "an identifier"),
			})
		}

		return &ident, nil
	}
}

// ============================================================================
// Operators
// ======================================================================================

func AddOp() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		r := parser.TryAnyRune(p, '+', '-', '|', '^')
		if r < 0 {
			return "", &diagnostic.Diagnostic{
				Message: "missing add op",
				Primary: quickanno.Expected(p, p.Pos(), "`+`, `-`, `|`, `^`"),
			}
		}

		return string(r), nil
	}
}

func MulOp() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if parser.TryToken(p, "<<") {
			return "<<", nil
		} else if parser.TryToken(p, ">>") {
			return ">>", nil
		} else if parser.TryToken(p, "&^") {
			return "&^", nil
		}

		r := parser.TryAnyRune(p, '*', '/', '%', '&')
		if r < 0 {
			return "", &diagnostic.Diagnostic{
				Message: "missing mul op",
				Primary: quickanno.Expected(p, p.Pos(), "`*`, `/`, `%`, `<<`, `>>`, `&`, `&^`"),
			}
		}

		return string(r), nil
	}
}
