package attribute

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Value() parser.Func[ast.AttributeValue] {
	return func(p *parser.Parser) (ast.AttributeValue, *fancyerr.Error) {
		if v := parser.Try(p, TypedAttributeValue()); v != nil {
			return v, nil
		} else if v := parser.Try(p, ExpressionValue()); v != nil {
			return v, nil
		}
		return nil, &fancyerr.Error{
			Message: "missing attribute value",
			Primary: quickanno.Expected(p, p.Pos(), "an attribute value"),
			Examples: []fancyerr.Example{
				{Title: "expression", Example: "`class=\"woof\"`"},
				{Title: "typed attribute", Example: "`data-website=url(\"https://mavolin.co\")`"},
			},
		}
	}
}

func ExpressionValue() parser.Func[*ast.ExpressionAttributeValue] {
	return func(p *parser.Parser) (*ast.ExpressionAttributeValue, *fancyerr.Error) {
		expr := parser.Try(p, code.Expression(code.Regular))
		if expr == nil {
			return nil, &fancyerr.Error{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return (*ast.ExpressionAttributeValue)(expr), nil
	}
}

func TypedAttributeValue() parser.Func[*ast.TypedAttributeValue] {
	return func(p *parser.Parser) (*ast.TypedAttributeValue, *fancyerr.Error) {
		var v ast.TypedAttributeValue

		v.Type = parser.Try(p, Type())
		if v.Type == nil {
			return nil, &fancyerr.Error{
				Message: "missing typed attribute value",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute type"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		v.LParen = parser.TryRuneAt(p, '(')
		if v.LParen == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "typed attribute value: missing opening parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "an opening parenthesis"),
			})
			return &v, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		v.Value = parser.Must(p, ExpressionValue())
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		v.RParen = parser.TryRuneAt(p, ')')
		if v.RParen == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "missing closing parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "a closing parenthesis"),
			})
		}

		return &v, nil
	}
}
