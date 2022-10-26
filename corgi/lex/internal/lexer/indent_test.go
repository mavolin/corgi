package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

func TestLexer_ConsumeIndent(t *testing.T) {
	// t.Parallel()

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
		// t.Parallel()

		for _, c := range successCases {
			c := c

			if !c.IgnoreNoIncrease {
				t.Run("no increase/"+c.Name,
					consumeIndentSuccessTest(ConsumeNoIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl,
						c.ExpectNoIncrease, c.ExpectIndent, c.ExpectIndentLen))
			}

			if !c.IgnoreSingleIncrease {
				t.Run("single increase/"+c.Name,
					consumeIndentSuccessTest(ConsumeSingleIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl,
						c.ExpectSingleIncrease, c.ExpectIndent, c.ExpectIndentLen))
			}

			if !c.IgnoreAllIndents {
				t.Run("allIndents/"+c.Name,
					consumeIndentSuccessTest(ConsumeAllIndents, c.In, c.Indent, c.IndentLen, c.IndentLvl,
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
			ExpectSingleIncrease: lexerr.ErrMixedIndentation,
			ExpectAllIndents:     lexerr.ErrMixedIndentation,
		},
		{
			Name:      "ErrMixedIndentation",
			In:        "\t  abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: lexerr.ErrMixedIndentation,
			ExpectAllIndents:     lexerr.ErrMixedIndentation,
		},
		{
			Name:      "ErrMixedIndentation",
			In:        "\t  abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     lexerr.ErrMixedIndentation,
			ExpectSingleIncrease: lexerr.ErrMixedIndentation,
			ExpectAllIndents:     lexerr.ErrMixedIndentation,
		},
		{
			Name:      "ErrIndentationIncrease",
			In:        "    abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: nil,
			ExpectAllIndents:     lexerr.ErrIndentationIncrease,
		},
		{
			Name:      "ErrIndentationIncrease",
			In:        "      abc",
			Indent:    ' ',
			IndentLen: 2,
			IndentLvl: 1,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: nil,
			ExpectAllIndents:     lexerr.ErrIndentationIncrease,
		},
		{
			Name:      "inconsistent indentation",
			In:        "   abc",
			Indent:    ' ',
			IndentLen: 2,

			ExpectNoIncrease:     nil,
			ExpectSingleIncrease: nil,
			ExpectAllIndents: &lexerr.IndentationError{
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
			ExpectSingleIncrease: &lexerr.IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
			ExpectAllIndents: &lexerr.IndentationError{
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

			ExpectNoIncrease: &lexerr.IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
			ExpectSingleIncrease: &lexerr.IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    ' ',
				ActualLen: 3,
			},
			ExpectAllIndents: &lexerr.IndentationError{
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
			ExpectSingleIncrease: &lexerr.IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
			ExpectAllIndents: &lexerr.IndentationError{
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

			ExpectNoIncrease: &lexerr.IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
			ExpectSingleIncrease: &lexerr.IndentationError{
				Expect:    ' ',
				ExpectLen: 2,
				Actual:    '\t',
				ActualLen: 1,
			},
			ExpectAllIndents: &lexerr.IndentationError{
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
				consumeIndentFailureTest(ConsumeNoIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl,
					c.ExpectNoIncrease))

			t.Run("single increase/"+c.Name,
				consumeIndentFailureTest(
					ConsumeSingleIncrease, c.In, c.Indent, c.IndentLen, c.IndentLvl, c.ExpectSingleIncrease))

			t.Run("allIndents/"+c.Name,
				consumeIndentFailureTest(ConsumeAllIndents, c.In, c.Indent, c.IndentLen, c.IndentLvl,
					c.ExpectAllIndents))
		}
	})
}

func consumeIndentSuccessTest(
	mode IndentConsumptionMode,
	in string, indent rune, indentLen int, indentLvl int, expectDelta int,
	expectIndent rune, expectIndentLen int,
) func(t *testing.T) {
	return func(t *testing.T) {
		// t.Parallel()

		l := New[token.Token](in, nil)
		l.indent = indent
		l.indentLen = indentLen
		l.indentLvl = indentLvl

		actual, _, err := l.ConsumeIndent(mode)
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
	mode IndentConsumptionMode,
	in string, indent rune, indentLen int, indentLvl int, expect error,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		l := New[token.Token](in, nil)
		l.indent = indent
		l.indentLen = indentLen
		l.indentLvl = indentLvl

		_, _, actual := l.ConsumeIndent(mode)
		assert.Equal(t, expect, actual)
	}
}
