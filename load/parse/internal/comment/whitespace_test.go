package comment

import (
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestOrHorizontalWhitespace(t *testing.T) {
	t.Parallel()

	successCases := []struct {
		in           string
		wantComments []*ast.Comment
	}{
		{in: "  "},
		{in: "\t"},
		{in: "  \t  \t\t "},
		{
			in: "/* test */",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 9},
					Until:   ast.Position{Line: 1, Col: 11},
				},
			},
		},
		{
			in: "  /* test */ \t/* test2 */\t",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 11},
					Until:   ast.Position{Line: 1, Col: 13},
				}, {
					Open:    &ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 24},
					Until:   ast.Position{Line: 1, Col: 26},
				},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			t.Run(testName(c.in), func(t *testing.T) {
				t.Parallel()

				p := parsetest.NewParser(t, c.in)
				skipped := parser.TrySkip(p, OrHorizontalWhitespace())
				should.True(t, skipped)
				parsetest.AssertEOF(t, p)

				wantGroups := make([]*ast.CommentGroup, len(c.wantComments))
				for i, comment := range c.wantComments {
					wantGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
				}
				should.Equal(t, wantGroups, p.Comments())
			})
		}
	})

	failureCases := []string{"// sdafads"}

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		for _, c := range failureCases {
			t.Run(testName(c), func(t *testing.T) {
				t.Parallel()
				p := parsetest.NewParser(t, c)
				skipped := parser.TrySkip(p, OrHorizontalWhitespace())
				should.False(t, skipped) // want match error
			})
		}
	})
}

func TestAndEOS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in           string
		wantIndex    int
		wantComments []*ast.Comment
	}{
		{in: "\n", wantIndex: 0},
		{in: "\r\n", wantIndex: 0},
		{in: "  \t  \t\t\n", wantIndex: 7},
		{in: "  \t  \t\t\r\n", wantIndex: 7},
		{in: "  \t  \t\t }", wantIndex: 8},
		{
			in:        "/* test */ ;",
			wantIndex: -1,
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 9},
					Until:   ast.Position{Line: 1, Col: 11},
				},
			},
		},
		{
			in:        " /* test\n */",
			wantIndex: -1,
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 2},
					Comment: " test\n ",
					General: true,
					Close:   &ast.Position{Line: 2, Col: 2},
					Until:   ast.Position{Line: 2, Col: 4},
				},
			},
		},
		{
			in:        "  /* test */ \t/* test2 */\n",
			wantIndex: 25,
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 11},
					Until:   ast.Position{Line: 1, Col: 13},
				}, {
					Open:    &ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 24},
					Until:   ast.Position{Line: 1, Col: 26},
				},
			},
		},
		{
			in:        " // foo",
			wantIndex: 1,
		},
	}

	for _, c := range tests {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			if c.wantIndex == -1 {
				c.wantIndex = len(c.in)
			}

			p := parsetest.NewParser(t, c.in)
			matches := parser.Try(p, AndEOS())
			should.Equal(t, true, matches)
			should.Equal(t, c.wantIndex, p.Index())

			wantGroups := make([]*ast.CommentGroup, len(c.wantComments))
			for i, comment := range c.wantComments {
				wantGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			should.Equal(t, wantGroups, p.Comments())
		})
	}
}

func TestAndEOL(t *testing.T) {
	t.Parallel()

	t.Run("special", func(t *testing.T) {
		t.Parallel()

		t.Run("eof", func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, "")
			skipped := parser.TrySkip(p, AndEOL())
			should.True(t, skipped)
			parsetest.AssertEOF(t, p)
		})
	})

	testOrEOL(t, AndEOL())
}

func testOrEOL(t *testing.T, f parser.WhitespaceFunc) {
	tests := []struct {
		in           string
		wantComments []*ast.Comment
	}{
		{in: "\n"},
		{in: "\r\n"},
		{in: "  \t  \t\t\n"},
		{in: "  \t  \t\t\r\n"},
		{
			in: "/* test */",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 9},
					Until:   ast.Position{Line: 1, Col: 11},
				},
			},
		},
		{
			in: "  /* test */ \t/* test2 */\n",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 11},
					Until:   ast.Position{Line: 1, Col: 13},
				}, {
					Open:    &ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 24},
					Until:   ast.Position{Line: 1, Col: 26},
				},
			},
		},
		{
			in: " // foo",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 2},
					Comment: " foo",
					General: false,
					Until:   ast.Position{Line: 1, Col: 8},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in)
			skipped := parser.TrySkip(p, f)
			should.True(t, skipped)
			parsetest.AssertEOF(t, p)

			wantGroups := make([]*ast.CommentGroup, len(c.wantComments))
			for i, comment := range c.wantComments {
				wantGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			should.Equal(t, wantGroups, p.Comments())
		})
	}
}

func TestOrAnyWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in           string
		wantComments []*ast.Comment
	}{
		{
			in: "  /* test */ \t/* test2 */\n" +
				"// foo\n" +
				"/* bar */\n",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 11},
					Until:   ast.Position{Line: 1, Col: 13},
				}, {
					Open:    &ast.Position{Line: 1, Col: 15},
					Comment: " test2 ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 24},
					Until:   ast.Position{Line: 1, Col: 26},
				}, {
					Open:    &ast.Position{Line: 2, Col: 1},
					Comment: " foo",
					General: false,
					Until:   ast.Position{Line: 2, Col: 7},
				}, {
					Open:    &ast.Position{Line: 3, Col: 1},
					Comment: " bar ",
					General: true,
					Close:   &ast.Position{Line: 3, Col: 8},
					Until:   ast.Position{Line: 3, Col: 10},
				},
			},
		}, {
			in: " // foo\n// bar",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 2},
					Comment: " foo",
					General: false,
					Until:   ast.Position{Line: 1, Col: 8},
				}, {
					Open:    &ast.Position{Line: 2, Col: 1},
					Comment: " bar",
					General: false,
					Until:   ast.Position{Line: 2, Col: 7},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in)
			skipped := parser.TrySkip(p, OrAnyWhitespace())
			should.True(t, skipped)
			parsetest.AssertEOF(t, p)

			wantGroups := make([]*ast.CommentGroup, len(c.wantComments))
			for i, comment := range c.wantComments {
				wantGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			should.Equal(t, wantGroups, p.Comments())
		})
	}

	t.Run("eol", func(t *testing.T) {
		testOrEOL(t, OrAnyWhitespace())
	})
}

func TestOrLoneWS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in           string
		wantComments []*ast.Comment
	}{
		{in: "  "},
		{in: "\t"},
		{in: "  \t  \t\t "},
		{
			in: "/* test */",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 1},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 9},
					Until:   ast.Position{Line: 1, Col: 11},
				},
			},
		},
		{
			in: "  /* test */\n /* test2 */\n",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 3},
					Comment: " test ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 11},
					Until:   ast.Position{Line: 1, Col: 13},
				}, {
					Open:    &ast.Position{Line: 2, Col: 2},
					Comment: " test2 ",
					General: true,
					Close:   &ast.Position{Line: 2, Col: 11},
					Until:   ast.Position{Line: 2, Col: 13},
				},
			},
		},
		{
			in: " // foo\n/* bar\n baz */",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 2},
					Comment: " foo",
					General: false,
					Until:   ast.Position{Line: 1, Col: 8},
				}, {
					Open:    &ast.Position{Line: 2, Col: 1},
					Comment: " bar\n baz ",
					General: true,
					Close:   &ast.Position{Line: 3, Col: 6},
					Until:   ast.Position{Line: 3, Col: 8},
				},
			},
		},
		{
			in: "/* foo */ // bar",
			wantComments: []*ast.Comment{
				{
					Open:    &ast.Position{Line: 1, Col: 1},
					Comment: " foo ",
					General: true,
					Close:   &ast.Position{Line: 1, Col: 8},
					Until:   ast.Position{Line: 1, Col: 10},
				}, {
					Open:    &ast.Position{Line: 1, Col: 11},
					Comment: " bar",
					General: false,
					Until:   ast.Position{Line: 1, Col: 17},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(testName(c.in), func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in)
			skipped := parser.TrySkip(p, OrLoneWS())
			should.True(t, skipped)
			parsetest.AssertEOF(t, p)

			wantGroups := make([]*ast.CommentGroup, len(c.wantComments))
			for i, comment := range c.wantComments {
				wantGroups[i] = &ast.CommentGroup{Comments: []*ast.Comment{comment}}
			}
			should.Equal(t, wantGroups, p.Comments())
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
				name.WriteString("<ln comment>")
				for ; i < len(input); i++ {
					if input[i] == '\n' {
						i--
						break
					}
				}
			case '*':
				name.WriteString("<block comment>")
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
