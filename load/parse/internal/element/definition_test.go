package element

import (
	"testing"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDefinition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ElementDefinition
	}{
		{
			name: "single",
			in:   "elem foo normal",
			expect: &ast.ElementDefinition{
				Elem: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 1, Col: 6}},
						Type: &ast.BasicElementType{
							Type: &ast.ElementTypeName{
								Name:     "normal",
								Type:     elemtype.Normal,
								Position: &ast.Position{Line: 1, Col: 10},
							},
						},
					},
				},
			},
		}, {
			name: "single with prefix",
			in:   "elem x foo normal",
			expect: &ast.ElementDefinition{
				Elem:   &ast.Position{Line: 1, Col: 1},
				Prefix: &ast.ElementName{Name: "x", Position: &ast.Position{Line: 1, Col: 6}},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 1, Col: 8}},
						Type: &ast.BasicElementType{
							Type: &ast.ElementTypeName{
								Name:     "normal",
								Type:     elemtype.Normal,
								Position: &ast.Position{Line: 1, Col: 12},
							},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in: "elem (\n" +
				"\tfoo normal\n" +
				"\tbar = div\n" +
				")",
			expect: &ast.ElementDefinition{
				Elem:   &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 6},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 2, Col: 2}},
						Type: &ast.BasicElementType{
							Type: &ast.ElementTypeName{
								Name:     "normal",
								Type:     elemtype.Normal,
								Position: &ast.Position{Line: 2, Col: 6},
							},
						},
					}, {
						Name: &ast.ElementName{Name: "bar", Position: &ast.Position{Line: 3, Col: 2}},
						Type: &ast.AliasElementType{
							EqualSign: &ast.Position{Line: 3, Col: 6},
							Name: &ast.ElementReference{
								Name: &ast.ElementName{Name: "div", Position: &ast.Position{Line: 3, Col: 8}},
							},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
			},
		}, {
			name: "multiple with prefix",
			in: "elem x (\n" +
				"\tfoo normal\n" +
				"\tbar = div\n" +
				")",
			expect: &ast.ElementDefinition{
				Elem:   &ast.Position{Line: 1, Col: 1},
				Prefix: &ast.ElementName{Name: "x", Position: &ast.Position{Line: 1, Col: 6}},
				LParen: &ast.Position{Line: 1, Col: 8},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 2, Col: 2}},
						Type: &ast.BasicElementType{
							Type: &ast.ElementTypeName{
								Name:     "normal",
								Type:     elemtype.Normal,
								Position: &ast.Position{Line: 2, Col: 6},
							},
						},
					}, {
						Name: &ast.ElementName{Name: "bar", Position: &ast.Position{Line: 3, Col: 2}},
						Type: &ast.AliasElementType{
							EqualSign: &ast.Position{Line: 3, Col: 6},
							Name: &ast.ElementReference{
								Name: &ast.ElementName{Name: "div", Position: &ast.Position{Line: 3, Col: 8}},
							},
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
			actual := testutil.ParsesFully(t, c.in, Definition())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ElementSpec
	}{
		{
			name: "basic",
			in:   "foo normal",
			expect: &ast.ElementSpec{
				Name: &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				Type: &ast.BasicElementType{
					Type: &ast.ElementTypeName{
						Name:     "normal",
						Type:     elemtype.Normal,
						Position: &ast.Position{Line: 1, Col: 5},
					},
				},
			},
		}, {
			name: "alias",
			in:   "foo = div",
			expect: &ast.ElementSpec{
				Name: &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				Type: &ast.AliasElementType{
					EqualSign: &ast.Position{Line: 1, Col: 5},
					Name: &ast.ElementReference{
						Name: &ast.ElementName{Name: "div", Position: &ast.Position{Line: 1, Col: 7}},
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := testutil.NewParser(t, c.in+"; 1other stuff")
			actual := testutil.AssertNoError(t, p, Spec())

			line, col, index := testutil.CalcEnd(1, 1, 0, c.in)
			testutil.AssertPosition(t, p, line, col, index)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestType(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, Type(), testBasicType)
	testutil.AssertAlsoFulfils(t, Type(), testAliasType)
}

func TestBasicType(t *testing.T) {
	t.Parallel()
	testBasicType(t, BasicType())
}

func testBasicType(t *testing.T, f parser.Func[*ast.BasicElementType]) {
	in := "normal"
	expect := &ast.BasicElementType{
		Type: &ast.ElementTypeName{
			Name:     "normal",
			Type:     elemtype.Normal,
			Position: &ast.Position{Line: 1, Col: 1},
		},
	}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestAliasType(t *testing.T) {
	t.Parallel()
	testAliasType(t, AliasType())
}

func testAliasType(t *testing.T, f parser.Func[*ast.AliasElementType]) {
	in := "= div"
	expect := &ast.AliasElementType{
		EqualSign: &ast.Position{Line: 1, Col: 1},
		Name: &ast.ElementReference{
			Name: &ast.ElementName{Name: "div", Position: &ast.Position{Line: 1, Col: 3}},
		},
	}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)

}

func TestTypeName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  elemtype.Type
	}{
		{
			name: "void",
			typ:  elemtype.Void,
		}, {
			name: "nothing",
			typ:  elemtype.Nothing,
		}, {
			name: "normal",
			typ:  elemtype.Normal,
		}, {
			name: "text",
			typ:  elemtype.Text,
		}, {
			name: "css",
			typ:  elemtype.CSS,
		}, {
			name: "js",
			typ:  elemtype.JS,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			expect := &ast.ElementTypeName{
				Name:     c.name,
				Type:     c.typ,
				Position: &ast.Position{Line: 1, Col: 1},
			}

			actual := testutil.ParsesFully(t, c.name, TypeName())
			assert.Equal(t, expect, actual)
		})
	}
}
