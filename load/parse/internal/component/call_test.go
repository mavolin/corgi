package component

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentCall
	}{
		{
			name: "no body",
			in:   ":foo",
			expect: &ast.ComponentCall{
				Colon: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ComponentCallHeader{
					Name: &ast.Ident{
						Ident:    "foo",
						Position: &ast.Position{Line: 1, Col: 2},
					},
				},
			},
		}, {
			name: "with body",
			in: ":foo {\n" +
				"\tbar\n" +
				"}",
			expect: &ast.ComponentCall{
				Colon: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ComponentCallHeader{
					Name: &ast.Ident{
						Ident:    "foo",
						Position: &ast.Position{Line: 1, Col: 2},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: 6},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: &ast.ElementHeader{
								Name: &ast.ElementReference{
									Name: &ast.ElementName{
										Name:     "bar",
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
			actual := testutil.ParsesFully(t, c.in, Call())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestCallHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentCallHeader
	}{
		{
			name: "only name",
			in:   "foo",
			expect: &ast.ComponentCallHeader{
				Name: &ast.Ident{
					Ident:    "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "with type arguments",
			in:   "foo[bar]",
			expect: &ast.ComponentCallHeader{
				Name: &ast.Ident{
					Ident:    "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				TypeArguments: &ast.TypeArguments{
					LBracket: &ast.Position{Line: 1, Col: 4},
					Types: []*ast.Type{
						{
							Parsed: &ast.NamedType{
								Name: &ast.Ident{
									Ident:    "bar",
									Position: &ast.Position{Line: 1, Col: 5},
								},
							},
							Type:  "bar",
							From:  ast.Position{Line: 1, Col: 5},
							Until: ast.Position{Line: 1, Col: 8},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 8},
				},
			},
		}, {
			name: "with arguments",
			in:   "foo(bar: baz)",
			expect: &ast.ComponentCallHeader{
				Name: &ast.Ident{
					Ident:    "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Arguments: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 4},
					Args: []ast.Argument{
						&ast.ComponentArgument{
							Name:  &ast.Ident{Ident: "bar", Position: &ast.Position{Line: 1, Col: 5}},
							Colon: &ast.Position{Line: 1, Col: 8},
							Value: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "baz",
										Position: &ast.Position{Line: 1, Col: 10},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 13},
				},
			},
		}, {
			name: "with type arguments and arguments",
			in:   "foo[bar](baz: qux)",
			expect: &ast.ComponentCallHeader{
				Name: &ast.Ident{
					Ident:    "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				TypeArguments: &ast.TypeArguments{
					LBracket: &ast.Position{Line: 1, Col: 4},
					Types: []*ast.Type{
						{
							Parsed: &ast.NamedType{
								Name: &ast.Ident{
									Ident:    "bar",
									Position: &ast.Position{Line: 1, Col: 5},
								},
							},
							Type:  "bar",
							From:  ast.Position{Line: 1, Col: 5},
							Until: ast.Position{Line: 1, Col: 8},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 8},
				},
				Arguments: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 9},
					Args: []ast.Argument{
						&ast.ComponentArgument{
							Name:  &ast.Ident{Ident: "baz", Position: &ast.Position{Line: 1, Col: 10}},
							Colon: &ast.Position{Line: 1, Col: 13},
							Value: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "qux",
										Position: &ast.Position{Line: 1, Col: 15},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 18},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, CallHeader())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestWith(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.With
	}{
		{
			name: "default block",
			in: "with {\n" +
				"\tbr" +
				"\n}",
			expect: &ast.With{
				With: &ast.Position{Line: 1, Col: 1},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: len("with ") + 1},
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
			name: "named block",
			in: "with foo {\n" +
				"\tbr\n" +
				"}",
			expect: &ast.With{
				With: &ast.Position{Line: 1, Col: 1},
				Identifier: &ast.Ident{
					Ident:    "foo",
					Position: &ast.Position{Line: 1, Col: len("with ") + 1},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 1, Col: len("with foo ") + 1},
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
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, With())
			assert.Equal(t, c.expect, actual)
		})
	}
}
