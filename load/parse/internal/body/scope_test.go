package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestScope(t *testing.T) {
	t.Parallel()
	testScope(t, Scope())
}

func testScope(t *testing.T, f parser.Func[*ast.Scope]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.Scope
	}{
		{
			name: "empty",
			in:   "{}",
			expect: &ast.Scope{
				LBrace: &ast.Position{Line: 1, Col: 1},
				RBrace: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "empty with newline",
			in:   "{\n}",
			expect: &ast.Scope{
				LBrace: &ast.Position{Line: 1, Col: 1},
				RBrace: &ast.Position{Line: 2, Col: 1},
			},
		}, {
			name: "element",
			in: "{\n" +
				"\tbr\n" +
				"}",
			expect: &ast.Scope{
				LBrace: &ast.Position{Line: 1, Col: 1},
				Nodes: []ast.ScopeNode{
					&ast.Element{
						Header: &ast.ElementHeader{
							Name: &ast.ElementName{
								Name:     "br",
								Position: &ast.Position{Line: 2, Col: 2},
							},
						},
					},
				},
				RBrace: &ast.Position{Line: 3, Col: 1},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestBadScopeNode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
	}{
		{
			name: "gibberish",
			in:   "$%#",
		}, {
			name: "with parens",
			in:   "(br)",
		}, {
			name: "with braces",
			in:   "{br}",
		}, {
			name: "with brackets",
			in:   "[br]",
		}, {
			name: "mix",
			in:   "{br}() %44 [br]",
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			expect := &ast.BadScopeNode{
				From:  ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 1 + len(c.in)},
			}

			p := testutil.NewParser(t, c.in+" foo")
			actual := testutil.AssertNoError(t, p, BadScopeNode())
			assert.Equal(t, expect, actual)

			testutil.AssertPosition(t, p, expect.Until.Line, expect.Until.Col, len(c.in))
		})
	}
}
