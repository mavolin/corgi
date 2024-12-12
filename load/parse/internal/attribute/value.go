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
		if v, ok := parser.TryOk(p, TypedAttributeValue()); ok {
			return v, nil
		} else if v, ok := parser.TryOk(p, ExpressionValue()); ok {
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
		expr, ok := parser.TryOk(p, code.Expression())
		if !ok {
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
		t, ok := parser.TryOk(p, Type())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing typed attribute value",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute type"),
			}
		}

		v := &ast.TypedAttributeValue{Type: *t}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		lParenPos := p.Pos()
		if !parser.TryRune(p, '(') {
			p.CaptureError(&fancyerr.Error{
				Message: "missing opening parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "an opening parenthesis"),
			})
			return v, nil
		}
		v.LParen = &lParenPos

		parser.TrySkip(p, comment.OrAnyWhitespace())

		v.Value = parser.Must(p, ExpressionValue())

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		rParenPos := p.Pos()
		if parser.TryRune(p, ')') {
			v.RParen = &rParenPos
		} else {
			p.CaptureError(&fancyerr.Error{
				Message: "missing closing parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "a closing parenthesis"),
			})
		}

		return v, nil
	}
}
