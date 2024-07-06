package argument

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestArgument(t *testing.T) {
	t.Parallel()
	// can't test attribute because it's in a different package
	testutil.AssertAlsoFulfils(t, Argument(), testComponentArgument)
}

func TestArguments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.Arguments
	}{
		{
			name: "empty",
			in:   "()",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args:   []ast.Argument{},
				RParen: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "single class shorthand",
			in:   "(.foo)",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.ClassShorthand{
						Name: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "foo",
								Position: ast.Position{Line: 1, Col: 3},
							},
						},
						Position: ast.Position{Line: 1, Col: 2},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 6},
			},
		}, {
			name: "single id shorthand",
			in:   "(#foo)",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.IDShorthand{
						ID: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "foo",
								Position: ast.Position{Line: 1, Col: 3},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 6},
			},
		}, {
			name: "single and placeholder",
			in:   "(&)",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.AndPlaceholder{
						Position: ast.Position{Line: 1, Col: 2},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 4},
			},
		}, {
			name: "single named boolean attribute",
			in:   "(disabled)",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.NamedAttribute{
						Name: ast.AttributeName{
							Name:     "disabled",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "single named value attribute",
			in:   "(class=foo)",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.NamedAttribute{
						Name: ast.AttributeName{
							Name:     "class",
							Position: ast.Position{Line: 1, Col: 2},
						},
						Assign: &ast.Position{Line: 1, Col: 7},
						Value: &ast.ExpressionAttributeValue{
							Nodes: []ast.ExpressionNode{
								&ast.GoCode{
									Code:     "foo",
									Position: ast.Position{Line: 1, Col: 8},
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
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.ComponentArgument{
						Name: &ast.Ident{
							Ident:    "foo",
							Position: ast.Position{Line: 1, Col: 2},
						},
						Colon: &ast.Position{Line: 1, Col: 5},
						Value: &ast.Expression{
							Nodes: []ast.ExpressionNode{
								&ast.GoCode{
									Code:     "bar",
									Position: ast.Position{Line: 1, Col: 7},
								},
							},
						},
						Position: ast.Position{Line: 1, Col: 2},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 11},
			},
		}, {
			name: "mix",
			in:   "(arg1: arg1Value, arg2: arg2Value, &, .class1, .class2, #id, booleanAttr, valueAttr=valueAttrValue)",
			expect: &ast.Arguments{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []ast.Argument{
					&ast.ComponentArgument{
						Name: &ast.Ident{
							Ident:    "arg1",
							Position: ast.Position{Line: 1, Col: 2},
						},
						Colon: &ast.Position{Line: 1, Col: 6},
						Value: &ast.Expression{
							Nodes: []ast.ExpressionNode{
								&ast.GoCode{
									Code:     "arg1Value",
									Position: ast.Position{Line: 1, Col: 8},
								},
							},
						},
						Position: ast.Position{Line: 1, Col: 2},
					},
					&ast.ComponentArgument{
						Name: &ast.Ident{
							Ident:    "arg2",
							Position: ast.Position{Line: 1, Col: 19},
						},
						Colon: &ast.Position{Line: 1, Col: 23},
						Value: &ast.Expression{
							Nodes: []ast.ExpressionNode{
								&ast.GoCode{
									Code:     "arg2Value",
									Position: ast.Position{Line: 1, Col: 25},
								},
							},
						},
						Position: ast.Position{Line: 1, Col: 19},
					},
					&ast.AndPlaceholder{
						Position: ast.Position{Line: 1, Col: 36},
					},
					&ast.ClassShorthand{
						Name: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "class1",
								Position: ast.Position{Line: 1, Col: 40},
							},
						},
						Position: ast.Position{Line: 1, Col: 39},
					},
					&ast.ClassShorthand{
						Name: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "class2",
								Position: ast.Position{Line: 1, Col: 49},
							},
						},
						Position: ast.Position{Line: 1, Col: 48},
					},
					&ast.IDShorthand{
						ID: ast.Shorthand{
							&ast.ShorthandText{
								Text:     "id",
								Position: ast.Position{Line: 1, Col: 58},
							},
						},
						Position: ast.Position{Line: 1, Col: 57},
					},
					&ast.NamedAttribute{
						Name: ast.AttributeName{
							Name:     "booleanAttr",
							Position: ast.Position{Line: 1, Col: 62},
						},
					},
					&ast.NamedAttribute{
						Name: ast.AttributeName{
							Name:     "valueAttr",
							Position: ast.Position{Line: 1, Col: 75},
						},
						Assign: &ast.Position{Line: 1, Col: 83},
						Value: &ast.ExpressionAttributeValue{
							Nodes: []ast.ExpressionNode{
								&ast.GoCode{
									Code:     "valueAttrValue",
									Position: ast.Position{Line: 1, Col: 84},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 99},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in+"other")
			actual := testutil.AssertNoError(t, p, Arguments())
			if assert.Equal(t, c.expect, actual) {
				testutil.AssertPosition(t, p, c.expect.End().Line, c.expect.End().Col, len(c.in))
			}
		})
	}
}
