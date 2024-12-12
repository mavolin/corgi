package whitespace

import (
	"strings"
	"testing"

	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAny(t *testing.T) {
	t.Parallel()

	testCases := []string{" ", "\t", "\n", "\r\n", " \t\n\r\n", "  \t  \n  \r\n  \t  \n  \r\n"}

	for _, in := range testCases {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, in)
			assert.Nil(t, Any()(p))
			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
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
	testCases := []string{" ", "\t", "  ", "\t\t", " \t  \t "}

	for _, in := range testCases {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, in+trail)
			assert.Nil(t, f(p))
			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
		})
	}
}

func TestVertical(t *testing.T) {
	t.Parallel()
	testVertical(t, Vertical(), " other")
}

func testVertical(t *testing.T, f parser.WhitespaceFunc, trail string) {
	testCases := []string{"\n", "\r\n", "\n\n", "\r\n\n", "\n\r\n"}

	for _, in := range testCases {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, in+trail)
			assert.Nil(t, f(p))
			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
		})
	}
}

func TestSingleVertical(t *testing.T) {
	t.Parallel()

	testCases := []string{"\n", "\r\n"}

	for _, in := range testCases {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, in+"\n")
			assert.Nil(t, SingleVertical()(p))
			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
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
