package component

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/attribute"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Component() parser.Func[*ast.Component] {
	return func(p *parser.Parser) *ast.Component {
		var c ast.Component

		c.Comp = parser.TryKeywordAt(p, "comp", comment.OrAnyWhitespace())
		if c.Comp == nil {
			return nil
		}

		c.Header = parser.TryOptional(p, Header(), comment.OrHorizontalWhitespace())
		if c.Header == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component: missing header",
				Primary: quickanno.Expected(p, p.Pos(), "a component header"),
				Examples: []diagnostic.Example{
					{Example: "comp Hello(name string)"},
				},
			})
		}

		c.Body = parser.Try(p, Body())
		if c.Body == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a component body"),
				Examples: []diagnostic.Example{
					{Example: "comp Hello(name string) { ... }"},
				},
			})
		}
		return &c
	}
}

func Header() parser.Func[*ast.ComponentHeader] {
	return func(p *parser.Parser) *ast.ComponentHeader {
		var h ast.ComponentHeader

		h.Name = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		if h.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component: header: missing name",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
			})
		}
		h.TypeParameters = parser.TryOptional(p, golang.TypeParameters(), comment.OrHorizontalWhitespace())
		h.Parameters = parser.TryOptional(p, Parameters(), nil)
		if h.Parameters == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component: header: missing parameters",
				Primary: quickanno.Expected(p, p.Pos(), "a parameter list"),
				Examples: []diagnostic.Example{
					{Example: "Hello(name string)"},
				},
			})
		}

		if h.Name == nil && h.Parameters == nil {
			return nil
		}

		return &h
	}
}

func Parameters() parser.Func[*ast.ComponentParameters] {
	return func(p *parser.Parser) *ast.ComponentParameters {
		l := parser.Try(p, list.ParenList("parameter", "component parameters", Parameter()))
		if l == nil {
			return nil
		}
		return &ast.ComponentParameters{
			LParen: l.Open,
			List:   l.Elems,
			RParen: l.Close,
		}
	}
}

func Parameter() parser.Func[*ast.ComponentParameter] {
	return func(p *parser.Parser) *ast.ComponentParameter {
		var param ast.ComponentParameter

		param.Name = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		if param.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component parameter: missing name",
				Primary: quickanno.Expected(p, param.Type.Start(), "a parameter name"),
			})
		}
		param.Type = parser.TryOptional(p, ParameterType(), comment.OrHorizontalWhitespace())
		param.Colon = parser.TryOptionalRuneAt(p, ':', comment.OrAnyWhitespace())

		if param.Name == nil && param.Type == nil && param.Colon == nil {
			return nil
		}

		param.Default = parser.Try(p, code.Expression(code.Regular))
		if param.Default == nil {
			if param.Colon != nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "component parameter: missing default value",
					Primary: quickanno.Expected(p, p.Pos(), "a default value"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, *param.Colon, "because of this colon"),
					},
					Hints: []diagnostic.Hint{
						{
							Hint: "I'm only expecting a default value because of the colon. " +
								"If you don't want to provide a default value, remove the colon and provide a type instead.",
						},
					},
				})
			}
			if param.Type == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "component parameter: missing type/default",
					Primary: quickanno.Expected(p, p.Pos(), "either a type or a colon followed by a default value"),
					Explanation: "Every parameter must have a type. " +
						"That type can either be specified explicitly behind the parameter name, " +
						"just like for function parameters, " +
						"or implicitly by providing a default value.",
					Examples: []diagnostic.Example{
						{Example: "bark: string", Title: "explicit type"},
						{Example: "bark: \"woof\"", Title: "implicit type"},
					},
				})
			}
		} else if param.Colon == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component parameter: missing colon before default",
				Primary: quickanno.Expected(p, param.Default.Start(), "a colon here"),
			})
		}

		return &param
	}
}

func ParameterType() parser.Func[*ast.Type] {
	return func(p *parser.Parser) *ast.Type {
		startI := p.Index()
		startPos := p.Pos()
		at := parser.Try(p, attribute.Type())
		if at != nil {
			return &ast.Type{
				Type:   p.AST.Raw[startI:p.Index()],
				Parsed: at,
				From:   startPos,
				Until:  p.Pos(),
			}
		}

		return parser.Try(p, golang.Type())
	}
}

func Block() parser.Func[*ast.Block] {
	return func(p *parser.Parser) *ast.Block {
		var b ast.Block

		b.Block = parser.TryKeywordAt(p, "block", comment.OrAnyWhitespace())
		if b.Block == nil {
			return nil
		}

		b.Identifier = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		b.Default = parser.Try(p, body.Body())

		return &b
	}
}
