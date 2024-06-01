package comment

import (
	"testing"

	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInlineGroup(t *testing.T) {
	t.Parallel()

	t.Run("block comment", func(t *testing.T) {
		t.Parallel()

		t.Run("single-line", func(t *testing.T) {
			t.Parallel()

			expect := &ast.CommentGroup{
				Comments: []*ast.Comment{
					{
						Open:    ast.Position{Line: 1, Col: 1},
						Comment: " hello ",
						Block:   true,
						Close:   &ast.Position{Line: 1, Col: 10},
					},
				},
			}

			actual := testutil.AssertParsesFully(t, "/* hello */", InlineGroup())
			assert.Equal(t, expect, actual)
		})

		t.Run("multi-line", func(t *testing.T) {
			t.Parallel()

			testutil.AssertMatchesButError(t, "/* hello\nworld */", InlineGroup())
		})
	})

	t.Run("line comment", func(t *testing.T) {
		t.Parallel()

		testutil.AssertMatchesButError(t, "// hello", InlineGroup())
	})
}

func TestLoneGroup(t *testing.T) {
	t.Parallel()

	t.Run("block comment", func(t *testing.T) {
		t.Parallel()

		t.Run("single-line", func(t *testing.T) {
			t.Parallel()

			expect := &ast.CommentGroup{
				Comments: []*ast.Comment{
					{
						Open:    ast.Position{Line: 1, Col: 1},
						Comment: " hello ",
						Block:   true,
						Close:   &ast.Position{Line: 1, Col: 10},
					},
				},
			}

			in := "/* hello */"
			actual := testutil.AssertParsesFully(t, in, LoneGroup())
			assert.Equal(t, expect, actual)

			p := testutil.NewParser(t, in+"/* ignore me */")
			actual = testutil.AssertNoError(t, p, LoneGroup())
			assert.Equal(t, expect, actual)
		})

		t.Run("multi-line", func(t *testing.T) {
			t.Parallel()

			expect := &ast.CommentGroup{
				Comments: []*ast.Comment{
					{
						Open:    ast.Position{Line: 1, Col: 1},
						Comment: " hello\n   world ",
						Block:   true,
						Close:   &ast.Position{Line: 2, Col: 10},
					},
				},
			}

			in := "/* hello\n   world */"
			actual := testutil.AssertParsesFully(t, in, LoneGroup())
			assert.Equal(t, expect, actual)

			p := testutil.NewParser(t, in+"/* ignore me */")
			actual = testutil.AssertNoError(t, p, LoneGroup())
			assert.Equal(t, expect, actual)
		})
	})

	t.Run("line comment", func(t *testing.T) {
		t.Parallel()

		t.Run("single", func(t *testing.T) {
			t.Parallel()

			expect := &ast.CommentGroup{
				Comments: []*ast.Comment{
					{
						Open:    ast.Position{Line: 1, Col: 1},
						Comment: " hello",
						Block:   false,
						Close:   &ast.Position{Line: 1, Col: 9},
					},
				},
			}

			in := "// hello"
			actual := testutil.AssertParsesFully(t, in, LoneGroup())
			assert.Equal(t, expect, actual)

			actual = testutil.AssertParsesFully(t, in+"\n", LoneGroup())
			assert.Equal(t, expect, actual)

			p := testutil.NewParser(t, in+"\n/* ignore me */")
			actual = testutil.AssertNoError(t, p, LoneGroup())
			assert.Equal(t, expect, actual)
		})

		t.Run("multiple", func(t *testing.T) {
			t.Parallel()

			expect := &ast.CommentGroup{
				Comments: []*ast.Comment{
					{
						Open:    ast.Position{Line: 1, Col: 1},
						Comment: " hello",
						Block:   false,
						Close:   &ast.Position{Line: 1, Col: 9},
					},
					{
						Open:    ast.Position{Line: 2, Col: 1},
						Comment: " world",
						Block:   false,
						Close:   &ast.Position{Line: 2, Col: 9},
					},
				},
			}

			in := "// hello\n// world"

			actual := testutil.AssertParsesFully(t, in, LoneGroup())
			assert.Equal(t, expect, actual)

			actual = testutil.AssertParsesFully(t, in+"\n", LoneGroup())
			assert.Equal(t, expect, actual)

			p := testutil.NewParser(t, in+"\n/* ignore me */")
			actual = testutil.AssertNoError(t, p, LoneGroup())
			assert.Equal(t, expect, actual)

			p = testutil.NewParser(t, in+"\n\n// ignore me")
			actual = testutil.AssertNoError(t, p, LoneGroup())
			assert.Equal(t, expect, actual)
		})
	})
}
