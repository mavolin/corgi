package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, Value(), testExpressionValue)
	testutil.AssertAlsoFulfils(t, Value(), testTypedAttributeValue)
}

func TestExpressionValue(t *testing.T) {
	t.Parallel()
	testExpressionValue(t, ExpressionValue())
}

func testExpressionValue(t *testing.T, f parser.Func[*ast.ExpressionAttributeValue]) {
	in := "woof"
	expect := &ast.ExpressionAttributeValue{
		Nodes: []ast.ExpressionNode{
			&ast.GoCode{
				Code:     "woof",
				Position: ast.Position{Line: 1, Col: 1},
			},
		},
	}

	p := testutil.NewParser(t, in+", 1other stuff")
	actual := testutil.AssertNoError(t, p, f)

	line, col, index := testutil.CalcEnd(1, 1, 0, in)
	testutil.AssertPosition(t, p, line, col, index)

	assert.Equal(t, expect, actual)
}

func TestTypedAttributeValue(t *testing.T) {
	t.Parallel()
	testTypedAttributeValue(t, TypedAttributeValue())

	t.Run("false positive", func(t *testing.T) {
		t.Parallel()

		testCases := []string{
			`'w'`,
			`'\"'`,
		}

		for _, in := range testCases {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				testutil.NoMatch(t, in, TypedAttributeValue())
			})
		}
	})
}

func testTypedAttributeValue(t *testing.T, f parser.Func[*ast.TypedAttributeValue]) {
	for _, c := range attrTypes {
		t.Run(c.name, func(t *testing.T) {
			in := "'" + c.name + "(woof)"
			expect := &ast.TypedAttributeValue{
				Type: ast.AttributeType{
					Quote: ast.Position{Line: 1, Col: 1},
					Name: &ast.AttributeTypeName{
						Name:     c.name,
						Type:     c.typ,
						Position: ast.Position{Line: 1, Col: 2},
					},
				},
				LParen: &ast.Position{Line: 1, Col: 1 + len("'") + len(c.name)},
				Value: &ast.ExpressionAttributeValue{
					Nodes: []ast.ExpressionNode{
						&ast.GoCode{
							Code:     "woof",
							Position: ast.Position{Line: 1, Col: 1 + len("'") + len(c.name) + len("(")},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 1 + len("'") + len(c.name) + len("(woof")},
			}

			p := testutil.NewParser(t, in+", 1other stuff")
			actual := testutil.AssertNoError(t, p, f)

			line, col, index := testutil.CalcEnd(1, 1, 0, in)
			testutil.AssertPosition(t, p, line, col, index)
			assert.Equal(t, expect, actual)
		})
	}
}
