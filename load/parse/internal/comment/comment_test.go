package comment

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestComment(t *testing.T) {
	t.Parallel()

	t.Run("line", func(t *testing.T) {
		t.Parallel()
		testLineComment(t, Comment())
	})

	t.Run("block", func(t *testing.T) {
		t.Parallel()
		testBlockComment(t, Comment())
	})
}

func TestLineComment(t *testing.T) {
	t.Parallel()
	testLineComment(t, LineComment())
}

func testLineComment(t *testing.T, f parser.Func[*ast.Comment]) {
	expect := &ast.Comment{
		Open:    ast.Position{Line: 1, Col: 1},
		Comment: " foo",
		Block:   false,
		Close:   &ast.Position{Line: 1, Col: 7},
	}

	actual := testutil.ParsesFully(t, "// foo\n", f)
	assert.Equal(t, expect, actual)
}

func TestBlockComment(t *testing.T) {
	t.Parallel()
	testBlockComment(t, BlockComment())
}

func testBlockComment(t *testing.T, f parser.Func[*ast.Comment]) {
	t.Run("single line", func(t *testing.T) {
		t.Parallel()

		expect := &ast.Comment{
			Open:    ast.Position{Line: 1, Col: 1},
			Comment: " foo ",
			Block:   true,
			Close:   &ast.Position{Line: 1, Col: 8},
		}

		actual := testutil.ParsesFully(t, "/* foo */", f)
		assert.Equal(t, expect, actual)
	})

	t.Run("multi line", func(t *testing.T) {
		t.Parallel()

		expect := &ast.Comment{
			Open:    ast.Position{Line: 1, Col: 1},
			Comment: " foo\n   bar ",
			Block:   true,
			Close:   &ast.Position{Line: 2, Col: 8},
		}

		actual := testutil.ParsesFully(t, "/* foo\n   bar */", f)
		assert.Equal(t, expect, actual)
	})

	t.Run("missing closing", func(t *testing.T) {
		t.Parallel()

		p := testutil.NewParser(t, "/* foo")
		p.DoInline(func() {
			testutil.AssertMatchesButError(t, p, f)
		})
	})

	t.Run("inline", func(t *testing.T) {
		t.Parallel()

		t.Run("single line", func(t *testing.T) {
			t.Parallel()

			expect := &ast.Comment{
				Open:    ast.Position{Line: 1, Col: 1},
				Comment: " foo ",
				Block:   true,
				Close:   &ast.Position{Line: 1, Col: 8},
			}

			p := testutil.NewParser(t, "/* foo */")
			p.DoInline(func() {
				actual := testutil.AssertNoError(t, p, f)
				testutil.AssertEOF(t, p)
				assert.Equal(t, expect, actual)
			})
		})

		t.Run("error on multi line", func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, "/* foo\n   bar */")
			p.DoInline(func() {
				testutil.AssertMatchesButError(t, p, f)
			})
		})

		t.Run("missing closing", func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, "/* foo\n*/")
			p.DoInline(func() {
				testutil.AssertMatchesButError(t, p, f)
			})
		})
	})
}
