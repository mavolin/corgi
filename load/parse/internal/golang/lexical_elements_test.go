package golang

import (
	"testing"

	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

// ============================================================================
// Rune literals
// ======================================================================================

func TestRuneLit(t *testing.T) {
	t.Parallel()

	tests := []string{
		`'a'`, `'ä'`, `'🐶'`, `'\n'`, `'\''`, `'"'`, `'\u1234'`, `'\U12345678'`,
		`'\x12'`, `'\123'`,
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesExact(t, c, RuneLit())
		})
	}
}

func TestUnicodeValue(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, UnicodeValue('\''), testLittleUValue)
	parsetest.AlsoFulfils(t, UnicodeValue('\''), testBigUValue)
	parsetest.AlsoFulfils(t, UnicodeValue('\''), testEscapedChar('\''))
	parsetest.AlsoFulfils(t, UnicodeValue('\''), func(t *testing.T, f parser.Func[bool]) {
		testUnicodeChar(t, '\'', func(p *parser.Parser) bool {
			return parser.Try(p, f)
		})
	})

	parsetest.AlsoFulfils(t, UnicodeValue('"'), testLittleUValue)
	parsetest.AlsoFulfils(t, UnicodeValue('"'), testBigUValue)
	parsetest.AlsoFulfils(t, UnicodeValue('"'), testEscapedChar('"'))
	parsetest.AlsoFulfils(t, UnicodeValue('"'), func(t *testing.T, f parser.Func[bool]) {
		testUnicodeChar(t, '"', func(p *parser.Parser) bool {
			return parser.Try(p, f)
		})
	})
}

func TestByteValue(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, ByteValue(), testOctalByteValue)
	parsetest.AlsoFulfils(t, ByteValue(), testHexByteValue)
}

func TestOctalByteValue(t *testing.T) {
	t.Parallel()
	testOctalByteValue(t, OctalByteValue())
}

func testOctalByteValue(t *testing.T, f parser.Func[bool]) {
	tests := []string{
		`\123`,
		`\567`,
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesExact(t, c, f)
		})
	}
}

func TestHexByteValue(t *testing.T) {
	t.Parallel()
	testHexByteValue(t, HexByteValue())
}

func testHexByteValue(t *testing.T, f parser.Func[bool]) {
	tests := []string{
		`\x12`,
		`\xef`,
		`\xEF`,
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesExact(t, c, f)
		})
	}
}

func TestUnicodeChar(t *testing.T) {
	t.Parallel()

	testUnicodeChar(t, '\'', UnicodeChar('\''))
	testUnicodeChar(t, '"', UnicodeChar('"'))
}

func testUnicodeChar(t *testing.T, except rune, f parser.Func[bool]) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tests := []rune{'a', 'ä', '🐶'}
		switch except {
		case '\'':
			tests = append(tests, '"')
		case '"':
			tests = append(tests, '\'')
		}

		for _, c := range tests {
			t.Run(string(c), func(t *testing.T) {
				t.Parallel()
				parsetest.ParsesExact(t, string(c), f)
			})
		}
	})
	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		tests := []rune{'\n', except}

		for _, c := range tests {
			t.Run(string(c), func(t *testing.T) {
				t.Parallel()
				parsetest.NoMatch(t, string(c), f)
			})
		}
	})
}

func TestLittleUValue(t *testing.T) {
	t.Parallel()
	testLittleUValue(t, LittleUValue())
}

func testLittleUValue(t *testing.T, f parser.Func[bool]) {
	tests := []string{`\u1234`, `\uefef`, `\uEFEF`}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesExact(t, c, f)
		})
	}
}

func TestBigUValue(t *testing.T) {
	t.Parallel()
	testBigUValue(t, BigUValue())
}

func testBigUValue(t *testing.T, f parser.Func[bool]) {
	tests := []string{`\U12345678`, `\Uefefefef`, `\UEFEFEFEF`}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesExact(t, c, f)
		})
	}
}

func TestEscapedChar(t *testing.T) {
	t.Parallel()

	testEscapedChar('\'')
	testEscapedChar('"')
}

func testEscapedChar(term rune) func(t *testing.T, f parser.Func[bool]) {
	return func(t *testing.T, f parser.Func[bool]) {
		tests := []string{`\a`, `\b`, `\f`, `\n`, `\r`, `\t`, `\v`, `\\`, `\` + string(term)}

		for _, c := range tests {
			t.Run(c, func(t *testing.T) {
				t.Parallel()
				parsetest.ParsesExact(t, c, f)
			})
		}
	}
}
