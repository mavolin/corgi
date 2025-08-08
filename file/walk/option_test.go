package walk

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

//goland:noinspection GoSnakeCaseUsage
func TestDontDive(t *testing.T) {
	scope_elem_scope_doctype := &ast.Doctype{}

	scope_elem_scope := &ast.Scope{Nodes: []ast.ScopeNode{scope_elem_scope_doctype}}
	scope_elem := &ast.Element{
		Body: scope_elem_scope,
	}

	scope_block := &ast.Block{}
	scope := &ast.Scope{
		Nodes: []ast.ScopeNode{scope_block, scope_elem},
	}

	var (
		seenScope       bool
		seenScope_block bool
		seenScope_elem  bool
	)

	Walk(scope, func(ctx *Context) Action {
		switch ctx.Node {
		case scope:
			should.False(t, seenScope) // scope seen twice
			seenScope = true
		case scope_block:
			should.False(t, seenScope_block) // scope_block seen twice
			seenScope_block = true
		case scope_elem:
			should.False(t, seenScope_elem) // scope_elem seen twice
			seenScope_elem = true
		default:
			t.Errorf("unexpected node %T", ctx.Node)
		}
		return Continue
	}, DontDive[*ast.Element]())

	should.True(t, seenScope)
	should.True(t, seenScope_block)
	should.True(t, seenScope_elem)
}
