package walk

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestClosest(t *testing.T) {
	ctx1 := &Context{Node: &ast.Element{}}
	ctx2 := &Context{Node: &ast.Doctype{}}
	ctx3 := &Context{Node: &ast.Element{}}

	parents := []*Context{ctx2, ctx1, ctx3}
	result := Closest[*ast.Element](parents)
	should.True(t, result == ctx3.Node)

	parents = []*Context{ctx2}
	result = Closest[*ast.Element](parents)
	should.Equal(t, result, nil)
}

func TestIsChildOf(t *testing.T) {
	ctx1 := &Context{Node: &ast.Element{}}
	ctx2 := &Context{Node: &ast.Doctype{}}
	parents := []*Context{ctx2, ctx1}

	should.True(t, IsChildOf[*ast.Element](parents))

	parents = []*Context{ctx2}
	should.False(t, IsChildOf[*ast.Element](parents))
}
