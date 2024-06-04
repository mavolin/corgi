package whitespace

import (
	"testing"

	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
)

func TestEOL(t *testing.T) {
	t.Parallel()

	testEOL(t, EOL())

	t.Run("EOF", func(t *testing.T) {
		t.Parallel()
		testEOF(t, EOL())
	})
}

func testEOL(t *testing.T, f parser.Func[struct{}]) {
	testCases := []string{"\n", "\r\n"}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, tc, f)
		})
	}
}

func TestEOF(t *testing.T) {
	t.Parallel()
	testEOF(t, EOF())
}

func testEOF(t *testing.T, f parser.Func[struct{}]) {
	testCases := []string{"", " ", "\t", "\t  ", "   \t  \t "}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, tc, f)
		})
	}
}
