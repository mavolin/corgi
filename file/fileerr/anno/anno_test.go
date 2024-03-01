package anno

import (
	"fmt"
	"testing"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/stretchr/testify/assert"
)

func TestAnno(t *testing.T) {
	t.Parallel()
	t.Run("only highlight", func(t *testing.T) {
		t.Parallel()

		expectFile := new(file.File)
		expectHighlight := Highlight{
			Line:  2,
			Start: 3,
			End:   4,
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
		assert.Equal(t, expectHighlight.Line, anno.Line)
		assert.Equal(t, expectHighlight.Start, anno.Start)
		assert.Equal(t, expectHighlight.End, anno.End)
		assert.Equal(t, expectAnno, anno.Annotation)
	})
	t.Run("highlight and context", func(t *testing.T) {
		t.Parallel()

		expectFile := new(file.File)
		expectHighlight := Highlight{
			Line:  2,
			Start: 3,
			End:   4,
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
		assert.Equal(t, expectHighlight.Line, anno.Line)
		assert.Equal(t, expectHighlight.Start, anno.Start)
		assert.Equal(t, expectHighlight.End, anno.End)
		assert.Equal(t, expectAnno, anno.Annotation)
	})
}

func TestRange(t *testing.T) {
	t.Parallel()

	expect := fileerr.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Line:         2,
		Start:        3,
		End:          5,
		Annotation:   "anno",
	}
	start := ast.Position{Line: expect.Line, Col: expect.Start}
	end := ast.Position{Line: expect.Line, Col: expect.End}

	anno := Range(expect.File, start, end, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestToEOL(t *testing.T) {
	t.Parallel()

	expectFile := &file.File{
		AST: &ast.AST{
			Lines: []string{
				"foo",
				"foobar",
				"bar",
			},
		},
	}
	expect := fileerr.Annotation{
		File:         expectFile,
		ContextStart: 2,
		ContextEnd:   3,
		Line:         2,
		Start:        3,
		End:          len(expectFile.Lines[1]) + 1,
		Annotation:   "anno",
	}
	start := ast.Position{Line: expect.Line, Col: expect.Start}

	anno := ToEOL(expect.File, start, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestPosition(t *testing.T) {
	t.Parallel()

	expect := fileerr.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Line:         2,
		Start:        3,
		End:          4,
		Annotation:   "anno",
	}
	start := ast.Position{Line: expect.Line, Col: expect.Start}

	anno := Position(expect.File, start, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestNChars(t *testing.T) {
	t.Parallel()

	n := 3
	expect := fileerr.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Line:         2,
		Start:        3,
		End:          3 + n,
		Annotation:   "anno",
	}
	start := ast.Position{Line: expect.Line, Col: expect.Start}

	anno := NChars(expect.File, start, n, expect.Annotation)
	assert.Equal(t, expect, anno)
	assert.Same(t, expect.File, anno.File)
}

func TestNode(t *testing.T) {

	t.Parallel()

	start := ast.Position{Line: 3, Col: 4} // to make things simpler

	testCases := []struct {
		node ast.Node
		end  int
	}{
		// handpicked selection
		{
			node: &ast.Ident{
				Ident:    "foo",
				Position: start,
			},
			end: start.Col + len("foo"),
		},
		{
			node: &ast.String{
				Open:  start,
				Quote: '"',
				Contents: []ast.StringContent{
					&ast.StringText{
						Text:     "foo",
						Position: delta(start, 1),
					},
				},
				Close: &ast.Position{Line: start.Line, Col: start.Col + len(`"foo`)},
			},
			end: start.Col + len(`"foo"`),
		},
		{
			node: &ast.State{
				LParen: &ast.Position{Line: start.Line, Col: start.Col + len("state ")},
				Vars: []ast.StateNode{
					&ast.StateVar{
						Names:  []*ast.Ident{{Ident: "foo", Position: ast.Position{Line: start.Line + 1, Col: 3}}},
						Assign: &ast.Position{Line: start.Line + 1, Col: 3 + len("foo ")},
						Values: []*ast.GoCode{
							{
								Expressions: []ast.GoCodeNode{
									&ast.RawGoCode{
										Code:     "bar",
										Position: ast.Position{Line: start.Line + 1, Col: 3 + len("foo = ")},
									},
								},
							},
						},
					},
				},
				RParen:   &ast.Position{Line: start.Line + 2, Col: 3},
				Position: start,
			},
			end: start.Col + len("state"),
		},
	}

	for _, c := range testCases {
		t.Run(fmt.Sprintf("%T", c.node), func(t *testing.T) {
			expect := fileerr.Annotation{
				File:         new(file.File),
				ContextStart: start.Line,
				ContextEnd:   start.Line + 1,
				Line:         start.Line,
				Start:        start.Col,
				End:          c.end,
				Annotation:   "anno",
			}

			anno := Node(expect.File, c.node, expect.Annotation)
			assert.Equal(t, expect, anno)
			assert.Same(t, expect.File, anno.File)
		})
	}
}
