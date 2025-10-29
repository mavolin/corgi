package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
	"golang.org/x/mod/module"
)

func PackageDirective() parser.Func[*ast.PackageDirective] {
	return func(p *parser.Parser) *ast.PackageDirective {
		pkg := parser.TryKeywordAt(p, "package")
		if pkg == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		name := parser.Try(p, golang.Identifier())
		if name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "package directive: missing package name",
				Primary: quickanno.Expected(p, p.Pos(), "a package name"),
			})
			return nil
		}

		return &ast.PackageDirective{Package: pkg, Name: name}
	}
}

func Import() parser.Func[*ast.Import] {
	return func(p *parser.Parser) *ast.Import {
		pos := parser.TryTokenAt(p, "import")
		if pos == nil {
			return nil
		}

		hasWS := parser.TrySkip(p, comment.OrAnyWhitespace())

		lParen := parser.TryOptionalRuneAt(p, '(', nil)
		if lParen == nil {
			if !hasWS {
				return nil
			}

			var imp ast.Import
			imp.Import = pos

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

		var imp ast.Import
		imp.Import = pos
		imp.LParen = lParen

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
		alias := parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		path := parser.Try(p, golang.StringLit())
		if path == nil {
			if alias == nil {
				return nil
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "import spec: missing path",
				Primary: quickanno.Expected(p, p.Pos(), "an import path"),
			})
		} else if err := module.CheckImportPath(path.Unquote()); err != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "import spec: invalid import path",
				Primary: []diagnostic.Annotation{
					anno.Node(p.File, path, err.Error()),
				},
			})
		}

		return &ast.ImportSpec{Alias: alias, Path: path}
	}
}
