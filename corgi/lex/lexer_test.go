package lex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Exported
// ======================================================================================

func TestLexer_Next(t *testing.T) {
	t.Parallel()

	in := "abc"
	expect := Item{
		Type: Code,
		Val:  "abc",
		Line: 1,
		Col:  1,
	}

	l := New(in)

	for i := 0; i < len(in); i++ {
		l.next()
	}

	go l.emit(Code)

	actual := l.Next()
	assert.Equal(t, expect, actual)
}

// ============================================================================
// Helpers
// ======================================================================================

func TestLexer_next(t *testing.T) {
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

	l := New(in)

	assert.Equal(t, 1, l.line)
	assert.Equal(t, 1, l.col)
	assert.Equal(t, 0, l.prevLineLen)
	assert.Equal(t, 0, l.pos)

	for _, e := range expect {
		actualRune := l.next()
		assert.Equalf(t, e.Rune, actualRune, "expected rune %q, but got %q", e.Rune, actualRune)
		assert.Equalf(t, e.Line, l.line, "expected line %d, but got %d", e.Line, l.line)
		assert.Equalf(t, e.Col, l.col, "expected col %d, but got %d", e.Col, l.col)
		assert.Equalf(t, e.PrevLineLen, l.prevLineLen,
			"expected prevLineLen %d, but got %d", e.PrevLineLen, l.prevLineLen)
		assert.Equalf(t, e.Pos, l.pos, "expected pos %d, but got %d", e.Pos, l.pos)
		assert.Equalf(t, e.PrevWidth, l.prevWidth, "expected prevWidth %d, but got %d", e.PrevWidth, l.prevWidth)
	}
}

func TestLexer_nextUntil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        string
		Args      []rune
		Expect    rune
		ExpectPos int
	}{
		{
			Name:      "match rune",
			In:        "abc\ndef",
			Args:      []rune{'\n', 'd'},
			Expect:    '\n',
			ExpectPos: 3,
		},
		{
			Name:      "eof",
			In:        "abcdef",
			Args:      []rune{'g'},
			Expect:    eof,
			ExpectPos: 6,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New(c.In)

			actual := l.nextUntil(c.Args...)
			assert.Equal(t, c.Expect, actual)
			assert.Equal(t, c.ExpectPos, l.pos)
		})
	}
}

func TestLexer_nextString(t *testing.T) {
	t.Parallel()

	in := "abcdef"

	l := New(in)
	l.nextString("abc")

	assert.Equalf(t, 3, l.pos, "expected pos 3, but got %d", l.pos)
	assert.Equalf(t, 4, l.col, "expected col 4, but got %d", l.col)
}

func TestLexer_nextRunes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		In        string
		Runes     []rune
		ExpectPos int
	}{
		{In: "ab", Runes: []rune{'b'}, ExpectPos: 0},
		{In: "ab", Runes: []rune{'a'}, ExpectPos: 1},
		{In: "aab", Runes: []rune{'a'}, ExpectPos: 2},
		{In: "aabc", Runes: []rune{'a', 'b'}, ExpectPos: 3},
		{In: "", Runes: []rune{'a'}, ExpectPos: 0},
		{In: "aaa", Runes: []rune{'a'}, ExpectPos: 3},
	}

	for _, c := range testCases {
		c := c

		t.Run("", func(t *testing.T) {
			t.Parallel()

			l := New(c.In)
			l.nextRunes(c.Runes...)

			assert.Equalf(t, c.ExpectPos, l.pos, "expected pos %d, but got %d", c.ExpectPos, l.pos)
		})
	}
}

// ======================================= backup =======================================

func TestLexer_backup(t *testing.T) {
	in := "a🌵b"

	l := New(in)
	l.next()
	l.next()

	assert.Equalf(t, 5, l.pos, "expected pos to be 5 before backup, but got %d", l.pos)
	assert.Equalf(t, 3, l.col, "expected col to be 2 before backup, but got %d", l.col)

	l.backup()

	assert.Equalf(t, 1, l.pos, "expected pos to be 1 after backup, but got %d", l.pos)
	assert.Equalf(t, 2, l.col, "expected col to be 1 after backup, but got %d", l.col)
}

// ======================================== peek ========================================

func TestLexer_peek(t *testing.T) {
	t.Parallel()

	in := "abc"

	l := New(in)

	var (
		expectPos  = l.pos
		expectLine = l.line
		expectCol  = l.col

		expectPeek = 'a'
		actualPeek = l.peek()
	)

	assert.Equalf(t, expectPeek, actualPeek, "expected peek %q, but got %q", expectPeek, actualPeek)
	assert.Equalf(t, expectPos, l.pos, "expected pos %d, but got %d", expectPos, l.pos)
	assert.Equalf(t, expectLine, l.line, "expected line %d, but got %d", expectLine, l.line)
	assert.Equalf(t, expectCol, l.col, "expected col %d, but got %d", expectCol, l.col)

	l.next()

	expectPos = l.pos
	expectLine = l.line
	expectCol = l.col

	expectPeek = 'b'
	actualPeek = l.peek()

	assert.Equalf(t, expectPeek, actualPeek, "expected peek %q, but got %q", expectPeek, actualPeek)
	assert.Equalf(t, expectPos, l.pos, "expected pos %d, but got %d", expectPos, l.pos)
	assert.Equalf(t, expectLine, l.line, "expected line %d, but got %d", expectLine, l.line)
	assert.Equalf(t, expectCol, l.col, "expected col %d, but got %d", expectCol, l.col)
}

func TestLexer_peekIsString(t *testing.T) {
	t.Parallel()

	in := "abc"

	l := New(in)

	assert.True(t, l.peekIsString("a"))
	assert.True(t, l.peekIsString("ab"))
	assert.True(t, l.peekIsString("abc"))

	assert.False(t, l.peekIsString("b"))
	assert.False(t, l.peekIsString("bc"))
	assert.False(t, l.peekIsString("abcd"))

	l.next()

	assert.True(t, l.peekIsString("b"))
	assert.True(t, l.peekIsString("bc"))

	assert.False(t, l.peekIsString("a"))
	assert.False(t, l.peekIsString("c"))
	assert.False(t, l.peekIsString("cd"))
	assert.False(t, l.peekIsString("abc"))
	assert.False(t, l.peekIsString("bcd"))
}

// ======================================= ignore =======================================

func TestLexer_ignore(t *testing.T) {
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

			l := New(c.In)

			for i := 0; i < c.NumNextCalls; i++ {
				l.next()
			}
			l.ignore()

			assert.Equalf(t, c.ExpectPos, l.startPos,
				"expected startPos to be %d, but got %d", c.ExpectPos, l.startPos)
			assert.Equalf(t, c.ExpectLine, l.startLine,
				"expected startLine to be %d, but got %d", c.ExpectLine, l.startLine)
			assert.Equalf(t, c.ExpectCol, l.startCol,
				"expected startCol to be %d, but got %d", c.ExpectCol, l.startCol)
		})
	}
}

func TestLexer_ignoreRunes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        string
		Runes     []rune
		ExpectPos int
	}{
		{
			Name:      "no occurrence",
			In:        "abbccc",
			Runes:     []rune{'b'},
			ExpectPos: 0,
		},
		{
			Name:      "single rune",
			In:        "abbccc",
			Runes:     []rune{'a'},
			ExpectPos: 1,
		},
		{
			Name:      "multiple runes",
			In:        "abbccc",
			Runes:     []rune{'a', 'b'},
			ExpectPos: 3,
		},
		{
			Name:      "eof",
			In:        "abbccc",
			Runes:     []rune{'a', 'b', 'c'},
			ExpectPos: 6,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New(c.In)

			l.ignoreRunes(c.Runes...)

			assert.Equalf(t, c.ExpectPos, l.pos, "expected pos to be %d, but got %d", c.ExpectPos, l.pos)
			assert.Equalf(t, c.ExpectPos, l.startPos,
				"expected start pos to be %d, but got %d", c.ExpectPos, l.startPos)
		})
	}
}

func TestLexer_ignoreWhitespace(t *testing.T) {
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

			l := New(c.In)

			l.ignoreWhitespace()

			assert.Equalf(t, c.ExpectPos, l.pos, "expected pos to be %d, but got %d", c.ExpectPos, l.pos)
			assert.Equalf(t, c.ExpectPos, l.startPos,
				"expected start pos to be %d, but got %d", c.ExpectPos, l.startPos)
		})
	}
}

func TestLexer_ignoreUntil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        string
		Runes     []rune
		ExpectPos int
	}{
		{
			Name:      "immediate stop",
			In:        "abc",
			Runes:     []rune{'a'},
			ExpectPos: 0,
		},
		{
			Name:      "skip some",
			In:        "ababbc",
			Runes:     []rune{'c'},
			ExpectPos: 5,
		},
		{
			Name:      "eof",
			In:        "abcbac",
			Runes:     []rune{'d'},
			ExpectPos: 6,
		},
	}

	for _, c := range testCases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			l := New(c.In)

			l.ignoreUntil(c.Runes...)

			assert.Equalf(t, c.ExpectPos, l.pos, "expected pos to be %d, but got %d", c.ExpectPos, l.pos)
			assert.Equalf(t, c.ExpectPos, l.startPos, "expected start pos to be %d, but got %d", c.ExpectPos,
				l.startPos)
		})
	}
}

// ======================================= Indent =======================================

func TestLexer_consumeIndent(t *testing.T) {
	t.Parallel()

	successCases := []struct {
		Name string
		In   string

		Indent          rune
		ExpectIndent    rune
		IndentLen       int
		ExpectIndentLen int
		IndentLvl       int

		ExpectNoIncrease int
		IgnoreNoIncrease bool

		ExpectSingleIncrease int
		IgnoreSingleIncrease bool

		ExpectAllIndents int
		IgnoreAllIndents bool
	}{
		{
			Name: "no indent",
			In:   "abc",

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:      "no indent",
			In:        "abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:      "no indent",
			In:        "  abc",
			Indent:    ' ',
			IndentLen: 2,

			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:      "single increase",
			In:        "  abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name:      "single increase",
			In:        "    abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name:      "multi increase",
			In:        "    abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 1,
			IgnoreAllIndents:     true,
		},
		{
			Name:      "multi increase",
			In:        "    abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 1,
			IgnoreAllIndents:     true,
		},
		{
			Name:      "single decrease",
			In:        "    abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 3,

			ExpectNoIncrease:     -1,
			ExpectSingleIncrease: -1,
			ExpectAllIndents:     -1,
		},
		{
			Name:      "multi decrease",
			In:        "  abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 3,

			ExpectNoIncrease:     -2,
			ExpectSingleIncrease: -2,
			ExpectAllIndents:     -2,
		},
		{
			Name:      "multi decrease",
			In:        "abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 3,

			ExpectNoIncrease:     -3,
			ExpectSingleIncrease: -3,
			ExpectAllIndents:     -3,
		},
		{
			Name:            "spaces",
			In:              "    abc",
			ExpectIndent:    ' ',
			ExpectIndentLen: 4,

			IgnoreNoIncrease:     true,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name: "spaces",
			In:   "    abc",

			ExpectSingleIncrease: 0,
			IgnoreSingleIncrease: true,
			IgnoreAllIndents:     true,
		},
		{
			Name:            "tabs",
			In:              "\tabc",
			ExpectIndent:    '\t',
			ExpectIndentLen: 1,

			IgnoreNoIncrease:     true,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name: "tabs",
			In:   "\tabc",

			ExpectSingleIncrease: 0,
			IgnoreSingleIncrease: true,
			IgnoreAllIndents:     true,
		},
		{
			Name:      "empty line",
			In:        "\n  abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name:      "empty line",
			In:        "  \nabc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     -1,
			ExpectSingleIncrease: -1,
			ExpectAllIndents:     -1,
		},
		{
			Name:      "empty line",
			In:        "      \n  abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name:      "empty line",
			In:        "      \n  abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:            "empty line",
			In:              "      \n  abc",
			ExpectIndent:    ' ',
			ExpectIndentLen: 2,

			IgnoreNoIncrease:     true,
			ExpectSingleIncrease: 1,
			ExpectAllIndents:     1,
		},
		{
			Name:      "eof",
			In:        "",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 3,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:      "eof",
			In:        "  ",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:      "eof",
			In:        " ",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name:      "eof",
			In:        "    ",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
		{
			Name: "eof",
			In:   "  ",

			ExpectNoIncrease:     0,
			ExpectSingleIncrease: 0,
			ExpectAllIndents:     0,
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			c := c

			if !c.IgnoreNoIncrease {
				t.Run("no increase/"+c.Name,
					consumeIndentSuccessTest(noIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl,
						c.ExpectNoIncrease, c.ExpectIndent, c.ExpectIndentLen))
			}

			if !c.IgnoreSingleIncrease {
				t.Run("single increase/"+c.Name,
					consumeIndentSuccessTest(singleIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl,
						c.ExpectSingleIncrease, c.ExpectIndent, c.ExpectIndentLen))
			}

			if !c.IgnoreAllIndents {
				t.Run("allIndents/"+c.Name,
					consumeIndentSuccessTest(allIndents, c.In, c.Indent, c.IndentLen, c.IndentLvl,
						c.ExpectAllIndents, c.ExpectIndent, c.ExpectIndentLen))
			}
		}
	})

	failureCases := []struct {
		Name      string
		In        string
		Indent    rune
		IndentLen int
		IndentLvl int

		ExpectNoIncrease     error
		ExpectSingleIncrease error
		ExpectAllIndents     error
	}{
		{
			Name: "ErrMixedIndentation",
			In:   "\t  abc",

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: ErrMixedIndentation,
			ExpectAllIndents:     ErrMixedIndentation,
		},
		{
			Name:      "ErrMixedIndentation",
			In:        "\t  abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: ErrMixedIndentation,
			ExpectAllIndents:     ErrMixedIndentation,
		},
		{
			Name:      "ErrMixedIndentation",
			In:        "\t  abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     ErrMixedIndentation,
			ExpectSingleIncrease: ErrMixedIndentation,
			ExpectAllIndents:     ErrMixedIndentation,
		},
		{
			Name:      "ErrIndentationIncrease",
			In:        "    abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: nil,
			ExpectAllIndents:     ErrIndentationIncrease,
		},
		{
			Name:      "ErrIndentationIncrease",
			In:        "      abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: nil,
			ExpectAllIndents:     ErrIndentationIncrease,
		},
		{
			Name:      "inconsistent indentation",
			In:        "   abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: nil,
			ExpectAllIndents: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
		},
		{
			Name:      "inconsistent indentation",
			In:        "   abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease: nil,
			ExpectSingleIncrease: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
			ExpectAllIndents: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
		},
		{
			Name:      "inconsistent indentation",
			In:        "   abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 2,

			ExpectNoIncrease: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
			ExpectSingleIncrease: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
			ExpectAllIndents: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
		},
		{
			Name:      "wrong indent rune: tab",
			In:        "\tabc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease: nil,
			ExpectSingleIncrease: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
			ExpectAllIndents: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
		},
		{
			Name:      "wrong indent rune: tab",
			In:        "\tabc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
			ExpectSingleIncrease: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
			ExpectAllIndents: &IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
		},
	}

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		for _, c := range failureCases {
			c := c

			t.Run("no increase/"+c.Name,
				consumeIndentFailureTest(noIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl, c.ExpectNoIncrease))

			t.Run("single increase/"+c.Name,
				consumeIndentFailureTest(
					singleIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl, c.ExpectSingleIncrease))

			t.Run("allIndents/"+c.Name,
				consumeIndentFailureTest(allIndents, c.In, c.Indent, c.IndentLen, c.IndentLvl, c.ExpectAllIndents))
		}
	})
}

func consumeIndentSuccessTest(
	mode indentConsumptionMode,
	in string, indent rune, indentLen int, indentLvl int, expectDelta int,
	expectIndent rune, expectIndentLen int,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		l := New(in)
		l.indent = indent
		l.indentLen = indentLen
		l.indentLvl = indentLvl

		actual, _, err := l.consumeIndent(mode)
		require.NoError(t, err)

		assert.Equalf(t, expectDelta, actual,
			"expected indent delta to be %d, but got %d", expectDelta, actual)
		assert.Equalf(t, indentLvl+expectDelta, l.indentLvl,
			"expected indent level to be %d, but got %d", indentLvl+expectDelta, l.indentLvl)

		if expectIndent != 0 {
			assert.Equalf(t, expectIndent, l.indent,
				"expected indent to be %q, but got %q", expectIndent, l.indent)
			assert.Equalf(t, expectIndentLen, l.indentLen,
				"expected indent len to be %d, but got %d", expectIndentLen, l.indentLen)
		}
	}
}

func consumeIndentFailureTest(
	mode indentConsumptionMode,
	in string, indent rune, indentLen int, indentLvl int, expect error,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		l := New(in)
		l.indent = indent
		l.indentLen = indentLen
		l.indentLvl = indentLvl

		_, _, actual := l.consumeIndent(mode)
		assert.Equal(t, expect, actual)
	}
}
