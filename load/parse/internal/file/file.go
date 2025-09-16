package file

import (
	"fmt"
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
	"github.com/mavolin/corgi/v2/load/parse/internal/control"
	"github.com/mavolin/corgi/v2/load/parse/internal/element"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/state"
	"github.com/mavolin/corgi/v2/load/parse/internal/text"
)

//nolint:gochecknoinits
func init() {
	code.SetComponentCall(component.Call())

	interpolation.SetExpression(code.Expression())
	interpolation.SetComponentCall(component.Call())

	attribute.SetElementReference(element.Reference())

	body.SetTextLine(text.Line)
	body.SetVerbatimTextLine(text.VerbatimLine)
	body.SetScopeNode(scopeNode)
}

func scopeNode(p *parser.Parser) ast.ScopeNode {
	if n := parser.Try(p, component.Call()); n != nil {
		return n
	} else if n := parser.Try(p, component.Block()); n != nil {
		return n
	} else if n := parser.Try(p, component.With()); n != nil {
		return n
	} else if n := parser.Try(p, control.Conditional()); n != nil {
		return n
	} else if n := parser.Try(p, control.Switch()); n != nil {
		return n
	} else if n := parser.Try(p, control.For()); n != nil {
		return n
	} else if n := parser.Try(p, text.ArrowBlock()); n != nil {
		return n
	} else if n := parser.Try(p, element.And()); n != nil {
		return n
	} else if n := parser.Try(p, element.Doctype()); n != nil {
		return n
	} else if n := parser.Try(p, element.Raw()); n != nil {
		return n
	} else if n := parser.Try(p, code.ExplicitCodeLine()); n != nil {
		return n
	} else if b := parser.Try(p, control.ElseIf()); b != nil {
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "unexpected `else if`",
			Primary: []diagnostic.Annotation{
				anno.Range(p.File, *b.Else, quickanno.DeltaPos(*b.If, 0, len("if")), "unexpected `else if"),
			},
			Explanation: "This `else if` is not part of an if statement.",
		})
		return &ast.BadNode{
			From:  b.Start(),
			Until: b.End(),
		}
	} else if b := parser.Try(p, control.Else()); b != nil {
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "unexpected `else`",
			Primary: []diagnostic.Annotation{
				anno.Range(p.File, *b.Else, quickanno.DeltaPos(*b.Else, 0, len("else")), "unexpected `else`"),
			},
			Explanation: "This `else` is not part of an if statement.",
		})
		return &ast.BadNode{
			From:  b.Start(),
			Until: b.End(),
		}
	} else if n := parser.Try(p, element.Element()); n != nil {
		return n
	} else if n := parser.Try(p, code.ImplicitCodeLine()); n != nil {
		return n
	} else if n := parser.Try(p, TopLevelNode()); n != nil {
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "unexpected top level node",
			Primary: []diagnostic.Annotation{
				anno.Node(p.File, n, fmt.Sprintf("cannot place %T here", n)),
			},
			Explanation: "Top level nodes can only be place at the top level of a file, " +
				"not inside of components.",
		})
		return &ast.BadNode{
			From:  n.Start(),
			Until: n.End(),
		}
	}

	return nil
}

func File() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		parser.TrySkip(p, comment.OrAnyWhitespace())

		p.AST.Package = parser.TryOptional(p, PackageDirective(), nil)
		if p.AST.Package == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "missing package directive",
				Primary:  quickanno.Expected(p, p.Pos(), "a package directive"),
				Examples: []diagnostic.Example{{Example: "`package main`"}},
			})
		} else {
			parser.Try(p, comment.AndForceEOS())
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}

		for {
			imp := parser.TryOptional(p, Import(), nil)
			if imp == nil {
				break
			}
			p.AST.Imports = append(p.AST.Imports, imp)
			parser.Try(p, comment.AndForceEOS())
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}
		p.AST.Imports = slices.Clip(p.AST.Imports)

		p.AST.TopLevel = parser.Try(p, TopLevel())
		parser.TrySkip(p, comment.OrAnyWhitespace())

		if !parser.MatchesAnyRune(p, parser.EOF) {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unexpected tokens",
				Primary: quickanno.Expected(p, p.Pos(), "end of file"),
			})
		}

		p.AST.Comments = p.Comments()
		return true
	}
}

func TopLevel() parser.Func[ast.TopLevel] {
	return func(p *parser.Parser) ast.TopLevel {
		scope := make(ast.TopLevel, 0, 36)

		for {
			if n := parser.TryOptional(p, TopLevelNode(), nil); n != nil {
				scope = append(scope, n)
			} else if s := parser.TryOptional(p, implicitTopLevelCodeLine(), nil); s != nil {
				scope = append(scope, s)
			} else if imp := parser.TryOptional(p, Import(), nil); imp != nil {
				scope = append(scope, &ast.BadNode{
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
			} else if n := parser.TryOptional(p, body.ScopeNode(), nil); n != nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "unexpected node",
					Primary: []diagnostic.Annotation{
						anno.Node(p.File, n, fmt.Sprintf("cannot place %T here", n)),
					},
					Hints: []diagnostic.Hint{
						{Hint: "Expected a state declaration, a component, an attribute or element definition, or code"},
					},
				})
				scope = append(scope, &ast.BadNode{
					From:  n.Start(),
					Until: n.End(),
				})
			} else if bn := parser.TryOptional(p, body.BadNode(), nil); bn != nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "bad node",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, bn.From, bn.Until, "unexpected tokens"),
					},
					Hints: []diagnostic.Hint{
						{Hint: "Expected a state declaration, a component, an attribute or element definition, or code"},
					},
				})
				scope = append(scope, bn)
			} else if parser.MatchesToken(p, "}") {
				from := p.Pos()
				parser.TryRune(p, '}')
				until := p.Pos()
				scope = append(scope, &ast.BadNode{
					From:  from,
					Until: until,
				})

				p.CaptureError(&diagnostic.Diagnostic{
					Message: "unexpected closing brace",
					Primary: []diagnostic.Annotation{
						anno.Position(p.File, from, "this `}` belongs to no opening `{`"),
					},
				})
			} else {
				break
			}

			parser.Try(p, comment.AndForceEOS())
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}

		return slices.Clip(scope)
	}
}

func TopLevelNode() parser.Func[ast.TopLevelNode] {
	return func(p *parser.Parser) ast.TopLevelNode {
		if sd := parser.TryOptional(p, state.Declaration(), nil); sd != nil {
			return sd
		} else if c := parser.TryOptional(p, component.Component(), nil); c != nil {
			return c
		} else if ad := parser.TryOptional(p, attribute.Definition(), nil); ad != nil {
			return ad
		} else if ed := parser.TryOptional(p, element.Definition(), nil); ed != nil {
			return ed
		}
		return nil
	}
}

func implicitTopLevelCodeLine() parser.Func[*ast.ImplicitCodeLine] {
	return func(p *parser.Parser) *ast.ImplicitCodeLine {
		s := code.Statement()(p)
		if s == nil {
			return nil
		}
		return &ast.ImplicitCodeLine{Statement: s}
	}
}
