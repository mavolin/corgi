package walk

import (
	"testing"

	"github.com/mavolin/corgi/file/ast"
	"github.com/stretchr/testify/assert"
)

func TestIsTopLevel(t *testing.T) {
	t.Parallel()
	var called int
	Walk(nil, scope, func(parents []Context, wctx Context) error {
		switch wctx.Node {
		case elseIf_p_arrowBlock:
			assert.False(t, IsTopLevel(parents))
			called++
		case elseIf_p:
			assert.True(t, IsTopLevel(parents))
			called++
		case if_:
			assert.True(t, IsTopLevel(parents))
			called++
		case div_span_table:
			assert.False(t, IsTopLevel(parents))
			called++
		case div_span:
			assert.False(t, IsTopLevel(parents))
			called++
		case div:
			assert.True(t, IsTopLevel(parents))
			called++
		}
		return nil
	})
	assert.Equal(t, 6, called)
}

func TestChildIsTopLevel(t *testing.T) {
	t.Parallel()
	var called int
	Walk(nil, scope, func(parents []Context, wctx Context) error {
		switch wctx.Node {
		case elseIf_p_arrowBlock:
			assert.False(t, ChildIsTopLevel(parents, wctx))
			called++
		case elseIf_p:
			assert.False(t, ChildIsTopLevel(parents, wctx))
			called++
		case if_:
			assert.True(t, ChildIsTopLevel(parents, wctx))
			called++
		case div_span_table:
			assert.False(t, ChildIsTopLevel(parents, wctx))
			called++
		case div_span:
			assert.False(t, ChildIsTopLevel(parents, wctx))
			called++
		case div:
			assert.False(t, ChildIsTopLevel(parents, wctx))
			called++
		}
		return nil
	})
	assert.Equal(t, 6, called)
}

func TestClosest(t *testing.T) {
	t.Parallel()
	var called int
	Walk(nil, scope, func(parents []Context, wctx Context) error {
		switch wctx.Node {
		case elseIf_p_arrowBlock:
			expect := if_
			actual := Closest[*ast.If](parents)
			assert.Same(t, expect, actual)
			called++
		case div_span_table:
			expect := div_span
			actual := Closest[*ast.Element](parents)
			assert.Same(t, expect, actual)
			called++
		case div_span:
			expect := div
			actual := Closest[*ast.Element](parents)
			assert.Same(t, expect, actual)
			called++
		}
		return nil
	})
	assert.Equal(t, 3, called)
}
