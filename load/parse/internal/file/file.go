package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/attribute"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/component"
	"github.com/mavolin/corgi/v2/load/parse/internal/element"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/state"
	"github.com/mavolin/corgi/v2/load/parse/internal/text"
)

func init() {
	interpolation.SetExpression(code.Expression())
	interpolation.SetElementHeader(element.Header())
	interpolation.SetComponentCallHeader(component.CallHeader())

	body.SetTextLine(text.Line)
	body.SetScopeNode(scopeNode)
}

func scopeNode(p *parser.Parser) (ast.ScopeNode, *fancyerr.Error) {
	panic("implement me")
}

func File() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		parser.TrySkip(p, comment.OrAnyWhitespace())
		p.File.File.Package = parser.Must(p, PackageDirective())
		p.File.File.Imports = parser.Collect(p, Import(), 8, comment.OrAnyWhitespace())
		p.File.TopLevel = parser.Must(p, TopLevel())
		parser.TrySkip(p, comment.OrAnyWhitespace())
		p.File.Comments = p.CloneState().Comments()
		return struct{}{}, nil
	}
}

func TopLevel() parser.Func[[]ast.ScopeNode] {
	return func(p *parser.Parser) ([]ast.ScopeNode, *fancyerr.Error) {
		scope := make([]ast.ScopeNode, 0, 36)

		for {
			if sd := parser.TryOptional(p, state.Declaration(), nil); sd != nil {
				scope = append(scope, sd)
			} else if c := parser.TryOptional(p, component.Component(), nil); c != nil {
				scope = append(scope, c)
			} else if ad := parser.TryOptional(p, attribute.Definition(), nil); ad != nil {
				scope = append(scope, ad)
			} else if ed := parser.TryOptional(p, element.Defintion(), nil); ed != nil {
				scope = append(scope, ed)
			} else if s := parser.TryOptional(p, code.Statement(), nil); s != nil {
				scope = append(scope, &ast.ImplicitCodeLine{Statement: s})
			} else if imp := parser.TryOptional(p, Import(), nil); imp != nil {
				scope = append(scope, &ast.BadScopeNode{
					From:  imp.Start(),
					Until: imp.End(),
				})
				p.CaptureError(&fancyerr.Error{
					Message: "unexpected import",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, imp.Start(), imp.End(), "cannot place import here"),
					},
					Hints: []fancyerr.Hint{
						{Hint: "Imports must be placed at the top of the file, right below the package directive."},
					},
				})
			} else if bn := parser.Try(p, body.BadScopeNode()); bn != nil {
				p.CaptureError(&fancyerr.Error{
					Message: "bad scope node",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, bn.From, bn.Until, "unexpected tokens"),
					},
					Hints: []fancyerr.Hint{
						{Hint: "Expected a state declaration, a component, or code"},
					},
				})
				scope = append(scope, bn)
			} else {
				break
			}

			parser.TrySkip(p, comment.OrAnyWhitespace())
		}

		return slices.Clip(scope), nil
	}
}
