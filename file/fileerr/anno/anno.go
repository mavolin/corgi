// Package anno provides helpers for creating annotations.
//
// Throughout the package, it assumed that supplied positions are valid positions
// in the file.
// If end > start or if start and end are on different lines even though they
// aren't allowed to be, start is preferred over end and a reasonable end is
// calculated.
package anno

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
)

type Annotation struct {
	Context    ContextFunc
	Highlight  HighlightFunc
	Annotation string
}

func Anno(f *file.File, a Annotation) fileerr.Annotation {
	c, h := a.Highlight(f)
	if a.Context != nil {
		c = a.Context(f, c, h)
	}
	return fileerr.Annotation{
		File:         f,
		ContextStart: c.Start,
		ContextEnd:   c.End,
		Line:         h.Line,
		Start:        h.Start,
		End:          h.End,
		Annotation:   a.Annotation,
	}
}

// Range is a shorthand for [HighlightRange].
func Range(f *file.File, start, end ast.Position, anno string) fileerr.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightRange(start, end),
		Annotation: anno,
	})
}

// ToEOL is a shorthand for [HighlightToEOL].
func ToEOL(f *file.File, start ast.Position, anno string) fileerr.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightToEOL(start),
		Annotation: anno,
	})
}

// Position is a shorthand for NChars(f, pos, 1, anno).
func Position(f *file.File, pos ast.Position, anno string) fileerr.Annotation {
	return NChars(f, pos, 1, anno)
}

// NChars is a shorthand for [HighlightNChars].
func NChars(f *file.File, start ast.Position, n int, anno string) fileerr.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightNChars(start, n),
		Annotation: anno,
	})
}

// Node is a shorthand for [HighlightNode].
func Node(f *file.File, itm ast.Node, anno string) fileerr.Annotation {
	return Anno(f, Annotation{
		Highlight:  HighlightNode(itm),
		Annotation: anno,
	})
}

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

// ContextLines returns a ContextFunc that includes the lines from start to end.
func ContextLines(start, end ast.Position) ContextFunc {
	return func(f *file.File, _ Context, _ Highlight) Context {
		start, end = normalizePos(f, start), normalizePos(f, end)
		s, e := start.Line, end.Line+1
		return Context{s, max(e, s+1)}
	}
}

// ContextDelta applies the given deltas to the interval returned by the
// HighlightFunc.
func ContextDelta(dStart, dEnd int) ContextFunc {
	return func(_ *file.File, c Context, _ Highlight) Context {
		s := max(1, c.Start+dStart)
		e := c.End + dEnd
		return Context{s, max(e, s)}
	}
}

func singleLine(pos ast.Position) Context {
	return Context{pos.Line, pos.Line + 1}
}

type (
	// Highlight is the line and the column interval [start, end) representing
	// the area to be highlighted and annotated.
	Highlight struct {
		Line       int
		Start, End int
	}
	HighlightFunc func(*file.File) (Context, Highlight)
)

var InvalidHighlight = Highlight{0, 0, 0}

// HighlightRange highlights the area in the interval [start, end).
// If start and end are on different lines, HighlightRange will highlight from start to
// EOL, and use [start, end) as the context.
func HighlightRange(start, end ast.Position) HighlightFunc {
	return func(f *file.File) (Context, Highlight) {
		start, end = normalizePos(f, start), normalizePos(f, end)
		if start.Line != end.Line {
			return Context{start.Line, end.Line}, toEOL(f, start)
		}

		return singleLine(start), Highlight{start.Line, start.Col, end.Col}
	}
}

// HighlightToEOL highlights the area from start to the end of the line, but at least
// one character
func HighlightToEOL(start ast.Position) HighlightFunc {
	return func(f *file.File) (Context, Highlight) {
		return singleLine(start), toEOL(f, start)
	}
}

func toEOL(f *file.File, start ast.Position) Highlight {
	if start.Line < 1 || start.Line > len(f.Lines) {
		return InvalidHighlight
	}

	s := start.Col
	e := max(len(f.Lines[start.Line-1])+1, s+1)
	return Highlight{start.Line, s, e}
}

// HighlightNChars highlights the area from start to start+n.
// n must be at least 1.
func HighlightNChars(start ast.Position, n int) HighlightFunc {
	if n <= 0 {
		panic("anno: HighlightNChars: n must be at least 1")
	}
	return func(*file.File) (Context, Highlight) {
		return singleLine(start), Highlight{start.Line, start.Col, start.Col + n}
	}
}

// HighlightNode highlights an opinionated area of the passed node and uses the
// start and end of the node as context.
func HighlightNode(node ast.Node) HighlightFunc {
	return func(f *file.File) (Context, Highlight) {
		start, end := normalizePos(f, node.Pos()), normalizePos(f, node.End())
		c, h := Context{start.Line, end.Line + 1}, Highlight{start.Line, start.Col, end.Col}

		switch node := node.(type) {
		// === component.go ===
		case *ast.Component:
			if node.RParen != nil && h.Line == node.RParen.Line {
				h.End = node.RParen.Col + len(")")
			} else if node.RBracket != nil && h.Line == node.RBracket.Line {
				h.End = node.RBracket.Col + len("]")
			} else if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			} else {
				h.End = h.Start + len("comp")
			}
		case *ast.ComponentCall:
			if node.RParen != nil && h.Line == node.RParen.Line {
				h.End = node.RParen.Col + len(")")
			} else if node.RBracket != nil && h.Line == node.RBracket.Line {
				h.End = node.RBracket.Col + len("]")
			} else if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			} else {
				h.End = h.Start + len("+")
			}
		}

		if start.Line == end.Line {
			return c, h
		}

		switch node := node.(type) {
		// === attribute.go ===
		case *ast.And:
			h.End = h.Start + len("&")
		case *ast.AttributeList:
			h.End = h.Start + len("(")
		case *ast.SimpleAttribute:
			h.End = h.Start + len(node.Name)
		case *ast.TypedAttributeValue:
			if node.LParen != nil {
				h.End = node.LParen.Col + 1
			} else {
				h.End = h.Start + len(node.Type.String())
			}
		case *ast.ComponentCallAttributeValue:
			return HighlightNode(node.ComponentCall)(f)

		// === base.go ===
		case *ast.StaticString:
			h.End = h.Start + len(`"`)

		// === body.go ===
		case *ast.Scope:
			h.End = h.Start + len("{")
		case *ast.BadNode:
			h = toEOL(f, node.Position)

		// === code.go ===
		case *ast.Code:
			if node.Implicit {
				h = toEOL(f, node.Position)
			} else {
				h.End = h.Start + len("-")
			}
		case *ast.Return:
			h.End = h.Start + len("return")
		case *ast.Break:
			h.End = h.Start + len("break")
		case *ast.Continue:
			h.End = h.Start + len("continue")

		// === component.go ===
		case *ast.Component:
			if node.RParen != nil && h.Line == node.RParen.Line {
				h.End = node.RParen.Col + len(")")
			} else if node.RBracket != nil && h.Line == node.RBracket.Line {
				h.End = node.RBracket.Col + len("]")
			} else if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			} else {
				h.End = h.Start + len("comp")
			}
		case *ast.TypeParam:
			if len(node.Names) > 0 {
				h.End = node.Names[0].End().Col
			}
		case *ast.ComponentParam:
			if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			}
		case *ast.ComponentCall:
			if node.RParen != nil && h.Line == node.RParen.Line {
				h.End = node.RParen.Col + len(")")
			} else if node.RBracket != nil && h.Line == node.RBracket.Line {
				h.End = node.RBracket.Col + len("]")
			} else if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			} else {
				h.End = h.Start + len("+")
			}
		case *ast.ComponentArg:
			if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			}
		case *ast.Block:
			if node.Name != nil && h.Line == node.Name.Position.Line {
				h.End = node.Name.End().Col
			}
		case *ast.UnderscoreBlockShorthand:
			h.End = h.Start + len("_{")

		// === control_structures.go ===
		case *ast.If:
			if node.Header != nil && h.Line == node.Header.End().Line {
				h.End = node.Header.End().Col
			} else {
				h.End = h.Start + len("if")
			}
		case *ast.ElseIf:
			if node.Header != nil && h.Line == node.Header.End().Line {
				h.End = node.Header.End().Col
			} else {
				h.End = h.Start + len("else if")
			}
		case *ast.Else:
			h.End = h.Start + len("else")
		case *ast.IfHeader:
			h = toEOL(f, start)
		case *ast.Switch:
			if node.Comparator != nil && h.Line == node.Comparator.End().Line {
				h.End = node.Comparator.End().Col
			} else {
				h.End = h.Start + len("switch")
			}
		case *ast.Case:
			if node.Colon != nil && h.Line == node.Colon.Line {
				h.End = node.Colon.Col + 1
			} else if node.Expression != nil && h.Line == node.Expression.End().Line {
				h.End = node.Expression.End().Col
			} else {
				h.End = h.Start + len("case")
			}
		case *ast.For:
			if node.Header != nil && h.Line == node.Header.End().Line {
				h.End = node.Header.End().Col
			} else {
				h.End = h.Start + len("for")
			}
		case *ast.ForRangeHeader:
			if node.Range != nil {
				h = Highlight{node.Range.Line, node.Range.Col, node.Range.Col + len("range")}
			}
			if node.Expression != nil && h.Line == node.Expression.End().Line {
				h.End = node.Expression.End().Col
			}

		// === element.go ===
		case *ast.Element:
			if len(node.Attributes) > 0 && h.Line == node.Attributes[len(node.Attributes)-1].End().Line {
				h.End = node.Attributes[len(node.Attributes)-1].End().Col
			} else {
				h.End = h.Start + len(node.Name)
				if node.Void {
					h.End++
				}
			}
		case *ast.RawElement:
			h.End = h.Start + len("!raw")

		// === expression.go ===
		case *ast.ChainExpression:
			if len(node.Chain) > 0 && h.Line == node.Chain[len(node.Chain)-1].End().Line {
				h.End = node.Chain[len(node.Chain)-1].End().Col
			} else if node.Root != nil && h.Line == node.Root.End().Line {
				h.End = node.Root.End().Col
				if node.CheckRoot {
					h.End++
				}
			}
		case *ast.IndexExpression:
			h.End = h.Start + len("[")
		case *ast.DotIdentExpression:
			h.End = h.Start + len(".")
		case *ast.ParenExpression:
			h.End = h.Start + len("(")
		case *ast.TypeAssertionExpression:
			h.End = h.Start + len(".(")

		// === go_code.go ===
		case *ast.GoCode:
			if len(node.Expressions) > 0 {
				h = toEOL(f, node.Expressions[0].Pos())
			}
		case *ast.RawGoCode:
			h = toEOL(f, node.Position)
		case *ast.BlockFunction:
			if node.LParen != nil && h.Line == node.LParen.Line {
				h.End = node.LParen.Col + 1
			} else {
				h.End = h.Start + len("block")
			}
		case *ast.String:
			h.End = h.Start + len(`"`)
		case *ast.StringText:
			h = toEOL(f, node.Position)

		// === preamble.go ===
		case *ast.PackageDirective:
			h.End = h.Start + len("package")
		case *ast.Import:
			h.End = h.Start + len("import")
		case *ast.ImportSpec:
			if node.Path != nil {
				return HighlightNode(node.Path)(f)
			} else if node.Alias != nil {
				h.End = h.Start + node.Alias.End().Col
			}
		case *ast.State:
			h.End = h.Start + len("state")
		case *ast.StateVar:
			if len(node.Names) > 0 {
				h = toEOL(f, node.Names[0].Position)
			}

		// === text.go ===
		case *ast.ArrowBlock:
			h = toEOL(f, node.Position)
		case *ast.BracketText:
			h.End = h.Start + len("[")
		case *ast.InterpolationValue:
			h.End = h.Start + len("[")
		}

		return c, h
	}
}

// Delta applies the passed deltas to the original highlighted area.
// Delta does not allow the change of lines, and will set start to 1,
// if dStart is too small.
func (f HighlightFunc) Delta(dStart, dEnd int) HighlightFunc {
	return func(file *file.File) (Context, Highlight) {
		c, h := f(file)

		s := max(h.Start+dStart, 1)
		e := max(h.End+dEnd, s+1)
		return c, Highlight{h.Line, s, e}
	}
}

func delta(pos ast.Position, delta int) ast.Position {
	pos.Col += delta
	return pos
}

func normalizePos(f *file.File, pos ast.Position) ast.Position {
	if pos.Col != 0 {
		return pos
	}
	pos.Line--
	pos.Col = len(f.Lines[pos.Line-1]) + 1
	return pos
}
