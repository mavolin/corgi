package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestBracketText(t *testing.T) {
	t.Parallel()
	testBracketText(t, BracketText())
}

func testBracketText(t *testing.T, f parser.Func[*ast.BracketText]) {
	tests := []struct {
		name string
		in   string
		want *ast.BracketText
	}{
		{
			name: "empty",
			in:   "[]",
			want: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				RBracket: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "single line",
			in:   "[ foo ]",
			want: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					ast.TextLine{
						&ast.Text{
							Text:     "foo",
							Position: &ast.Position{Line: 1, Col: 3},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "multi line",
			in: "[\n" +
				"\tfoo\n" +
				"\tbar\n" +
				"]",
			want: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					ast.TextLine{
						&ast.Text{
							Text:     "foo",
							Position: &ast.Position{Line: 2, Col: 2},
						},
					}, ast.TextLine{
						&ast.Text{
							Text:     "bar",
							Position: &ast.Position{Line: 3, Col: 2},
						},
					},
				},
				RBracket: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, got, c.want)
		})
	}
}

func TestVerbatimBracketText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.BracketText
	}{
		{
			name: "empty",
			in:   "[]",
			want: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				RBracket: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "single line",
			in:   "[ foo #bar ]",
			want: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					ast.TextLine{
						&ast.Text{
							Text:     "foo #bar",
							Position: &ast.Position{Line: 1, Col: len("[ ") + 1},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: len("[ foo #bar ") + 1},
			},
		}, {
			name: "multi line",
			in: "[\n" +
				"\tfoo #bar \n" +
				"\tbar #:baz()\n" +
				"]",
			want: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					ast.TextLine{
						&ast.Text{
							Text:     "foo #bar",
							Position: &ast.Position{Line: 2, Col: 2},
						},
					}, ast.TextLine{
						&ast.Text{
							Text:     "bar #:baz()",
							Position: &ast.Position{Line: 3, Col: 2},
						},
					},
				},
				RBracket: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, VerbatimBracketText())
			should.Equal(t, got, c.want)
		})
	}
}
