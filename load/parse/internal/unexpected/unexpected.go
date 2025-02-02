package unexpected

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// UntilAnyRune captures any unexpected runes, until any of the passed runes is
// detected.
//
// You may optionally provide a [parser.WhitespaceFunc] to capture whitespace
// between subsequent sets of unexpected runes.
// If you provide none, UntilAnyRune will consume no whitespace.
//
// If UntilAnyRune finds any non-whitespace rune not contained in runes, it
// returns with an error highlighting that segment.
// Leading and trailing whitespace, as determined by the whitespace func is
// ignored.
//
// It is assumed that wsFunc captures all consecutive whitespace and that upon
// returning, the next rune is a non-whitespace rune.
func UntilAnyRune(p *parser.Parser, wsFunc parser.WhitespaceFunc, runes ...rune) *fancyerr.Error {
	var start, end ast.Position

	state := p.CloneState()

	if wsFunc == nil {
		start = p.Pos()
		s := parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, runes...) && !parser.MatchesAnyRune(p, whitespace.Runes...)
		})
		if s == "" {
			p.RestoreState(state)
			return nil
		}
		end = p.Pos()
	} else {
		parser.TrySkip(p, wsFunc)
		start = p.Pos()
		for {
			s := parser.TokenWhile(p, func() bool {
				return !parser.MatchesAnyRune(p, runes...) && !parser.MatchesAnyRune(p, whitespace.Runes...)
			})
			if s == "" {
				p.RestoreState(state)
				return nil
			}
			end = p.Pos()

			hasWS := parser.TrySkipOk(p, wsFunc)
			if !hasWS {
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
