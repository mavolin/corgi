package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestZeroCoalescing(t *testing.T) {
	t.Parallel()
	testZeroCoalescing(t, ZeroCoalescing())
}

func testZeroCoalescing(t *testing.T, f parser.Func[*ast.ZeroCoalescing]) {
	in := "**foo?.bar?(baz, faz)?[1?]?.(**qux?)?"
	want := &ast.ZeroCoalescing{
		DerefCount: 2,
		Root: &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{
					Code:     "foo",
					Position: &ast.Position{Line: 1, Col: 3},
				},
			},
		},
		CheckRoot: &ast.Position{Line: 1, Col: 6},
		Chain: []ast.ZeroCoalescingNode{
			&ast.ZCSelectorExpression{
				Dot:   &ast.Position{Line: 1, Col: 7},
				Ident: &ast.Identifier{Name: "bar", Position: &ast.Position{Line: 1, Col: 8}},
				Check: &ast.Position{Line: 1, Col: 11},
			}, &ast.ZCParenExpression{
				LParen: &ast.Position{Line: 1, Col: 12},
				Args: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "baz",
								Position: &ast.Position{Line: 1, Col: 13},
							},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "faz",
								Position: &ast.Position{Line: 1, Col: 18},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 21},
				Check:  &ast.Position{Line: 1, Col: 22},
			}, &ast.ZCIndexExpression{
				LBracket: &ast.Position{Line: 1, Col: 23},
				Index: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: &ast.Position{Line: 1, Col: 24},
						},
					},
				},
				CheckIndex: &ast.Position{Line: 1, Col: 25},
				RBracket:   &ast.Position{Line: 1, Col: 26},
				CheckValue: &ast.Position{Line: 1, Col: 27},
			}, &ast.ZCTypeAssertionExpression{
				Dot:          &ast.Position{Line: 1, Col: 28},
				LParen:       &ast.Position{Line: 1, Col: 29},
				PointerCount: 2,
				Type: &ast.Identifier{
					Name:     "qux",
					Position: &ast.Position{Line: 1, Col: 32},
				},
				CheckType:  &ast.Position{Line: 1, Col: 35},
				RParen:     &ast.Position{Line: 1, Col: 36},
				CheckValue: &ast.Position{Line: 1, Col: 37},
			},
		},
		DerefPosition: &ast.Position{Line: 1, Col: 1},
	}

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want)
}

func TestZCIndexExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ZCIndexExpression
	}{
		{
			name: "check nothing",
			in:   "[1]",
			want: &ast.ZCIndexExpression{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: &ast.Position{Line: 1, Col: 2},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 3},
			},
		}, {
			name: "check index",
			in:   "[1?]",
			want: &ast.ZCIndexExpression{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: &ast.Position{Line: 1, Col: 2},
						},
					},
				},
				CheckIndex: &ast.Position{Line: 1, Col: 3},
				RBracket:   &ast.Position{Line: 1, Col: 4},
			},
		}, {
			name: "check value",
			in:   "[1]?",
			want: &ast.ZCIndexExpression{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: &ast.Position{Line: 1, Col: 2},
						},
					},
				},
				RBracket:   &ast.Position{Line: 1, Col: 3},
				CheckValue: &ast.Position{Line: 1, Col: 4},
			},
		}, {
			name: "check both",
			in:   "[1?]?",
			want: &ast.ZCIndexExpression{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: &ast.Position{Line: 1, Col: 2},
						},
					},
				},
				CheckIndex: &ast.Position{Line: 1, Col: 3},
				RBracket:   &ast.Position{Line: 1, Col: 4},
				CheckValue: &ast.Position{Line: 1, Col: 5},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, ZCIndexExpression())
			should.Equal(t, got, c.want)
		})
	}
}

func TestZCSelectorExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ZCSelectorExpression
	}{
		{
			name: "check nothing",
			in:   ".foo",
			want: &ast.ZCSelectorExpression{
				Dot:   &ast.Position{Line: 1, Col: 1},
				Ident: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 2}},
			},
		}, {
			name: "check",
			in:   ".foo?",
			want: &ast.ZCSelectorExpression{
				Dot:   &ast.Position{Line: 1, Col: 1},
				Ident: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 2}},
				Check: &ast.Position{Line: 1, Col: 5},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, ZCSelectorExpression())
			should.Equal(t, got, c.want)
		})
	}
}

func TestZCParenExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ZCParenExpression
	}{
		{
			name: "no args",
			in:   "()",
			want: &ast.ZCParenExpression{
				LParen: &ast.Position{Line: 1, Col: 1},
				RParen: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "one arg",
			in:   "(foo)",
			want: &ast.ZCParenExpression{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 5},
			},
		}, {
			name: "multiple args",
			in:   "(foo, bar)",
			want: &ast.ZCParenExpression{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					}, {
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "bar",
								Position: &ast.Position{Line: 1, Col: 7},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "check",
			in:   "(foo)?",
			want: &ast.ZCParenExpression{
				LParen: &ast.Position{Line: 1, Col: 1},
				Args: []*ast.Expression{
					{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 5},
				Check:  &ast.Position{Line: 1, Col: 6},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, ZCParenExpression())
			should.Equal(t, got, c.want)
		})
	}
}

func TestZCTypeAssertionExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ZCTypeAssertionExpression
	}{
		{
			name: "unqualified type",
			in:   ".(foo)",
			want: &ast.ZCTypeAssertionExpression{
				Dot:    &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 3},
				},
				RParen: &ast.Position{Line: 1, Col: 6},
			},
		}, {
			name: "qualified type",
			in:   ".(foo.Bar)",
			want: &ast.ZCTypeAssertionExpression{
				Dot:    &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.QualifiedIdentifier{
					Package: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 3},
					},
					Dot: &ast.Position{Line: 1, Col: 6},
					Name: &ast.Identifier{
						Name:     "Bar",
						Position: &ast.Position{Line: 1, Col: 7},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "pointer type",
			in:   ".(*foo)",
			want: &ast.ZCTypeAssertionExpression{
				Dot:          &ast.Position{Line: 1, Col: 1},
				LParen:       &ast.Position{Line: 1, Col: 2},
				PointerCount: 1,
				Type: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 4},
				},
				RParen: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "check type",
			in:   ".(foo?)",
			want: &ast.ZCTypeAssertionExpression{
				Dot:    &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 3},
				},
				CheckType: &ast.Position{Line: 1, Col: 6},
				RParen:    &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "check value",
			in:   ".(foo)?",
			want: &ast.ZCTypeAssertionExpression{
				Dot:    &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 3},
				},
				RParen:     &ast.Position{Line: 1, Col: 6},
				CheckValue: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "check both",
			in:   ".(foo?)?",
			want: &ast.ZCTypeAssertionExpression{
				Dot:    &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: 3},
				},
				CheckType:  &ast.Position{Line: 1, Col: 6},
				RParen:     &ast.Position{Line: 1, Col: 7},
				CheckValue: &ast.Position{Line: 1, Col: 8},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, ZCTypeAssertionExpression())
			should.Equal(t, got, c.want)
		})
	}
}
