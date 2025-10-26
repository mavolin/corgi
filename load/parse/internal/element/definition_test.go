package element

import (
	"testing"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
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
			name: "single without prefix",
			in:   "elem foo normal",
			want: &ast.ElementDefinition{
				Elem: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 1, Col: 6},
						},
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
				Elem: &ast.Position{Line: 1, Col: 1},
				Prefix: &ast.ElementName{
					Name:          "x",
					CanonicalName: "x",
					Position:      &ast.Position{Line: 1, Col: 6},
				},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 1, Col: 8},
						},
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
			name: "multiple without prefix",
			in: "elem (\n" +
				"\tfoo normal\n" +
				"\tbar = div\n" +
				")",
			want: &ast.ElementDefinition{
				Elem:   &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 6},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 2, Col: 2},
						},
						Type: &ast.BasicElementType{
							Type: &ast.ElementTypeName{
								Name:     "normal",
								Type:     elemtype.Normal,
								Position: &ast.Position{Line: 2, Col: 6},
							},
						},
					}, {
						Name: &ast.ElementName{
							Name:          "bar",
							CanonicalName: "bar",
							Position:      &ast.Position{Line: 3, Col: 2},
						},
						Type: &ast.AliasElementType{
							EqualSign: &ast.Position{Line: 3, Col: 6},
							Name: &ast.ElementReference{
								Name: &ast.ElementName{
									Name:          "div",
									CanonicalName: "div",
									Position:      &ast.Position{Line: 3, Col: 8},
								},
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
				Elem: &ast.Position{Line: 1, Col: 1},
				Prefix: &ast.ElementName{
					Name:          "x",
					CanonicalName: "x",
					Position:      &ast.Position{Line: 1, Col: 6},
				},
				LParen: &ast.Position{Line: 1, Col: 8},
				Specs: []*ast.ElementSpec{
					{
						Name: &ast.ElementName{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 2, Col: 2},
						},
						Type: &ast.BasicElementType{
							Type: &ast.ElementTypeName{
								Name:     "normal",
								Type:     elemtype.Normal,
								Position: &ast.Position{Line: 2, Col: 6},
							},
						},
					}, {
						Name: &ast.ElementName{
							Name:          "bar",
							CanonicalName: "bar",
							Position:      &ast.Position{Line: 3, Col: 2},
						},
						Type: &ast.AliasElementType{
							EqualSign: &ast.Position{Line: 3, Col: 6},
							Name: &ast.ElementReference{
								Name: &ast.ElementName{
									Name:          "div",
									CanonicalName: "div",
									Position:      &ast.Position{Line: 3, Col: 8},
								},
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
			got := parsetest.ParsesUntilEOS(t, c.in, Definition())
			should.Equal(t, got, c.want)
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
				Name: &ast.ElementName{
					Name:          "foo",
					CanonicalName: "foo",
					Position:      &ast.Position{Line: 1, Col: 1},
				},
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
				Name: &ast.ElementName{
					Name:          "foo",
					CanonicalName: "foo",
					Position:      &ast.Position{Line: 1, Col: 1},
				},
				Type: &ast.AliasElementType{
					EqualSign: &ast.Position{Line: 1, Col: 5},
					Name: &ast.ElementReference{
						Name: &ast.ElementName{
							Name:          "div",
							CanonicalName: "div",
							Position:      &ast.Position{Line: 1, Col: 7},
						},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, Spec())
			should.Equal(t, got, c.want)
		})
	}
}

func TestType(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, Type(), testBasicType)
	parsetest.AlsoFulfils(t, Type(), testAliasType)
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

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want)
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
			Name: &ast.ElementName{
				Name:          "div",
				CanonicalName: "div",
				Position:      &ast.Position{Line: 1, Col: 3},
			},
		},
	}

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want)
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	for _, et := range elemtype.All {
		t.Run(et.String(), func(t *testing.T) {
			t.Parallel()

			want := &ast.ElementTypeName{
				Name:     et.String(),
				Type:     et,
				Position: &ast.Position{Line: 1, Col: 1},
			}

			got := parsetest.ParsesExact(t, et.String(), TypeName())
			should.Equal(t, got, want)
		})
	}
}
