package component

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestComponentCallBody(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, CallBody(), testDefaultBlockShorthand)
	t.Run("Body", func(t *testing.T) {
		t.Parallel()
		in := "{\n" +
			"\tbr\n" +
			"}"
		want := &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 1},
			Nodes: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementReference{
							Name: &ast.ElementName{
								Name:     "br",
								Position: &ast.Position{Line: 2, Col: 1 + len("\t")},
							},
						},
					},
				},
			},
			RBrace: &ast.Position{Line: 3, Col: 1},
		}

		got := parsetest.ParsesFully(t, in, Body())
		should.Equal[ast.ComponentBody](t, want, got)
	})
}

func TestDefaultBlockShorthand(t *testing.T) {
	t.Parallel()
	testDefaultBlockShorthand(t, DefaultBlockShorthand())
}

func testDefaultBlockShorthand(t *testing.T, f parser.Func[*ast.DefaultBlockShorthand]) {
	tests := []struct {
		name string
		in   string
		want *ast.DefaultBlockShorthand
	}{
		{
			name: "bracket text",
			in:   "_[ foo ]",
			want: &ast.DefaultBlockShorthand{
				Body: &ast.BracketText{
					LBracket: &ast.Position{Line: 1, Col: 2},
					Lines: ast.TextBlock{
						ast.TextLine{
							&ast.Text{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 4},
							},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 8},
				},
				Position: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "scope",
			in:   "_{}",
			want: &ast.DefaultBlockShorthand{
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 2},
					RBrace: &ast.Position{Line: 1, Col: 3},
				},
				Position: &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}
