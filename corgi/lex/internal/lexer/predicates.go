package lexer

var (
	IsHorizontalWhitespace = Matches('\t', ' ')
	IsWhitespace           = Matches('\t', ' ', '\n')
	IsNewline              = Matches('\n')
)

// Matches returns a predicate that reports true whenever the rune the
// predicate is called with matches any of rs.
func Matches(rs ...rune) func(r rune) bool {
	return func(r rune) bool {
		for _, cmp := range rs {
			if r == cmp {
				return true
			}
		}

		return false
	}
}

// MatchesNot returns a predicate that reports false whenever the rune the
// predicate is called with matches any of rs.
func MatchesNot(rs ...rune) func(r rune) bool {
	return func(r rune) bool {
		for _, cmp := range rs {
			if r == cmp {
				return false
			}
		}

		return true
	}
}
