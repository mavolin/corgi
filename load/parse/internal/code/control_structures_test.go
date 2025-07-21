package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestConditional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.Conditional
	}{
		{
			name: "only if",
			in: "if i < 10 {\n" +
				"\tbr\n" +
				"}",
			want: &ast.Conditional{
				If: &ast.If{
					If: &ast.Position{Line: 1, Col: 1},
					Header: &ast.IfHeader{
						Condition: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
							},
						},
					},
					Then: &ast.Scope{
						LBrace: &ast.Position{Line: 1, Col: 11},
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
			},
		}, {
			name: "if else",
			in: "if i < 10 {\n" +
				"\tbr\n" +
				"} else {\n" +
				"\tdiv\n" +
				"}",
			want: &ast.Conditional{
				If: &ast.If{
					If: &ast.Position{Line: 1, Col: 1},
					Header: &ast.IfHeader{
						Condition: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
							},
						},
					},
					Then: &ast.Scope{
						LBrace: &ast.Position{Line: 1, Col: 11},
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
				Else: &ast.Else{
					Else: &ast.Position{Line: 3, Col: 3},
					Then: &ast.Scope{
						LBrace: &ast.Position{Line: 3, Col: 8},
						Nodes: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementReference{
										Name: &ast.ElementName{
											Name:     "div",
											Position: &ast.Position{Line: 4, Col: 2},
										},
									},
								},
							},
						},
						RBrace: &ast.Position{Line: 5, Col: 1},
					},
				},
			},
		}, {
			name: "if else if else",
			in: "if i < 10 {\n" +
				"\tbr\n" +
				"} else if i < 20 {\n" +
				"\tdiv\n" +
				"} else {\n" +
				"\tspan\n" +
				"}",
			want: &ast.Conditional{
				If: &ast.If{
					If: &ast.Position{Line: 1, Col: 1},
					Header: &ast.IfHeader{
						Condition: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
							},
						},
					},
					Then: &ast.Scope{
						LBrace: &ast.Position{Line: 1, Col: 11},
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
				ElseIfs: []*ast.ElseIf{
					{
						Else: &ast.Position{Line: 3, Col: 3},
						If:   &ast.Position{Line: 3, Col: 8},
						Header: &ast.IfHeader{
							Condition: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{Code: "i < 20", Position: &ast.Position{Line: 3, Col: 11}},
								},
							},
						},
						Then: &ast.Scope{
							LBrace: &ast.Position{Line: 3, Col: 18},
							Nodes: []ast.ScopeNode{
								&ast.Element{
									Header: &ast.ElementHeader{
										Name: &ast.ElementReference{
											Name: &ast.ElementName{
												Name:     "div",
												Position: &ast.Position{Line: 4, Col: 2},
											},
										},
									},
								},
							},
							RBrace: &ast.Position{Line: 5, Col: 1},
						},
					},
				},
				Else: &ast.Else{
					Else: &ast.Position{Line: 5, Col: 3},
					Then: &ast.Scope{
						LBrace: &ast.Position{Line: 5, Col: 8},
						Nodes: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementReference{
										Name: &ast.ElementName{
											Name:     "span",
											Position: &ast.Position{Line: 6, Col: 2},
										},
									},
								},
							},
						},
						RBrace: &ast.Position{Line: 7, Col: 1},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsesCodeNodeFully(t, c.in, Conditional())
			should.Equal(t, c.want, got)
		})
	}
}

func TestIf(t *testing.T) {
	t.Parallel()

	in := "if i < 10 {\n" +
		"\tbr\n" +
		"}"
	want := &ast.If{
		If: &ast.Position{Line: 1, Col: 1},
		Header: &ast.IfHeader{
			Condition: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
				},
			},
		},
		Then: &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 11},
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
	}

	got := parsesCodeNodeFully(t, in, If())
	should.Equal(t, want, got)
}

func TestElseIf(t *testing.T) {
	t.Parallel()

	in := "else if i < 10 {\n" +
		"\tbr\n" +
		"}"
	want := &ast.ElseIf{
		Else: &ast.Position{Line: 1, Col: 1},
		If:   &ast.Position{Line: 1, Col: 6},
		Header: &ast.IfHeader{
			Condition: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 9}},
				},
			},
		},
		Then: &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 16},
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
	}

	got := parsesCodeNodeFully(t, in, ElseIf())
	should.Equal(t, want, got)
}

func TestElse(t *testing.T) {
	t.Parallel()

	in := "else {\n" +
		"\tbr\n" +
		"}"
	want := &ast.Else{
		Else: &ast.Position{Line: 1, Col: 1},
		Then: &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 6},
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
	}

	got := parsesCodeNodeFully(t, in, Else())
	should.Equal(t, want, got)
}

func TestIfHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.IfHeader
	}{
		{
			name: "condition",
			in:   "i < 10",
			want: &ast.IfHeader{
				Condition: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
			},
		}, {
			name: "with statement",
			in:   "i := 0; i < 10",
			want: &ast.IfHeader{
				Statement: &ast.SimpleStatement{
					Parsed: &ast.ShortVarDeclaration{
						Names: []*ast.Identifier{
							{Name: "i", Position: &ast.Position{Line: 1, Col: 1}},
						},
						ColonEqualSign: &ast.Position{Line: 1, Col: 3},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
								},
							},
						},
					},
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
						&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 3}},
						&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
				Condition: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 9}},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in+" { 1other stuff }")
			got := parsetest.AssertNoError(t, p, IfHeader())

			line, col, index := parsetest.CalcEnd(1, 1, 0, c.in)
			parsetest.AssertPosition(t, p, line, col, index)
			should.Equal(t, c.want, got)
		})
	}
}

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
			should.Equal(t, c.want, got)
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
	should.Equal(t, want, got)
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
	should.Equal(t, want, got)
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
		}, {
			name:   "empty with case",
			suffix: "case 1:",
		}, {
			name:   "empty with default",
			suffix: "default:",
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
			should.Equal(t, c.want, got)
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

func TestFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.For
	}{
		{
			name: "infinite",
			in: "for {\n" +
				"\tbr\n" +
				"}",
			want: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 5},
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
		}, {
			name: "condition",
			in: "for i < 10 {\n" +
				"\tbr\n" +
				"}",
			want: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ForConditionHeader{
					Condition: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 5}},
						},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 12},
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
		}, {
			name: "clause",
			in: "for i := 0; i < 10; i++ {\n" +
				"\tbr\n" +
				"}",
			want: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ForClauseHeader{
					Init: &ast.SimpleStatement{
						Parsed: &ast.ShortVarDeclaration{
							Names: []*ast.Identifier{
								{Name: "i", Position: &ast.Position{Line: 1, Col: 5}},
							},
							ColonEqualSign: &ast.Position{Line: 1, Col: 7},
							Values: []*ast.Expression{
								{
									Nodes: ast.Code{
										&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 10}},
									},
								},
							},
						},
						Nodes: ast.Code{
							&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 5}},
							&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 7}},
							&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 10}},
						},
					},
					Condition: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 13}},
						},
					},
					Post: &ast.SimpleStatement{
						Parsed: &ast.IncDec{
							Expression: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 21}},
								},
							},
							IncrPos: &ast.Position{Line: 1, Col: 22},
						},
						Nodes: ast.Code{
							&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 21}},
							&ast.GoCode{Code: "++", Position: &ast.Position{Line: 1, Col: 22}},
						},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 25},
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
		}, {
			name: "range",
			in: "for i := range s {\n" +
				"\tbr\n" +
				"}",
			want: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ForRangeHeader{
					Var1: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 5}},
						},
					},
					Colon:     &ast.Position{Line: 1, Col: 7},
					EqualSign: &ast.Position{Line: 1, Col: 8},
					Range:     &ast.Position{Line: 1, Col: 10},
					Expression: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 16}},
						},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 18},
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
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsesCodeNodeFully(t, c.in, For())
			should.Equal(t, c.want, got)
		})
	}
}

func TestForHeader(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, ForHeader(), testForConditionHeader)
	parsetest.AssertAlsoFulfils(t, ForHeader(), testForClauseHeader)
	parsetest.AssertAlsoFulfils(t, ForHeader(), testForRangeHeader)
}

func TestForConditionHeader(t *testing.T) {
	t.Parallel()
	testForConditionHeader(t, ForConditionHeader())
}

func testForConditionHeader(t *testing.T, f parser.Func[*ast.ForConditionHeader]) {
	in := "i < 10"
	want := &ast.ForConditionHeader{
		Condition: &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 1}},
			},
		},
	}

	p := parsetest.NewParser(t, in+" { 1other stuff }")
	got := parsetest.AssertNoError(t, p, f)

	line, col, index := parsetest.CalcEnd(1, 1, 0, in)
	parsetest.AssertPosition(t, p, line, col, index)
	should.Equal(t, want, got)
}

func TestForClauseHeader(t *testing.T) {
	t.Parallel()
	testForClauseHeader(t, ForClauseHeader())
}

func testForClauseHeader(t *testing.T, f parser.Func[*ast.ForClauseHeader]) {
	tests := []struct {
		name string
		in   string
		want *ast.ForClauseHeader
	}{
		{
			name: "empty",
			in:   ";;",
			want: &ast.ForClauseHeader{},
		}, {
			name: "with init",
			in:   "i := 0;;",
			want: &ast.ForClauseHeader{
				Init: &ast.SimpleStatement{
					Parsed: &ast.ShortVarDeclaration{
						Names: []*ast.Identifier{
							{Name: "i", Position: &ast.Position{Line: 1, Col: 1}},
						},
						ColonEqualSign: &ast.Position{Line: 1, Col: 3},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
								},
							},
						},
					},
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
						&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 3}},
						&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
			},
		}, {
			name: "with condition",
			in:   "; i < 10;",
			want: &ast.ForClauseHeader{
				Condition: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 3}},
					},
				},
			},
		}, {
			name: "with post",
			in:   ";; i++",
			want: &ast.ForClauseHeader{
				Post: &ast.SimpleStatement{
					Parsed: &ast.IncDec{
						Expression: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 4}},
							},
						},
						IncrPos: &ast.Position{Line: 1, Col: 5},
					},
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 4}},
						&ast.GoCode{Code: "++", Position: &ast.Position{Line: 1, Col: 5}},
					},
				},
			},
		}, {
			name: "full",
			in:   "i := 0; i < 10; i++",
			want: &ast.ForClauseHeader{
				Init: &ast.SimpleStatement{
					Parsed: &ast.ShortVarDeclaration{
						Names: []*ast.Identifier{
							{Name: "i", Position: &ast.Position{Line: 1, Col: 1}},
						},
						ColonEqualSign: &ast.Position{Line: 1, Col: 3},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
								},
							},
						},
					},
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
						&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 3}},
						&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
				Condition: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 9}},
					},
				},
				Post: &ast.SimpleStatement{
					Parsed: &ast.IncDec{
						Expression: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 17}},
							},
						},
						IncrPos: &ast.Position{Line: 1, Col: 18},
					},
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 17}},
						&ast.GoCode{Code: "++", Position: &ast.Position{Line: 1, Col: 18}},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsesCodeNodeFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}

func TestForRangeHeader(t *testing.T) {
	t.Parallel()
	testForRangeHeader(t, ForRangeHeader())
}

func testForRangeHeader(t *testing.T, f parser.Func[*ast.ForRangeHeader]) {
	tests := []struct {
		name string
		in   string
		want *ast.ForRangeHeader
	}{
		{
			name: "basic",
			in:   "range s",
			want: &ast.ForRangeHeader{
				Range: &ast.Position{Line: 1, Col: 1},
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 7}},
					},
				},
			},
		}, {
			name: "ordered",
			in:   "ordered range s",
			want: &ast.ForRangeHeader{
				Ordered: &ast.Position{Line: 1, Col: 1},
				Range:   &ast.Position{Line: 1, Col: 9},
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 15}},
					},
				},
			},
		}, {
			name: "with index",
			in:   "i = range s",
			want: &ast.ForRangeHeader{
				Var1: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 3},
				Range:     &ast.Position{Line: 1, Col: 5},
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 11}},
					},
				},
			},
		}, {
			name: "with index and value",
			in:   "i, v = range s",
			want: &ast.ForRangeHeader{
				Var1: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				Var2: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "v", Position: &ast.Position{Line: 1, Col: 4}},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 6},
				Range:     &ast.Position{Line: 1, Col: 8},
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 14}},
					},
				},
			},
		}, {
			name: "declares",
			in:   "i, v := range s",
			want: &ast.ForRangeHeader{
				Var1: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				Var2: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "v", Position: &ast.Position{Line: 1, Col: 4}},
					},
				},
				Colon:     &ast.Position{Line: 1, Col: 6},
				EqualSign: &ast.Position{Line: 1, Col: 7},
				Range:     &ast.Position{Line: 1, Col: 9},
				Expression: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 15}},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsesCodeNodeFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}
