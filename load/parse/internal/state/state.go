package state

import (
	"fmt"
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Declaration() parser.Func[*ast.StateDeclaration] {
	return func(p *parser.Parser) (*ast.StateDeclaration, *fancyerr.Error) {
		d := &ast.StateDeclaration{State: p.Pos()}
		if !parser.TryToken(p, "state") {
			return nil, &fancyerr.Error{
				Message: "missing state declaration",
				Primary: quickanno.Expected(p, d.Pos(), "a state declaration"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		d.LParen = p.PosPtr()
		if !parser.TryOptionalRune(p, '(') {
			d.LParen = nil

			d.Specs = make([]*ast.StateSpec, 1)
			d.Specs[0] = parser.Must(p, Spec())

			parser.MustSkip(p, comment.AndMustEOS())
			return d, nil
		}

		d.Specs = make([]*ast.StateSpec, 0, 18)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			spec, ok := parser.TryOk(p, Spec())
			if !ok {
				break
			}
			d.Specs = append(d.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.MustSkip(p, comment.AndEOS())
		}
		d.Specs = slices.Clip(d.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "state declaration: unexpected runes"
			p.CaptureError(err)
		}

		d.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			d.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "state declaration: missing ')'",
				Primary: quickanno.Expected(p, *d.LParen, "a closing ')' for the '(' here"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return d, nil
	}
}

func Spec() parser.Func[*ast.StateSpec] {
	return func(p *parser.Parser) (*ast.StateSpec, *fancyerr.Error) {
		names, ok := parser.TryOk(p, list.CommaList("state name", "state names", golang.Identifier()))
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing state spec",
				Primary: quickanno.Expected(p, p.Pos(), "one or more identifiers"),
				Examples: []fancyerr.Example{
					{Example: "`bark = \"woof\"`"},
				},
			}
		}

		s := &ast.StateSpec{Names: names}

		var pos ast.Position
		if parser.TrySkipOk(p, comment.OrHorizontalWhitespace()) {
			pos = p.Pos()

			s.Type, ok = parser.TryOptionalOk(p, golang.Type())
			if ok {
				parser.TrySkip(p, comment.OrHorizontalWhitespace())
			}
		}

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ';', ')')
		if err != nil {
			err.Message = "state spec: unexpected runes"
			p.CaptureError(err)
		}

		s.EqualSign = p.PosPtr()
		if !parser.TryRune(p, '=') {
			s.EqualSign = nil
			if s.Type == nil {
				p.CaptureError(&fancyerr.Error{
					Message: "state spec: missing type or value",
					Primary: quickanno.Expected(p, pos, "either a type or an equal sign"),
				})
			}
			return s, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		s.Values, ok = parser.TryOk(p, list.CommaList("state value", "state values", code.Expression()))
		if !ok {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "state spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "state spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(names), " expressions")),
				})
			}
		}

		if len(s.Values) > 1 && len(s.Names) != len(s.Values) {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "state spec: mismatched number of values and variables",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, s.Values[0].Pos(), s.Values[len(s.Values)-1].End(),
							fmt.Sprint("expected a single expression, but found ", len(s.Values))),
					},
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "state spec: mismatched number of values and variables",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, s.Values[0].Pos(), s.Values[len(s.Values)-1].End(),
							fmt.Sprint("a single or ", len(names), " expressions, but found ", len(s.Values))),
					},
				})
			}
		}

		return s, nil
	}
}
