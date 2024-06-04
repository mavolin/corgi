package attribute

import (
	"testing"

	"github.com/mavolin/corgi/escape/attrtype"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
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

func testExpressionValue(t *testing.T, f parser.Func[ast.ExpressionAttributeValue]) {
	expect := &ast.ExpressionAttributeValue{
		&ast.GoCode{
			Code:     "woof",
			Position: ast.Position{Line: 1, Col: 1},
		},
	}

	actual := testutil.ParsesFully(t, "woof", ExpressionValue())
	assert.Equal(t, expect, actual)
}

func TestTypedAttributeValue(t *testing.T) {
	t.Parallel()
	testTypedAttributeValue(t, TypedAttributeValue())
}

func testTypedAttributeValue(t *testing.T, f parser.Func[*ast.TypedAttributeValue]) {
	testCases := []struct {
		name string
		typ  attrtype.Type
	}{
		{name: "unsafeBool", typ: attrtype.UnsafeBool},
		{name: "unsafe", typ: attrtype.Unsafe},
		{name: "bool", typ: attrtype.Bool},
		{name: "text", typ: attrtype.Text},
		{name: "css", typ: attrtype.CSS},
		{name: "js", typ: attrtype.JS},
		{name: "url", typ: attrtype.URL},
		{name: "urlList", typ: attrtype.URLList},
		{name: "resourceURL", typ: attrtype.ResourceURL},
		{name: "srcset", typ: attrtype.Srcset},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			in := c.name + "(woof)"
			expect := &ast.TypedAttributeValue{
				Type:   c.typ,
				LParen: &ast.Position{Line: 1, Col: 1 + len(c.name)},
				Value: &ast.ExpressionAttributeValue{
					&ast.GoCode{
						Code:     "woof",
						Position: ast.Position{Line: 1, Col: 1 + len(c.name) + len("(")},
					},
				},
				RParen:   &ast.Position{Line: 1, Col: 1 + len(c.name) + len("(woof")},
				Position: ast.Position{Line: 1, Col: 1},
			}

			actual := testutil.ParsesFully(t, in, TypedAttributeValue())
			assert.Equal(t, expect, actual)
		})
	}
}
