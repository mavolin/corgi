package argument

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ComponentArgument() parser.Func[*ast.ComponentArgument] {
	return func(p *parser.Parser) *ast.ComponentArgument {
		name := parser.TryOptional(p, golang.Identifier(), nil)
		if name == nil {
			return nil
		}
		hasPreColonWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		colon := parser.TryOptionalRuneAt(p, ':', nil)
		if colon == nil {
			return nil
		}
		hasPostColonWS := parser.TrySkip(p, comment.OrAnyWhitespace())

		if name == nil && !hasPostColonWS {
			return nil
		}

		if !hasPostColonWS {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "missing whitespace after colon",
				Primary: quickanno.Expected(p, *colon, "a space, tab, or an inline block comment"),
			})

			// A string directly after the colon is one of two cases where we
			// can be reasonably sure that this is not a named attribute.
			if parser.MatchesToken(p, `"`) {
				hasPostColonWS = true
			}
		}

		start := p.RuneIndex()
		value := parser.Try(p, code.Expression())
		if value == nil {
			if name == nil { // only a colon
				return nil
			}

			// just as likely a named argument with a trailing colon
			if !hasPreColonWS && (parser.MatchesWS(p, whitespace.EOL()) || parser.MatchesAnyRune(p, ',', ')')) {
				return nil
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component argument: missing value",
				Primary: quickanno.Expected(p, p.Pos(), "a value for the argument"),
			})
		}

		var arg ast.ComponentArgument
		arg.Name = name
		arg.Colon = colon
		arg.Value = value

		if !hasPreColonWS && !hasPostColonWS {
			// The only other way we can be sure this is not a named attribute,
			// is if the value contains no `=` and at least one non-trailing
			// whitespace.
			end := p.RuneIndex()

			var haveWS bool
			for i := start; i < end; i++ {
				prev := p.RuneAt(i - 1) // safe because we know start > 0
				switch p.RuneAt(i) {
				case '=':
					switch prev {
					case '!', '<', '>', '=':
					default:
						return nil
					}
				case ' ', '\t', '\n':
					if !haveWS {
						haveWS = true
					}
				default:
					if haveWS {
						// the ws we have is non-trailing
						return &arg
					}
				}
			}
		}

		return &arg
	}
}
