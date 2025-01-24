package attribute

import (
	"fmt"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Attribute() parser.Func[ast.Attribute] {
	return func(p *parser.Parser) (ast.Attribute, *fancyerr.Error) {
		if a, ok := parser.TryOk(p, AndPlaceholder()); ok {
			return a, nil
		} else if a, ok := parser.TryOk(p, IDShorthand()); ok {
			return a, nil
		} else if a, ok := parser.TryOk(p, ClassShorthand()); ok {
			return a, nil
		} else if a, ok := parser.TryOk(p, NamedAttribute()); ok {
			return a, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing attribute",
			Primary: quickanno.Expected(p, p.Pos(), "an attribute"),
			Examples: []fancyerr.Example{
				{Title: "value attribute", Example: "`class=\"woof\"`"},
				{Title: "boolean attribute", Example: "`async`"},
				{Title: "class shorthand", Example: "`.bark`"},
			},
		}
	}
}

func AndPlaceholder() parser.Func[*ast.AndPlaceholder] {
	return func(p *parser.Parser) (*ast.AndPlaceholder, *fancyerr.Error) {
		pos := p.Pos()
		if !parser.TryRune(p, '&') {
			return nil, &fancyerr.Error{
				Message: "missing `&`",
				Primary: quickanno.Expected(p, pos, "an and placeholder (`&`)"),
			}
		}

		return &ast.AndPlaceholder{Position: pos}, nil
	}
}

func NamedAttribute() parser.Func[*ast.NamedAttribute] {
	return func(p *parser.Parser) (*ast.NamedAttribute, *fancyerr.Error) {
		attr := &ast.NamedAttribute{}

		attrName, ok := parser.TryOk(p, Name())
		if ok {
			attr.Name = *attrName
		} else {
			attr.Name.Position = p.Pos()
			// let this slide, as long as there is an equal sign following
			p.CaptureError(&fancyerr.Error{
				Message:  "missing attribute name",
				Primary:  quickanno.Expected(p, p.Pos(), "an attribute name before the `=`"),
				Examples: []fancyerr.Example{{Example: "`class=\"woof\"`"}},
			})
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ',', ')')
		if err != nil {
			err.Message = "unexpected runes after attribute name"

			// check if this could possibly be a component argument, i.e.
			// if the name contains a colon and no parentheses
			for i, r := range attr.Name.Name {
				if r == '(' || r == ')' {
					break
				} else if r == ':' && i > 0 { // this could be a comp arg
					err.Hints = append(err.Hints, fancyerr.Hint{
						Hint:    "If this is supposed to be a component argument, add a space after the colon.",
						Example: "`" + attr.Name.Name[:i] + ": ...`",
					})
					break
				}
			}

			p.CaptureError(err)
		}

		assignPos := p.Pos()
		if !parser.TryRune(p, '=') {
			if attr.Name.Name == "" { // we have neither a name nor a =, this is not an attr
				return nil, &fancyerr.Error{
					Message: "missing named attribute",
					Primary: quickanno.Expected(p, attr.Pos(), "an attribute"),
					Examples: []fancyerr.Example{
						{Title: "value attribute", Example: "`class=\"woof\"`"},
						{Title: "boolean attribute", Example: "`async`"},
					},
				}
			}

			return attr, nil
		}
		attr.Assign = &assignPos

		parser.TrySkip(p, comment.OrAnyWhitespace())

		attr.Value = parser.Must(p, Value())
		return attr, nil
	}
}

func Name() parser.Func[*ast.AttributeName] {
	return func(p *parser.Parser) (*ast.AttributeName, *fancyerr.Error) {
		name := &ast.AttributeName{Position: p.Pos()}

		var parenCount int
		name.Name = parser.TokenWhile(p, func() bool {
			if !parser.Matches(p, html.AttributeNameRune()) {
				return false
			}

			if parser.MatchesToken(p, "(") {
				parenCount++
				return true
			} else if parser.MatchesToken(p, ")") {
				parenCount--
				return parenCount >= 0
			}

			return !parser.MatchesToken(p, ",")
		})

		if name.Name == "" {
			return nil, &fancyerr.Error{
				Message: "missing attribute name",
				Primary: quickanno.Expected(p, name.Position, "an attribute name"),
			}
		} else if parenCount > 0 {
			p.CaptureError(&fancyerr.Error{
				Message: "unbalanced parentheses",
				Primary: quickanno.Expected(p, name.Position, fmt.Sprintf("%d closing parenthesis", parenCount)),
				Explanation: fmt.Sprint("Attributes may contain parentheses, but they must be balanced to "+
					"help the parser distinguish between the end of an attribute list and an "+
					"attribute name. You currently have an excess of ", parenCount, " opening parentheses, "+
					"which need to be closed to make this a valid attribute name."),
			})
		}

		return name, nil
	}
}
