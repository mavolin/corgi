package unexpected

import (
	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/fancyerr/anno"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/whitespace"
)

// UntilAnyRune captures unexpected runes, that are still on the same
// line, until any of the passed runes is detected.
//
// You may optionally provide a [parser.Func] to capture whitespace between
// subsequent sets of unexpected runes.
// If you provide none, UntilAnyRune will consume no whitespace.
//
// If UntilAnyRune finds any non-whitespace rune not contained in runes, it
// returns with an error highlighting that segment.
// Leading and trailing whitespace, as determined by the whitespace func is
// ignored.
//
// It is assumed that wsFunc captures all consecutive whitespace and that upon
// returning, the next rune is a non-whitespace rune.
func UntilAnyRune(p *parser.Parser, wsFunc parser.Func[struct{}], runes ...rune) *fancyerr.Error {
	var start, end ast.Position

	if wsFunc == nil {
		start = p.Pos()
		s := parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, runes...) && !parser.MatchesAnyRune(p, whitespace.Runes...)
		})
		if s == "" {
			return nil
		}
		end = p.Pos()
	} else {
		parser.Try(p, wsFunc)
		start = p.Pos()
		for {
			s := parser.TokenWhile(p, func() bool {
				return !parser.MatchesAnyRune(p, runes...) && !parser.MatchesAnyRune(p, whitespace.Runes...)
			})
			if s == "" {
				return nil
			}
			end = p.Pos()

			if _, ok := parser.Try(p, wsFunc); !ok {
				break
			}
		}
	}

	return &fancyerr.Error{
		Message: "unexpected runes",
		Primary: []fancyerr.Annotation{
			anno.Range(p.File, start, end, "remove this"),
		},
	}
}
