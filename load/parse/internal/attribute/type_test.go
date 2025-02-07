package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

var attrTypes = []struct {
	name string
	typ  attrtype.Type
}{
	{name: "unsafeBool", typ: attrtype.UnsafeBool},
	{name: "unsafe", typ: attrtype.Unsafe},
	{name: "bool", typ: attrtype.Bool},
	{name: "text", typ: attrtype.Text},
	{name: "innocuous", typ: attrtype.Innocuous},
	{name: "css", typ: attrtype.CSS},
	{name: "js", typ: attrtype.JS},
	{name: "url", typ: attrtype.URL},
	{name: "urlList", typ: attrtype.URLList},
	{name: "resourceURL", typ: attrtype.ResourceURL},
	{name: "srcset", typ: attrtype.Srcset},
}

func TestType(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range attrTypes {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				expect := &ast.AttributeType{
					Quote: &ast.Position{Line: 1, Col: 1},
					Name: &ast.AttributeTypeName{
						Name:     c.name,
						Type:     c.typ,
						Position: &ast.Position{Line: 1, Col: 2},
					},
				}
				actual := testutil.ParsesFully(t, "'"+c.name, Type())
				assert.Equal(t, expect, actual)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			expect := &ast.AttributeType{
				Quote: &ast.Position{Line: 1, Col: 1},
				Name: &ast.AttributeTypeName{
					Name:     "foo",
					Type:     attrtype.Unknown,
					Position: &ast.Position{Line: 1, Col: 2},
				},
			}
			actual := testutil.MatchesButError(t, "'foo", Type())
			assert.Equal(t, expect, actual)
		})
	})
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range attrTypes {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				expect := &ast.AttributeTypeName{
					Name:     c.name,
					Type:     c.typ,
					Position: &ast.Position{Line: 1, Col: 1},
				}
				actual := testutil.ParsesFully(t, c.name, TypeName())
				assert.Equal(t, expect, actual)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			expect := &ast.AttributeTypeName{
				Name:     "foo",
				Type:     attrtype.Unknown,
				Position: &ast.Position{Line: 1, Col: 1},
			}
			actual := testutil.MatchesButError(t, "foo", TypeName())
			assert.Equal(t, expect, actual)
		})
	})
}
