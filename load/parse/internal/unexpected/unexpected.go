package unexpected

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// UntilAnyRune captures any unexpected runes, until any of the passed runes is
// detected.
//
// You may optionally provide a [parser.WhitespaceFunc] to capture whitespace
// between subsequent sets of unexpected runes.
// If you provide none, UntilAnyRune will consume no whitespace.
// Upon returning, the parser will have skipped any occurring whitespace.
//
// If UntilAnyRune finds any non-whitespace rune not contained in runes, it
// returns with an error highlighting that segment.
// Trailing whitespace, as determined by the whitespace func is
// ignored.
//
// It is assumed that wsFunc captures all consecutive whitespace and that upon
// returning, the next rune is a non-whitespace rune.
func UntilAnyRune(p *parser.Parser, ws parser.WhitespaceFunc, runes ...rune) *diagnostic.Diagnostic {
	var start, end ast.Position

	if ws == nil {
		start = p.Pos()
		s := parser.OptionalTokenWhile(p, nil, func() bool {
			r := parser.PeekRune(p)
			return r != parser.EOF && !slices.Contains(runes, r) && !parser.MatchesAnyRune(p, whitespace.Runes...) &&
				!parser.MatchesAnyToken(p, "//", "/*")
		})
		if s == "" {
			return nil
		}
		end = p.Pos()
	} else {
		parser.TrySkip(p, ws)
		start = p.Pos()
		for {
			s := parser.OptionalTokenWhile(p, nil, func() bool {
				r := parser.PeekRune(p)
				return r != parser.EOF && !slices.Contains(runes, r) && !parser.MatchesAnyRune(p, whitespace.Runes...) &&
					!parser.MatchesAnyToken(p, "//", "/*")
			})
			if s == "" {
				break
			}
			end = p.Pos()

			if !parser.TrySkip(p, ws) {
				break
			}
		}
		if end == (ast.Position{}) {
			return nil
		}
	}

	return &diagnostic.Diagnostic{
		Message: "unexpected runes",
		Primary: []diagnostic.Annotation{
			anno.Range(p.File, start, end, "remove this"),
		},
	}
}

// UntilAnyToken is like [UntilAnyRune], but uses tokens instead of runes.
func UntilAnyToken(p *parser.Parser, ws parser.WhitespaceFunc, tokens ...string) *diagnostic.Diagnostic {
	var start, end ast.Position

	if ws == nil {
		start = p.Pos()
		s := parser.OptionalTokenWhile(p, nil, func() bool {
			return !parser.MatchesAnyToken(p, tokens...) && !parser.MatchesAnyRune(p, whitespace.Runes...)
		})
		if s == "" {
			return nil
		}
		end = p.Pos()
	} else {
		start = p.Pos()
		for {
			s := parser.OptionalTokenWhile(p, nil, func() bool {
				return !parser.MatchesAnyToken(p, tokens...) && !parser.MatchesAnyRune(p, whitespace.Runes...)
			})
			if s == "" {
				break
			}
			end = p.Pos()

			if !parser.TrySkip(p, ws) {
				break
			}
		}
		if end == (ast.Position{}) {
			return nil
		}
	}

	return &diagnostic.Diagnostic{
		Message: "unexpected runes",
		Primary: []diagnostic.Annotation{
			anno.Range(p.File, start, end, "remove this"),
		},
	}
}
