package text

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestArrowBlock(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ArrowBlock
	}{
		{
			name: "empty",
			in:   ">",
			expect: &ast.ArrowBlock{
				Arrow: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "single line",
			in:   "> foo",
			expect: &ast.ArrowBlock{
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
			expect: &ast.ArrowBlock{
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := parsesTextFully(t, c.in, ArrowBlock())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestLine(t *testing.T) {
	t.Parallel()

	in := "foo #{bar}##baz #_"
	expect := ast.TextLine{
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

	actual := parsesTextFully(t, in, Line('\n'))
	assert.Equal(t, expect, actual)
}

func TestVerbatimLine(t *testing.T) {
	t.Parallel()

	in := "foo #{bar}##baz #? #_"
	expect := ast.TextLine{
		&ast.Text{
			Text:     "foo #{bar}##baz #? #_",
			Position: &ast.Position{Line: 1, Col: 1},
		},
	}

	actual := parsesTextFully(t, in, VerbatimLine('\n'))
	assert.Equal(t, expect, actual)
}

func TestText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			expect := &ast.Text{
				Text:     c.in,
				Position: &ast.Position{Line: 1, Col: 1},
			}
			actual := parsesTextFully(t, c.in, Text('\n'))
			assert.Equal(t, expect, actual)
		})
	}
}

func parsesTextFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := testutil.NewParser(t, input+"\n1other stuff")
	v := testutil.AssertNoError(t, p, f)

	line, col, index := testutil.CalcEnd(1, 1, 0, input)
	testutil.AssertPosition(t, p, line, col, index)

	return v
}
