package attribute

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestAttribute(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, Attribute(), testAndPlaceholder)
	parsetest.AssertAlsoFulfils(t, Attribute(), testIDShorthand)
	parsetest.AssertAlsoFulfils(t, Attribute(), testClassShorthand)
	parsetest.AssertAlsoFulfils(t, Attribute(), testNamedAttribute)
}

func TestAndPlaceholder(t *testing.T) {
	t.Parallel()
	testAndPlaceholder(t, AndPlaceholder())
}

func testAndPlaceholder(t *testing.T, f parser.Func[*ast.AndPlaceholder]) {
	want := &ast.AndPlaceholder{
		And: &ast.Position{Line: 1, Col: 1},
	}

	p := parsetest.NewParser(t, "&, other")
	got := parsetest.AssertNoError(t, p, f)
	if should.Equal(t, want, got) {
		parsetest.AssertPosition(t, p, want.End().Line, want.End().Col, 1)
	}
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

			// we add ", other" to the input to ensure that the parser stops at
			// the correct position
			p := parsetest.NewParser(t, c.in+", other")
			got := parsetest.AssertNoError(t, p, f)
			if should.Equal(t, c.want, got) {
				parsetest.AssertPosition(t, p, c.want.End().Line, c.want.End().Col, len(c.in))
			}
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

			got := parsetest.ParsesFully(t, c.in, Reference())
			should.Equal(t, c.want, got)
		})
	}
}
