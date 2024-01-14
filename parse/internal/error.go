package internal

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileerr"
)

func invalidIdent(c *current, name string, start file.Position, ident string) (file.Ident, error) {
	return file.Ident{
			Ident:    ident,
			Position: start,
		}, &fileerr.Error{
			Message: name + ": invalid name",
			ErrorAnnotation: anno(c, annotation{
				Start:      start,
				Len:        len(ident),
				Annotation: "this is not a valid identifier",
			}),
			ShouldBe: "a letter, or `_`, optionally followed by `_`, letters, and numbers",
		}
}

func missingIdent(c *current, name, example string, startOffset int) (file.Ident, error) {
	return file.Ident{Position: pos(c)}, &fileerr.Error{
		Message: name + ": missing name",
		ErrorAnnotation: anno(c, annotation{
			Start:       pos(c),
			StartOffset: startOffset,
			Annotation:  "expected the name of the " + name,
		}),
		Example: example,
	}
}

func unclosedList(c *current, listName string) (file.Position, error) {
	return pos(c), &fileerr.Error{
		Message: listName + ": unclosed `(` or missing `,`",
		ErrorAnnotation: anno(c, annotation{
			Start:      pos(c),
			Annotation: "expected a `,` or `)`",
		}),
		HintAnnotations: []fileerr.Annotation{
			anno(c, annotation{
				Start:      popStart(c),
				Annotation: "for the `(` you opened here",
			}),
		},
	}
}

func unclosedIndex(c *current, listName string) (file.Position, error) {
	return pos(c), &fileerr.Error{
		Message: listName + ": unclosed `[` or missing `,`",
		ErrorAnnotation: anno(c, annotation{
			Start:      pos(c),
			Annotation: "expected a `,` or `]`",
		}),
		HintAnnotations: []fileerr.Annotation{
			anno(c, annotation{
				Start:      popStart(c),
				Annotation: "for the `[` you opened here",
			}),
		},
	}
}

func unclosedParen(c *current, open, close string) (file.GoCodeItem, error) {
	start := popStart(c)
	return nil, &fileerr.Error{
		Message: "go code: unclosed `" + open + "`",
		ErrorAnnotation: anno(c, annotation{
			ContextLen: 3,
			Start:      start,
			Annotation: "expected a `" + close + "` for this `" + open + "`",
		}),
	}
}

func newUnexpectedTokensErr(c *current, start, end file.Position, errAnno string) *fileerr.Error {
	return &fileerr.Error{
		Message: "unexpected tokens",
		ErrorAnnotation: anno(c, annotation{
			Start:      start,
			End:        end,
			EndOffset:  -1,
			Annotation: errAnno,
		}),
	}
}