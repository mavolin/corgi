// Package diagnostic provides an abstraction for diagnostics such as errors
// or lints, as well as a formatter to render them in a human-friendly way.
package diagnostic

import (
	"strconv"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
)

type (
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
	Diagnostic struct { //nolint:errname
		// Type of diagnostic, used in the header, e.g. "error", "warning",
		// "lint" etc.
		// The empty string is equivalent to "error".
		Type Type
		// Message is the static message of the diagnostic.
		// It must not change based on the context of the error, i.e. it must
		// not include any information like file name, position, etc.
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
		//
		// Annotations must not overlap.
		Primary, Secondary []Annotation

		// Cause is the optional cause of the error.
		//
		// Rendered as:
		//  Cause: #{diagnostic.Cause.Error()}
		Cause error

		// Explanation is a longer explanation of the error.
		// Where annotations should give you the facts, the explanation should
		// give the required context to deeper understand the problem.
		//
		// Explanations are, however, not the place to give dynamically
		// generated information regarding the error.
		// After once reading the explanation of a diagnostic with the same
		// message, a user should not have to read it again if they encounter
		// the same error elsewhere.
		// That, of course, doesn't mean an explanation can't include
		// information from message and annotations to make it easier to
		// make the connection between the error and the explanation.
		// What it does mean, however, is that the explanation should remain
		// the same in its essence, for every error of the same type.
		// If you are in a situation where there are different explanations,
		// you should probably also use different error messages.
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
		// Values that start with "https://" are treated as absolute and will
		// be rendered as-is.
		// Otherwise, they are treated as a bang-path relative to the base URL
		// of the documentation.
		//
		// Defaults to "internal-error" for InternalErrors.
		// Set to " " to prevent this behavior.
		//
		// Rendered as:
		//  See: #{baseURL}/!#{diagnostic.Docs} (for relative URLs)
		//
		//  See: #{diagnostic.Docs} (for absolute URLs)
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
		ContextStart, ContextEnd ast.Line
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
	// An InternalError is an error with tooling instead of user code.
	// If present and Docs is unset, Docs defaults to "internal-error".
	InternalError Type = "internal error"
	Error         Type = "error"
	Warning       Type = "warning"
	Lint          Type = "lint"
)

func (d *Diagnostic) Short() string {
	typ := d.Type
	if typ == "" {
		typ = Error
	}

	return string(typ) + ": " + d.Error()
}

func (d *Diagnostic) Error() string {
	var sb strings.Builder
	if len(d.Primary) > 0 {
		p := d.Primary[0]
		sb.WriteString(string(p.File.PathInModule()))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(int(p.Start.Line)))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(int(p.Start.Col)))
		sb.WriteString(": ")
	}
	sb.WriteString(d.Message)
	if d.Cause != nil {
		sb.WriteString(": ")
		sb.WriteString(d.Cause.Error())
	}

	return sb.String()
}
