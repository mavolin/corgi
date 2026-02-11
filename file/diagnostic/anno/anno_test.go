package anno

import (
	"fmt"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestAnno(t *testing.T) {
	t.Parallel()
	t.Run("only highlight", func(t *testing.T) {
		t.Parallel()

		wantFile := new(file.File)
		wantHighlight := Highlight{
			Start: ast.Position{Line: 2, Col: 3},
			End:   ast.Position{Line: 2, Col: 4},
		}
		wantContext := Context{
			Start: 1,
			End:   5,
		}
		wantAnno := "annotation"

		anno := Anno(wantFile, Annotation{
			Highlight: func(f *file.File) (Context, Highlight) {
				should.True(t, wantFile == f)
				return wantContext, wantHighlight
			},
			Annotation: wantAnno,
		})
		should.True(t, wantFile == anno.File)                 // file
		should.Equal(t, anno.ContextStart, wantContext.Start) // context start
		should.Equal(t, anno.ContextEnd, wantContext.End)     // context end
		should.Equal(t, anno.Start, wantHighlight.Start)      // highlight start
		should.Equal(t, anno.End, wantHighlight.End)          // highlight end
		should.Equal(t, anno.Annotation, wantAnno)            // annotation
	})
	t.Run("highlight and context", func(t *testing.T) {
		t.Parallel()

		wantFile := new(file.File)
		wantHighlight := Highlight{
			Start: ast.Position{Line: 2, Col: 3},
			End:   ast.Position{Line: 2, Col: 4},
		}
		wantContext := Context{
			Start: 1,
			End:   5,
		}
		otherContext := Context{
			Start: 0,
			End:   6,
		}
		wantAnno := "annotation"

		anno := Anno(wantFile, Annotation{
			Context: func(f *file.File, context Context, highlight Highlight) Context {
				should.True(t, wantFile == f)
				should.Equal(t, context, otherContext)
				should.Equal(t, highlight, wantHighlight)
				return wantContext
			},
			Highlight: func(f *file.File) (Context, Highlight) {
				should.True(t, wantFile == f)
				return otherContext, wantHighlight
			},
			Annotation: wantAnno,
		})
		should.Equal(t, anno.File, wantFile)
		should.Equal(t, anno.ContextStart, wantContext.Start)
		should.Equal(t, anno.ContextEnd, wantContext.End)
		should.Equal(t, anno.Start, wantHighlight.Start)
		should.Equal(t, anno.End, wantHighlight.End)
		should.Equal(t, anno.Annotation, wantAnno)
	})
}

func TestRange(t *testing.T) {
	t.Parallel()

	want := diagnostic.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: 5},
		Annotation:   "anno",
	}

	anno := Range(want.File, want.Start, want.End, want.Annotation)
	should.Equal(t, anno, want)
	should.True(t, want.File == anno.File)
}

func TestToEOL(t *testing.T) {
	t.Parallel()

	wantFile := &file.File{
		Lines: []string{
			"foo",
			"foobar",
			"bar",
		},
	}
	want := diagnostic.Annotation{
		File:         wantFile,
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: len(wantFile.Lines[1]) + 1},
		Annotation:   "anno",
	}

	anno := ToEOL(want.File, want.Start, want.Annotation)
	should.Equal(t, anno, want)
	should.True(t, want.File == anno.File)
}

func TestPosition(t *testing.T) {
	t.Parallel()

	want := diagnostic.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: 4},
		Annotation:   "anno",
	}

	anno := Position(want.File, want.Start, want.Annotation)
	should.Equal(t, anno, want)
	should.True(t, want.File == anno.File)
}

func TestNChars(t *testing.T) {
	t.Parallel()

	n := 3
	want := diagnostic.Annotation{
		File:         new(file.File),
		ContextStart: 2,
		ContextEnd:   3,
		Start:        ast.Position{Line: 2, Col: 3},
		End:          ast.Position{Line: 2, Col: 3 + n},
		Annotation:   "anno",
	}

	anno := NRunes(want.File, want.Start, n, want.Annotation)
	should.Equal(t, anno, want)
	should.True(t, want.File == anno.File)
}

func TestNode(t *testing.T) {
	t.Parallel()

	start := ast.Position{Line: 3, Col: 4} // to make things simpler

	tests := []struct {
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
		},
	}

	for _, c := range tests {
		t.Run(fmt.Sprintf("%T", c.node), func(t *testing.T) {
			t.Parallel()

			want := diagnostic.Annotation{
				File:         new(file.File),
				ContextStart: c.node.Start().Line,
				ContextEnd:   c.node.End().Line + 1,
				Start:        c.node.Start(),
				End:          c.node.End(),
				Annotation:   "anno",
			}

			anno := Node(want.File, c.node, want.Annotation)
			should.Equal(t, anno, want)
		})
	}
}
