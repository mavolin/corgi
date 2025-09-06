package state

import (
	"fmt"
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Declaration() parser.Func[*ast.StateDeclaration] {
	return func(p *parser.Parser) *ast.StateDeclaration {
		state := parser.TryTokenAt(p, "state")
		if state == nil {
			return nil
		}

		var d ast.StateDeclaration
		d.State = state

		hasWS := parser.TrySkip(p, comment.OrAnyWhitespace())

		d.LParen = parser.TryOptionalRuneAt(p, '(', nil)
		if d.LParen == nil {
			if !hasWS {
				return nil
			}

			spec := parser.Try(p, Spec())
			if spec != nil {
				d.Specs = []*ast.StateSpec{spec}
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "missing state spec",
					Primary: quickanno.Expected(p, p.Pos(), "one or more identifiers"),
					Examples: []diagnostic.Example{
						{Example: "`bark = \"woof\"`"},
					},
				})
			}

			return &d
		}

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			spec := parser.Try(p, Spec())
			if spec == nil {
				break
			}
			d.Specs = append(d.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.Try(p, comment.AndMustEOS())
		}
		d.Specs = slices.Clip(d.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "state declaration: unexpected runes"
			p.CaptureError(err)
		}

		d.RParen = parser.TryOptionalRuneAt(p, ')', nil)
		if d.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "state declaration: missing ')'",
				Primary: quickanno.Expected(p, *d.LParen, "a closing ')' for the '(' here"),
			})
		}

		return &d
	}
}

func Spec() parser.Func[*ast.StateSpec] {
	return func(p *parser.Parser) *ast.StateSpec {
		names := parser.Try(p, list.CommaList("state name", "state names", golang.Identifier()))
		if names == nil {
			return nil
		}

		var s ast.StateSpec
		s.Names = names

		var pos ast.Position
		if parser.TrySkip(p, comment.OrHorizontalWhitespace()) {
			pos = p.Pos()
			s.Type = parser.TryOptional(p, golang.Type(), comment.OrHorizontalWhitespace())
		}

		s.EqualSign = parser.TryRuneAt(p, '=')
		if s.EqualSign == nil {
			if s.Type == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "state spec: missing type or value",
					Primary: quickanno.Expected(p, pos, "either a type or an equal sign"),
				})
			}
			return &s
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		valuesStart := p.Pos()
		s.Values = parser.Try(p, list.CommaList("state value", "state values", code.Expression(code.Regular)))
		valuesEnd := p.Pos()
		if s.Values == nil {
			if len(s.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "state spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "state spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(s.Names), " expressions")),
				})
			}
		}

		if len(s.Values) > 1 && len(s.Names) != len(s.Values) {
			if len(s.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "state spec: mismatched number of values and variables",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("expected a single expression, but found ", len(s.Values))),
					},
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "state spec: mismatched number of values and variables",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("a single or ", len(s.Names), " expressions, but found ", len(s.Values))),
					},
				})
			}
		}

		return &s
	}
}
