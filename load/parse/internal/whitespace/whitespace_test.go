package whitespace

import (
	"strings"
	"testing"

	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
)

func TestAny(t *testing.T) {
	t.Parallel()

	testCases := []string{" ", "\t", "\n", "\r\n", " \t\n\r\n", "  \t  \n  \r\n  \t  \n  \r\n"}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, tc, Any())
		})
	}

	t.Run("horizontal", func(t *testing.T) {
		t.Parallel()
		testHorizontal(t, Any())
	})

	t.Run("vertical", func(t *testing.T) {
		t.Parallel()
		testVertical(t, Any())
	})
}

func TestHorizontal(t *testing.T) {
	t.Parallel()
	testHorizontal(t, Horizontal())
}

func testHorizontal(t *testing.T, f parser.Func[struct{}]) {
	testCases := []string{" ", "\t", "  ", "\t\t", " \t  \t "}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, tc, f)
		})
	}
}

func TestVertical(t *testing.T) {
	t.Parallel()
	testVertical(t, Vertical())
}

func testVertical(t *testing.T, f parser.Func[struct{}]) {
	testCases := []string{"\n", "\r\n", "\n\n", "\r\n\n", "\n\r\n"}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, tc, f)
		})
	}
}

func TestSingleVertical(t *testing.T) {
	t.Parallel()

	testCases := []string{"\n", "\r\n"}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, tc, SingleVertical())
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
