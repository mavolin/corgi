package walk

import (
	"testing"

	"github.com/mavolin/corgi/file/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//goland:noinspection GoSnakeCaseUsage
var (
	scope = &ast.Scope{
		Nodes: []ast.ScopeNode{
			&ast.If{
				Then: &ast.Scope{
					Nodes: []ast.ScopeNode{
						&ast.Element{Name: "br"},
					},
				},
				ElseIfs: []*ast.ElseIf{
					{
						Then: &ast.Scope{
							Nodes: []ast.ScopeNode{
								&ast.Element{
									Name: "p",
									Body: &ast.Scope{
										Nodes: []ast.ScopeNode{
											&ast.ArrowBlock{
												Lines: []ast.TextLine{
													{&ast.Text{Text: "Hello, World!"}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.Element{
				Name: "div",
				Body: &ast.Scope{
					Nodes: []ast.ScopeNode{
						&ast.Element{
							Name: "span",
							Body: &ast.Scope{
								Nodes: []ast.ScopeNode{
									&ast.ComponentCall{
										Name: &ast.Ident{Ident: "foo"},
									},
									&ast.Element{Name: "table"},
								},
							},
						},
						&ast.Element{Name: "img"},
					},
				},
			},
		},
	}

	_ = 0 // so the first /**/ stays in place
	/**/ if_ = scope.Nodes[0].(*ast.If)
	/*  */ ifScope = if_.Then.(*ast.Scope)
	/*    */ if_br = ifScope.Nodes[0].(*ast.Element)
	/**/ elseIf = if_.ElseIfs[0]
	/*  */ elseIfScope = elseIf.Then.(*ast.Scope)
	/*    */ elseIf_p = elseIfScope.Nodes[0].(*ast.Element)
	/*      */ elseIf_pScope = elseIf_p.Body.(*ast.Scope)
	/*        */ elseIf_p_arrowBlock = elseIf_pScope.Nodes[0].(*ast.ArrowBlock)
	/*          */ elseIf_p_arrowBlock_textLine = elseIf_p_arrowBlock.Lines[0]
	/*            */ elseIf_p_arrowBlock_textLine_text = elseIf_p_arrowBlock_textLine[0].(*ast.Text)
	/**/ div = scope.Nodes[1].(*ast.Element)
	/*  */ divScope = div.Body.(*ast.Scope)
	/*    */ div_span = divScope.Nodes[0].(*ast.Element)
	/*      */ div_spanScope = div_span.Body.(*ast.Scope)
	/*        */ div_span_componentCall = div_spanScope.Nodes[0].(*ast.ComponentCall)
	/*        */ div_span_table = div_spanScope.Nodes[1].(*ast.Element)
	/*    */ div_img = divScope.Nodes[1].(*ast.Element)
)

func TestWalk(t *testing.T) {
	expectOrder := []ast.Node{
		if_, if_br,
		elseIf, elseIf_p, elseIf_p_arrowBlock, elseIf_p_arrowBlock_textLine,
		elseIf_p_arrowBlock_textLine_text,
		div, div_span, div_span_componentCall, div_span_table, div_img,
	}

	var i int

	err := Walk(scope, func(wctx *Context) error {
		require.Less(t, i, len(expectOrder), "unexpected node")
		assert.Equal(t, expectOrder[i], wctx.Node)
		i++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, len(expectOrder), i)
}

func TestWalkT(t *testing.T) {
	expectOrder := []ast.Node{if_br, elseIf_p, div, div_span, div_span_table, div_img}

	var i int

	err := WalkT(scope, func(wctx *ContextT[*ast.Element]) error {
		require.Less(t, i, len(expectOrder), "unexpected node")
		assert.Equal(t, expectOrder[i], wctx.Node)
		i++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, len(expectOrder), i)
}
