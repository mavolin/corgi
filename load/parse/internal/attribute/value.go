package attribute

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Value() parser.Func[ast.AttributeValue] {
	return func(p *parser.Parser) ast.AttributeValue {
		if v := parser.Try(p, TypedAttributeValue()); v != nil {
			return v
		} else if v := parser.Try(p, ExpressionValue()); v != nil {
			return v
		}
		return nil
	}
}

func ExpressionValue() parser.Func[*ast.ExpressionAttributeValue] {
	return func(p *parser.Parser) *ast.ExpressionAttributeValue {
		expr := parser.Try(p, code.Expression(code.Regular))
		if expr == nil {
			return nil
		}
		return (*ast.ExpressionAttributeValue)(expr)
	}
}

func TypedAttributeValue() parser.Func[*ast.TypedAttributeValue] {
	return func(p *parser.Parser) *ast.TypedAttributeValue {
		attrType := parser.Try(p, Type())
		if attrType == nil {
			return nil
		}

		var v ast.TypedAttributeValue
		v.Type = attrType

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		v.LParen = parser.TryRuneAt(p, '(')
		if v.LParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "typed attribute value: missing opening parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "an opening parenthesis"),
			})
			return &v
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		if ev := parser.Try(p, ExpressionValue()); ev != nil {
			v.Value = ev // typed nil
		} else {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "typed attribute value: missing value",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute value"),
			})
			return &v
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		v.RParen = parser.TryRuneAt(p, ')')
		if v.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "missing closing parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "a closing parenthesis"),
			})
		}

		return &v
	}
}
