package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func TestGoCode(t *testing.T) {
	t.Parallel()
	testGoCode()(t, GoCode(Regular))
}

func testGoCode() func(t *testing.T, f parser.Func[ast.Code]) {
	return func(t *testing.T, f parser.Func[ast.Code]) {
		tests := []struct {
			name string
			code string
			want []ast.CodeNode
		}{
			{
				name: "identifier",
				code: "woof",
			}, {
				name: "comparison",
				code: "len(woof) >= len(bark)",
			}, {
				name: "comma in parentheses",
				code: "(woof, bark)",
			}, {
				name: "comma in brackets",
				code: "a[woof, bark]",
			}, {
				name: "comma in braces",
				code: "func(){woof, bark}",
			}, {
				name: "semicolon in braces",
				code: "func(){woof; bark}",
			}, {
				name: "rune literal",
				code: "';'",
			}, {
				name: "mix",
				code: "foo(bar, baz) + string(';')",
			}, {
				name: "string in parentheses",
				code: "(\"foo\")",
				want: []ast.CodeNode{
					&ast.GoCode{Code: "(", Position: &ast.Position{Line: 1, Col: 1}},
					&ast.String{
						Open:  &ast.Position{Line: 1, Col: 2},
						Quote: '"',
						Contents: []ast.StringNode{
							&ast.StringText{Text: "foo", Position: &ast.Position{Line: 1, Col: 3}},
						},
						Close: &ast.Position{Line: 1, Col: 6},
					},
					&ast.GoCode{Code: ")", Position: &ast.Position{Line: 1, Col: 7}},
				},
			},
		}

		for _, c := range tests {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				want := c.want
				if want == nil {
					want = []ast.CodeNode{&ast.GoCode{Code: c.code, Position: &ast.Position{Line: 1, Col: 1}}}
				}

				got := parsesCodeNodeFully(t, c.code, f)
				should.Equal(t, got, want)
			})
		}
	}
}
