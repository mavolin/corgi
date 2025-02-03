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
				Header: ast.ComponentCallHeader{
					Colon: ast.Position{Line: 1, Col: 1},
					Name: &ast.Ident{
						Ident:    "foo",
						Position: ast.Position{Line: 1, Col: 2},
					},
				},
			},
		}, {
			name: "with body",
			in: ":foo {\n" +
				"\tbar\n" +
				"}",
			expect: &ast.ComponentCall{
				Header: ast.ComponentCallHeader{
					Colon: ast.Position{Line: 1, Col: 1},
					Name: &ast.Ident{
						Ident:    "foo",
						Position: ast.Position{Line: 1, Col: 2},
					},
				},
				Body: &ast.Scope{
					LBrace: ast.Position{Line: 1, Col: 6},
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Header: ast.ElementHeader{
								Name: ast.ElementName{
									Name:     "bar",
									Position: ast.Position{Line: 2, Col: 2},
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
			in:   ":foo",
			expect: &ast.ComponentCallHeader{
				Colon: ast.Position{Line: 1, Col: 1},
				Name: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 2},
				},
			},
		}, {
			name: "with type arguments",
			in:   ":foo[bar]",
			expect: &ast.ComponentCallHeader{
				Colon: ast.Position{Line: 1, Col: 1},
				Name: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 2},
				},
				TypeArguments: &ast.TypeArguments{
					LBracket: ast.Position{Line: 1, Col: 5},
					Types: []*ast.Type{
						{
							Parsed: &ast.NamedType{
								Name: &ast.Ident{
									Ident:    "bar",
									Position: ast.Position{Line: 1, Col: 6},
								},
							},
							Type:     "bar",
							Position: ast.Position{Line: 1, Col: 6},
							Until:    ast.Position{Line: 1, Col: 9},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 9},
				},
			},
		}, {
			name: "with arguments",
			in:   ":foo(bar: baz)",
			expect: &ast.ComponentCallHeader{
				Colon: ast.Position{Line: 1, Col: 1},
				Name: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 2},
				},
				Arguments: &ast.Arguments{
					LParen: ast.Position{Line: 1, Col: 5},
					Args: []ast.Argument{
						&ast.ComponentArgument{
							Name:  ast.Ident{Ident: "bar", Position: ast.Position{Line: 1, Col: 6}},
							Colon: &ast.Position{Line: 1, Col: 9},
							Value: &ast.Expression{
								Code: ast.Code{
									&ast.GoCode{
										Code:     "baz",
										Position: ast.Position{Line: 1, Col: 11},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 14},
				},
			},
		}, {
			name: "with type arguments and arguments",
			in:   ":foo[bar](baz: qux)",
			expect: &ast.ComponentCallHeader{
				Colon: ast.Position{Line: 1, Col: 1},
				Name: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 2},
				},
				TypeArguments: &ast.TypeArguments{
					LBracket: ast.Position{Line: 1, Col: 5},
					Types: []*ast.Type{
						{
							Parsed: &ast.NamedType{
								Name: &ast.Ident{
									Ident:    "bar",
									Position: ast.Position{Line: 1, Col: 6},
								},
							},
							Type:     "bar",
							Position: ast.Position{Line: 1, Col: 6},
							Until:    ast.Position{Line: 1, Col: 9},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 9},
				},
				Arguments: &ast.Arguments{
					LParen: ast.Position{Line: 1, Col: 10},
					Args: []ast.Argument{
						&ast.ComponentArgument{
							Name:  ast.Ident{Ident: "baz", Position: ast.Position{Line: 1, Col: 11}},
							Colon: &ast.Position{Line: 1, Col: 14},
							Value: &ast.Expression{
								Code: ast.Code{
									&ast.GoCode{
										Code:     "qux",
										Position: ast.Position{Line: 1, Col: 16},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 19},
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

	in := "with foo {\n" +
		"\tbr\n" +
		"}"
	expect := &ast.With{
		With: ast.Position{Line: 1, Col: 1},
		Name: &ast.Ident{
			Ident:    "foo",
			Position: ast.Position{Line: 1, Col: 6},
		},
		Body: &ast.Scope{
			LBrace: ast.Position{Line: 1, Col: 10},
			Nodes: []ast.ScopeNode{
				&ast.Element{
					Header: ast.ElementHeader{
						Name: ast.ElementName{
							Name:     "br",
							Position: ast.Position{Line: 2, Col: 2},
						},
					},
				},
			},
			RBrace: &ast.Position{Line: 3, Col: 1},
		},
	}

	actual := testutil.ParsesFully(t, in, With())
	assert.Equal(t, expect, actual)
}
