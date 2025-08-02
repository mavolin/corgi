package whitespace

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
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
	tests := []string{"\n", "\r\n"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, in)
			should.True(t, f(p))
			line, col, index := parsetest.CalcEnd(1, 1, 0, in)
			parsetest.AssertPosition(t, p, line, col, index)
		})
	}
}

func TestEOF(t *testing.T) {
	t.Parallel()
	testEOF(t, EOF())
}

func testEOF(t *testing.T, f parser.WhitespaceFunc) {
	tests := []string{"", " ", "\t", "\t  ", "   \t  \t "}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, in)
			should.True(t, f(p))
			line, col, index := parsetest.CalcEnd(1, 1, 0, in)
			parsetest.AssertPosition(t, p, line, col, index)
		})
	}
}
