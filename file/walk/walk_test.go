package walk

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

var testNode = &ast.Element{
	Header: &ast.ElementHeader{
		Name: &ast.ElementReference{},
	},
	Body: &ast.Scope{
		LBrace: &ast.Position{},
		Nodes: []ast.ScopeNode{
			&ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementReference{},
				},
			},
			&ast.Doctype{},
		},
		RBrace: &ast.Position{},
	},
}

func TestWalk(t *testing.T) {
	t.Parallel()

	t.Run("visit all nodes", func(t *testing.T) {
		t.Parallel()

		visited := make(map[ast.Node]bool)
		Walk(testNode, func(ctx *Context) Action {
			visited[ctx.Node] = true
			return Continue
		})

		// Should've visited the main element, its header, name, body, and the 2 content nodes
		expectedVisits := 8
		should.Equal(t, len(visited), expectedVisits)
	})

	t.Run("early termination", func(t *testing.T) {
		t.Parallel()

		var visitCount int
		Walk(testNode, func(ctx *Context) Action {
			visitCount++
			if _, ok := ctx.Node.(*ast.Element); len(ctx.Parents) > 0 && ok {
				return Break
			}
			return Continue
		})

		should.Equal(t, visitCount, 5)
	})

	t.Run("no dive", func(t *testing.T) {
		t.Parallel()

		var visitCount int
		Walk(testNode, func(ctx *Context) Action {
			visitCount++
			if _, ok := ctx.Node.(*ast.Element); len(ctx.Parents) > 0 && ok {
				return NoDive
			}
			return Continue
		})

		should.Equal(t, visitCount, 6)
	})

	t.Run("parent relationships", func(t *testing.T) {
		t.Parallel()

		parentCheckPassed := false
		Walk(testNode, func(ctx *Context) Action {
			// Check that child nodes have their parent in the parents slice
			if _, ok := ctx.Node.(*ast.Doctype); ok {
				if len(ctx.Parents) > 0 {
					if scope, ok := ctx.Parents[len(ctx.Parents)-1].Node.(*ast.Scope); ok {
						if scope == testNode.Body {
							parentCheckPassed = true
						}
					}
				}
			}
			return Continue
		})

		should.Equal(t, parentCheckPassed, true)
	})
}

func TestWalkT(t *testing.T) {
	t.Parallel()

	t.Run("visits doctype", func(t *testing.T) {
		t.Parallel()

		// Test WalkT with specific node type
		var visitCount int
		WalkT[*ast.Doctype](testNode, func(*ContextT[*ast.Doctype]) Action {
			visitCount++
			return Continue
		})

		should.Equal(t, visitCount, 1)
	})

	t.Run("VisitsElementNodes", func(t *testing.T) {
		t.Parallel()

		var visitCount int
		WalkT[*ast.Element](testNode, func(*ContextT[*ast.Element]) Action {
			visitCount++
			return Continue
		})

		should.Equal(t, visitCount, 2)
	})

	t.Run("early termination", func(t *testing.T) {
		t.Parallel()

		var visitCount int
		WalkT[*ast.Element](testNode, func(*ContextT[*ast.Element]) Action {
			visitCount++
			return Break
		})

		should.Equal(t, visitCount, 1)
	})

	t.Run("no dive", func(t *testing.T) {
		t.Parallel()

		var visitCount int
		WalkT[*ast.Element](testNode, func(*ContextT[*ast.Element]) Action {
			visitCount++
			return NoDive
		})

		should.Equal(t, visitCount, 1)
	})
}
