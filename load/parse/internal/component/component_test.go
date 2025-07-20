package component

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestComponent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.Component
	}{
		{
			name: "simple",
			in: "comp foo() {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.Component{
				Comp: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ComponentHeader{
					Name: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 6},
					},
					Parameters: &ast.ComponentParameters{
						LParen: &ast.Position{Line: 1, Col: 9},
						RParen: &ast.Position{Line: 1, Col: 10},
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
			name: "with extend",
			in: "comp foo() : bar() {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.Component{
				Comp: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ComponentHeader{
					Name: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 6},
					},
					Parameters: &ast.ComponentParameters{
						LParen: &ast.Position{Line: 1, Col: 9},
						RParen: &ast.Position{Line: 1, Col: 10},
					},
				},
				Colon: &ast.Position{Line: 1, Col: 12},
				Extend: &ast.ComponentCallHeader{
					Name: &ast.Identifier{
						Name:     "bar",
						Position: &ast.Position{Line: 1, Col: 14},
					},
					Arguments: &ast.Arguments{
						LParen: &ast.Position{Line: 1, Col: 17},
						RParen: &ast.Position{Line: 1, Col: 18},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 20},
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, Component())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentHeader
	}{
		{
			name: "simple",
			in:   "foo()",
			expect: &ast.ComponentHeader{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Parameters: &ast.ComponentParameters{
					LParen: &ast.Position{Line: 1, Col: 4},
					RParen: &ast.Position{Line: 1, Col: 5},
				},
			},
		}, {
			name: "with type params",
			in:   "foo[T any](val T)",
			expect: &ast.ComponentHeader{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				TypeParams: &ast.TypeParameters{
					LBracket: &ast.Position{Line: 1, Col: 4},
					Params: []*ast.TypeParameter{
						{
							Names: []*ast.Identifier{
								{Name: "T", Position: &ast.Position{Line: 1, Col: 5}},
							},
							Type: &ast.Type{
								Parsed: &ast.NamedType{
									Name: &ast.Identifier{
										Name:     "any",
										Position: &ast.Position{Line: 1, Col: 7},
									},
								},
								Type:  "any",
								From:  ast.Position{Line: 1, Col: 7},
								Until: ast.Position{Line: 1, Col: 10},
							},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 10},
				},
				Parameters: &ast.ComponentParameters{
					LParen: &ast.Position{Line: 1, Col: 11},
					Params: []*ast.ComponentParameter{
						{
							Name: &ast.Identifier{
								Name:     "val",
								Position: &ast.Position{Line: 1, Col: 12},
							},
							Type: &ast.Type{
								Parsed: &ast.NamedType{
									Name: &ast.Identifier{
										Name:     "T",
										Position: &ast.Position{Line: 1, Col: 16},
									},
								},
								Type:  "T",
								From:  ast.Position{Line: 1, Col: 16},
								Until: ast.Position{Line: 1, Col: 17},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 17},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, Header())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestParameters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentParameters
	}{
		{
			name: "empty",
			in:   "()",
			expect: &ast.ComponentParameters{
				LParen: &ast.Position{Line: 1, Col: 1},
				RParen: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "trailing comma",
			in: "(foo any,\n" +
				")",
			expect: &ast.ComponentParameters{
				LParen: &ast.Position{Line: 1, Col: 1},
				Params: []*ast.ComponentParameter{
					{
						Name: &ast.Identifier{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 2},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "any",
									Position: &ast.Position{Line: 1, Col: 6},
								},
							},
							Type:  "any",
							From:  ast.Position{Line: 1, Col: 6},
							Until: ast.Position{Line: 1, Col: 9},
						},
					},
				},
				RParen: &ast.Position{Line: 2, Col: 1},
			},
		}, {
			name: "multiple",
			in:   "(foo any, bar string)",
			expect: &ast.ComponentParameters{
				LParen: &ast.Position{Line: 1, Col: 1},
				Params: []*ast.ComponentParameter{
					{
						Name: &ast.Identifier{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 2},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "any",
									Position: &ast.Position{Line: 1, Col: 6},
								},
							},
							Type:  "any",
							From:  ast.Position{Line: 1, Col: 6},
							Until: ast.Position{Line: 1, Col: 9},
						},
					}, {
						Name: &ast.Identifier{
							Name:     "bar",
							Position: &ast.Position{Line: 1, Col: 11},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "string",
									Position: &ast.Position{Line: 1, Col: 15},
								},
							},
							Type:  "string",
							From:  ast.Position{Line: 1, Col: 15},
							Until: ast.Position{Line: 1, Col: 21},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 21},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, Parameters())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestParameter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentParameter
	}{
		{
			name: "with type",
			in:   "param string",
			expect: &ast.ComponentParameter{
				Name: &ast.Identifier{
					Name:     "param",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{
							Name:     "string",
							Position: &ast.Position{Line: 1, Col: 7},
						},
					},
					Type:  "string",
					From:  ast.Position{Line: 1, Col: 7},
					Until: ast.Position{Line: 1, Col: 13},
				},
			},
		}, {
			name: "with default",
			in:   "param: \"default\"",
			expect: &ast.ComponentParameter{
				Name: &ast.Identifier{
					Name:     "param",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Colon: &ast.Position{Line: 1, Col: 6},
				Default: &ast.Expression{
					Nodes: ast.Code{
						&ast.String{
							Open:  &ast.Position{Line: 1, Col: 8},
							Quote: '"',
							Contents: []ast.StringNode{
								&ast.StringText{
									Text:     "default",
									Position: &ast.Position{Line: 1, Col: 9},
								},
							},
							Close: &ast.Position{Line: 1, Col: 16},
						},
					},
				},
			},
		}, {
			name: "with type and default",
			in:   "param string: \"default\"",
			expect: &ast.ComponentParameter{
				Name: &ast.Identifier{
					Name:     "param",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{
							Name:     "string",
							Position: &ast.Position{Line: 1, Col: 7},
						},
					},
					Type:  "string",
					From:  ast.Position{Line: 1, Col: 7},
					Until: ast.Position{Line: 1, Col: 13},
				},
				Colon: &ast.Position{Line: 1, Col: 13},
				Default: &ast.Expression{
					Nodes: ast.Code{
						&ast.String{
							Open:  &ast.Position{Line: 1, Col: 15},
							Quote: '"',
							Contents: []ast.StringNode{
								&ast.StringText{
									Text:     "default",
									Position: &ast.Position{Line: 1, Col: 16},
								},
							},
							Close: &ast.Position{Line: 1, Col: 23},
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
			actual := testutil.AssertNoError(t, p, Parameter())

			line, col, index := testutil.CalcEnd(1, 1, 0, c.in)
			testutil.AssertPosition(t, p, line, col, index)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestAlias(t *testing.T) {
	t.Parallel()

	in := "alias foo(s string) : bar(baz: s)"
	expect := &ast.Alias{
		Alias: &ast.Position{Line: 1, Col: 1},
		Header: &ast.ComponentHeader{
			Name: &ast.Identifier{
				Name:     "foo",
				Position: &ast.Position{Line: 1, Col: 7},
			},
			Parameters: &ast.ComponentParameters{
				LParen: &ast.Position{Line: 1, Col: 10},
				Params: []*ast.ComponentParameter{
					{
						Name: &ast.Identifier{
							Name:     "s",
							Position: &ast.Position{Line: 1, Col: 11},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "string",
									Position: &ast.Position{Line: 1, Col: 13},
								},
							},
							Type:  "string",
							From:  ast.Position{Line: 1, Col: 13},
							Until: ast.Position{Line: 1, Col: 19},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 19},
			},
		},
		ComponentCall: &ast.ComponentCall{
			Colon: &ast.Position{Line: 1, Col: 21},
			Header: &ast.ComponentCallHeader{
				Name: &ast.Identifier{
					Name:     "bar",
					Position: &ast.Position{Line: 1, Col: 23},
				},
				Arguments: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 26},
					Args: []ast.Argument{
						&ast.ComponentArgument{
							Name: &ast.Identifier{
								Name:     "baz",
								Position: &ast.Position{Line: 1, Col: 27},
							},
							Colon: &ast.Position{Line: 1, Col: 30},
							Value: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "s",
										Position: &ast.Position{Line: 1, Col: 32},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 33},
				},
			},
		},
	}

	actual := testutil.ParsesFully(t, in, Alias())
	assert.Equal(t, expect, actual)
}

func TestBlock(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.Block
	}{
		{
			name: "default block without default",
			in:   "block",
			expect: &ast.Block{
				Block: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "default block with default",
			in: "block {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.Block{
				Block: &ast.Position{Line: 1, Col: 1},
				Default: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: len("block ") + 1},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementReference{
									Name: &ast.ElementName{
										Name:     "br",
										Position: &ast.Position{Line: 2, Col: len("\t") + 1},
									},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 3, Col: 1},
				},
			},
		}, {
			name: "named block with default",
			in: "block foo {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.Block{
				Block: &ast.Position{Line: 1, Col: 1},
				Identifier: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: len("block ") + 1},
				},
				Default: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: len("block foo ") + 1},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementReference{
									Name: &ast.ElementName{
										Name:     "br",
										Position: &ast.Position{Line: 2, Col: len("\t") + 1},
									},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 3, Col: 1},
				},
			},
		}, {
			name: "named block without default",
			in:   "block foo",
			expect: &ast.Block{
				Block: &ast.Position{Line: 1, Col: 1},
				Identifier: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: len("block ") + 1},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, Block())
			assert.Equal(t, c.expect, actual)
		})
	}
}
