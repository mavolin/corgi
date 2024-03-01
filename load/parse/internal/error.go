package internal

import (
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/fileerr/anno"
)

func invalidIdent(c *current, name string, start ast.Position, ident string) (*ast.Ident, error) {
	return &ast.Ident{
			Ident:    ident,
			Position: start,
		}, &fileerr.Error{
			Message: name + ": invalid name",
			ErrorAnnotation: anno.Anno(file(c), anno.Annotation{
				Highlight:  anno.HighlightNChars(start, len(ident)),
				Annotation: "this is not a valid identifier",
			}),
			ShouldBe: "a letter, or `_`, optionally followed by `_`, letters, and numbers",
		}
}

func missingIdent(c *current, name, example string, startOffset int) (*ast.Ident, error) {
	return &ast.Ident{Position: pos(c)}, &fileerr.Error{
		Message: name + ": missing name",
		ErrorAnnotation: anno.Anno(file(c), anno.Annotation{
			Highlight:  anno.HighlightNChars(pos(c), startOffset),
			Annotation: "expected the name of the " + name,
		}),
		Example: example,
	}
}

func unclosedList(c *current, listName string) (ast.Position, error) {
	return pos(c), &fileerr.Error{
		Message:         listName + ": unclosed `(` or missing `,`",
		ErrorAnnotation: anno.Position(file(c), pos(c), "expected a `,` or `)`\""),
		HintAnnotations: []fileerr.Annotation{
			anno.Position(file(c), popStart(c), "for the `(` you opened here"),
		},
	}
}

func unclosedIndex(c *current, listName string) (ast.Position, error) {
	return pos(c), &fileerr.Error{
		Message:         listName + ": unclosed `[` or missing `,`",
		ErrorAnnotation: anno.Position(file(c), pos(c), "expected a `,` or `]`"),
		HintAnnotations: []fileerr.Annotation{
			anno.Position(file(c), popStart(c), "for the `[` you opened here"),
		},
	}
}

func unclosedParen(c *current, open, close string) (*ast.GoCodeNode, error) {
	start := popStart(c)
	return nil, &fileerr.Error{
		Message: "go code: unclosed `" + open + "`",
		ErrorAnnotation: anno.Anno(file(c), anno.Annotation{
			Context:    anno.ContextDelta(0, 3),
			Highlight:  anno.HighlightPosition(start),
			Annotation: "expected a `" + close + "` for this `" + open + "`",
		}),
	}
}

func newUnexpectedTokensErr(c *current, start, end ast.Position, errAnno string) *fileerr.Error {
	return &fileerr.Error{
		Message:         "unexpected tokens",
		ErrorAnnotation: anno.Range(file(c), start, end, errAnno),
	}
}
