package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Rune literals
// ======================================================================================

func TestRuneLit(t *testing.T) {
	t.Parallel()

	testCases := []string{
		`'a'`, `'ä'`, `'🐶'`, `'\n'`, `'\''`, `'"'`, `'\u1234'`, `'\U12345678'`,
		`'\x12'`, `'\123'`,
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c, RuneLit())
			assert.Equal(t, c, actual)
		})
	}
}

func TestUnicodeValue(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, UnicodeValue('\''), testLittleUValue)
	testutil.AssertAlsoFulfils(t, UnicodeValue('\''), testBigUValue)
	testutil.AssertAlsoFulfils(t, UnicodeValue('\''), testEscapedChar('\''))
	testutil.AssertAlsoFulfils(t, UnicodeValue('\''), func(t *testing.T, f parser.Func[string]) {
		testUnicodeChar(t, '\'', func(p *parser.Parser) (rune, *fancyerr.Error) {
			s, err := parser.Try(p, f)
			if err != nil {
				return 0, err
			}
			rs := []rune(s)
			if len(rs) != 1 {
				return 0, &fancyerr.Error{
					Message: "expected a single rune",
				}
			}
			return rs[0], nil
		})
	})

	testutil.AssertAlsoFulfils(t, UnicodeValue('"'), testLittleUValue)
	testutil.AssertAlsoFulfils(t, UnicodeValue('"'), testBigUValue)
	testutil.AssertAlsoFulfils(t, UnicodeValue('"'), testEscapedChar('"'))
	testutil.AssertAlsoFulfils(t, UnicodeValue('"'), func(t *testing.T, f parser.Func[string]) {
		testUnicodeChar(t, '"', func(p *parser.Parser) (rune, *fancyerr.Error) {
			s, err := parser.Try(p, f)
			if err != nil {
				return 0, err
			}
			rs := []rune(s)
			if len(rs) != 1 {
				return 0, &fancyerr.Error{
					Message: "expected a single rune",
				}
			}
			return rs[0], nil
		})
	})

}

func TestByteValue(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, ByteValue(), testOctalByteValue)
	testutil.AssertAlsoFulfils(t, ByteValue(), testHexByteValue)
}

func TestOctalByteValue(t *testing.T) {
	t.Parallel()
	testOctalByteValue(t, OctalByteValue())
}

func testOctalByteValue(t *testing.T, f parser.Func[string]) {
	testCases := []string{
		`\123`,
		`\567`,
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, c, f)
		})
	}
}

func TestHexByteValue(t *testing.T) {
	t.Parallel()
	testHexByteValue(t, HexByteValue())
}

func testHexByteValue(t *testing.T, f parser.Func[string]) {
	testCases := []string{
		`\x12`,
		`\xef`,
		`\xEF`,
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, c, f)
		})
	}
}

func TestUnicodeChar(t *testing.T) {
	t.Parallel()

	testUnicodeChar(t, '\'', UnicodeChar('\''))
	testUnicodeChar(t, '"', UnicodeChar('"'))
}

func testUnicodeChar(t *testing.T, except rune, f parser.Func[rune]) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testCases := []rune{'a', 'ä', '🐶'}
		switch except {
		case '\'':
			testCases = append(testCases, '"')
		case '"':
			testCases = append(testCases, '\'')
		}

		for _, c := range testCases {
			t.Run(string(c), func(t *testing.T) {
				t.Parallel()
				testutil.ParsesFully(t, string(c), f)
			})
		}
	})
	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []rune{'\n', except}

		for _, c := range testCases {
			t.Run(string(c), func(t *testing.T) {
				t.Parallel()
				testutil.NoMatch(t, string(c), f)
			})
		}
	})
}

func TestLittleUValue(t *testing.T) {
	t.Parallel()
	testLittleUValue(t, LittleUValue())
}

func testLittleUValue(t *testing.T, f parser.Func[string]) {
	testCases := []string{`\u1234`, `\uefef`, `\uEFEF`}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, c, f)
		})
	}
}

func TestBigUValue(t *testing.T) {
	t.Parallel()
	testBigUValue(t, BigUValue())
}

func testBigUValue(t *testing.T, f parser.Func[string]) {
	testCases := []string{`\U12345678`, `\Uefefefef`, `\UEFEFEFEF`}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, c, f)
		})
	}
}

func TestEscapedChar(t *testing.T) {
	t.Parallel()

	testEscapedChar('\'')
	testEscapedChar('"')
}

func testEscapedChar(term rune) func(t *testing.T, f parser.Func[string]) {
	return func(t *testing.T, f parser.Func[string]) {
		testCases := []string{`\a`, `\b`, `\f`, `\n`, `\r`, `\t`, `\v`, `\\`, `\` + string(term)}

		for _, c := range testCases {
			t.Run(c, func(t *testing.T) {
				t.Parallel()
				testutil.ParsesFully(t, c, f)
			})
		}
	}
}

// ============================================================================
// String literals
// ======================================================================================

func TestStringLit(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, StringLit(), testRawStringLit)
	testutil.AssertAlsoFulfils(t, StringLit(), testInterpretedStringLit)
}

func TestRawStringLit(t *testing.T) {
	t.Parallel()
	testRawStringLit(t, RawStringLit())
}

func testRawStringLit(t *testing.T, f parser.Func[*ast.StaticString]) {
	testCases := []string{
		"``",
		"`woof`",
		"`woof\n`",
		"`woof\nbark`",
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, c, f)
		})
	}
}

func TestInterpretedStringLit(t *testing.T) {
	t.Parallel()
	testInterpretedStringLit(t, InterpretedStringLit())
}

func testInterpretedStringLit(t *testing.T, f parser.Func[*ast.StaticString]) {
	testCases := []string{
		`""`,
		`"woof"`,
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			testutil.ParsesFully(t, c, f)
		})
	}
}
