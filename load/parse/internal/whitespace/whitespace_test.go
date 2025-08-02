package whitespace

import (
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestAny(t *testing.T) {
	t.Parallel()

	tests := []string{" ", "\t", "\n", "\r\n", " \t\n\r\n", "  \t  \n  \r\n  \t  \n  \r\n"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, in)
			should.True(t, Any()(p))
			line, col, index := parsetest.CalcEnd(1, 1, 0, in)
			parsetest.AssertPosition(t, p, line, col, index)
		})
	}

	t.Run("horizontal", func(t *testing.T) {
		t.Parallel()
		testHorizontal(t, Any(), "other")
	})

	t.Run("vertical", func(t *testing.T) {
		t.Parallel()
		testVertical(t, Any(), "other")
	})
}

func TestHorizontal(t *testing.T) {
	t.Parallel()
	testHorizontal(t, Horizontal(), "\nother")
}

func testHorizontal(t *testing.T, f parser.WhitespaceFunc, trail string) {
	tests := []string{" ", "\t", "  ", "\t\t", " \t  \t "}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, in+trail)
			should.True(t, f(p))
			line, col, index := parsetest.CalcEnd(1, 1, 0, in)
			parsetest.AssertPosition(t, p, line, col, index)
		})
	}
}

func TestVertical(t *testing.T) {
	t.Parallel()
	testVertical(t, Vertical(), " other")
}

func testVertical(t *testing.T, f parser.WhitespaceFunc, trail string) {
	tests := []string{"\n", "\r\n", "\n\n", "\r\n\n", "\n\r\n"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, in+trail)
			should.True(t, f(p))
			line, col, index := parsetest.CalcEnd(1, 1, 0, in)
			parsetest.AssertPosition(t, p, line, col, index)
		})
	}
}

func TestSingleVertical(t *testing.T) {
	t.Parallel()

	tests := []string{"\n", "\r\n"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, in+"\n")
			should.True(t, SingleVertical()(p))
			line, col, index := parsetest.CalcEnd(1, 1, 0, in)
			parsetest.AssertPosition(t, p, line, col, index)
		})
	}
}

func testName(input string) string {
	var name strings.Builder
	for _, r := range input {
		switch r {
		case ' ':
			name.WriteString("_")
		case '\t':
			name.WriteString("\\t")
		case '\n':
			name.WriteString("\\n")
		case '\r':
			name.WriteString("\\r")
		}
	}
	return name.String()
}
