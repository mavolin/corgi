package parser

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
)

// TryToken attempts to match the given token verbatim.
func TryToken(p *Parser, s string) (ok bool) {
	restore := p.state.takeWSStart()
	if !MatchesToken(p, s) {
		p.RestoreState(restore)
		return false
	}

	for range s {
		p.next()
	}
	return true
}

func TryOptionalToken(p *Parser, s string, ws WhitespaceFunc) (ok bool) {
	state := p.CloneState()
	if !MatchesToken(p, s) {
		p.RestoreState(state)
		return false
	}

	p.state.ws = nil
	for range s {
		p.next()
	}
	if ws != nil {
		TrySkip(p, ws)
	}
	return true
}

func TryAnyToken(p *Parser, ss ...string) string {
	restore := p.state.takeWSStart()
	for _, s := range ss {
		if MatchesToken(p, s) {
			for range s {
				p.next()
			}
			return s
		}
	}
	p.RestoreState(restore)
	return ""
}

func TryKeywordAt(p *Parser, k string, ws WhitespaceFunc) *ast.Position {
	pos := p.Pos()
	if !TryToken(p, k) || (!MatchesAnyRune(p, '(', ';', '}', '\n', '\r') && !TrySkip(p, ws)) {
		return nil
	}
	return &pos
}

func TryOptionalKeywordAt(p *Parser, k string, ws WhitespaceFunc) *ast.Position {
	pos := p.Pos()
	if !TryOptionalToken(p, k, nil) || (!MatchesAnyRune(p, '(', ';', '}', '\n', '\r') && !TrySkip(p, ws)) {
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

func TryOptionalTokenAt(p *Parser, s string, ws WhitespaceFunc) *ast.Position {
	pos := p.Pos()
	if !TryOptionalToken(p, s, ws) {
		return nil
	}
	return &pos
}

func TryRune(p *Parser, r rune) (ok bool) {
	restore := p.state.takeWSStart()
	if r != p.peek() {
		p.RestoreState(restore)
		return false
	}
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
	state := p.CloneState()
	if r != p.peek() {
		p.RestoreState(state)
		return false
	}
	p.state.ws = nil
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
	restore := p.state.takeWSStart()
	peek := p.peek()
	if peek == EOF {
		return -1
	}
	if !pred(peek) {
		p.RestoreState(restore)
		return -1
	}
	return p.next()
}

// TokenWhile consumes runes as long as the predicate returns true.
// The predicate may invoke the parser inside the predicate, but must not
// consume any runes itself.
//
// If TokenWhile doesn't consume any runes, previously consumed whitespace is
// rolled back.
func TokenWhile(p *Parser, pred func() bool) string {
	restore := p.state.takeWSStart()
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
	p.state.ws = nil // the predicate might've set a restore point
	return p.File.Raw[start:p.Index()]
}

func Collect[T any](p *Parser, f Func[T], cap int, ws WhitespaceFunc) []T {
	ts := make([]T, 0, cap)
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
