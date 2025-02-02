package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func PackageDirective() parser.Func[*ast.PackageDirective] {
	return func(p *parser.Parser) (*ast.PackageDirective, *fancyerr.Error) {
		d := &ast.PackageDirective{Package: p.Pos()}
		if !parser.TryToken(p, "package") || !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing package directive",
				Primary: quickanno.Expected(p, p.Pos(), "a package directive"),
			}
		}

		d.Name = parser.Must(p, golang.Identifier())
		parser.MustSkip(p, comment.AndMustEOS())
		return d, nil
	}
}

func Import() parser.Func[*ast.Import] {
	return func(p *parser.Parser) (*ast.Import, *fancyerr.Error) {
		imp := &ast.Import{Import: p.Pos()}
		if !parser.TryToken(p, "import") {
			return nil, &fancyerr.Error{
				Message: "missing import directive",
				Primary: quickanno.Expected(p, p.Pos(), "an import directive"),
			}
		}

		hasWS := parser.TrySkipOk(p, comment.OrAnyWhitespace())

		imp.LParen = p.PosPtr()
		if !parser.TryOptionalRune(p, '(') {
			imp.LParen = nil
			if !hasWS {
				return nil, &fancyerr.Error{
					Message: "missing import directive",
					Primary: quickanno.Expected(p, imp.Import, "an import directive"),
				}
			}

			spec := parser.Must(p, ImportSpec())
			if spec != nil {
				imp.Specs = []*ast.ImportSpec{spec}
			}

			parser.MustSkip(p, comment.AndMustEOS())
			return imp, nil
		}

		imp.Specs = make([]*ast.ImportSpec, 0, 36)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			spec, ok := parser.TryOk(p, ImportSpec())
			if !ok {
				break
			}
			imp.Specs = append(imp.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.MustSkip(p, comment.AndEOS())
		}
		imp.Specs = slices.Clip(imp.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "import: unexpected runes"
			p.CaptureError(err)
		}

		imp.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			imp.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "import: missing ')'",
				Primary: quickanno.Expected(p, *imp.LParen, "a closing ')' for the '(' here"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return imp, nil
	}
}

func ImportSpec() parser.Func[*ast.ImportSpec] {
	return func(p *parser.Parser) (*ast.ImportSpec, *fancyerr.Error) {
		spec := &ast.ImportSpec{}

		var ok bool
		spec.Alias, ok = parser.TryOk(p, golang.Identifier())
		if ok {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		spec.Path, ok = parser.TryOk(p, golang.StringLit())
		if !ok {
			if spec.Alias == nil {
				return nil, &fancyerr.Error{
					Message: "import spec: missing path",
					Primary: quickanno.Expected(p, p.Pos(), "a path"),
				}
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "import spec: missing path",
					Primary: quickanno.Expected(p, p.Pos(), "a path"),
				})
			}
		}

		return spec, nil
	}
}
