package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestIDShorthand(t *testing.T) {
	t.Parallel()
	testIDShorthand(t, IDShorthand())
}

func testIDShorthand(t *testing.T, f parser.Func[*ast.IDShorthand]) {
	want := &ast.IDShorthand{
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

	got := parsetest.ParsesFully(t, "#foo#{bar}", f)
	should.Equal(t, want, got)
}

func TestClassShorthand(t *testing.T) {
	t.Parallel()
	testClassShorthand(t, ClassShorthand())
}

func testClassShorthand(t *testing.T, f parser.Func[*ast.ClassShorthand]) {
	tests := []struct {
		name string
		in   string
		want *ast.ClassShorthand
	}{
		{
			name: "only constant",
			in:   ".foo",
			want: &ast.ClassShorthand{
				Dot: &ast.Position{Line: 1, Col: 1},
				Names: []ast.Shorthand{
					{&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 2}}},
				},
			},
		}, {
			name: "constant and interpolation",
			in:   ".foo#{bar}",
			want: &ast.ClassShorthand{
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
			want: &ast.ClassShorthand{
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in+", 1other stuff")
			got := parsetest.AssertNoError(t, p, f)

			line, col, index := parsetest.CalcEnd(1, 1, 0, c.in)
			parsetest.AssertPosition(t, p, line, col, index)
			should.Equal(t, c.want, got)
		})
	}
}

func TestShorthand(t *testing.T) {
	t.Parallel()
	testShorthand(t, Shorthand())
}

func testShorthand(t *testing.T, f parser.Func[ast.Shorthand]) {
	tests := []struct {
		name string
		in   string
		want ast.Shorthand
	}{
		{
			name: "text",
			in:   "foo",
			want: ast.Shorthand{
				&ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 1}},
			},
		}, {
			name: "interpolation",
			in:   "#{bar}",
			want: ast.Shorthand{
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
			want: ast.Shorthand{
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}

func TestShorthandNode(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, ShorthandNode(), testShorthandText)
	parsetest.AssertAlsoFulfils(t, ShorthandNode(), testShorthandInterpolation)
}

func TestShorthandText(t *testing.T) {
	t.Parallel()
	testShorthandText(t, ShorthandText())
}

func testShorthandText(t *testing.T, f parser.Func[*ast.ShorthandText]) {
	t.Run("text", func(t *testing.T) {
		t.Parallel()
		want := &ast.ShorthandText{Text: "foo", Position: &ast.Position{Line: 1, Col: 1}}
		got := parsetest.ParsesFully(t, "foo", f)
		should.Equal(t, want, got)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		parsetest.NoMatch(t, "", f)
	})
}

func TestShorthandInterpolation(t *testing.T) {
	t.Parallel()
	testShorthandInterpolation(t, ShorthandInterpolation())
}

func testShorthandInterpolation(t *testing.T, f parser.Func[*ast.ShorthandInterpolation]) {
	t.Run("interpolation", func(t *testing.T) {
		t.Parallel()
		want := &ast.ShorthandInterpolation{
			LBrace: &ast.Position{Line: 1, Col: 2},
			Expression: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{Code: "foo", Position: &ast.Position{Line: 1, Col: 3}},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 6},
			Hash:   &ast.Position{Line: 1, Col: 1},
		}

		got := parsetest.ParsesFully(t, "#{foo}", f)
		should.Equal(t, want, got)
	})

	t.Run("missing expression", func(t *testing.T) {
		t.Parallel()
		parsetest.NoMatch(t, "", f)
	})
}
