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
		arg := &ast.ComponentArgument{Position: p.Pos()}

		var ok bool
		arg.Name, ok = parser.TryOk(p, golang.Identifier())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing component argument",
				Primary: quickanno.Expected(p, arg.Position, "an argument name"),
			}
		}

		hasPreColonWS := parser.TrySkipOk(p, comment.OrHorizontalWhitespace())

		colon := p.Pos()
		if !parser.TryRune(p, ':') {
			return nil, &fancyerr.Error{
				Message: "missing colon",
				Primary: quickanno.Expected(p, arg.Position, "a colon separating the argument name and value"),
			}
		}
		arg.Colon = &colon

		hasPostColonWS := parser.TrySkipOk(p, comment.OrHorizontalWhitespace())
		if !hasPostColonWS {
			p.CaptureError(&fancyerr.Error{
				Message: "missing whitespace after colon",
				Primary: quickanno.Expected(p, colon, "a space, tab, or an inline block comment"),
			})

			// A string directly after the colon is one of two cases where we
			// can be reasonably sure that this is not a named attribute.
			if parser.MatchesToken(p, `"`) {
				hasPostColonWS = true
			}
		}

		pos := p.Pos()
		start := p.Index()
		arg.Value, ok = parser.TryOk(p, code.Expression())
		if !ok {
			// just as likely a named argument with a trailing colon
			if !hasPreColonWS && (parser.MatchesWS(p, whitespace.EOL()) || parser.MatchesAnyRune(p, ',', ')')) {
				return nil, &fancyerr.Error{
					Message: "missing component argument",
					Primary: quickanno.Expected(p, arg.Position, "a valid component argument"),
				}
			}
			p.CaptureError(&fancyerr.Error{
				Message: "missing component argument value",
				Primary: quickanno.Expected(p, pos, "a value for the argument"),
			})
		}

		if !hasPreColonWS && !hasPostColonWS {
			// The only other way we can be sure this is not a named attribute,
			// is if this value contains no `=` and at least one non-trailing
			// whitespace.
			var haveWS bool
			for i := start; i < p.Index(); i++ {
				prev := p.File.Raw[i-1] // safe because we know start > 0
				switch p.File.Raw[i] {
				case '=':
					switch prev {
					case '!', '<', '>', '=':
					default:
						goto err
					}
				case ' ', '\t', '\n':
					if !haveWS {
						haveWS = true
					}
				default:
					if haveWS {
						// the ws we have is non-trailing
						return arg, nil
					}
				}
			}

		err:
			return nil, &fancyerr.Error{
				Message: "missing component argument",
				Primary: quickanno.Expected(p, arg.Position, "an argument name"),
				Hints: []fancyerr.Hint{
					{Hint: "If this is supposed to be a named attribute, add a space after the colon."},
				},
			}
		}

		return arg, nil
	}
}
