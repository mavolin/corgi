package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestScope(t *testing.T) {
	t.Parallel()
	testScope(t, Scope())
}

func testScope(t *testing.T, f parser.Func[*ast.Scope]) {
	tests := []struct {
		name string
		in   string
		want *ast.Scope
	}{
		{
			name: "empty",
			in:   "{}",
			want: &ast.Scope{
				LBrace: &ast.Position{Line: 1, Col: 1},
				RBrace: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "empty with newline",
			in:   "{\n}",
			want: &ast.Scope{
				LBrace: &ast.Position{Line: 1, Col: 1},
				RBrace: &ast.Position{Line: 2, Col: 1},
			},
		}, {
			name: "element",
			in: "{\n" +
				"\tbr\n" +
				"}",
			want: &ast.Scope{
				LBrace: &ast.Position{Line: 1, Col: 1},
				Nodes: []ast.ScopeNode{
					&ast.Element{
						Header: &ast.ElementHeader{
							Name: &ast.ElementReference{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
								},
							},
						},
					},
				},
				RBrace: &ast.Position{Line: 3, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, f)
			should.Equal(t, got, c.want)
		})
	}
}

func TestBadNode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{
			name: "gibberish",
			in:   "$%#",
		}, {
			name: "with parens",
			in:   "(br)",
		}, {
			name: "with braces",
			in:   "{br}",
		}, {
			name: "with brackets",
			in:   "[br]",
		}, {
			name: "mix",
			in:   "{br}() %44 [br]",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			want := &ast.BadNode{
				From:  ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 1 + len(c.in)},
			}

			got := parsetest.ParsesUntilEOS(t, c.in, BadNode())
			should.Equal(t, got, want)
		})
	}
}
