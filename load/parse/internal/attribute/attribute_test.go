package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestAttribute(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, Attribute(), testAndPlaceholder)
	parsetest.AlsoFulfils(t, Attribute(), testIDShorthand)
	parsetest.AlsoFulfils(t, Attribute(), testClassShorthand)
	parsetest.AlsoFulfils(t, Attribute(), testNamedAttribute)
}

func TestAndPlaceholder(t *testing.T) {
	t.Parallel()
	testAndPlaceholder(t, AndPlaceholder())
}

func testAndPlaceholder(t *testing.T, f parser.Func[*ast.AndPlaceholder]) {
	want := &ast.AndPlaceholder{
		And: &ast.Position{Line: 1, Col: 1},
	}

	got := parsetest.ParsesUntilComma(t, "&", f)
	should.Equal(t, got, want)
}

func TestNamedAttribute(t *testing.T) {
	t.Parallel()
	testNamedAttribute(t, NamedAttribute())
}

func testNamedAttribute(t *testing.T, f parser.Func[*ast.NamedAttribute]) {
	tests := []struct {
		name string
		in   string
		want *ast.NamedAttribute
	}{
		{
			name: "boolean",
			in:   "async",
			want: &ast.NamedAttribute{
				Name: &ast.AttributeReference{
					Name: &ast.AttributeName{
						Name:     "async",
						Position: &ast.Position{Line: 1, Col: 1},
					},
				},
			},
		}, {
			name: "value",
			in:   `value=woof`,
			want: &ast.NamedAttribute{
				Name: &ast.AttributeReference{
					Name: &ast.AttributeName{
						Name:     "value",
						Position: &ast.Position{Line: 1, Col: 1},
					},
				},
				EqualSign: &ast.Position{Line: 1, Col: 6},
				Value: &ast.ExpressionAttributeValue{
					Nodes: ast.Code{
						&ast.GoCode{
							Code:     "woof",
							Position: &ast.Position{Line: 1, Col: 7},
						},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesUntilComma(t, c.in, f)
			should.Equal(t, got, c.want)
		})
	}
}

func TestReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.AttributeReference
	}{
		{
			name: "local",
			in:   "name",
			want: &ast.AttributeReference{
				Name: &ast.AttributeName{
					Name:     "name",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "external",
			in:   "package1.Name",
			want: &ast.AttributeReference{
				Package: &ast.Identifier{
					Name:     "package1",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Dot: &ast.Position{Line: 1, Col: 9},
				Name: &ast.AttributeName{
					Name:     "Name",
					Position: &ast.Position{Line: 1, Col: 10},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, Reference())
			should.Equal(t, got, c.want)
		})
	}
}
