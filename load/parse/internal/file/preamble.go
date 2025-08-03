package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func PackageDirective() parser.Func[*ast.PackageDirective] {
	return func(p *parser.Parser) *ast.PackageDirective {
		var d ast.PackageDirective

		d.Package = parser.TryKeywordAt(p, "package", comment.OrAnyWhitespace())
		if d.Package == nil {
			return nil
		}

		d.Name = parser.Try(p, golang.Identifier())
		if d.Name == nil {
			return nil
		}
		return &d
	}
}

func Import() parser.Func[*ast.Import] {
	return func(p *parser.Parser) *ast.Import {
		var imp ast.Import

		imp.Import = p.PosPtr()
		if !parser.TryToken(p, "import") {
			return nil
		}

		hasWS := parser.TrySkip(p, comment.OrAnyWhitespace())

		imp.LParen = parser.TryOptionalRuneAt(p, '(', nil)
		if imp.LParen == nil {
			if !hasWS {
				return nil
			}

			spec := parser.Try(p, ImportSpec())
			if spec != nil {
				imp.Specs = []*ast.ImportSpec{spec}
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message:  "missing import spec",
					Primary:  quickanno.Expected(p, p.Pos(), "an import specs"),
					Examples: []diagnostic.Example{{Example: "`import bark \"woof\"`"}},
				})
			}
			return &imp
		}

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			spec := parser.Try(p, ImportSpec())
			if spec == nil {
				break
			}
			imp.Specs = append(imp.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.Try(p, comment.AndMustEOS())
		}
		imp.Specs = slices.Clip(imp.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "import: unexpected runes"
			p.CaptureError(err)
		}

		imp.RParen = parser.TryOptionalRuneAt(p, ')', nil)
		if imp.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "import: missing ')'",
				Primary: quickanno.Expected(p, *imp.LParen, "a closing ')' for the '(' here"),
			})
		}

		return &imp
	}
}

func ImportSpec() parser.Func[*ast.ImportSpec] {
	return func(p *parser.Parser) *ast.ImportSpec {
		var spec ast.ImportSpec

		spec.Alias = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		spec.Path = parser.Try(p, golang.StringLit())
		if spec.Path == nil {
			if spec.Alias == nil {
				return nil
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "import spec: missing path",
				Primary: quickanno.Expected(p, p.Pos(), "an import path"),
			})
		}

		return &spec
	}
}
