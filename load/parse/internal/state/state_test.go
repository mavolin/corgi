package state

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestDeclaration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.StateDeclaration
	}{
		{
			name: "single",
			in:   "state foo int",
			want: &ast.StateDeclaration{
				State: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.StateSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 11}},
							},
							Type:  "int",
							From:  ast.Position{Line: 1, Col: 11},
							Until: ast.Position{Line: 1, Col: 14},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in: "state (\n" +
				"\tfoo = 42\n" +
				"\tbar int\n" +
				")",
			want: &ast.StateDeclaration{
				State:  &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 7},
				Specs: []*ast.StateSpec{
					{
						Names: []*ast.Identifier{
							{Name: "foo", Position: &ast.Position{Line: 2, Col: 2}},
						},
						EqualSign: &ast.Position{Line: 2, Col: 6},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{Code: "42", Position: &ast.Position{Line: 2, Col: 8}},
								},
							},
						},
					}, {
						Names: []*ast.Identifier{
							{Name: "bar", Position: &ast.Position{Line: 3, Col: 2}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 3, Col: 6}},
							},
							Type:  "int",
							From:  ast.Position{Line: 3, Col: 6},
							Until: ast.Position{Line: 3, Col: 9},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesFully(t, c.in, Declaration())
			should.Equal(t, got, c.want)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.StateSpec
	}{
		{
			name: "only type",
			in:   "foo int",
			want: &ast.StateSpec{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 5}},
					},
					Type:  "int",
					From:  ast.Position{Line: 1, Col: 5},
					Until: ast.Position{Line: 1, Col: 8},
				},
			},
		}, {
			name: "only value",
			in:   "foo = 42",
			want: &ast.StateSpec{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				},
				EqualSign: &ast.Position{Line: 1, Col: 5},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 7}},
						},
					},
				},
			},
		}, {
			name: "type and value",
			in:   "foo int = 42",
			want: &ast.StateSpec{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 5}},
					},
					Type:  "int",
					From:  ast.Position{Line: 1, Col: 5},
					Until: ast.Position{Line: 1, Col: 8},
				},
				EqualSign: &ast.Position{Line: 1, Col: 9},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 11}},
						},
					},
				},
			},
		}, {
			name: "multiple names with type",
			in:   "foo, bar int",
			want: &ast.StateSpec{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: &ast.Position{Line: 1, Col: 6}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 10}},
					},
					Type:  "int",
					From:  ast.Position{Line: 1, Col: 10},
					Until: ast.Position{Line: 1, Col: 13},
				},
			},
		}, {
			name: "multiple names with multiple values",
			in:   "foo, bar = 42, 43",
			want: &ast.StateSpec{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: &ast.Position{Line: 1, Col: 6}},
				},
				EqualSign: &ast.Position{Line: 1, Col: 10},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 12}},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{Code: "43", Position: &ast.Position{Line: 1, Col: 16}},
						},
					},
				},
			},
		}, {
			name: "multiple names with type and multiple values",
			in:   "foo, bar int = 42, 43",
			want: &ast.StateSpec{
				Names: []*ast.Identifier{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: &ast.Position{Line: 1, Col: 6}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Identifier{Name: "int", Position: &ast.Position{Line: 1, Col: 10}},
					},
					Type:  "int",
					From:  ast.Position{Line: 1, Col: 10},
					Until: ast.Position{Line: 1, Col: 13},
				},
				EqualSign: &ast.Position{Line: 1, Col: 14},
				Values: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{Code: "42", Position: &ast.Position{Line: 1, Col: 16}},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{Code: "43", Position: &ast.Position{Line: 1, Col: 20}},
						},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in+"\n 1other stuff")
			got := parsetest.AssertNoError(t, p, Spec())

			line, col, index := parsetest.CalcEnd(1, 1, 0, c.in)
			parsetest.AssertPosition(t, p, line, col, index)
			should.Equal(t, got, c.want)
		})
	}
}
