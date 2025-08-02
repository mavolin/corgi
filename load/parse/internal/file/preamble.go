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
	return func(p *parser.Parser) (*ast.PackageDirective, *diagnostic.Diagnostic) {
		var d ast.PackageDirective

		d.Package = parser.TryKeywordAt(p, "package", comment.OrAnyWhitespace())
		if d.Package == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing package directive",
				Primary: quickanno.Expected(p, p.Pos(), "a package directive"),
			}
		}

		d.Name = parser.Must(p, golang.Identifier())
		return &d, nil
	}
}

func Import() parser.Func[*ast.Import] {
	return func(p *parser.Parser) (*ast.Import, *diagnostic.Diagnostic) {
		var imp ast.Import

		imp.Import = p.PosPtr()
		if !parser.TryToken(p, "import") {
			return nil, &diagnostic.Diagnostic{
				Message: "missing import directive",
				Primary: quickanno.Expected(p, p.Pos(), "an import directive"),
			}
		}

		hasWS := parser.TrySkip(p, comment.OrAnyWhitespace())

		imp.LParen = parser.TryOptionalRuneAt(p, '(', nil)
		if imp.LParen == nil {
			if !hasWS {
				return nil, &diagnostic.Diagnostic{
					Message: "missing import directive",
					Primary: quickanno.Expected(p, *imp.Import, "an import directive"),
				}
			}

			imp.Specs = []*ast.ImportSpec{parser.Must(p, ImportSpec())}
			return &imp, nil
		}

		imp.Specs = make([]*ast.ImportSpec, 0, 36)
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

			parser.Must(p, comment.AndEOS())
		}
		if len(imp.Specs) == 0 {
			imp.Specs = nil
		} else {
			imp.Specs = slices.Clip(imp.Specs)
		}

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

		return &imp, nil
	}
}

func ImportSpec() parser.Func[*ast.ImportSpec] {
	return func(p *parser.Parser) (*ast.ImportSpec, *diagnostic.Diagnostic) {
		var spec ast.ImportSpec

		spec.Alias = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		spec.Path = parser.Try(p, golang.StringLit())
		if spec.Path == nil {
			if spec.Alias == nil {
				return nil, &diagnostic.Diagnostic{
					Message: "missing import spec",
					Primary: quickanno.Expected(p, p.Pos(), "an import path"),
				}
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "import spec: missing path",
				Primary: quickanno.Expected(p, p.Pos(), "an import path"),
			})
		}

		return &spec, nil
	}
}
