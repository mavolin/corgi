package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/attribute"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/component"
	"github.com/mavolin/corgi/v2/load/parse/internal/element"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/state"
	"github.com/mavolin/corgi/v2/load/parse/internal/text"
)

func init() {
	interpolation.SetExpression(code.Expression(code.Regular))
	interpolation.SetElementHeader(element.Header())
	interpolation.SetComponentCallHeader(component.CallHeader())

	body.SetTextLine(text.Line)
	body.SetScopeNode(scopeNode)
}

func scopeNode(p *parser.Parser) (ast.ScopeNode, *diagnostic.Diagnostic) {
	if n := parser.Try(p, attribute.Definition()); n != nil {
		return n, nil
	} else if n := parser.Try(p, code.ImplicitCodeLine()); n != nil {
		return n, nil
	} else if n := parser.Try(p, code.ExplicitCodeLine()); n != nil {
		return n, nil
	} else if n := parser.Try(p, component.Component()); n != nil {
		return n, nil
	} else if n := parser.Try(p, component.Alias()); n != nil {
		return n, nil
	} else if n := parser.Try(p, component.Block()); n != nil {
		return n, nil
	} else if n := parser.Try(p, code.Conditional()); n != nil {
		return n, nil
	} else if n := parser.Try(p, code.Switch()); n != nil {
		return n, nil
	} else if n := parser.Try(p, code.For()); n != nil {
		return n, nil
	} else if n := parser.Try(p, state.Declaration()); n != nil {
		return n, nil
	} else if n := parser.Try(p, text.ArrowBlock()); n != nil {
		return n, nil
	} else if n := parser.Try(p, element.Definition()); n != nil {
		return n, nil
	} else if n := parser.Try(p, element.And()); n != nil {
		return n, nil
	} else if n := parser.Try(p, element.Doctype()); n != nil {
		return n, nil
	} else if n := parser.Try(p, element.Raw()); n != nil {
		return n, nil
	} else if n := parser.Try(p, component.Call()); n != nil {
		return n, nil
	}

	if b := parser.Try(p, code.Else()); b != nil {
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "unexpected `else`",
			Primary: []diagnostic.Annotation{
				anno.Range(p.File, *b.Else, quickanno.DeltaPos(*b.Else, 0, len("else")), "unexpected `else`"),
			},
			Explanation: "This `else` is not part of an if statement.",
		})
		return &ast.BadScopeNode{
			From:  b.Start(),
			Until: b.End(),
		}, nil
	} else if b := parser.Try(p, code.ElseIf()); b != nil {
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "unexpected `else if`",
			Primary: []diagnostic.Annotation{
				anno.Range(p.File, *b.Else, quickanno.DeltaPos(*b.If, 0, len("if")), "unexpected `else if`"),
			},
			Explanation: "This `else if` is not part of an if statement.",
		})
		return &ast.BadScopeNode{
			From:  b.Start(),
			Until: b.End(),
		}, nil
	}

	if n := parser.Try(p, element.Element()); n != nil {
		return n, nil
	}

	return nil, &diagnostic.Diagnostic{
		Message: "missing scope node",
		Primary: quickanno.Expected(p, p.Pos(), "a scope node"),
	}
}

func File() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *diagnostic.Diagnostic) {
		parser.TrySkip(p, comment.OrAnyWhitespace())
		p.File.File.Package = parser.Must(p, PackageDirective())
		p.File.File.Imports = parser.Collect(p, Import(), 8, comment.OrAnyWhitespace())
		for _, imp := range p.File.File.Imports {
			for _, spec := range imp.Specs {
				if spec.Path != nil {
					p.Preload(spec.Path.Unquote())
				}
			}
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		p.File.TopLevel = parser.Must(p, TopLevel())
		parser.TrySkip(p, comment.OrAnyWhitespace())
		p.File.Comments = p.CloneState().Comments()
		if !parser.MatchesAnyRune(p, parser.EOF) {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unexpected tokens",
				Primary: quickanno.Expected(p, p.Pos(), "end of file"),
			})
		}
		return struct{}{}, nil
	}
}

func TopLevel() parser.Func[[]ast.ScopeNode] {
	return func(p *parser.Parser) ([]ast.ScopeNode, *diagnostic.Diagnostic) {
		scope := make([]ast.ScopeNode, 0, 36)

		for {
			if sd := parser.TryOptional(p, state.Declaration(), nil); sd != nil {
				scope = append(scope, sd)
			} else if c := parser.TryOptional(p, component.Component(), nil); c != nil {
				scope = append(scope, c)
			} else if n := parser.TryOptional(p, component.Alias(), nil); n != nil {
				scope = append(scope, n)
			} else if ad := parser.TryOptional(p, attribute.Definition(), nil); ad != nil {
				scope = append(scope, ad)
			} else if ed := parser.TryOptional(p, element.Definition(), nil); ed != nil {
				scope = append(scope, ed)
			} else if s := parser.TryOptional(p, code.Statement(code.Regular), nil); s != nil {
				scope = append(scope, &ast.ImplicitCodeLine{Statement: s})
			} else if imp := parser.TryOptional(p, Import(), nil); imp != nil {
				scope = append(scope, &ast.BadScopeNode{
					From:  imp.Start(),
					Until: imp.End(),
				})
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "unexpected import",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, imp.Start(), imp.End(), "cannot place import here"),
					},
					Hints: []diagnostic.Hint{
						{Hint: "Imports must be placed at the top of the file, right below the package directive."},
					},
				})
			} else if bn := parser.Try(p, body.BadScopeNode()); bn != nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "bad scope node",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, bn.From, bn.Until, "unexpected tokens"),
					},
					Hints: []diagnostic.Hint{
						{Hint: "Expected a state declaration, a component, or code"},
					},
				})
				scope = append(scope, bn)
			} else {
				break
			}

			parser.MustSkip(p, comment.AndMustEOS())
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}

		return slices.Clip(scope), nil
	}
}
