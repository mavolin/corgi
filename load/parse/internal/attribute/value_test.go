package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestValue(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, Value(), testExpressionValue)
	parsetest.AlsoFulfils(t, Value(), testTypedAttributeValue)
}

func TestExpressionValue(t *testing.T) {
	t.Parallel()
	testExpressionValue(t, ExpressionValue())
}

func testExpressionValue(t *testing.T, f parser.Func[*ast.ExpressionAttributeValue]) {
	in := "woof"
	want := &ast.ExpressionAttributeValue{
		Nodes: ast.Code{
			&ast.GoCode{
				Code:     "woof",
				Position: &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	got := parsetest.ParsesUntilComma(t, in, f)
	should.Equal(t, got, want)
}

func TestTypedAttributeValue(t *testing.T) {
	t.Parallel()
	testTypedAttributeValue(t, TypedAttributeValue())

	t.Run("false positive", func(t *testing.T) {
		t.Parallel()

		tests := []string{
			`'w'`,
			`'\"'`,
		}

		for _, in := range tests {
			t.Run(in, func(t *testing.T) {
				t.Parallel()

				parsetest.NoMatch(t, in, TypedAttributeValue())
			})
		}
	})
}

func testTypedAttributeValue(t *testing.T, f parser.Func[*ast.TypedAttributeValue]) {
	for _, c := range attrTypes {
		t.Run(c.name, func(t *testing.T) {
			in := "'" + c.name + "(woof)"
			want := &ast.TypedAttributeValue{
				Type: &ast.AttributeType{
					Quote: &ast.Position{Line: 1, Col: 1},
					Name: &ast.AttributeTypeName{
						Name:     c.name,
						Type:     c.typ,
						Position: &ast.Position{Line: 1, Col: 2},
					},
				},
				LParen: &ast.Position{Line: 1, Col: 1 + len("'") + len(c.name)},
				Value: &ast.ExpressionAttributeValue{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "woof",
							Position: &ast.Position{Line: 1, Col: 1 + len("'") + len(c.name) + len("(")},
						},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 1 + len("'") + len(c.name) + len("(woof")},
			}

			got := parsetest.ParsesUntilComma(t, in, f)
			should.Equal(t, got, want)
		})
	}
}
