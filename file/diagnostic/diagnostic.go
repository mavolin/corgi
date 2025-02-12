// Package diagnostic produces preformatted errors/warnings/etc
package diagnostic

import (
	"fmt"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
)

type (
	line = int
	col  = int

	// Diagnostic is an optionally annotated error.
	//
	// It must have a message.
	//
	// For most errors it is also conventional to have at least one primary
	// annotation and an explanation.
	//
	// The type and message serve as a header.
	//
	// Users mustn't preformat strings to have artificial line breaks.
	// The formatter will ensure line breaks are placed so that the error
	// conforms to a chosen width.
	// Instead, use line breaks as you would when writing prose; to indicate a
	// new paragraph or to separate different thoughts.
	//
	// Wrap code in backticks, e.g. `code`.
	Diagnostic struct {
		// Type of diagnostic, used in the header, e.g. "error", "warning",
		// "lint" etc.
		// The empty string is equivalent to "error".
		Type    Type
		Message string // e.g. "missing type"

		// Primary and Secondary annotations are used to highlight the error.
		//
		// Usually, you have one or more primary annotations that highlight
		// erroneous code, and possibly secondary annotations that add context,
		// e.g. an import statement relevant to the error.
		//
		// They are distinguished by the color of the annotation and the
		// character used to underline:
		// Primary annotation use the caret `^`, secondary annotations use the
		// tilde `~`.
		Primary   []Annotation
		Secondary []Annotation

		// Cause is the optional cause of the error.
		//
		// Rendered as:
		//  Cause: #{diagnostic.Cause.Error()}
		Cause error

		// Explanation is a longer explanation of the error.
		// Where annotations should give you the facts, the explanation should
		// give the required context to deeper understand the problem.
		//
		// Explanations are, however, not the place to give additional
		// information about the context surrounding the error.
		// After once reading the explanation, users should not have to read it
		// again if they encounter the same error elsewhere.
		// That, of course, doesn't mean an explanation can't restate
		// information from message and annotations to adapt it to the context.
		//
		// A good explanation is succinct and only present when necessary.
		// Suddenly seeing a huge block of text, for a missing semicolon is
		// just as discouraging as a complicated error with no explanation.
		//
		// Rendered directly below the annotation without title.
		Explanation string

		// Examples provide a short example of the correct usage.
		//
		// Use concrete names where appropriate (e.g. `class` instead of
		// `attribute`).
		// When using placeholders, feel free to use something
		// corgi/dog-related, e.g. `woof` and `bark`, over `foo` and `bar`.
		//
		// If there is more than one example, titles may be used to give the
		// user keywords to search for in the documentation to find more
		// information about an example.
		// Titles use in-sentence capitalization.
		//
		// Rendered as:
		//  Example: `.woof`
		//
		//  Example: `.woof` (class shorthand)
		//
		//  Examples: `value="woof"`, `async`, `.bark`
		//
		//  Examples: `value="foo"` (attribute with value)
		//            `async`       (boolean attribute)
		//            `.foo`        (class shorthand)
		Examples []Example

		// Hints provide information about common causes and solutions to the
		// error faced.
		//
		// They may also suggest conditional action like "Remove whitespace, if
		// there is any".
		//
		// Rendered as:
		//  Hint: Remove whitespace, if here is any
		//
		//  Hint: Remove whitespace, if here is any
		//        |> `. foo` -> `.foo`
		Hints []Hint

		// Docs is a relative path to the page in the documentation that
		// has further context.
		//
		// Use it only when appropriate; an unresolved component name is hardly
		// the place to link the documentation about components.
		// We can expect common sense.
		//
		// Rendered as:
		//  See: #{baseURL}/#{diagnostic.Docs}
		Docs string
	}

	Annotation struct {
		File *file.File
		// ContextStart and ContextEnd are the lines of input relevant to the
		// annotation, which are printed when the Pretty is called.
		//
		// Usually that is the line on which the error occurred, however, the
		// context may be larger if there is more relevant information to show.
		// For example a component argument error may choose to include the point when
		// the component was called.
		//
		// ContextStart is inclusive and ContextEnd is exclusive.
		ContextStart, ContextEnd line
		// Start and End specify the col range to be highlighted.
		//
		// Note that Start and End may exceed the actual line length.
		//
		// This is, for example, useful to highlight a missing token at the end of
		// a line.
		//
		// Start is inclusive and End is exclusive.
		Start, End ast.Position
		Annotation string // optional
	}

	Example struct {
		Title   string
		Example string
	}

	Hint struct {
		Hint    string
		Example string
	}
)

type Type string

const (
	Error   Type = "error"
	Warning Type = "warning"
	Lint    Type = "lint"
)

func (d *Diagnostic) Short() string {
	typ := d.Type
	if typ == "" {
		typ = Error
	}
	return fmt.Sprint(typ, ": ", d.Message)
}

func (d *Diagnostic) Error() string {
	if len(d.Primary) > 0 {
		f := d.Primary[0]
		return fmt.Sprint(f.File.PathInModule, ":", f.Start, ": ", d.Message)
	}

	return d.Message
}
