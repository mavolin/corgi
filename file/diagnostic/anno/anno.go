// Package anno provides helpers that reduce common boilerplate when creating
// annotations.
//
// Throughout the package, it assumed that supplied positions are valid
// positions in the file.
// If end > start or if start and end are on different lines even though they
// aren't allowed to be, start is preferred over end and a reasonable end is
// calculated.
package anno

import (
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

type Annotation struct {
	Context    ContextFunc
	Highlight  HighlightFunc
	Annotation string
}

func Anno(f *file.File, a Annotation) diagnostic.Annotation {
	c, h := a.Highlight(f)
	if a.Context != nil {
		c = a.Context(f, c, h)
	}
	return diagnostic.Annotation{
		File:         f,
		ContextStart: c.Start,
		ContextEnd:   c.End,
		Start:        h.Start,
		End:          h.End,
		Annotation:   a.Annotation,
	}
}

// Node is a shorthand for [HighlightNode].
func Node(f *file.File, n ast.Node, anno string) diagnostic.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightNode(n),
		Annotation: anno,
	})
}

// Range is a shorthand for [HighlightRange].
func Range(f *file.File, start, end ast.Position, anno string) diagnostic.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightRange(start, end),
		Annotation: anno,
	})
}

// ToEOL is a shorthand for [HighlightToEOL].
func ToEOL(f *file.File, start ast.Position, anno string) diagnostic.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightToEOL(start),
		Annotation: anno,
	})
}

// FirstWord is a shorthand for [HighlightFirstWord].
func FirstWord(f *file.File, start ast.Position, s, anno string) diagnostic.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightFirstWord(start, s),
		Annotation: anno,
	})
}

// Position is a shorthand for [HighlightPosition].
func Position(f *file.File, pos ast.Position, anno string) diagnostic.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightPosition(pos),
		Annotation: anno,
	})
}

// NRunes is a shorthand for [HighlightNRunes].
func NRunes(f *file.File, start ast.Position, n int, anno string) diagnostic.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightNRunes(start, n),
		Annotation: anno,
	})
}

// Node is a shorthand for [HighlightNode].
// func Node(f *file.File, itm ast.Node, anno string) diagnostic.Annotation {
// 	return Anno(f, Annotation{
// 		Highlight:  HighlightNode(itm),
// 		Annotation: anno,
// 	})
// }

type (
	// Context is the interval [start, end) of the lines which are to be
	// included in the annotation for context.
	// It must at least span the annotated line.
	Context struct {
		Start, End int
	}
	ContextFunc func(*file.File, Context, Highlight) Context
)

var InvalidContext = Context{0, 0}

// StaticContext returns a ContextFunc that always returns the given interval.
func StaticContext(start, end int) ContextFunc {
	return func(*file.File, Context, Highlight) Context { return Context{start, max(end, start+1)} }
}

// ContextNode is a shorthand for [ContextRange] that uses the start and end
// positions of the given node.
func ContextNode(n ast.Node) ContextFunc {
	return ContextRange(n.Start(), n.End())
}

// ContextRange returns a ContextFunc that includes the lines from start to end.
func ContextRange(start, end ast.Position) ContextFunc {
	return func(f *file.File, _ Context, _ Highlight) Context {
		start, end = normalizePos(f, start), normalizePos(f, end)
		s, e := start.Line, end.Line+1
		return Context{s, max(e, s+1)}
	}
}

// ContextDelta applies the given deltas to the interval returned by the
// HighlightFunc.
func ContextDelta(dStart, dEnd int) ContextFunc {
	return func(f *file.File, c Context, _ Highlight) Context {
		s := max(1, c.Start+dStart)
		e := c.End + dEnd
		return Context{min(s, len(f.AST.Lines)), min(max(e, s), len(f.AST.Lines)+1)}
	}
}

func singleLine(pos ast.Position) Context {
	return Context{pos.Line, pos.Line + 1}
}

type (
	// Highlight is the line and the column interval [start, end) representing
	// the area to be highlighted and annotated.
	Highlight struct {
		Start, End ast.Position
	}
	HighlightFunc func(*file.File) (Context, Highlight)
)

var InvalidHighlight = Highlight{ast.Position{}, ast.Position{}}

// HighlightNode highlights from the start to the end of the node.
func HighlightNode(n ast.Node) HighlightFunc {
	return HighlightRange(ast.Highlight(n))
}

// HighlightRange highlights the area in the interval [start, end).
// If start and end are on different lines, HighlightRange will highlight from
// start to EOL, and use [start, end) as the context.
func HighlightRange(start, end ast.Position) HighlightFunc {
	return func(f *file.File) (Context, Highlight) {
		start, end = normalizePos(f, start), normalizePos(f, end)
		return Context{start.Line, end.Line + 1}, Highlight{start, end}
	}
}

// HighlightToEOL highlights the area from start to the end of the line, but at least
// one character
func HighlightToEOL(start ast.Position) HighlightFunc {
	return func(f *file.File) (Context, Highlight) {
		return singleLine(start), toEOL(f, start)
	}
}

func toEOL(f *file.File, p ast.Position) Highlight {
	if p.Line < 1 || p.Line > len(f.AST.Lines) {
		return InvalidHighlight
	}

	e := max(len(f.AST.Lines[p.Line-1])+1, p.Col+1)
	return Highlight{p, ast.Position{Line: p.Line, Col: e}}
}

// HighlightFirstWord highlights the first word in s, as determined by the
// space, tab, carriage return and newline characters.
// If s is the empty string, a single char is highlighted.
// If s contains only a single word, it is highlighted in its entirety.
func HighlightFirstWord(start ast.Position, s string) HighlightFunc {
	length := len(s)
	if length == 0 {
		return HighlightPosition(start)
	}
	for i, b := range []byte(s[1:]) {
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			length = i
		}
	}
	return HighlightNRunes(start, length)
}

// HighlightPosition is a shorthand for HighlightNRunes(start, 1).
func HighlightPosition(start ast.Position) HighlightFunc {
	return HighlightNRunes(start, 1)
}

// HighlightNRunes highlights the area from start to start+n.
// n must be at least 1.
// n must be specified in bytes.
func HighlightNRunes(start ast.Position, n int) HighlightFunc {
	if n <= 0 {
		panic("anno: HighlightNRunes: n must be at least 1")
	}
	return func(*file.File) (Context, Highlight) {
		return singleLine(start), Highlight{start, ast.Position{Line: start.Line, Col: start.Col + n}}
	}
}

func normalizePos(f *file.File, pos ast.Position) ast.Position {
	if pos.Col != 0 {
		return pos
	}
	pos.Line--
	pos.Col = len(f.AST.Lines[pos.Line-1]) + 1
	return pos
}
