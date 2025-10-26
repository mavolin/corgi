package attribute

import (
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

var attrTypes = []struct {
	typ  attrtype.Type
	attr string
}{
	{typ: attrtype.UnsafeBool},
	{typ: attrtype.Unsafe},
	{typ: attrtype.UnsafeBool, attr: "woof"},
	{typ: attrtype.Unsafe, attr: "woof"},
	{typ: attrtype.Bool},
	{typ: attrtype.Int},
	{typ: attrtype.Float},
	{typ: attrtype.String},
	{typ: attrtype.Text},
	{typ: attrtype.CSS},
	{typ: attrtype.JS},
	{typ: attrtype.URL},
	{typ: attrtype.ResourceURL},
	{typ: attrtype.Srcset},
	{typ: attrtype.SpaceList{Element: attrtype.Int}},
	{typ: attrtype.SpaceList{Element: attrtype.Float}},
	{typ: attrtype.SpaceList{Element: attrtype.String}},
	{typ: attrtype.SpaceList{Element: attrtype.URL}},
	{typ: attrtype.SpaceList{Element: attrtype.ResourceURL}},
	{typ: attrtype.CommaList{Element: attrtype.Int}},
	{typ: attrtype.CommaList{Element: attrtype.Float}},
	{typ: attrtype.CommaList{Element: attrtype.String}},
	{typ: attrtype.CommaList{Element: attrtype.URL}},
	{typ: attrtype.CommaList{Element: attrtype.ResourceURL}},
}

func TestType(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range attrTypes {
			name := c.typ.String()
			s := name
			if c.attr != "" {
				s += "[" + c.attr + "]"
			}
			t.Run(s, func(t *testing.T) {
				t.Parallel()

				want := &ast.AttributeType{
					Quote: &ast.Position{Line: 1, Col: 1},
					Name:  wantTypeName(c.typ, ast.Position{Line: 1, Col: 1 + len("'")}),
				}
				if c.attr != "" {
					want.LBracket = &ast.Position{Line: 1, Col: 1 + len("'") + len(name)}
					want.Attribute = &ast.AttributeName{
						Name:          c.attr,
						CanonicalName: c.attr,
						Position:      &ast.Position{Line: 1, Col: 1 + len("'") + len(name) + len("[")},
					}
					want.RBracket = &ast.Position{
						Line: 1,
						Col:  1 + len("'") + len(name) + len("[") + len(c.attr),
					}
				}
				got := parsetest.ParsesExact(t, "'"+s, Type())
				should.Equal(t, got, want, parsetest.CmpOpts...)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			want := &ast.AttributeType{
				Quote: &ast.Position{Line: 1, Col: 1},
				Name: &ast.AttributeTypeName{
					Name:     "foo",
					Type:     nil,
					Position: &ast.Position{Line: 1, Col: 2},
				},
			}
			wantError := "unknown attribute type"

			got := parsetest.ParsesExact(t, "'"+want.Name.Name, Type(), parsetest.WantErrors(wantError))
			should.Equal(t, got, want, parsetest.CmpOpts...)
		})
	})
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range attrTypes {
			t.Run(c.typ.String(), func(t *testing.T) {
				t.Parallel()

				want := wantTypeName(c.typ, ast.Position{Line: 1, Col: 1})
				got := parsetest.ParsesExact(t, c.typ.String(), TypeName())
				should.Equal(t, got, want, parsetest.CmpOpts...)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			want := &ast.AttributeTypeName{
				Name:     "foo",
				Type:     nil,
				Position: &ast.Position{Line: 1, Col: 1},
			}
			wantError := "unknown attribute type"

			got := parsetest.ParsesExact(t, want.Name, TypeName(), parsetest.WantErrors(wantError))
			should.Equal(t, got, want, parsetest.CmpOpts...)
		})
	})
}

func wantTypeName(t attrtype.Type, start ast.Position) *ast.AttributeTypeName {
	name := t.String()
	listName, elemName, _ := strings.Cut(name, "[")
	elemName = strings.TrimSuffix(elemName, "]")

	atn := &ast.AttributeTypeName{
		Name:     listName,
		Type:     t,
		Position: &start,
	}
	if elemName != "" {
		atn.LBracket = &ast.Position{Line: start.Line, Col: start.Col + len(listName)}
		atn.Element = elemName
		atn.RBracket = &ast.Position{Line: start.Line, Col: start.Col + len(listName) + len("["+elemName)}
	}
	return atn
}
