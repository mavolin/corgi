package whitespace

import (
	"testing"

	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEOL(t *testing.T) {
	t.Parallel()

	testEOL(t, EOL())

	t.Run("EOF", func(t *testing.T) {
		t.Parallel()
		testEOF(t, EOL())
	})
}

func testEOL(t *testing.T, f parser.WhitespaceFunc) {
	testCases := []string{"\n", "\r\n"}

	for _, in := range testCases {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, in)
			assert.Nil(t, f(p))
			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
		})
	}
}

func TestEOF(t *testing.T) {
	t.Parallel()
	testEOF(t, EOF())
}

func testEOF(t *testing.T, f parser.WhitespaceFunc) {
	testCases := []string{"", " ", "\t", "\t  ", "   \t  \t "}

	for _, in := range testCases {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, in)
			assert.Nil(t, f(p))
			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
		})
	}
}
