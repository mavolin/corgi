package control

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
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

			got := parsetest.ParsesUntilEOS(t, c.in, Conditional())
			should.Equal(t, got, c.want)
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

	got := parsetest.ParsesUntilEOS(t, in, If())
	should.Equal(t, got, want)
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

	got := parsetest.ParsesUntilEOS(t, in, ElseIf())
	should.Equal(t, got, want)
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

	got := parsetest.ParsesUntilEOS(t, in, Else())
	should.Equal(t, got, want)
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

			got := parsetest.ParsesUntilBody(t, c.in, IfHeader())
			should.Equal(t, got, c.want)
		})
	}
}
