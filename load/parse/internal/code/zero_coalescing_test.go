package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestZeroCoalescing(t *testing.T) {
	t.Parallel()
	testZeroCoalescing(t, ZeroCoalescing())
}

func testZeroCoalescing(t *testing.T, f parser.Func[*ast.ZeroCoalescing]) {
	in := "**foo?.bar?(baz, faz)?[1?]?.(**qux?)?"
	expect := &ast.ZeroCoalescing{
		DerefCount: 2,
		Root: &ast.Expression{
			Code: ast.Code{
				&ast.GoCode{
					Code:     "foo",
					Position: ast.Position{Line: 1, Col: 3},
				},
			},
		},
		CheckRoot: &ast.Position{Line: 1, Col: 6},
		Chain: []ast.ZeroCoalescingNode{
			&ast.ZCSelectorExpression{
				Dot:   ast.Position{Line: 1, Col: 7},
				Ident: &ast.Ident{Ident: "bar", Position: ast.Position{Line: 1, Col: 8}},
				Check: &ast.Position{Line: 1, Col: 11},
			}, &ast.ZCParenExpression{
				LParen: ast.Position{Line: 1, Col: 12},
				Args: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "baz",
								Position: ast.Position{Line: 1, Col: 13},
							},
						},
					}, {
						Code: ast.Code{
							&ast.GoCode{
								Code:     "faz",
								Position: ast.Position{Line: 1, Col: 18},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 21},
				Check:  &ast.Position{Line: 1, Col: 22},
			}, &ast.ZCIndexExpression{
				LBracket: ast.Position{Line: 1, Col: 23},
				Index: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: ast.Position{Line: 1, Col: 24},
						},
					},
				},
				CheckIndex: &ast.Position{Line: 1, Col: 25},
				RBracket:   &ast.Position{Line: 1, Col: 26},
				CheckValue: &ast.Position{Line: 1, Col: 27},
			}, &ast.ZCTypeAssertionExpression{
				Dot:          ast.Position{Line: 1, Col: 28},
				LParen:       &ast.Position{Line: 1, Col: 29},
				PointerCount: 2,
				Type: &ast.Ident{
					Ident:    "qux",
					Position: ast.Position{Line: 1, Col: 32},
				},
				CheckType:  &ast.Position{Line: 1, Col: 35},
				RParen:     &ast.Position{Line: 1, Col: 36},
				CheckValue: &ast.Position{Line: 1, Col: 37},
			},
		},
		Position: ast.Position{Line: 1, Col: 1},
	}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestZCIndexExpression(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ZCIndexExpression
	}{
		{
			name: "check nothing",
			in:   "[1]",
			expect: &ast.ZCIndexExpression{
				LBracket: ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 3},
			},
		}, {
			name: "check index",
			in:   "[1?]",
			expect: &ast.ZCIndexExpression{
				LBracket: ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
				},
				CheckIndex: &ast.Position{Line: 1, Col: 3},
				RBracket:   &ast.Position{Line: 1, Col: 4},
			},
		}, {
			name: "check value",
			in:   "[1]?",
			expect: &ast.ZCIndexExpression{
				LBracket: ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
				},
				RBracket:   &ast.Position{Line: 1, Col: 3},
				CheckValue: &ast.Position{Line: 1, Col: 4},
			},
		}, {
			name: "check both",
			in:   "[1?]?",
			expect: &ast.ZCIndexExpression{
				LBracket: ast.Position{Line: 1, Col: 1},
				Index: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{
							Code:     "1",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
				},
				CheckIndex: &ast.Position{Line: 1, Col: 3},
				RBracket:   &ast.Position{Line: 1, Col: 4},
				CheckValue: &ast.Position{Line: 1, Col: 5},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, ZCIndexExpression())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestZCSelectorExpression(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ZCSelectorExpression
	}{
		{
			name: "check nothing",
			in:   ".foo",
			expect: &ast.ZCSelectorExpression{
				Dot:   ast.Position{Line: 1, Col: 1},
				Ident: &ast.Ident{Ident: "foo", Position: ast.Position{Line: 1, Col: 2}},
			},
		}, {
			name: "check",
			in:   ".foo?",
			expect: &ast.ZCSelectorExpression{
				Dot:   ast.Position{Line: 1, Col: 1},
				Ident: &ast.Ident{Ident: "foo", Position: ast.Position{Line: 1, Col: 2}},
				Check: &ast.Position{Line: 1, Col: 5},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, ZCSelectorExpression())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestZCParenExpression(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ZCParenExpression
	}{
		{
			name: "no args",
			in:   "()",
			expect: &ast.ZCParenExpression{
				LParen: ast.Position{Line: 1, Col: 1},
				Args:   []*ast.Expression{},
				RParen: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "one arg",
			in:   "(foo)",
			expect: &ast.ZCParenExpression{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "foo",
								Position: ast.Position{Line: 1, Col: 2},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 5},
			},
		}, {
			name: "multiple args",
			in:   "(foo, bar)",
			expect: &ast.ZCParenExpression{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "foo",
								Position: ast.Position{Line: 1, Col: 2},
							},
						},
					}, {
						Code: ast.Code{
							&ast.GoCode{
								Code:     "bar",
								Position: ast.Position{Line: 1, Col: 7},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "check",
			in:   "(foo)?",
			expect: &ast.ZCParenExpression{
				LParen: ast.Position{Line: 1, Col: 1},
				Args: []*ast.Expression{
					{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "foo",
								Position: ast.Position{Line: 1, Col: 2},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 5},
				Check:  &ast.Position{Line: 1, Col: 6},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, ZCParenExpression())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestZCTypeAssertionExpression(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ZCTypeAssertionExpression
	}{
		{
			name: "unqualified type",
			in:   ".(foo)",
			expect: &ast.ZCTypeAssertionExpression{
				Dot:    ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 3},
				},
				RParen: &ast.Position{Line: 1, Col: 6},
			},
		}, {
			name: "qualified type",
			in:   ".(foo.Bar)",
			expect: &ast.ZCTypeAssertionExpression{
				Dot:    ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.QualifiedIdent{
					Package: ast.Ident{
						Ident:    "foo",
						Position: ast.Position{Line: 1, Col: 3},
					},
					Dot: &ast.Position{Line: 1, Col: 6},
					Name: &ast.Ident{
						Ident:    "Bar",
						Position: ast.Position{Line: 1, Col: 7},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 10},
			},
		}, {
			name: "pointer type",
			in:   ".(*foo)",
			expect: &ast.ZCTypeAssertionExpression{
				Dot:          ast.Position{Line: 1, Col: 1},
				LParen:       &ast.Position{Line: 1, Col: 2},
				PointerCount: 1,
				Type: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 4},
				},
				RParen: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "check type",
			in:   ".(foo?)",
			expect: &ast.ZCTypeAssertionExpression{
				Dot:    ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 3},
				},
				CheckType: &ast.Position{Line: 1, Col: 6},
				RParen:    &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "check value",
			in:   ".(foo)?",
			expect: &ast.ZCTypeAssertionExpression{
				Dot:    ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 3},
				},
				RParen:     &ast.Position{Line: 1, Col: 6},
				CheckValue: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "check both",
			in:   ".(foo?)?",
			expect: &ast.ZCTypeAssertionExpression{
				Dot:    ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 2},
				Type: &ast.Ident{
					Ident:    "foo",
					Position: ast.Position{Line: 1, Col: 3},
				},
				CheckType:  &ast.Position{Line: 1, Col: 6},
				RParen:     &ast.Position{Line: 1, Col: 7},
				CheckValue: &ast.Position{Line: 1, Col: 8},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, ZCTypeAssertionExpression())
			assert.Equal(t, c.expect, actual)
		})
	}
}
