package whitespace

import (
	"testing"

	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
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

func testEOL(t *testing.T, f parser.Func[struct{}]) {
	testCases := []string{"\n", "\r\n"}

	for _, tc := range testCases {
		t.Run(testName(tc), func(t *testing.T) {
			t.Parallel()
			testutil.AssertParsesFully(t, tc, f)
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
			testutil.AssertParsesFully(t, tc, f)
		})
	}
}

func TestEOS(t *testing.T) {
	t.Parallel()

	t.Run("}", func(t *testing.T) {
		t.Parallel()
		p := testutil.NewParser(t, " \t }")
		testutil.AssertNoError(t, p, EOS())
		assert.Equal(t, 0, p.Index(), "index mismatch")
	})

	t.Run("//", func(t *testing.T) {
		t.Parallel()
		p := testutil.NewParser(t, " \t\t //")
		testutil.AssertNoError(t, p, EOS())
		assert.Equal(t, 0, p.Index(), "index mismatch")
	})

	t.Run(";", func(t *testing.T) {
		t.Parallel()
		testutil.AssertParsesFully(t, "  \t \t\t;", EOS())
	})

	t.Run("EOL", func(t *testing.T) {
		t.Parallel()
		testEOL(t, EOS())
	})

	t.Run("EOF", func(t *testing.T) {
		t.Parallel()
		testEOF(t, EOS())
	})
}
