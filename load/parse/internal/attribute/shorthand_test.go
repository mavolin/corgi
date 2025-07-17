package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestIDShorthand(t *testing.T) {
	t.Parallel()
	testIDShorthand(t, IDShorthand())
}

func testIDShorthand(t *testing.T, f parser.Func[*ast.IDShorthand]) {
	expect := &ast.IDShorthand{
		Hash: &ast.Position{Line: 1, Col: 1},
		ID: ast.Shorthand{
			&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 2}},
			&ast.ShorthandInterpolation{
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
				Hash:   &ast.Position{Line: 1, Col: 5},
			},
		},
	}

	actual := testutil.ParsesFully(t, "#foo#{bar}", f)
	assert.Equal(t, expect, actual)
}

func TestClassShorthand(t *testing.T) {
	t.Parallel()
	testClassShorthand(t, ClassShorthand())
}

func testClassShorthand(t *testing.T, f parser.Func[*ast.ClassShorthand]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ClassShorthand
	}{
		{
			name: "only constant",
			in:   ".foo",
			expect: &ast.ClassShorthand{
				Dot: &ast.Position{Line: 1, Col: 1},
				Names: []ast.Shorthand{
					{&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 2}}},
				},
			},
		}, {
			name: "constant and interpolation",
			in:   ".foo#{bar}",
			expect: &ast.ClassShorthand{
				Dot: &ast.Position{Line: 1, Col: 1},
				Names: []ast.Shorthand{
					{
						&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 2}},
						&ast.ShorthandInterpolation{
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
							Hash:   &ast.Position{Line: 1, Col: 5},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in:   ".md foo #{bar}",
			expect: &ast.ClassShorthand{
				Dot: &ast.Position{Line: 1, Col: 1},
				Names: []ast.Shorthand{
					{
						&ast.ShorthandText{Text: "md", Position: &ast.Position{Line: 1, Col: 2}},
					}, {
						&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 5}},
					}, {
						&ast.ShorthandInterpolation{
							LBrace: &ast.Position{Line: 1, Col: 10},
							Expression: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "bar",
										Position: &ast.Position{Line: 1, Col: 11},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 14},
							Hash:   &ast.Position{Line: 1, Col: 9},
						},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in+", 1other stuff")
			actual := testutil.AssertNoError(t, p, f)

			line, col, index := testutil.CalcEnd(1, 1, 0, c.in)
			testutil.AssertPosition(t, p, line, col, index)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestShorthand(t *testing.T) {
	t.Parallel()
	testShorthand(t, Shorthand())
}

func testShorthand(t *testing.T, f parser.Func[ast.Shorthand]) {
	testCases := []struct {
		name   string
		in     string
		expect ast.Shorthand
	}{
		{
			name: "text",
			in:   "foo",
			expect: ast.Shorthand{
				&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 1}},
			},
		}, {
			name: "interpolation",
			in:   "#{bar}",
			expect: ast.Shorthand{
				&ast.ShorthandInterpolation{
					LBrace: &ast.Position{Line: 1, Col: 2},
					Expression: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 3}},
						},
					},
					RBrace: &ast.Position{Line: 1, Col: 6},
					Hash:   &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "mix",
			in:   "foo#{bar}foobar",
			expect: ast.Shorthand{
				&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				&ast.ShorthandInterpolation{
					LBrace: &ast.Position{Line: 1, Col: 5},
					Expression: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "bar", Position: &ast.Position{Line: 1, Col: 6}},
						},
					},
					RBrace: &ast.Position{Line: 1, Col: 9},
					Hash:   &ast.Position{Line: 1, Col: 4},
				},
				&ast.ShorthandText{Text: "foobar", Position: &ast.Position{Line: 1, Col: 10}},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestShorthandNode(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, ShorthandNode(), testShorthandText)
	testutil.AssertAlsoFulfils(t, ShorthandNode(), testShorthandInterpolation)
}

func TestShorthandText(t *testing.T) {
	t.Parallel()
	testShorthandText(t, ShorthandText())
}

func testShorthandText(t *testing.T, f parser.Func[*ast.ShorthandText]) {
	t.Run("text", func(t *testing.T) {
		t.Parallel()
		expect := &ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 1}}
		actual := testutil.ParsesFully(t, "foo", f)
		assert.Equal(t, expect, actual)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		testutil.NoMatch(t, "", f)
	})
}

func TestShorthandInterpolation(t *testing.T) {
	t.Parallel()
	testShorthandInterpolation(t, ShorthandInterpolation())
}

func testShorthandInterpolation(t *testing.T, f parser.Func[*ast.ShorthandInterpolation]) {
	t.Run("interpolation", func(t *testing.T) {
		t.Parallel()
		expect := &ast.ShorthandInterpolation{
			LBrace: &ast.Position{Line: 1, Col: 2},
			Expression: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 3}},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 6},
			Hash:   &ast.Position{Line: 1, Col: 1},
		}

		actual := testutil.ParsesFully(t, "#{foo}", f)
		assert.Equal(t, expect, actual)
	})

	t.Run("missing expression", func(t *testing.T) {
		t.Parallel()
		testutil.NoMatch(t, "", f)
	})
}
