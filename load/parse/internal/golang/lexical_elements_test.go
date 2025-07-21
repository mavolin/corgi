package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
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
			got := parsetest.ParsesFully(t, c, RuneLit())
			should.Equal(t, c, got)
		})
	}
}

func TestUnicodeValue(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, UnicodeValue('\''), testLittleUValue)
	parsetest.AssertAlsoFulfils(t, UnicodeValue('\''), testBigUValue)
	parsetest.AssertAlsoFulfils(t, UnicodeValue('\''), testEscapedChar('\''))
	parsetest.AssertAlsoFulfils(t, UnicodeValue('\''), func(t *testing.T, f parser.Func[string]) {
		testUnicodeChar(t, '\'', func(p *parser.Parser) (rune, *diagnostic.Diagnostic) {
			s, err := parser.TryErr(p, f)
			if err != nil {
				return 0, err
			}
			rs := []rune(s)
			if len(rs) != 1 {
				return 0, &diagnostic.Diagnostic{
					Message: "expected a single rune",
				}
			}
			return rs[0], nil
		})
	})

	parsetest.AssertAlsoFulfils(t, UnicodeValue('"'), testLittleUValue)
	parsetest.AssertAlsoFulfils(t, UnicodeValue('"'), testBigUValue)
	parsetest.AssertAlsoFulfils(t, UnicodeValue('"'), testEscapedChar('"'))
	parsetest.AssertAlsoFulfils(t, UnicodeValue('"'), func(t *testing.T, f parser.Func[string]) {
		testUnicodeChar(t, '"', func(p *parser.Parser) (rune, *diagnostic.Diagnostic) {
			s, err := parser.TryErr(p, f)
			if err != nil {
				return 0, err
			}
			rs := []rune(s)
			if len(rs) != 1 {
				return 0, &diagnostic.Diagnostic{
					Message: "expected a single rune",
				}
			}
			return rs[0], nil
		})
	})

}

func TestByteValue(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, ByteValue(), testOctalByteValue)
	parsetest.AssertAlsoFulfils(t, ByteValue(), testHexByteValue)
}

func TestOctalByteValue(t *testing.T) {
	t.Parallel()
	testOctalByteValue(t, OctalByteValue())
}

func testOctalByteValue(t *testing.T, f parser.Func[string]) {
	tests := []string{
		`\123`,
		`\567`,
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesFully(t, c, f)
		})
	}
}

func TestHexByteValue(t *testing.T) {
	t.Parallel()
	testHexByteValue(t, HexByteValue())
}

func testHexByteValue(t *testing.T, f parser.Func[string]) {
	tests := []string{
		`\x12`,
		`\xef`,
		`\xEF`,
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesFully(t, c, f)
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
				parsetest.ParsesFully(t, string(c), f)
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

func testLittleUValue(t *testing.T, f parser.Func[string]) {
	tests := []string{`\u1234`, `\uefef`, `\uEFEF`}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesFully(t, c, f)
		})
	}
}

func TestBigUValue(t *testing.T) {
	t.Parallel()
	testBigUValue(t, BigUValue())
}

func testBigUValue(t *testing.T, f parser.Func[string]) {
	tests := []string{`\U12345678`, `\Uefefefef`, `\UEFEFEFEF`}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesFully(t, c, f)
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
		tests := []string{`\a`, `\b`, `\f`, `\n`, `\r`, `\t`, `\v`, `\\`, `\` + string(term)}

		for _, c := range tests {
			t.Run(c, func(t *testing.T) {
				t.Parallel()
				parsetest.ParsesFully(t, c, f)
			})
		}
	}
}

// ============================================================================
// String literals
// ======================================================================================

func TestStringLit(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, StringLit(), testRawStringLit)
	parsetest.AssertAlsoFulfils(t, StringLit(), testInterpretedStringLit)
}

func TestRawStringLit(t *testing.T) {
	t.Parallel()
	testRawStringLit(t, RawStringLit())
}

func testRawStringLit(t *testing.T, f parser.Func[*ast.StaticString]) {
	tests := []string{
		"``",
		"`woof`",
		"`woof\n`",
		"`woof\nbark`",
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesFully(t, c, f)
		})
	}
}

func TestInterpretedStringLit(t *testing.T) {
	t.Parallel()
	testInterpretedStringLit(t, InterpretedStringLit())
}

func testInterpretedStringLit(t *testing.T, f parser.Func[*ast.StaticString]) {
	tests := []string{
		`""`,
		`"woof"`,
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			parsetest.ParsesFully(t, c, f)
		})
	}
}
