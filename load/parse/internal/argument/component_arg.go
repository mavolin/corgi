package argument

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ComponentArgument() parser.Func[*ast.ComponentArgument] {
	return func(p *parser.Parser) (*ast.ComponentArgument, *fancyerr.Error) {
		var arg ast.ComponentArgument

		arg.Name = parser.TryOptional(p, golang.Identifier(), nil)
		if arg.Name == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "component argument: missing name",
				Primary: quickanno.Expected(p, p.Pos(), "an argument name"),
			})
		}
		hasPreColonWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		arg.Colon = parser.TryOptionalRuneAt(p, ':', nil)
		if arg.Colon == nil {
			return nil, &fancyerr.Error{
				Message: "component argument: missing colon",
				Primary: quickanno.Expected(p, p.Pos(), "a colon separating the argument name and value"),
			}
		}
		hasPostColonWS := parser.TrySkip(p, comment.OrAnyWhitespace())

		if arg.Name == nil && !hasPostColonWS {
			return nil, &fancyerr.Error{
				Message: "missing component argument",
				Primary: quickanno.Expected(p, p.Pos(), "an argument name"),
			}
		}

		if !hasPostColonWS {
			p.CaptureError(&fancyerr.Error{
				Message: "missing whitespace after colon",
				Primary: quickanno.Expected(p, *arg.Colon, "a space, tab, or an inline block comment"),
			})

			// A string directly after the colon is one of two cases where we
			// can be reasonably sure that this is not a named attribute.
			if parser.MatchesToken(p, `"`) {
				hasPostColonWS = true
			}
		}

		pos := p.Pos()
		start := p.Index()
		arg.Value = parser.Try(p, code.Expression(code.Regular))
		if arg.Value == nil {
			if arg.Name == nil { // only a colon
				return nil, &fancyerr.Error{
					Message: "missing component argument",
					Primary: quickanno.Expected(p, arg.Start(), "a valid component argument"),
				}
			}

			// just as likely a named argument with a trailing colon
			if !hasPreColonWS && (parser.MatchesWS(p, whitespace.EOL()) || parser.MatchesAnyRune(p, ',', ')')) {
				return nil, &fancyerr.Error{
					Message: "missing component argument",
					Primary: quickanno.Expected(p, arg.Start(), "a valid component argument"),
				}
			}
			p.CaptureError(&fancyerr.Error{
				Message: "component argument: missing value",
				Primary: quickanno.Expected(p, pos, "a value for the argument"),
			})
		}

		if !hasPreColonWS && !hasPostColonWS {
			// The only other way we can be sure this is not a named attribute,
			// is if the value contains no `=` and at least one non-trailing
			// whitespace.
			var haveWS bool
			for i := start; i < p.Index(); i++ {
				prev := p.File.Raw[i-1] // safe because we know start > 0
				switch p.File.Raw[i] {
				case '=':
					switch prev {
					case '!', '<', '>', '=':
					default:
						return nil, &fancyerr.Error{
							Message: "missing component argument",
							Primary: quickanno.Expected(p, arg.Start(), "an argument name"),
							Hints: []fancyerr.Hint{
								{Hint: "If this is supposed to be a named attribute, add a space after the colon."},
							},
						}
					}
				case ' ', '\t', '\n':
					if !haveWS {
						haveWS = true
					}
				default:
					if haveWS {
						// the ws we have is non-trailing
						return &arg, nil
					}
				}
			}
		}

		return &arg, nil
	}
}
