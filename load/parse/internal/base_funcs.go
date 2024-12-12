package parser

// TryToken attempts to match the given token verbatim.
func TryToken(p *Parser, s string) (ok bool) {
	restore := p.takeRestore()
	if !MatchesToken(p, s) {
		p.RestoreState(restore)
		return false
	}

	for range s {
		p.next()
	}
	return true
}

func TryAnyToken(p *Parser, ss ...string) string {
	restore := p.takeRestore()
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

func TryRune(p *Parser, r rune) (ok bool) {
	restore := p.takeRestore()
	if r != p.peek() {
		p.RestoreState(restore)
		return false
	}
	p.next()
	return true
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
	restore := p.takeRestore()
	if !pred(p.peek()) {
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
	restore := p.takeRestore()
	start := p.Index()
	if !pred() {
		p.RestoreState(restore)
		return ""
	}
	p.next()
	for pred() {
		r := p.next()
		if r == EOF {
			break
		}
	}
	return p.File.Raw[start:p.Index()]
}
