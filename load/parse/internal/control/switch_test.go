package control

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestSwitch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.Switch
	}{
		{
			name: "comparator",
			in: "switch i {\n" +
				"case 1:\n" +
				"\tbr\n" +
				"}",
			want: &ast.Switch{
				Switch: &ast.Position{Line: 1, Col: 1},
				Comparator: &ast.SimpleStatement{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 8}},
					},
				},
				LBrace: &ast.Position{Line: 1, Col: 10},
				Cases: []*ast.Case{
					{
						Case: &ast.Position{Line: 2, Col: 1},
						Expression: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "1", Position: &ast.Position{Line: 2, Col: 6}},
							},
						},
						Colon: &ast.Position{Line: 2, Col: 7},
						Then: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementReference{
										Name: &ast.ElementName{
											Name:     "br",
											Position: &ast.Position{Line: 3, Col: 2},
										},
									},
								},
							},
						},
					},
				},
				RBrace: &ast.Position{Line: 4, Col: 1},
			},
		}, {
			name: "no comparator",
			in: "switch {\n" +
				"case 1:\n" +
				"\tbr\n" +
				"}",
			want: &ast.Switch{
				Switch: &ast.Position{Line: 1, Col: 1},
				LBrace: &ast.Position{Line: 1, Col: 8},
				Cases: []*ast.Case{
					{
						Case: &ast.Position{Line: 2, Col: 1},
						Expression: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "1", Position: &ast.Position{Line: 2, Col: 6}},
							},
						},
						Colon: &ast.Position{Line: 2, Col: 7},
						Then: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementReference{
										Name: &ast.ElementName{
											Name:     "br",
											Position: &ast.Position{Line: 3, Col: 2},
										},
									},
								},
							},
						},
					},
				},
				RBrace: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsesCodeNodeFully(t, c.in, Switch())
			should.Equal(t, got, c.want)
		})
	}
}

func TestSwitchCase(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, SwitchCase(), testCase)
	parsetest.AssertAlsoFulfils(t, SwitchCase(), testDefault)
}

func TestCase(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, Case(), testCase)
}

func testCase(t *testing.T, f parser.Func[*ast.Case]) {
	in := "case 1:\n" +
		"\tbr"
	want := &ast.Case{
		Case: &ast.Position{Line: 1, Col: 1},
		Expression: &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{Code: "1", Position: &ast.Position{Line: 1, Col: 6}},
			},
		},
		Colon: &ast.Position{Line: 1, Col: 7},
		Then: []ast.ScopeNode{
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
	}

	got := parsesSwitchCaseFully(t, in, "", f)
	should.Equal(t, got, want)
}

func TestDefault(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, Default(), testDefault)
}

func testDefault(t *testing.T, f parser.Func[*ast.Case]) {
	in := "default:\n" +
		"\tbr"
	want := &ast.Case{
		Default: &ast.Position{Line: 1, Col: 1},
		Colon:   &ast.Position{Line: 1, Col: 8},
		Then: []ast.ScopeNode{
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
	}

	got := parsesSwitchCaseFully(t, in, "", f)
	should.Equal(t, got, want)
}

func TestCaseBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     string
		suffix string
		want   []ast.ScopeNode
	}{
		{
			name: "empty",
			in:   "",
			want: []ast.ScopeNode{},
		}, {
			name:   "empty with case",
			suffix: "case 1:",
			want:   []ast.ScopeNode{},
		}, {
			name:   "empty with default",
			suffix: "default:",
			want:   []ast.ScopeNode{},
		}, {
			name:   "case",
			in:     "br",
			suffix: "\ncase 1:",
			want: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementReference{
							Name: &ast.ElementName{
								Name:     "br",
								Position: &ast.Position{Line: 1, Col: 1},
							},
						},
					},
				},
			},
		}, {
			name:   "default",
			in:     "br",
			suffix: "\ndefault:",
			want: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementReference{
							Name: &ast.ElementName{
								Name:     "br",
								Position: &ast.Position{Line: 1, Col: 1},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsesSwitchCaseFully(t, c.in, c.suffix, CaseBody())
			should.Equal(t, got, c.want)
		})
	}
}

func parsesSwitchCaseFully[T any](t *testing.T, in string, suffix string, f parser.Func[T]) T {
	p := parsetest.NewParser(t, in+suffix+"} 1other stuff")
	got := parsetest.AssertNoError(t, p, f)

	line, col, index := parsetest.CalcEnd(1, 1, 0, in)
	parsetest.AssertPosition(t, p, line, col, index)
	return got
}
