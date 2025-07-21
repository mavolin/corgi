package argument

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestArgument(t *testing.T) {
	t.Parallel()
	// can't test attribute because it's in a different package
	parsetest.AssertAlsoFulfils(t, Argument(), testComponentArgument)
}

func TestArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.Arguments
	}{
		{
			name: "empty",
			in:   "()",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				RParen: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "single class shorthand",
			in:   "(.foo)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.ClassShorthand{
						Dot: &ast.Position{Line: 1, Col: 2},
						Names: []ast.Shorthand{
							{
								&ast.ShorthandText{
									Text:     "foo",
									Position: &ast.Position{Line: 1, Col: 3},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 6},
			},
		}, {
			name: "single id shorthand",
			in:   "(#foo)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.IDShorthand{
						Hash: &ast.Position{Line: 1, Col: 2},
						ID: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 3},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 6},
			},
		}, {
			name: "single and placeholder",
			in:   "(&)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.AndPlaceholder{
						And: &ast.Position{Line: 1, Col: 2},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 3},
			},
		}, {
			name: "single named boolean attribute",
			in:   "(disabled)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.NamedAttribute{
						Name: &ast.AttributeReference{
							Name: &ast.AttributeName{
								Name:     "disabled",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "single named value attribute",
			in:   "(class=foo)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.NamedAttribute{
						Name: &ast.AttributeReference{
							Name: &ast.AttributeName{
								Name:     "class",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
						EqualSign: &ast.Position{Line: 1, Col: 7},
						Value: &ast.ExpressionAttributeValue{
							Nodes: ast.Code{
								&ast.GoCode{
									Code:     "foo",
									Position: &ast.Position{Line: 1, Col: 8},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 11},
			},
		}, {
			name: "single component argument",
			in:   "(foo: bar)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.ComponentArgument{
						Name: &ast.Identifier{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 2},
						},
						Colon: &ast.Position{Line: 1, Col: 5},
						Value: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{
									Code:     "bar",
									Position: &ast.Position{Line: 1, Col: 7},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "mix",
			in:   "(arg1: arg1Value, arg2: arg2Value, &, .class1 class2, #id, booleanAttr, valueAttr=valueAttrValue)",
			want: &ast.Arguments{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.ComponentArgument{
						Name: &ast.Identifier{
							Name:     "arg1",
							Position: &ast.Position{Line: 1, Col: 2},
						},
						Colon: &ast.Position{Line: 1, Col: 6},
						Value: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{
									Code:     "arg1Value",
									Position: &ast.Position{Line: 1, Col: 8},
								},
							},
						},
					},
					&ast.ComponentArgument{
						Name: &ast.Identifier{
							Name:     "arg2",
							Position: &ast.Position{Line: 1, Col: 19},
						},
						Colon: &ast.Position{Line: 1, Col: 23},
						Value: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{
									Code:     "arg2Value",
									Position: &ast.Position{Line: 1, Col: 25},
								},
							},
						},
					},
					&ast.AndPlaceholder{
						And: &ast.Position{Line: 1, Col: 36},
					},
					&ast.ClassShorthand{
						Dot: &ast.Position{Line: 1, Col: 39},
						Names: []ast.Shorthand{
							{
								&ast.ShorthandText{
									Text:     "class1",
									Position: &ast.Position{Line: 1, Col: 40},
								},
							}, {
								&ast.ShorthandText{
									Text:     "class2",
									Position: &ast.Position{Line: 1, Col: 47},
								},
							},
						},
					},
					&ast.IDShorthand{
						ID: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "id",
								Position: &ast.Position{Line: 1, Col: 56},
							},
						},
						Hash: &ast.Position{Line: 1, Col: 55},
					},
					&ast.NamedAttribute{
						Name: &ast.AttributeReference{
							Name: &ast.AttributeName{
								Name:     "booleanAttr",
								Position: &ast.Position{Line: 1, Col: 60},
							},
						},
					},
					&ast.NamedAttribute{
						Name: &ast.AttributeReference{
							Name: &ast.AttributeName{
								Name:     "valueAttr",
								Position: &ast.Position{Line: 1, Col: 73},
							},
						},
						EqualSign: &ast.Position{Line: 1, Col: 82},
						Value: &ast.ExpressionAttributeValue{
							Nodes: ast.Code{
								&ast.GoCode{
									Code:     "valueAttrValue",
									Position: &ast.Position{Line: 1, Col: 83},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 97},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in+"other")
			got := parsetest.AssertNoError(t, p, Arguments())
			if should.Equal(t, c.want, got) {
				parsetest.AssertPosition(t, p, c.want.End().Line, c.want.End().Col, len(c.in))
			}
		})
	}
}
