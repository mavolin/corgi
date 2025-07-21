package element

import (
	"testing"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ElementDefinition
	}{
		{
			name: "single",
			in:   "elem foo normal",
			want: &ast.ElementDefinition{
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
			want: &ast.ElementDefinition{
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
			want: &ast.ElementDefinition{
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
			want: &ast.ElementDefinition{
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesFully(t, c.in, Definition())
			should.Equal(t, c.want, got)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ElementSpec
	}{
		{
			name: "basic",
			in:   "foo normal",
			want: &ast.ElementSpec{
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
			want: &ast.ElementSpec{
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in+"; 1other stuff")
			got := parsetest.AssertNoError(t, p, Spec())

			line, col, index := parsetest.CalcEnd(1, 1, 0, c.in)
			parsetest.AssertPosition(t, p, line, col, index)
			should.Equal(t, c.want, got)
		})
	}
}

func TestType(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, Type(), testBasicType)
	parsetest.AssertAlsoFulfils(t, Type(), testAliasType)
}

func TestBasicType(t *testing.T) {
	t.Parallel()
	testBasicType(t, BasicType())
}

func testBasicType(t *testing.T, f parser.Func[*ast.BasicElementType]) {
	in := "normal"
	want := &ast.BasicElementType{
		Type: &ast.ElementTypeName{
			Name:     "normal",
			Type:     elemtype.Normal,
			Position: &ast.Position{Line: 1, Col: 1},
		},
	}

	got := parsetest.ParsesFully(t, in, f)
	should.Equal(t, want, got)
}

func TestAliasType(t *testing.T) {
	t.Parallel()
	testAliasType(t, AliasType())
}

func testAliasType(t *testing.T, f parser.Func[*ast.AliasElementType]) {
	in := "= div"
	want := &ast.AliasElementType{
		EqualSign: &ast.Position{Line: 1, Col: 1},
		Name: &ast.ElementReference{
			Name: &ast.ElementName{Name: "div", Position: &ast.Position{Line: 1, Col: 3}},
		},
	}

	got := parsetest.ParsesFully(t, in, f)
	should.Equal(t, want, got)
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
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

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			want := &ast.ElementTypeName{
				Name:     c.name,
				Type:     c.typ,
				Position: &ast.Position{Line: 1, Col: 1},
			}

			got := parsetest.ParsesFully(t, c.name, TypeName())
			should.Equal(t, want, got)
		})
	}
}
