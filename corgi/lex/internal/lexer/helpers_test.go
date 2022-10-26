package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// Next
// ======================================================================================

func TestLexer_NextRune(t *testing.T) {
	t.Parallel()

	in := "ab\n🌵."

	expect := []struct {
		Rune        rune
		Line, Col   int
		PrevLineLen int
		Pos         int
		PrevWidth   int
	}{
		{Rune: 'a', Line: 1, Col: 2, Pos: 1, PrevWidth: 1},
		{Rune: 'b', Line: 1, Col: 3, Pos: 2, PrevWidth: 1},
		{Rune: '\n', Line: 2, Col: 1, PrevLineLen: 3, Pos: 3, PrevWidth: 1},
		{Rune: '🌵', Line: 2, Col: 2, PrevLineLen: 3, Pos: 7, PrevWidth: 4},
		{Rune: '.', Line: 2, Col: 3, PrevLineLen: 3, Pos: 8, PrevWidth: 1},
	}

	l := New[token.Token](in, nil)

	assert.Equal(t, 1, l.line)
	assert.Equal(t, 1, l.col)
	assert.Equal(t, 0, l.prevLineLen)
	assert.Equal(t, 0, l.pos)

	for _, e := range expect {
		actualRune := l.Next()
		assert.Equalf(t, e.Rune, actualRune, "expected rune %q, but got %q", e.Rune, actualRune)
		assert.Equalf(t, e.Line, l.line, "expected line %d, but got %d", e.Line, l.line)
		assert.Equalf(t, e.Col, l.col, "expected col %d, but got %d", e.Col, l.col)
		assert.Equalf(t, e.PrevLineLen, l.prevLineLen,
			"expected prevLineLen %d, but got %d", e.PrevLineLen, l.prevLineLen)
		assert.Equalf(t, e.Pos, l.pos, "expected pos %d, but got %d", e.Pos, l.pos)
		assert.Equalf(t, e.PrevWidth, l.prevWidth, "expected prevWidth %d, but got %d", e.PrevWidth, l.prevWidth)
	}
}

func TestLexer_NextRuneWhile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        string
		Predicate func(rune) bool
		Expect    rune
		ExpectPos int
	}{
		{
			Name:      "Matches",
			In:        "ab",
			Predicate: Matches('b'),
			Expect:    'a',
			ExpectPos: 0,
		},
		{
			Name:      "Matches",
			In:        "ab",
			Predicate: Matches('a'),
			Expect:    'b',
			ExpectPos: 1,
		},
		{
			Name:      "Matches",
			In:        "aab",
			Predicate: Matches('a'),
			Expect:    'b',
			ExpectPos: 2,
		},
		{
			Name:      "Matches",
			In:        "aabc",
			Predicate: Matches('a', 'b'),
			Expect:    'c',
			ExpectPos: 3,
		},
		{
			Name:      "Matches",
			In:        "",
			Predicate: Matches('a'),
			Expect:    EOF,
			ExpectPos: 0,
		},
		{
			Name:      "Matches",
			In:        "aaa",
			Predicate: Matches('a'),
			Expect:    EOF,
			ExpectPos: 3,
		},
		{
			Name:      "IsNot/matches rune",
			In:        "abc\ndef",
			Predicate: IsNot('\n', 'd'),
			Expect:    '\n',
			ExpectPos: 3,
		},
		{
			Name:      "IsNot/encounters eof",
			In:        "abcdef",
			Predicate: IsNot('g'),
			Expect:    EOF,
			ExpectPos: 6,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New[token.Token](c.In, nil)

			actual := l.NextWhile(c.Predicate)
			assert.Equal(t, c.Expect, actual)
			assert.Equal(t, c.ExpectPos, l.pos)
		})
	}
}

func TestLexer_SkipString(t *testing.T) {
	t.Parallel()

	in := "abcdef"

	l := New[token.Token](in, nil)
	l.SkipString("abc")

	assert.Equalf(t, 3, l.pos, "expected pos 3, but got %d", l.pos)
	assert.Equalf(t, 4, l.col, "expected col 4, but got %d", l.col)
}

func TestLexer_Backup(t *testing.T) {
	in := "a🌵b"

	l := New[token.Token](in, nil)
	l.Next()
	l.Next()

	assert.Equalf(t, 5, l.pos, "expected pos to be 5 before backup, but got %d", l.pos)
	assert.Equalf(t, 3, l.col, "expected col to be 2 before backup, but got %d", l.col)

	l.Backup()

	assert.Equalf(t, 1, l.pos, "expected pos to be 1 after backup, but got %d", l.pos)
	assert.Equalf(t, 2, l.col, "expected col to be 1 after backup, but got %d", l.col)
}

// ============================================================================
// Peek
// ======================================================================================

func TestLexer_PeekRune(t *testing.T) {
	t.Parallel()

	in := "abc"

	l := New[token.Token](in, nil)

	var (
		expectPos  = l.pos
		expectLine = l.line
		expectCol  = l.col

		expectPeek = 'a'
		actualPeek = l.Peek()
	)

	assert.Equalf(t, expectPeek, actualPeek, "expected peek %q, but got %q", expectPeek, actualPeek)
	assert.Equalf(t, expectPos, l.pos, "expected pos %d, but got %d", expectPos, l.pos)
	assert.Equalf(t, expectLine, l.line, "expected line %d, but got %d", expectLine, l.line)
	assert.Equalf(t, expectCol, l.col, "expected col %d, but got %d", expectCol, l.col)

	l.Next()

	expectPos = l.pos
	expectLine = l.line
	expectCol = l.col

	expectPeek = 'b'
	actualPeek = l.Peek()

	assert.Equalf(t, expectPeek, actualPeek, "expected peek %q, but got %q", expectPeek, actualPeek)
	assert.Equalf(t, expectPos, l.pos, "expected pos %d, but got %d", expectPos, l.pos)
	assert.Equalf(t, expectLine, l.line, "expected line %d, but got %d", expectLine, l.line)
	assert.Equalf(t, expectCol, l.col, "expected col %d, but got %d", expectCol, l.col)
}

func TestLexer_PeekIsString(t *testing.T) {
	t.Parallel()

	in := "abc"

	l := New[token.Token](in, nil)

	assert.True(t, l.PeekIsString("a"))
	assert.True(t, l.PeekIsString("ab"))
	assert.True(t, l.PeekIsString("abc"))

	assert.False(t, l.PeekIsString("b"))
	assert.False(t, l.PeekIsString("bc"))
	assert.False(t, l.PeekIsString("abcd"))

	l.Next()

	assert.True(t, l.PeekIsString("b"))
	assert.True(t, l.PeekIsString("bc"))

	assert.False(t, l.PeekIsString("a"))
	assert.False(t, l.PeekIsString("c"))
	assert.False(t, l.PeekIsString("cd"))
	assert.False(t, l.PeekIsString("abc"))
	assert.False(t, l.PeekIsString("bcd"))
}

// ============================================================================
// Ignore
// ======================================================================================

func TestLexer_Ignore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name         string
		In           string
		NumNextCalls int
		ExpectPos    int
		ExpectLine   int
		ExpectCol    int
	}{
		{
			Name:         "noop",
			In:           "ab",
			NumNextCalls: 0,
			ExpectPos:    0,
			ExpectLine:   1,
			ExpectCol:    1,
		},
		{
			Name:         "ignore one",
			In:           "ab",
			NumNextCalls: 1,
			ExpectPos:    1,
			ExpectLine:   1,
			ExpectCol:    2,
		},
		{
			Name:         "ignore two",
			In:           "a\nb",
			NumNextCalls: 2,
			ExpectPos:    2,
			ExpectLine:   2,
			ExpectCol:    1,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New[token.Token](c.In, nil)

			for i := 0; i < c.NumNextCalls; i++ {
				l.Next()
			}
			l.Ignore()

			assert.Equalf(t, c.ExpectPos, l.startPos,
				"expected startPos to be %d, but got %d", c.ExpectPos, l.startPos)
			assert.Equalf(t, c.ExpectLine, l.startLine,
				"expected startLine to be %d, but got %d", c.ExpectLine, l.startLine)
			assert.Equalf(t, c.ExpectCol, l.startCol,
				"expected startCol to be %d, but got %d", c.ExpectCol, l.startCol)
		})
	}
}

func TestLexer_IgnoreRunesWhile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        string
		Predicate func(rune) bool
		ExpectPos int
	}{
		{
			Name:      "Matches/no occurrence",
			In:        "abbccc",
			Predicate: Matches('b'),
			ExpectPos: 0,
		},
		{
			Name:      "Matches/single rune",
			In:        "abbccc",
			Predicate: Matches('a'),
			ExpectPos: 1,
		},
		{
			Name:      "Matches/multiple runes",
			In:        "abbccc",
			Predicate: Matches('a', 'b'),
			ExpectPos: 3,
		},
		{
			Name:      "Matches/eof",
			In:        "abbccc",
			Predicate: Matches('a', 'b', 'c'),
			ExpectPos: 6,
		},
		{
			Name:      "IsNot/immediate stop",
			In:        "abc",
			Predicate: IsNot('a'),
			ExpectPos: 0,
		},
		{
			Name:      "IsNot/skip some",
			In:        "ababbc",
			Predicate: IsNot('c'),
			ExpectPos: 5,
		},
		{
			Name:      "IsNot/eof",
			In:        "abcbac",
			Predicate: IsNot('d'),
			ExpectPos: 6,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New[token.Token](c.In, nil)

			l.IgnoreWhile(c.Predicate)

			assert.Equalf(t, c.ExpectPos, l.pos, "expected pos to be %d, but got %d", c.ExpectPos, l.pos)
			assert.Equalf(t, c.ExpectPos, l.startPos,
				"expected start pos to be %d, but got %d", c.ExpectPos, l.startPos)
		})
	}
}

func TestLexer_IgnoreWhitespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        string
		ExpectPos int
	}{
		{
			Name:      "no occurrence",
			In:        "a",
			ExpectPos: 0,
		},
		{
			Name:      "single space",
			In:        " a",
			ExpectPos: 1,
		},
		{
			Name:      "single tab",
			In:        "\ta",
			ExpectPos: 1,
		},
		{
			Name:      "multiple spaces",
			In:        "   a",
			ExpectPos: 3,
		},
		{
			Name:      "multiple spaces and tabs",
			In:        " \t \t a",
			ExpectPos: 5,
		},
		{
			Name:      "eof",
			In:        "  ",
			ExpectPos: 2,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New[token.Token](c.In, nil)

			l.IgnoreWhile(IsHorizontalWhitespace)

			assert.Equalf(t, c.ExpectPos, l.pos, "expected pos to be %d, but got %d", c.ExpectPos, l.pos)
			assert.Equalf(t, c.ExpectPos, l.startPos,
				"expected start pos to be %d, but got %d", c.ExpectPos, l.startPos)
		})
	}
}
