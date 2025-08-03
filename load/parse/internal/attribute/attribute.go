package attribute

import (
	"fmt"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Attribute() parser.Func[ast.Attribute] {
	return func(p *parser.Parser) ast.Attribute {
		if a := parser.Try(p, AndPlaceholder()); a != nil {
			return a
		} else if a := parser.Try(p, IDShorthand()); a != nil {
			return a
		} else if a := parser.Try(p, ClassShorthand()); a != nil {
			return a
		} else if a := parser.Try(p, NamedAttribute()); a != nil {
			return a
		}

		return nil
	}
}

func AndPlaceholder() parser.Func[*ast.AndPlaceholder] {
	return func(p *parser.Parser) *ast.AndPlaceholder {
		var ap ast.AndPlaceholder

		ap.And = parser.TryRuneAt(p, '&')
		if ap.And == nil {
			return nil
		}
		return &ap
	}
}

func NamedAttribute() parser.Func[*ast.NamedAttribute] {
	return func(p *parser.Parser) *ast.NamedAttribute {
		var attr ast.NamedAttribute

		attr.Name = parser.TryOptional(p, Reference(), comment.OrHorizontalWhitespace())
		if attr.Name == nil {
			// let this slide, as long as there is an equal sign following
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "missing attribute name",
				Primary:  quickanno.Expected(p, p.Pos(), "an attribute name before the `=`"),
				Examples: []diagnostic.Example{{Example: "`class=\"woof\"`"}},
			})
		}
		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ',', ')')
		if err != nil {
			err.Message = "unexpected runes after attribute name"

			// check if this could possibly be a component argument, i.e.
			// if the name contains a colon and no parentheses
			if attr.Name != nil && attr.Name.Package == nil && attr.Name.Dot == nil {
				for i, r := range attr.Name.Name.Name {
					if r == '(' || r == ')' {
						break
					} else if r == ':' && i > 0 { // this could be a comp arg
						err.Hints = append(err.Hints, diagnostic.Hint{
							Hint:    "If this is supposed to be a component argument, add a space after the colon.",
							Example: "`" + attr.Name.Name.Name[:i] + ": ...`",
						})
						break
					}
				}
			}

			p.CaptureError(err)
		}

		attr.EqualSign = parser.TryOptionalRuneAt(p, '=', comment.OrAnyWhitespace())
		if attr.EqualSign == nil {
			if attr.Name == nil { // we have neither a name nor a =, this is not an attr
				return nil
			}

			return &attr
		}

		attr.Value = parser.Try(p, Value())
		if attr.Value == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "named attribute: missing value",
				Primary: quickanno.Expected(p, p.Pos(), "a value for the attribute"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *attr.EqualSign, "because of this equal sign"),
				},
				Hints: []diagnostic.Hint{{Hint: "If you don't want to provide a value, remove the equal sign."}},
			})
			return &attr
		}
		return &attr
	}
}

func Reference() parser.Func[*ast.AttributeReference] {
	return func(p *parser.Parser) *ast.AttributeReference {
		var ref ast.AttributeReference

		state := p.CloneState()

		ref.Package = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		ref.Dot = parser.TryOptionalRuneAt(p, '.', comment.OrAnyWhitespace())
		if ref.Dot == nil {
			ref.Package = nil
			p.RestoreState(state)
		} else if ref.Package == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute reference: missing package name",
				Primary: quickanno.Expected(p, p.Pos(), "a package name before the `.`"),
			})
		}

		ref.Name = parser.Try(p, Name())
		if ref.Name == nil {
			return nil
		}
		return &ref
	}
}

func Name() parser.Func[*ast.AttributeName] {
	return func(p *parser.Parser) *ast.AttributeName {
		var name ast.AttributeName
		name.Position = p.PosPtr()

		var parenCount int
		name.Name = parser.TokenWhile(p, func() bool {
			if !parser.Matches(p, html.AttributeNameRune()) {
				return false
			}

			if parser.MatchesAnyRune(p, '(', '[') {
				parenCount++
				return true
			} else if parser.MatchesAnyRune(p, ')', ']') {
				parenCount--
				return parenCount >= 0
			}

			return !parser.MatchesToken(p, ",")
		})

		if name.Name == "" {
			return nil
		} else if parenCount > 0 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute name: unbalanced parentheses/brackets",
				Primary: []diagnostic.Annotation{
					anno.Position(p.File, name.End(), fmt.Sprintf("expected %d closing parenthesis/brackets", parenCount)),
				},
				Explanation: fmt.Sprint("Attributes may contain parentheses/brackets, but they must be balanced to "+
					"help the parser distinguish between the end of an attribute list and an "+
					"attribute name. You currently have an excess of ", parenCount, " opening parentheses, "+
					"which need to be closed to make this a valid attribute name."),
			})
		}

		return &name
	}
}
