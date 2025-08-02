package parser

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
)

func NextRune(p *Parser) rune {
	CommitWS(p)
	return p.next()
}

// TryToken attempts to match the given token verbatim.
func TryToken(p *Parser, s string) (ok bool) {
	if !MatchesToken(p, s) {
		RestoreWS(p)
		return false
	}

	CommitWS(p)
	p.skipString(s)
	return true
}

func TryOptionalToken(p *Parser, s string, ws WhitespaceFunc) (ok bool) {
	if !MatchesToken(p, s) {
		return false
	}

	CommitWS(p)
	p.skipString(s)
	if ws != nil {
		TrySkip(p, ws)
	}
	return true
}

func TryAnyToken(p *Parser, ss ...string) string {
	for _, s := range ss {
		if MatchesToken(p, s) {
			CommitWS(p)
			p.skipString(s)
			return s
		}
	}
	RestoreWS(p)
	return ""
}

func TryKeywordAt(p *Parser, k string, ws WhitespaceFunc) *ast.Position {
	pos := p.Pos()
	if !TryToken(p, k) || (!MatchesAnyRune(p, EOF, ':', '(', ';', '}', '\n', '\r') && !TrySkip(p, ws)) {
		return nil
	}
	return &pos
}

func TryOptionalKeywordAt(p *Parser, k string, ws WhitespaceFunc) *ast.Position {
	pos := p.Pos()
	if !TryOptionalToken(p, k, nil) || (!MatchesAnyRune(p, EOF, '(', ';', '}', '\n', '\r') && !TrySkip(p, ws)) {
		return nil
	}
	return &pos
}

func TryTokenAt(p *Parser, s string) *ast.Position {
	pos := p.Pos()
	if !TryToken(p, s) {
		return nil
	}
	return &pos
}

func TryRune(p *Parser, r rune) (ok bool) {
	if r != p.peek() {
		RestoreWS(p)
		return false
	}

	CommitWS(p)
	p.next()
	return true
}

func TryRuneAt(p *Parser, r rune) *ast.Position {
	pos := p.Pos()
	if !TryRune(p, r) {
		return nil
	}
	return &pos
}

func TryOptionalRune(p *Parser, r rune, ws WhitespaceFunc) (ok bool) {
	if r != p.peek() {
		return false
	}

	CommitWS(p)
	p.next()
	if ws != nil {
		TrySkip(p, ws)
	}
	return true
}

func TryOptionalRuneAt(p *Parser, r rune, ws WhitespaceFunc) *ast.Position {
	pos := p.Pos()
	if !TryOptionalRune(p, r, ws) {
		return nil
	}
	return &pos
}

// TryAnyRune attempts to match the next rune against any of the passed runes.
//
// It returns the matched rune, or -1 if none matched.
func TryAnyRune(p *Parser, rs ...rune) rune {
	return TryRunePredicate(p, func(r rune) bool {
		for _, rr := range rs {
			if r == rr {
				return true
			}
		}
		return false
	})
}

// TryRunePredicate attempts to match the next rune against the predicate.
// If successful, it returns the matched rune, otherwise, it returns -1.
func TryRunePredicate(p *Parser, pred func(rune) bool) rune {
	peek := p.peek()
	if peek == EOF || !pred(peek) {
		RestoreWS(p)
		return -1
	}
	CommitWS(p)
	return p.next()
}

// TokenWhile consumes runes as long as the predicate returns true.
// The predicate may invoke the parser inside the predicate, but must not
// consume any runes itself.
//
// If TokenWhile doesn't consume any runes, previously consumed whitespace is
// rolled back.
func TokenWhile(p *Parser, pred func() bool) string {
	restore := p.takeWSStart()
	start := p.Index()
	if !pred() {
		if p.Index() != start {
			panic("TokenWhile: predicate consumed runes")
		}
		p.RestoreState(restore)
		return ""
	}
	p.next()
	i := p.Index()
	for pred() {
		if p.Index() != i {
			panic("TokenWhile: predicate consumed runes")
		}
		r := p.next()
		if r == EOF {
			break
		}
		i = p.Index()
	}
	CommitWS(p) // the predicate might've set a restore point
	p.pool.Put(restore)
	return p.AST.Raw[start:p.Index()]
}

func Collect[T any](p *Parser, f Func[T], capacity int, ws WhitespaceFunc) []T {
	ts := make([]T, 0, capacity)
	for {
		if ws != nil {
			TrySkip(p, ws)
		}
		t, err := TryErr(p, f)
		if err != nil {
			break
		}
		ts = append(ts, t)
	}
	if len(ts) == 0 {
		return nil
	}
	return slices.Clip(ts)
}
