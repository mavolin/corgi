package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestIf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.If
	}{
		{
			name: "only if",
			in: "if i < 10 {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.If{
				If: &ast.Position{Line: 1, Col: 1},
				Header: &ast.IfHeader{
					Condition: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
						},
					},
				},
				Then: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 11},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 3, Col: 1},
				},
			},
		}, {
			name: "if else",
			in: "if i < 10 {\n" +
				"\tbr\n" +
				"} else {\n" +
				"\tdiv\n" +
				"}",
			expect: &ast.If{
				If: &ast.Position{Line: 1, Col: 1},
				Header: &ast.IfHeader{
					Condition: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
						},
					},
				},
				Then: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 11},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 3, Col: 1},
				},
				Else: &ast.Else{
					Else: &ast.Position{Line: 3, Col: 3},
					Then: &ast.Scope{
						LBrace: &ast.Position{Line: 3, Col: 8},
						Nodes: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementName{
										Name:     "div",
										Position: &ast.Position{Line: 4, Col: 2},
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
			expect: &ast.If{
				If: &ast.Position{Line: 1, Col: 1},
				Header: &ast.IfHeader{
					Condition: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 4}},
						},
					},
				},
				Then: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 11},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 3, Col: 1},
				},
				ElseIfs: []*ast.ElseIf{
					{
						Else: &ast.Position{Line: 3, Col: 3},
						If:   &ast.Position{Line: 3, Col: 8},
						Header: &ast.IfHeader{
							Condition: &ast.Expression{
								Code: ast.Code{
									&ast.GoCode{Code: "i < 20", Position: &ast.Position{Line: 3, Col: 11}},
								},
							},
						},
						Then: &ast.Scope{
							LBrace: &ast.Position{Line: 3, Col: 18},
							Nodes: []ast.ScopeNode{
								&ast.Element{
									Header: &ast.ElementHeader{
										Name: &ast.ElementName{
											Name:     "div",
											Position: &ast.Position{Line: 4, Col: 2},
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
									Name: &ast.ElementName{
										Name:     "span",
										Position: &ast.Position{Line: 6, Col: 2},
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, If())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestElseIf(t *testing.T) {
	t.Parallel()

	in := "else if i < 10 {\n" +
		"\tbr\n" +
		"}"
	expect := &ast.ElseIf{
		Else: &ast.Position{Line: 1, Col: 1},
		If:   &ast.Position{Line: 1, Col: 6},
		Header: &ast.IfHeader{
			Condition: &ast.Expression{
				Code: ast.Code{
					&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 9}},
				},
			},
		},
		Then: &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 16},
			Nodes: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 2, Col: 2},
						},
					},
				},
			},
			RBrace: &ast.Position{Line: 3, Col: 1},
		},
	}

	actual := parsesCodeNodeFully(t, in, ElseIf())
	assert.Equal(t, expect, actual)
}

func TestElse(t *testing.T) {
	t.Parallel()

	in := "else {\n" +
		"\tbr\n" +
		"}"
	expect := &ast.Else{
		Else: &ast.Position{Line: 1, Col: 1},
		Then: &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 6},
			Nodes: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 2, Col: 2},
						},
					},
				},
			},
			RBrace: &ast.Position{Line: 3, Col: 1},
		},
	}

	actual := parsesCodeNodeFully(t, in, Else())
	assert.Equal(t, expect, actual)
}

func TestIfHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.IfHeader
	}{
		{
			name: "condition",
			in:   "i < 10",
			expect: &ast.IfHeader{
				Condition: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
			},
		}, {
			name: "with statement",
			in:   "i := 0; i < 10",
			expect: &ast.IfHeader{
				Statement: &ast.SimpleStatement{
					Parsed: &ast.ShortVarDeclaration{
						Names: []*ast.Ident{
							{Ident: "i", Position: &ast.Position{Line: 1, Col: 1}},
						},
						ColonEqualSign: &ast.Position{Line: 1, Col: 3},
						Values: []*ast.Expression{
							{
								Code: ast.Code{
									&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
								},
							},
						},
					},
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
						&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 3}},
						&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
				Condition: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 9}},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in+" { 1other stuff }")
			actual := testutil.AssertNoError(t, p, IfHeader())

			line, col, index := testutil.CalcEnd(1, 1, 0, c.in)
			testutil.AssertPosition(t, p, line, col, index)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestSwitch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.Switch
	}{
		{
			name: "comparator",
			in: "switch i {\n" +
				"case 1:\n" +
				"\tbr\n" +
				"}",
			expect: &ast.Switch{
				Switch: &ast.Position{Line: 1, Col: 1},
				Comparator: &ast.SimpleStatement{
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 8}},
					},
				},
				LBrace: &ast.Position{Line: 1, Col: 10},
				Cases: []*ast.Case{
					{
						Case: &ast.Position{Line: 2, Col: 1},
						Expression: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "1", Position: &ast.Position{Line: 2, Col: 6}},
							},
						},
						Colon: &ast.Position{Line: 2, Col: 7},
						Then: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementName{
										Name:     "br",
										Position: &ast.Position{Line: 3, Col: 2},
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
			expect: &ast.Switch{
				Switch: &ast.Position{Line: 1, Col: 1},
				LBrace: &ast.Position{Line: 1, Col: 8},
				Cases: []*ast.Case{
					{
						Case: &ast.Position{Line: 2, Col: 1},
						Expression: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "1", Position: &ast.Position{Line: 2, Col: 6}},
							},
						},
						Colon: &ast.Position{Line: 2, Col: 7},
						Then: []ast.ScopeNode{
							&ast.Element{
								Header: &ast.ElementHeader{
									Name: &ast.ElementName{
										Name:     "br",
										Position: &ast.Position{Line: 3, Col: 2},
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, Switch())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestSwitchCase(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, SwitchCase(), testCase)
	testutil.AssertAlsoFulfils(t, SwitchCase(), testDefault)
}

func TestCase(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, Case(), testCase)
}

func testCase(t *testing.T, f parser.Func[*ast.Case]) {
	in := "case 1:\n" +
		"\tbr"
	expect := &ast.Case{
		Case: &ast.Position{Line: 1, Col: 1},
		Expression: &ast.Expression{
			Code: ast.Code{
				&ast.GoCode{Code: "1", Position: &ast.Position{Line: 1, Col: 6}},
			},
		},
		Colon: &ast.Position{Line: 1, Col: 7},
		Then: []ast.ScopeNode{
			&ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementName{
						Name:     "br",
						Position: &ast.Position{Line: 2, Col: 2},
					},
				},
			},
		},
	}

	actual := parsesSwitchCaseFully(t, in, "", f)
	assert.Equal(t, expect, actual)
}

func TestDefault(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, Default(), testDefault)
}

func testDefault(t *testing.T, f parser.Func[*ast.Case]) {
	in := "default:\n" +
		"\tbr"
	expect := &ast.Case{
		Default: &ast.Position{Line: 1, Col: 1},
		Colon:   &ast.Position{Line: 1, Col: 8},
		Then: []ast.ScopeNode{
			&ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementName{
						Name:     "br",
						Position: &ast.Position{Line: 2, Col: 2},
					},
				},
			},
		},
	}

	actual := parsesSwitchCaseFully(t, in, "", f)
	assert.Equal(t, expect, actual)
}

func TestCaseBody(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		suffix string
		expect []ast.ScopeNode
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
			expect: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 1, Col: 1},
						},
					},
				},
			},
		}, {
			name:   "default",
			in:     "br",
			suffix: "\ndefault:",
			expect: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 1, Col: 1},
						},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesSwitchCaseFully(t, c.in, c.suffix, CaseBody())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func parsesSwitchCaseFully[T any](t *testing.T, in string, suffix string, f parser.Func[T]) T {
	p := testutil.NewParser(t, in+suffix+"} 1other stuff")
	actual := testutil.AssertNoError(t, p, f)

	line, col, index := testutil.CalcEnd(1, 1, 0, in)
	testutil.AssertPosition(t, p, line, col, index)
	return actual
}

func TestFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.For
	}{
		{
			name: "infinite",
			in: "for {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 5},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
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
			expect: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ForConditionHeader{
					Condition: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 5}},
						},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 12},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
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
			expect: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ForClauseHeader{
					Init: &ast.SimpleStatement{
						Parsed: &ast.ShortVarDeclaration{
							Names: []*ast.Ident{
								{Ident: "i", Position: &ast.Position{Line: 1, Col: 5}},
							},
							ColonEqualSign: &ast.Position{Line: 1, Col: 7},
							Values: []*ast.Expression{
								{
									Code: ast.Code{
										&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 10}},
									},
								},
							},
						},
						Code: ast.Code{
							&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 5}},
							&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 7}},
							&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 10}},
						},
					},
					Condition: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 13}},
						},
					},
					Post: &ast.SimpleStatement{
						Parsed: &ast.IncDec{
							Expression: &ast.Expression{
								Code: ast.Code{
									&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 21}},
								},
							},
							IncrPos: &ast.Position{Line: 1, Col: 22},
						},
						Code: ast.Code{
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
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
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
			expect: &ast.For{
				For: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ForRangeHeader{
					Var1: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 5}},
						},
					},
					Colon:     &ast.Position{Line: 1, Col: 7},
					EqualSign: &ast.Position{Line: 1, Col: 8},
					Range:     &ast.Position{Line: 1, Col: 10},
					Expression: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 16}},
						},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 18},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementName{
									Name:     "br",
									Position: &ast.Position{Line: 2, Col: 2},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 3, Col: 1},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, For())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestForHeader(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, ForHeader(), testForConditionHeader)
	testutil.AssertAlsoFulfils(t, ForHeader(), testForClauseHeader)
	testutil.AssertAlsoFulfils(t, ForHeader(), testForRangeHeader)
}

func TestForConditionHeader(t *testing.T) {
	t.Parallel()
	testForConditionHeader(t, ForConditionHeader())
}

func testForConditionHeader(t *testing.T, f parser.Func[*ast.ForConditionHeader]) {
	in := "i < 10"
	expect := &ast.ForConditionHeader{
		Condition: &ast.Expression{
			Code: ast.Code{
				&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 1}},
			},
		},
	}

	p := testutil.NewParser(t, in+" { 1other stuff }")
	actual := testutil.AssertNoError(t, p, f)

	line, col, index := testutil.CalcEnd(1, 1, 0, in)
	testutil.AssertPosition(t, p, line, col, index)
	assert.Equal(t, expect, actual)
}

func TestForClauseHeader(t *testing.T) {
	t.Parallel()
	testForClauseHeader(t, ForClauseHeader())
}

func testForClauseHeader(t *testing.T, f parser.Func[*ast.ForClauseHeader]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ForClauseHeader
	}{
		{
			name:   "empty",
			in:     ";;",
			expect: &ast.ForClauseHeader{},
		}, {
			name: "with init",
			in:   "i := 0;;",
			expect: &ast.ForClauseHeader{
				Init: &ast.SimpleStatement{
					Parsed: &ast.ShortVarDeclaration{
						Names: []*ast.Ident{
							{Ident: "i", Position: &ast.Position{Line: 1, Col: 1}},
						},
						ColonEqualSign: &ast.Position{Line: 1, Col: 3},
						Values: []*ast.Expression{
							{
								Code: ast.Code{
									&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
								},
							},
						},
					},
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
						&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 3}},
						&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
			},
		}, {
			name: "with condition",
			in:   "; i < 10;",
			expect: &ast.ForClauseHeader{
				Condition: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 3}},
					},
				},
			},
		}, {
			name: "with post",
			in:   ";; i++",
			expect: &ast.ForClauseHeader{
				Post: &ast.SimpleStatement{
					Parsed: &ast.IncDec{
						Expression: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 4}},
							},
						},
						IncrPos: &ast.Position{Line: 1, Col: 5},
					},
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 4}},
						&ast.GoCode{Code: "++", Position: &ast.Position{Line: 1, Col: 5}},
					},
				},
			},
		}, {
			name: "full",
			in:   "i := 0; i < 10; i++",
			expect: &ast.ForClauseHeader{
				Init: &ast.SimpleStatement{
					Parsed: &ast.ShortVarDeclaration{
						Names: []*ast.Ident{
							{Ident: "i", Position: &ast.Position{Line: 1, Col: 1}},
						},
						ColonEqualSign: &ast.Position{Line: 1, Col: 3},
						Values: []*ast.Expression{
							{
								Code: ast.Code{
									&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
								},
							},
						},
					},
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
						&ast.GoCode{Code: ":=", Position: &ast.Position{Line: 1, Col: 3}},
						&ast.GoCode{Code: "0", Position: &ast.Position{Line: 1, Col: 6}},
					},
				},
				Condition: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i < 10", Position: &ast.Position{Line: 1, Col: 9}},
					},
				},
				Post: &ast.SimpleStatement{
					Parsed: &ast.IncDec{
						Expression: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 17}},
							},
						},
						IncrPos: &ast.Position{Line: 1, Col: 18},
					},
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 17}},
						&ast.GoCode{Code: "++", Position: &ast.Position{Line: 1, Col: 18}},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestForRangeHeader(t *testing.T) {
	t.Parallel()
	testForRangeHeader(t, ForRangeHeader())
}

func testForRangeHeader(t *testing.T, f parser.Func[*ast.ForRangeHeader]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ForRangeHeader
	}{
		{
			name: "basic",
			in:   "range s",
			expect: &ast.ForRangeHeader{
				Range: &ast.Position{Line: 1, Col: 1},
				Expression: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 7}},
					},
				},
			},
		}, {
			name: "ordered",
			in:   "ordered range s",
			expect: &ast.ForRangeHeader{
				Ordered: &ast.Position{Line: 1, Col: 1},
				Range:   &ast.Position{Line: 1, Col: 9},
				Expression: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 15}},
					},
				},
			},
		}, {
			name: "with index",
			in:   "i = range s",
			expect: &ast.ForRangeHeader{
				Var1: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 3},
				Range:     &ast.Position{Line: 1, Col: 5},
				Expression: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 11}},
					},
				},
			},
		}, {
			name: "with index and value",
			in:   "i, v = range s",
			expect: &ast.ForRangeHeader{
				Var1: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				Var2: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "v", Position: &ast.Position{Line: 1, Col: 4}},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 6},
				Range:     &ast.Position{Line: 1, Col: 8},
				Expression: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 14}},
					},
				},
			},
		}, {
			name: "declares",
			in:   "i, v := range s",
			expect: &ast.ForRangeHeader{
				Var1: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "i", Position: &ast.Position{Line: 1, Col: 1}},
					},
				},
				Var2: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "v", Position: &ast.Position{Line: 1, Col: 4}},
					},
				},
				Colon:     &ast.Position{Line: 1, Col: 6},
				EqualSign: &ast.Position{Line: 1, Col: 7},
				Range:     &ast.Position{Line: 1, Col: 9},
				Expression: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "s", Position: &ast.Position{Line: 1, Col: 15}},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := parsesCodeNodeFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}
