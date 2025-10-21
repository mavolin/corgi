package component

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ComponentCall
	}{
		{
			name: "no body",
			in:   ":foo",
			want: &ast.ComponentCall{
				Colon: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ComponentCallHeader{
					Name: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 2},
					},
				},
			},
		}, {
			name: "with body",
			in: ":foo {\n" +
				"\tbar\n" +
				"}",
			want: &ast.ComponentCall{
				Colon: &ast.Position{Line: 1, Col: 1},
				Header: &ast.ComponentCallHeader{
					Name: &ast.Identifier{
						Name:     "foo",
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesExact(t, c.in, Call())
			should.Equal(t, got, c.want)
		})
	}
}

func TestCallHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ComponentCallHeader
	}{
		{
			name: "only name",
			in:   "foo",
			want: &ast.ComponentCallHeader{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "with type arguments",
			in:   "foo[bar]",
			want: &ast.ComponentCallHeader{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				TypeArguments: &ast.TypeArguments{
					LBracket: &ast.Position{Line: 1, Col: 4},
					Types: []*ast.Type{
						{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "bar",
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
			want: &ast.ComponentCallHeader{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Arguments: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 4},
					List: []ast.Argument{
						&ast.ComponentArgument{
							Name:  &ast.Identifier{Name: "bar", Position: &ast.Position{Line: 1, Col: 5}},
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
			want: &ast.ComponentCallHeader{
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				TypeArguments: &ast.TypeArguments{
					LBracket: &ast.Position{Line: 1, Col: 4},
					Types: []*ast.Type{
						{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{
									Name:     "bar",
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
					List: []ast.Argument{
						&ast.ComponentArgument{
							Name:  &ast.Identifier{Name: "baz", Position: &ast.Position{Line: 1, Col: 10}},
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesExact(t, c.in, CallHeader())
			should.Equal(t, got, c.want)
		})
	}
}

func TestWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.With
	}{
		{
			name: "default block",
			in: "with {\n" +
				"\tbr" +
				"\n}",
			want: &ast.With{
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
			want: &ast.With{
				With: &ast.Position{Line: 1, Col: 1},
				Identifier: &ast.Identifier{
					Name:     "foo",
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesExact(t, c.in, With())
			should.Equal(t, got, c.want)
		})
	}
}
