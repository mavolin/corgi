package state

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDeclaration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.StateDeclaration
	}{
		{
			name: "single",
			in:   "state foo int",
			expect: &ast.StateDeclaration{
				State: ast.Position{Line: 1, Col: 1},
				Specs: []*ast.StateSpec{
					{
						Names: []*ast.Ident{
							{Ident: "foo", Position: ast.Position{Line: 1, Col: 7}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Ident{Ident: "int", Position: ast.Position{Line: 1, Col: 11}},
							},
							Type:     "int",
							Position: ast.Position{Line: 1, Col: 11},
							Until:    ast.Position{Line: 1, Col: 14},
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
			expect: &ast.StateDeclaration{
				State:  ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 7},
				Specs: []*ast.StateSpec{
					{
						Names: []*ast.Ident{
							{Ident: "foo", Position: ast.Position{Line: 2, Col: 2}},
						},
						EqualSign: &ast.Position{Line: 2, Col: 6},
						Values: []*ast.Expression{
							{
								Code: ast.Code{
									&ast.GoCode{Code: "42", Position: ast.Position{Line: 2, Col: 8}},
								},
							},
						},
					}, {
						Names: []*ast.Ident{
							{Ident: "bar", Position: ast.Position{Line: 3, Col: 2}},
						},
						Type: &ast.Type{
							Parsed: &ast.NamedType{
								Name: &ast.Ident{Ident: "int", Position: ast.Position{Line: 3, Col: 6}},
							},
							Type:     "int",
							Position: ast.Position{Line: 3, Col: 6},
							Until:    ast.Position{Line: 3, Col: 9},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in+";", Declaration())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.StateSpec
	}{
		{
			name: "only type",
			in:   "foo int",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Ident{Ident: "int", Position: ast.Position{Line: 1, Col: 5}},
					},
					Type:     "int",
					Position: ast.Position{Line: 1, Col: 5},
					Until:    ast.Position{Line: 1, Col: 8},
				},
			},
		}, {
			name: "only value",
			in:   "foo = 42",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
				},
				EqualSign: &ast.Position{Line: 1, Col: 5},
				Values: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{Code: "42", Position: ast.Position{Line: 1, Col: 7}},
						},
					},
				},
			},
		}, {
			name: "type and value",
			in:   "foo int = 42",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Ident{Ident: "int", Position: ast.Position{Line: 1, Col: 5}},
					},
					Type:     "int",
					Position: ast.Position{Line: 1, Col: 5},
					Until:    ast.Position{Line: 1, Col: 8},
				},
				EqualSign: &ast.Position{Line: 1, Col: 9},
				Values: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{Code: "42", Position: ast.Position{Line: 1, Col: 11}},
						},
					},
				},
			},
		}, {
			name: "multiple names with type",
			in:   "foo, bar int",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
					{Ident: "bar", Position: ast.Position{Line: 1, Col: 6}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Ident{Ident: "int", Position: ast.Position{Line: 1, Col: 10}},
					},
					Type:     "int",
					Position: ast.Position{Line: 1, Col: 10},
					Until:    ast.Position{Line: 1, Col: 13},
				},
			},
		}, {
			name: "multiple names with multiple values",
			in:   "foo, bar = 42, 43",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
					{Ident: "bar", Position: ast.Position{Line: 1, Col: 6}},
				},
				EqualSign: &ast.Position{Line: 1, Col: 10},
				Values: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{Code: "42", Position: ast.Position{Line: 1, Col: 12}},
						},
					}, {
						Code: ast.Code{
							&ast.GoCode{Code: "43", Position: ast.Position{Line: 1, Col: 16}},
						},
					},
				},
			},
		}, {
			name: "multiple names with type and multiple values",
			in:   "foo, bar int = 42, 43",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
					{Ident: "bar", Position: ast.Position{Line: 1, Col: 6}},
				},
				Type: &ast.Type{
					Parsed: &ast.NamedType{
						Name: &ast.Ident{Ident: "int", Position: ast.Position{Line: 1, Col: 10}},
					},
					Type:     "int",
					Position: ast.Position{Line: 1, Col: 10},
					Until:    ast.Position{Line: 1, Col: 13},
				},
				EqualSign: &ast.Position{Line: 1, Col: 14},
				Values: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{Code: "42", Position: ast.Position{Line: 1, Col: 16}},
						},
					}, {
						Code: ast.Code{
							&ast.GoCode{Code: "43", Position: ast.Position{Line: 1, Col: 20}},
						},
					},
				},
			},
		}, {
			name: "multiple names with single value",
			in:   "foo, bar = baz()",
			expect: &ast.StateSpec{
				Names: []*ast.Ident{
					{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
					{Ident: "bar", Position: ast.Position{Line: 1, Col: 6}},
				},
				EqualSign: &ast.Position{Line: 1, Col: 10},
				Values: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{Code: "baz()", Position: ast.Position{Line: 1, Col: 12}},
						},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in+"\n 1other stuff")
			actual := testutil.AssertNoError(t, p, Spec())

			line, col, index := testutil.CalcEnd(1, 1, 0, c.in)
			testutil.AssertPosition(t, p, line, col, index)
			assert.Equal(t, c.expect, actual)
		})
	}
}
