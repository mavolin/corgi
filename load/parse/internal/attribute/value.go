package attribute

import (
	"github.com/mavolin/corgi/escape/attrtype"
	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/fancyerr/anno"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/code"
	"github.com/mavolin/corgi/load/parse/internal/comment"
	"github.com/mavolin/corgi/load/parse/internal/golang"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
)

func Value() parser.Func[ast.AttributeValue] {
	return func(p *parser.Parser) (ast.AttributeValue, *fancyerr.Error) {
		if v, ok := parser.Try(p, TypedAttributeValue()); ok {
			return v, nil
		} else if v, ok := parser.Try(p, ExpressionValue()); ok {
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

func ExpressionValue() parser.Func[ast.ExpressionAttributeValue] {
	return func(p *parser.Parser) (ast.ExpressionAttributeValue, *fancyerr.Error) {
		expr, ok := parser.Try(p, code.Expression())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return ast.ExpressionAttributeValue(expr), nil
	}
}

func TypedAttributeValue() parser.Func[*ast.TypedAttributeValue] {
	return func(p *parser.Parser) (*ast.TypedAttributeValue, *fancyerr.Error) {
		v := &ast.TypedAttributeValue{Position: p.Pos()}

		// remember that cases that are a superset of another case must come
		// first
		switch {
		case parser.TryToken(p, "unsafeBool"):
			v.Type = attrtype.UnsafeBool
		case parser.TryToken(p, "unsafe"):
			v.Type = attrtype.Unsafe
		case parser.TryToken(p, "bool"):
			v.Type = attrtype.Bool
		case parser.TryToken(p, "text"):
			v.Type = attrtype.Text
		case parser.TryToken(p, "css"):
			v.Type = attrtype.CSS
		case parser.TryToken(p, "js"):
			v.Type = attrtype.JS
		case parser.TryToken(p, "urlList"):
			v.Type = attrtype.URLList
		case parser.TryToken(p, "url"):
			v.Type = attrtype.URL
		case parser.TryToken(p, "resourceURL"):
			v.Type = attrtype.ResourceURL
		case parser.TryToken(p, "srcset"):
			v.Type = attrtype.Srcset
		}

		extra := parser.TokenWhile(p, func() bool {
			return parser.Matches(p, golang.AnyIdentifierRune())
		})
		end := p.Pos()

		parser.Try(p, comment.OrHorizontalWhitespace())

		lParenPos := p.Pos()
		v.LParen = &lParenPos

		if !v.Type.IsValid() || extra != "" || !parser.TryRune(p, '(') {
			return nil, &fancyerr.Error{
				Message: "invalid type",
				Primary: []fancyerr.Annotation{anno.Range(p.File, v.Position, end, "not a valid attribute type")},
			}
		}

		v.Value = parser.Must(p, ExpressionValue())

		pos := p.Pos()

		parser.Try(p, comment.OrHorizontalWhitespace())

		if parser.TryRune(p, ',') {
			parser.Try(p, comment.OrAnyWhitespace())
		}

		rParenPos := p.Pos()
		if !parser.TryRune(p, ')') {
			p.CaptureError(&fancyerr.Error{
				Message: "missing closing parenthesis",
				Primary: quickanno.Expected(p, pos, "a closing parenthesis"),
			})
		} else {
			v.RParen = &rParenPos
		}

		return v, nil
	}
}
