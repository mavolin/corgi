package attribute

import (
	"testing"

	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAttribute(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, Attribute(), testAndPlaceholder)
	testutil.AssertAlsoFulfils(t, Attribute(), testIDShorthand)
	testutil.AssertAlsoFulfils(t, Attribute(), testClassShorthand)
	testutil.AssertAlsoFulfils(t, Attribute(), testNamedAttribute)
}

func TestAndPlaceholder(t *testing.T) {
	t.Parallel()
	testAndPlaceholder(t, AndPlaceholder())
}

func testAndPlaceholder(t *testing.T, f parser.Func[*ast.AndPlaceholder]) {
	expect := &ast.AndPlaceholder{
		Position: ast.Position{Line: 1, Col: 1},
	}

	p := testutil.NewParser(t, "&, other")
	actual := testutil.AssertNoError(t, p, f)
	if assert.Equal(t, expect, actual) {
		testutil.AssertPosition(t, p, expect.End().Line, expect.End().Col, 1)
	}
}

func TestNamedAttribute(t *testing.T) {
	t.Parallel()
	testNamedAttribute(t, NamedAttribute())
}

func testNamedAttribute(t *testing.T, f parser.Func[*ast.NamedAttribute]) {
	testCases := []struct {
		name        string
		in          string
		expect      *ast.NamedAttribute
		expectIndex int
	}{
		{
			name: "boolean",
			// add the other value to simulate a real list and make sure the comma isn't consumed
			in: "async, other",
			expect: &ast.NamedAttribute{
				Position: ast.Position{Line: 1, Col: 1},
				Name:     "async",
			},
			expectIndex: 5,
		}, {
			name: "value",
			in:   `value=woof, other`,
			expect: &ast.NamedAttribute{
				Position: ast.Position{Line: 1, Col: 1},
				Name:     "class",
				Assign:   &ast.Position{Line: 1, Col: 6},
				Value: &ast.ExpressionAttributeValue{
					&ast.GoCode{
						Code:     "woof",
						Position: ast.Position{Line: 1, Col: 7},
					},
				},
			},
			expectIndex: 10,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			p := testutil.NewParser(t, c.in)
			actual := testutil.AssertNoError(t, p, f)
			if assert.Equal(t, c.expect, actual) {
				testutil.AssertPosition(t, p, c.expect.End().Line, c.expect.End().Col, c.expectIndex)
			}
		})

	}
}
