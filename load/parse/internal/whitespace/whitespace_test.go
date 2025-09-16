package whitespace

import (
	"strings"
	"testing"

	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestAny(t *testing.T) {
	t.Parallel()

	tests := []string{" ", "\t", "\n", "\r\n", " \t\n\r\n", "  \t  \n  \r\n  \t  \n  \r\n"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()
			parsetest.SkipsWhitespaceUntilIdentifier(t, in, Any())
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
			parsetest.SkipsWhitespaceUntil(t, in, trail, f)
		})
	}
}

func TestSingleHorizontal(t *testing.T) {
	t.Parallel()
	testSingleHorizontal(t, SingleHorizontal())
}

func testSingleHorizontal(t *testing.T, f parser.WhitespaceFunc) {
	tests := []string{" ", "\t"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()
			parsetest.SkipsWhitespaceUntil(t, in, " ", f)
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
			parsetest.SkipsWhitespaceUntil(t, in, trail, f)
		})
	}
}

func TestSingleVertical(t *testing.T) {
	t.Parallel()
	testSingleVertical(t, SingleVertical())
}

func testSingleVertical(t *testing.T, f parser.WhitespaceFunc) {
	tests := []string{"\n", "\r\n"}

	for _, in := range tests {
		t.Run(testName(in), func(t *testing.T) {
			t.Parallel()
			parsetest.SkipsWhitespaceUntil(t, in, "\n", f)
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
