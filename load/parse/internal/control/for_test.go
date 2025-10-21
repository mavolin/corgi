package control

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

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

			got := parsetest.ParsesUntilEOS(t, c.in, For())
			should.Equal(t, got, c.want)
		})
	}
}

func TestForHeader(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, ForHeader(), testForConditionHeader)
	parsetest.AlsoFulfils(t, ForHeader(), testForClauseHeader)
	parsetest.AlsoFulfils(t, ForHeader(), testForRangeHeader)
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

	got := parsetest.ParsesUntilBody(t, in, f)
	should.Equal(t, got, want)
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

			got := parsetest.ParsesUntilBody(t, c.in, f)
			should.Equal(t, got, c.want)
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

			got := parsetest.ParsesUntilEOS(t, c.in, f)
			should.Equal(t, got, c.want)
		})
	}
}
