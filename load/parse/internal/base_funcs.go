package parser

// TryToken attempts to match the given token verbatim.
func TryToken(p *Parser, s string) (ok bool) {
	if MatchesToken(p, s) {
		for range s {
			p.next()
		}
		return true
	}
	return false
}

func TryAnyTokens(p *Parser, ss ...string) (ok bool) {
	for _, s := range ss {
		if TryToken(p, s) {
			return true
		}
	}
	return false
}

func TryRune(p *Parser, r rune) (ok bool) {
	if r == p.peek() {
		p.next()
		return true
	}
	return false
}

// TryAnyRune attempts to match the next rune against any of the passed runes.
func TryAnyRune(p *Parser, rs ...rune) (ok bool) {
	peek := p.peek()
	for _, r := range rs {
		if peek == r {
			p.next()
			return true
		}
	}
	return false
}

func TryRunePredicate(p *Parser, pred func(rune) bool) (r rune, ok bool) {
	if pred(p.peek()) {
		return p.next(), true
	}
	return 0, false
}

// TokenWhile consumes runes as long as the predicate returns true.
// The predicate may invoke the parser inside the predicate, but must not
// consume any runes itself.
func TokenWhile(p *Parser, pred func() bool) string {
	start := p.Index()
	for pred() {
		r := p.next()
		if r == EOF {
			break
		}
	}
	return p.File.Raw[start:p.Index()]
}

func TryAtLeastOne[T any](p *Parser, f Func[T]) (_ []T, ok bool) {
	first, ok := Try(p, f)
	if !ok {
		return nil, false
	}

	res := make([]T, 1, 48)
	res[0] = first

	for {
		next, ok := Try(p, f)
		if !ok {
			return res, true
		}
		res = append(res, next)
	}
}
