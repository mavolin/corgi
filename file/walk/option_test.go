package walk

import (
	"testing"

	"github.com/mavolin/corgi/file/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChildOf(t *testing.T) {
	t.Parallel()
	t.Run("if", func(t *testing.T) {
		t.Parallel()
		expectOrder := []ast.Node{if_br}

		var i int

		err := Walk(scope, func(wctx *Context) error {
			require.Less(t, i, len(expectOrder), "unexpected node")
			assert.Equal(t, expectOrder[i], wctx.Node)
			i++
			return nil
		}, ChildOf(&ast.If{}))
		require.NoError(t, err)
		assert.Equal(t, len(expectOrder), i)
	})
	t.Run("element element", func(t *testing.T) {
		t.Parallel()
		expectOrder := []ast.Node{
			div_span_componentCall, div_span_table,
		}

		var i int

		err := Walk(scope, func(wctx *Context) error {
			require.Less(t, i, len(expectOrder), "unexpected node")
			assert.Equal(t, expectOrder[i], wctx.Node)
			i++
			return nil
		}, ChildOf(&ast.Element{}, &ast.Element{}))
		require.NoError(t, err)
		assert.Equal(t, len(expectOrder), i)
	})
}

func TestChildOfAny(t *testing.T) {
	t.Parallel()
	expectOrder := []ast.Node{
		if_br, elseIf_p_arrowBlock, elseIf_p_arrowBlock_textLine, elseIf_p_arrowBlock_textLine_text,
		div_span, div_span_componentCall, div_span_table, div_img,
	}

	var i int

	err := Walk(scope, func(wctx *Context) error {
		require.Less(t, i, len(expectOrder), "unexpected node")
		assert.Equal(t, expectOrder[i], wctx.Node)
		i++
		return nil
	}, ChildOfAny(&ast.If{}, &ast.Element{}))
	require.NoError(t, err)
	assert.Equal(t, len(expectOrder), i)
}

func TestNotChildOf(t *testing.T) {
	t.Parallel()
	t.Run("if", func(t *testing.T) {
		t.Parallel()
		expectOrder := []ast.Node{
			if_, elseIf, elseIf_p, elseIf_p_arrowBlock, elseIf_p_arrowBlock_textLine,
			elseIf_p_arrowBlock_textLine_text, div, div_span, div_span_componentCall,
			div_span_table, div_img,
		}

		var i int

		err := Walk(scope, func(wctx *Context) error {
			require.Less(t, i, len(expectOrder), "unexpected node")
			assert.Equal(t, expectOrder[i], wctx.Node)
			i++
			return nil
		}, NotChildOf(&ast.If{}))
		require.NoError(t, err)
		assert.Equal(t, len(expectOrder), i)
	})
	t.Run("element element", func(t *testing.T) {
		t.Parallel()
		expectOrder := []ast.Node{
			if_, if_br, elseIf, elseIf_p, elseIf_p_arrowBlock,
			elseIf_p_arrowBlock_textLine, elseIf_p_arrowBlock_textLine_text, div, div_span, div_img,
		}

		var i int

		err := Walk(scope, func(wctx *Context) error {
			require.Less(t, i, len(expectOrder), "unexpected node")
			assert.Equal(t, expectOrder[i], wctx.Node)
			i++
			return nil
		}, NotChildOf(&ast.Element{}, &ast.Element{}))
		require.NoError(t, err)
		assert.Equal(t, len(expectOrder), i)
	})
}

func TestNotChildOfAny(t *testing.T) {
	t.Parallel()
	expectOrder := []ast.Node{
		if_, elseIf, elseIf_p, div,
	}

	var i int

	err := Walk(scope, func(wctx *Context) error {
		require.Less(t, i, len(expectOrder), "unexpected node")
		assert.Equal(t, expectOrder[i], wctx.Node)
		i++
		return nil
	}, NotChildOfAny(&ast.If{}, &ast.Element{}))
	require.NoError(t, err)
	assert.Equal(t, len(expectOrder), i)
}

func TestDontDiveAny(t *testing.T) {
	t.Parallel()
	expectOrder := []ast.Node{
		if_, elseIf, div, div_span, div_span_componentCall, div_span_table, div_img,
	}

	var i int

	err := Walk(scope, func(wctx *Context) error {
		require.Less(t, i, len(expectOrder), "unexpected node")
		assert.Equal(t, expectOrder[i], wctx.Node)
		i++
		return nil
	}, DontDiveAny(&ast.If{}, &ast.ElseIf{}, &ast.Else{}, &ast.ComponentCall{}))
	require.NoError(t, err)
	assert.Equal(t, len(expectOrder), i)
}

func TestTopLevel(t *testing.T) {
	t.Parallel()
	expectOrder := []ast.Node{
		if_, if_br, elseIf, elseIf_p, div,
	}

	var i int

	err := Walk(scope, func(wctx *Context) error {
		require.Less(t, i, len(expectOrder), "unexpected node")
		assert.Equal(t, expectOrder[i], wctx.Node)
		i++
		return nil
	}, TopLevel())
	require.NoError(t, err)
	assert.Equal(t, len(expectOrder), i)
}
