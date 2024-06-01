package comment

import (
	"strings"
	"testing"

	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestOrHorizontalWhitespace(t *testing.T) {
	t.Parallel()

	successCases := []struct {
		in             string
		expectComments []*ast.Comment
	}{
		{in: "  "}, {in: "\t"}, {in: "  \t  \t\t "},
		{
			in: "/* test */",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 9},
				},
			},
		}, {
			in: "  /* test */ \t/* test2 */\t",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 11},
				}, {
					Open:    ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 24},
				},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			t.Run(testName(c.in), func(t *testing.T) {
				t.Parallel()

				p := testutil.NewParser(t, c.in)
				testutil.AssertNoError(t, p, OrHorizontalWhitespace())
				testutil.AssertEOF(t, p)

				expectGroups := make([]*ast.CommentGroup, len(c.expectComments))
				for i, comment := range c.expectComments {
					expectGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
				}
				assert.Equal(t, expectGroups, p.CloneState().Comments())
			})
		}
	})

	failureCases := []string{"/* dsafasd\nsda*/", "// sdafads"}

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		for _, c := range failureCases {
			t.Run(testName(c), func(t *testing.T) {
				t.Parallel()
				testutil.AssertMatchesButError(t, c, OrHorizontalWhitespace())
			})
		}
	})
}

func TestAndEOS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		in             string
		expectIndex    int
		expectComments []*ast.Comment
	}{
		{in: "\n", expectIndex: 0}, {in: "\r\n", expectIndex: 0}, {in: "  \t  \t\t\n", expectIndex: 7},
		{in: "  \t  \t\t\r\n", expectIndex: 7}, {in: "  \t  \t\t }", expectIndex: 8},
		{
			in:          "/* test */ ;",
			expectIndex: -1,
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 9},
				},
			},
		}, {
			in:          " /* test\n */",
			expectIndex: 1,
		}, {
			in:          "  /* test */ \t/* test2 */\n",
			expectIndex: 25,
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 11},
				}, {
					Open:    ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 24},
				},
			},
		}, {
			in:          " // foo",
			expectIndex: 1,
		},
	}

	for _, c := range testCases {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			if c.expectIndex == -1 {
				c.expectIndex = len(c.in)
			}

			p := testutil.NewParser(t, c.in)
			testutil.AssertNoError(t, p, AndEOS())
			assert.Equal(t, c.expectIndex, p.Index())

			expectGroups := make([]*ast.CommentGroup, len(c.expectComments))
			for i, comment := range c.expectComments {
				expectGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			assert.Equal(t, expectGroups, p.CloneState().Comments())
		})
	}
}

func TestOrEOL(t *testing.T) {
	t.Parallel()
	testOrEOL(t, OrEOL())
}

func testOrEOL(t *testing.T, f parser.Func[struct{}]) {
	testCases := []struct {
		in             string
		expectComments []*ast.Comment
	}{
		{in: ""}, {in: "\n"}, {in: "\r\n"}, {in: "  \t  \t\t\n"}, {in: "  \t  \t\t\r\n"},
		{
			in: "/* test */",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 9},
				},
			},
		}, {
			in: "  /* test */ \t/* test2 */\n",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 11},
				}, {
					Open:    ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 24},
				},
			},
		}, {
			in: " // foo",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 2},
					Comment: " foo",
					Block:   false,
					Close:   &ast.Position{Line: 1, Col: 8},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in)
			testutil.AssertNoError(t, p, f)
			testutil.AssertEOF(t, p)

			expectGroups := make([]*ast.CommentGroup, len(c.expectComments))
			for i, comment := range c.expectComments {
				expectGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			assert.Equal(t, expectGroups, p.CloneState().Comments())
		})
	}
}

func TestOrEOLWhitespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		in             string
		expectComments []*ast.Comment
	}{
		{
			in: "  /* test */ \t/* test2 */\n" +
				"// foo\n" +
				"/* bar */\n",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 11},
				}, {
					Open:    ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 24},
				}, {
					Open:    ast.Position{Line: 2, Col: 1},
					Comment: " foo",
					Block:   false,
					Close:   &ast.Position{Line: 2, Col: 7},
				}, {
					Open:    ast.Position{Line: 3, Col: 1},
					Comment: " bar ",
					Block:   true,
					Close:   &ast.Position{Line: 3, Col: 8},
				},
			},
		}, {
			in: " // foo\n// bar",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 2},
					Comment: " foo",
					Block:   false,
					Close:   &ast.Position{Line: 1, Col: 8},
				}, {
					Open:    ast.Position{Line: 2, Col: 1},
					Comment: " bar",
					Block:   false,
					Close:   &ast.Position{Line: 2, Col: 7},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in)
			testutil.AssertNoError(t, p, OrEOLWhitespace())
			testutil.AssertEOF(t, p)

			expectGroups := make([]*ast.CommentGroup, len(c.expectComments))
			for i, comment := range c.expectComments {
				expectGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			assert.Equal(t, expectGroups, p.CloneState().Comments())
		})
	}

	testOrEOL(t, OrEOLWhitespace())
}

func TestOrLoneWS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		in             string
		expectComments []*ast.Comment
	}{
		{in: "  "}, {in: "\t"}, {in: "  \t  \t\t "},
		{
			in: "/* test */",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 9},
				},
			},
		}, {
			in: "  /* test */\n /* test2 */\n",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 11},
				}, {
					Open:    ast.Position{Line: 2, Col: 2},
					Comment: " test2 ",
					Block:   true,
					Close:   &ast.Position{Line: 2, Col: 11},
				},
			},
		}, {
			in: " // foo\n/* bar\n baz */",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 2},
					Comment: " foo",
					Block:   false,
					Close:   &ast.Position{Line: 1, Col: 8},
				}, {
					Open:    ast.Position{Line: 2, Col: 1},
					Comment: " bar\n baz ",
					Block:   true,
					Close:   &ast.Position{Line: 3, Col: 6},
				},
			},
		},
		{
			in: "/* foo */ // bar",
			expectComments: []*ast.Comment{
				{
					Open:    ast.Position{Line: 1, Col: 1},
					Comment: " foo ",
					Block:   true,
					Close:   &ast.Position{Line: 1, Col: 8},
				}, {
					Open:    ast.Position{Line: 1, Col: 11},
					Comment: " bar",
					Block:   false,
					Close:   &ast.Position{Line: 1, Col: 17},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in)
			testutil.AssertNoError(t, p, OrLoneWS())
			testutil.AssertEOF(t, p)

			expectGroups := make([]*ast.CommentGroup, len(c.expectComments))
			for i, comment := range c.expectComments {
				expectGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			assert.Equal(t, expectGroups, p.CloneState().Comments())
		})
	}
}

func testName(input string) string {
	var name strings.Builder

	for i := 0; i < len(input); i++ {
		switch input[i] {
		case ' ':
			name.WriteString("_")
		case '\t':
			name.WriteString("\\t")
		case '\n':
			name.WriteString("\\n")
		case '\r':
			name.WriteString("\\r")
		case '/':
			if i+1 >= len(input) {
				break
			}
			i++
			switch input[i] {
			case '/':
				name.WriteString("<ln comm>")
				for ; i < len(input); i++ {
					if input[i] == '\n' {
						i--
						break
					}
				}
			case '*':
				name.WriteString("<blo comm>")
				for ; i < len(input); i++ {
					if input[i] == '*' && i+1 < len(input) && input[i+1] == '/' {
						i++
						break
					}
				}
			}
		}
	}
	return name.String()
}
