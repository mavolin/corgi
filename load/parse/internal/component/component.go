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
	return func(p *parser.Parser) (*ast.Component, *diagnostic.Diagnostic) {
		var c ast.Component

		c.Comp = parser.TryKeywordAt(p, "comp", comment.OrAnyWhitespace())
		if c.Comp == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing component",
				Primary: quickanno.Expected(p, p.Pos(), "a component"),
			}
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

		c.Colon = parser.TryOptionalRuneAt(p, ':', comment.OrAnyWhitespace())
		c.Extend = parser.TryOptional(p, CallHeader(), comment.OrHorizontalWhitespace())
		if c.Extend != nil && c.Colon == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component: missing colon before extend",
				Primary: quickanno.Expected(p, c.Extend.Start(), "a colon before the component call header"),
			})
		} else if c.Extend == nil && c.Colon != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component: missing extend",
				Primary: quickanno.Expected(p, p.Pos(), "a component call header"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *c.Colon, "because of this colon, indicating a following component call header"),
				},
				Hints: []diagnostic.Hint{
					{Hint: "If you don't want to extend another component, remove the colon."},
				},
			})
		}

		c.Body = parser.Must(p, body.Body())
		return &c, nil
	}
}

func Header() parser.Func[*ast.ComponentHeader] {
	return func(p *parser.Parser) (*ast.ComponentHeader, *diagnostic.Diagnostic) {
		var h ast.ComponentHeader

		h.Name = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		if h.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component header: missing name",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
			})
		}
		h.TypeParams = parser.TryOptional(p, golang.TypeParameters(), comment.OrHorizontalWhitespace())
		h.Params = parser.TryOptional(p, Parameters(), nil)
		if h.Params == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component header: missing parameters",
				Primary: quickanno.Expected(p, p.Pos(), "a parameter list"),
				Examples: []diagnostic.Example{
					{Example: "Hello(name string)"},
				},
			})
		}

		if h.Name == nil && h.Params == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing component header",
				Primary: quickanno.Expected(p, h.Start(), "an identifier and a list of parameters"),
			}
		}

		return &h, nil
	}
}

func Parameters() parser.Func[*ast.ComponentParameters] {
	return func(p *parser.Parser) (*ast.ComponentParameters, *diagnostic.Diagnostic) {
		l, err := parser.TryErr(p, list.ParenList("component parameters", Parameter()))
		if err != nil {
			return nil, err
		}
		return &ast.ComponentParameters{
			LParen: l.Open,
			Params: l.Elems,
			RParen: l.Close,
		}, nil
	}
}

func Parameter() parser.Func[*ast.ComponentParameter] {
	return func(p *parser.Parser) (*ast.ComponentParameter, *diagnostic.Diagnostic) {
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
			return nil, &diagnostic.Diagnostic{
				Message: "missing component parameter",
				Primary: quickanno.Expected(p, p.Pos(), "a parameter name"),
			}
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
		} else {
			if param.Colon == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "component parameter: missing colon before default",
					Primary: quickanno.Expected(p, param.Default.Start(), "a colon here"),
				})
			}
		}

		return &param, nil
	}
}

func ParameterType() parser.Func[*ast.Type] {
	return func(p *parser.Parser) (*ast.Type, *diagnostic.Diagnostic) {
		startI := p.Index()
		startPos := p.Pos()
		at := parser.Try(p, attribute.Type())
		if at != nil {
			return &ast.Type{
				Type:   p.AST.Raw[startI:p.Index()],
				Parsed: at,
				From:   startPos,
				Until:  p.Pos(),
			}, nil
		}

		return parser.TryErr(p, golang.Type())
	}
}

func Alias() parser.Func[*ast.Alias] {
	return func(p *parser.Parser) (*ast.Alias, *diagnostic.Diagnostic) {
		var a ast.Alias

		a.Alias = parser.TryKeywordAt(p, "alias", comment.OrAnyWhitespace())
		if a.Alias == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing alias",
				Primary: quickanno.Expected(p, p.Pos(), "an alias"),
			}
		}

		colon := parser.MatchesAnyRune(p, ':')
		if !colon {
			a.Header = parser.TryOptional(p, Header(), comment.OrAnyWhitespace())
		}
		if colon || a.Header == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "alias: missing component header",
				Primary: quickanno.Expected(p, p.Pos(), "a component header for the alias"),
			})
		}

		a.ComponentCall = parser.Try(p, call(true))
		if a.ComponentCall == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "alias: missing component call",
				Primary: quickanno.Expected(p, p.Pos(), "a component call to alias"),
			})
		}

		return &a, nil
	}
}

func Block() parser.Func[*ast.Block] {
	return func(p *parser.Parser) (*ast.Block, *diagnostic.Diagnostic) {
		var b ast.Block

		b.Block = parser.TryKeywordAt(p, "block", comment.OrAnyWhitespace())
		if b.Block == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing block",
				Primary: quickanno.Expected(p, p.Pos(), "a block"),
			}
		}

		b.Name = parser.Try(p, golang.Identifier())
		if b.Name == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "block: missing name",
				Primary: quickanno.Expected(p, *b.Block, "a name of a block"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		b.Default = parser.Try(p, body.Body())

		return &b, nil
	}
}
