package anno

import (
	"fmt"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/stretchr/testify/assert"
)

func TestAnno(t *testing.T) {
	t.Parallel()
	t.Run("only highlight", func(t *testing.T) {
		t.Parallel()

		expectFile := new(file.File)
		expectHighlight := Highlight{
			Start: ast.Position{Line: 2, Col: 3},
			End:   ast.Position{Line: 2, Col: 4},
		}
		expectContext := Context{
			Start: 1,
			End:   5,
		}
		expectAnno := "annotation"

		anno := Anno(expectFile, Annotation{
			Highlight: func(f *file.File) (Context, Highlight) {
				assert.Same(t, expectFile, f)
				return expectContext, expectHighlight
			},
			Annotation: expectAnno,
		})
		assert.Equal(t, expectFile, anno.File)
		assert.Equal(t, expectContext.Start, anno.ContextStart)
		assert.Equal(t, expectContext.End, anno.ContextEnd)
		assert.Equal(t, expectHighlight.Start, anno.Start)
		assert.Equal(t, expectHighlight.End, anno.End)
		assert.Equal(t, expectAnno, anno.Annotation)
	})
	t.Run("highlight and context", func(t *testing.T) {
		t.Parallel()

		expectFile := new(file.File)
		expectHighlight := Highlight{
			Start: ast.Position{Line: 2, Col: 3},
			End:   ast.Position{Line: 2, Col: 4},
		}
		expectContext := Context{
			Start: 1,
			End:   5,
		}
		otherContext := Context{
			Start: 0,
			End:   6,
		}
		expectAnno := "annotation"

		anno := Anno(expectFile, Annotation{
			Context: func(f *file.File, context Context, highlight Highlight) Context {
				assert.Same(t, expectFile, f)
				assert.Equal(t, otherContext, context)
				assert.Equal(t, expectHighlight, highlight)
				return expectContext
			},
			Highlight: func(f *file.File) (Context, Highlight) {
				assert.Same(t, expectFile, f)
				return otherContext, expectHighlight
			},
			Annotation: expectAnno,
		})
		assert.Equal(t, expectFile, anno.File)
		assert.Equal(t, expectContext.Start, anno.ContextStart)
		assert.Equal(t, expectContext.End, anno.ContextEnd)
		assert.Equal(t, expectHighlight.Start, anno.Start)
		assert.Equal(t, expectHighlight.End, anno.End)
		assert.Equal(t, expectAnno, anno.Annotation)
	})
}

func TestRange(t *testing.T) {
	t.Parallel()

	expect := diagnostic.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: 5},
		Annotation:   "anno",
	}

	anno := Range(expect.File, expect.Start, expect.End, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestToEOL(t *testing.T) {
	t.Parallel()

	expectFile := &file.File{
		AST: &ast.File{
			Lines: []string{
				"foo",
				"foobar",
				"bar",
			},
		},
	}
	expect := diagnostic.Annotation{
		File:         expectFile,
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: len(expectFile.AST.Lines[1]) + 1},
		Annotation:   "anno",
	}

	anno := ToEOL(expect.File, expect.Start, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestPosition(t *testing.T) {
	t.Parallel()

	expect := diagnostic.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: 4},
		Annotation:   "anno",
	}

	anno := Position(expect.File, expect.Start, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestNChars(t *testing.T) {
	t.Parallel()

	n := 3
	expect := diagnostic.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: 3 + n},
		Annotation:   "anno",
	}

	anno := NChars(expect.File, expect.Start, n, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestNode(t *testing.T) {
	t.Parallel()

	start := ast.Position{Line: 3, Col: 4} // to make things simpler

	testCases := []struct {
		node ast.Node
	}{
		{
			node: &ast.Identifier{
				Name:     "foo",
				Position: &start,
			},
		}, {
			node: &ast.String{
				Open:  &start,
				Quote: '"',
				Contents: []ast.StringNode{
					&ast.StringText{
						Text:     "foo",
						Position: &ast.Position{Line: start.Line, Col: start.Col + len(`"`)},
					},
				},
				Close: &ast.Position{Line: start.Line, Col: start.Col + len(`"foo`)},
			},
		}, {
			node: &ast.StateDeclaration{
				LParen: &ast.Position{Line: start.Line, Col: start.Col + len("state ")},
				Specs: []*ast.StateSpec{
					{
						Names: []*ast.Identifier{
							{
								Name: "foo", Position: &ast.Position{Line: start.Line + 1, Col: 3},
							},
						},
						EqualSign: &ast.Position{Line: start.Line + 1, Col: 3 + len("foo ")},
						Values: []*ast.Expression{
							{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "bar",
										Position: &ast.Position{Line: start.Line + 1, Col: 3 + len("foo = ")},
									},
								},
							},
						},
					},
				},
				RParen: &ast.Position{Line: start.Line + 2, Col: 3},
			},
		},
	}

	for _, c := range testCases {
		t.Run(fmt.Sprintf("%T", c.node), func(t *testing.T) {
			expect := diagnostic.Annotation{
				File:         new(file.File),
				ContextStart: c.node.Start().Line,
				ContextEnd:   c.node.End().Line + 1,
				Start:        c.node.Start(),
				End:          c.node.End(),
				Annotation:   "anno",
			}

			anno := Node(expect.File, c.node, expect.Annotation)
			assert.Equal(t, expect, anno)
		})
	}
}
