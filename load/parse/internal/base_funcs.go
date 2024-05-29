package parser

// TryToken attempts to match the given token verbatim.
func TryToken(p *Parser, s string) (ok bool) {
	if MatchesString(p, s) {
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

func While(p *Parser, pred func() bool) string {
	start := p.Index()
	for pred() {
		r := p.next()
		if r == EOF {
			break
		}
	}
	return p.File.Raw[start:p.Index()]
}
