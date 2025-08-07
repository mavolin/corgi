package text

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestArrowBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ArrowBlock
	}{
		{
			name: "empty",
			in:   ">",
			want: &ast.ArrowBlock{
				Arrow: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "single line",
			in:   "> foo",
			want: &ast.ArrowBlock{
				Arrow: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					{
						&ast.Text{
							Text:     "foo",
							Position: &ast.Position{Line: 1, Col: 3},
						},
					},
				},
			},
		}, {
			name: "multi line",
			in: "> foo\n" +
				"\n" +
				"  bar",
			want: &ast.ArrowBlock{
				Arrow: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					{
						&ast.Text{
							Text:     "foo",
							Position: &ast.Position{Line: 1, Col: 3},
						},
					}, {
						&ast.Text{
							Text:     "bar",
							Position: &ast.Position{Line: 3, Col: 3},
						},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsesTextFully(t, c.in, ArrowBlock())
			should.Equal(t, got, c.want)
		})
	}
}

func TestLine(t *testing.T) {
	t.Parallel()

	in := "foo #{bar}##baz #_"
	want := ast.TextLine{
		&ast.Text{
			Text:     "foo",
			Position: &ast.Position{Line: 1, Col: 1},
		},
		&ast.ExpressionInterpolation{
			Hash:   &ast.Position{Line: 1, Col: 5},
			LBrace: &ast.Position{Line: 1, Col: 6},
			Expression: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{
						Code:     "bar",
						Position: &ast.Position{Line: 1, Col: 7},
					},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 10},
		},
		&ast.EscapedHash{Hash: &ast.Position{Line: 1, Col: 11}},
		&ast.Text{
			Text:     "baz",
			Position: &ast.Position{Line: 1, Col: 13},
		},
		&ast.HashSpace{Hash: &ast.Position{Line: 1, Col: 17}},
	}

	got := parsesTextFully(t, in, Line('\n'))
	should.Equal(t, got, want)
}

func TestVerbatimLine(t *testing.T) {
	t.Parallel()

	in := "foo #{bar}##baz #? #_"
	want := ast.TextLine{
		&ast.Text{
			Text:     "foo #{bar}##baz #? #_",
			Position: &ast.Position{Line: 1, Col: 1},
		},
	}

	got := parsesTextFully(t, in, VerbatimLine('\n'))
	should.Equal(t, got, want)
}

func TestText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{
			name: "simple",
			in:   "foo",
		}, {
			name: "unambiguous hash at end",
			in:   "foo#",
		}, {
			name: "unambiguous hash within",
			in:   "foo # bar",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			want := &ast.Text{
				Text:     c.in,
				Position: &ast.Position{Line: 1, Col: 1},
			}
			got := parsesTextFully(t, c.in, Text('\n'))
			should.Equal(t, got, want)
		})
	}
}

func parsesTextFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := parsetest.NewParser(t, input+"\n1other stuff")
	v := parsetest.AssertNoError(t, p, f)

	line, col, index := parsetest.CalcEnd(1, 1, 0, input)
	parsetest.AssertPosition(t, p, line, col, index)

	return v
}
