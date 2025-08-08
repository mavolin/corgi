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
		and := parser.TryRuneAt(p, '&')
		if and == nil {
			return nil
		}

		var ap ast.AndPlaceholder
		ap.And = and
		return &ap
	}
}

func NamedAttribute() parser.Func[*ast.NamedAttribute] {
	return func(p *parser.Parser) *ast.NamedAttribute {
		name := parser.TryOptional(p, Reference(), comment.OrHorizontalWhitespace())
		if name == nil {
			// Check if there's an equal sign, otherwise not an attribute
			// We'll check this later after trying to capture unexpected runes
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
			if name != nil && name.Package == nil && name.Dot == nil {
				for i, r := range name.Name.Name {
					if r == '(' || r == ')' {
						break
					} else if r == ':' && i > 0 { // this could be a comp arg
						err.Hints = append(err.Hints, diagnostic.Hint{
							Hint:    "If this is supposed to be a component argument, add a space after the colon.",
							Example: "`" + name.Name.Name[:i] + ": ...`",
						})
						break
					}
				}
			}

			p.CaptureError(err)
		}

		equalSign := parser.TryOptionalRuneAt(p, '=', comment.OrAnyWhitespace())
		if equalSign == nil {
			if name == nil { // we have neither a name nor a =, this is not an attr
				return nil
			}

			var attr ast.NamedAttribute
			attr.Name = name
			return &attr
		}

		value := parser.Try(p, Value())

		var attr ast.NamedAttribute
		attr.Name = name
		attr.EqualSign = equalSign
		attr.Value = value

		if value == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "named attribute: missing value",
				Primary: quickanno.Expected(p, p.Pos(), "a value for the attribute"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *attr.EqualSign, "because of this equal sign"),
				},
				Hints: []diagnostic.Hint{{Hint: "If you don't want to provide a value, remove the equal sign."}},
			})
		}

		return &attr
	}
}

func Reference() parser.Func[*ast.AttributeReference] {
	return func(p *parser.Parser) *ast.AttributeReference {
		state := p.CloneState()

		pkg := parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		dot := parser.TryOptionalRuneAt(p, '.', comment.OrAnyWhitespace())

		if dot == nil {
			pkg = nil
			p.RestoreState(state)
		} else if pkg == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute reference: missing package name",
				Primary: quickanno.Expected(p, p.Pos(), "a package name before the `.`"),
			})
		}

		name := parser.Try(p, Name())
		if name == nil {
			return nil
		}

		var ref ast.AttributeReference
		ref.Package = pkg
		ref.Dot = dot
		ref.Name = name
		return &ref
	}
}

func Name() parser.Func[*ast.AttributeName] {
	return func(p *parser.Parser) *ast.AttributeName {
		pos := p.Pos()

		var parenCount int
		nameStr := parser.TokenWhile(p, func() bool {
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

		if nameStr == "" {
			return nil
		}

		var name ast.AttributeName
		name.Position = &pos
		name.Name = nameStr

		if parenCount > 0 {
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
