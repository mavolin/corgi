package fileerr

import (
	"strconv"
	"strings"

	"github.com/mavolin/corgi/file"
)

type Error struct {
	Message string

	ErrorAnnotation Annotation
	HintAnnotations []Annotation

	Example     string
	ShouldBe    string
	Suggestions []Suggestion

	// Cause is the cause of the error, if it has one
	Cause error
}

type Annotation struct {
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
	ContextStart, ContextEnd int
	// Line is the line of the annotation.
	// It must lie between ContextStart and ContextEnd.
	Line int
	// Start and End specify the col range to be highlighted.
	//
	// Note that Start and End may exceed the actual line length.
	//
	// This is, for example, useful to highlight a missing token at the end of
	// a line.
	//
	// Start is inclusive and End is exclusive.
	Start, End int
	Annotation string

	// Lines are the lines that are annotated, starting with the line with the
	// number ContextStart.
	Lines []string
}

type Suggestion struct {
	Suggestion string
	Example    string
	ShouldBe   string
	Code       string
}

func (err *Error) Error() string {
	var sb strings.Builder
	sb.Grow(len(err.Message) + 100)

	if err.ErrorAnnotation.File != nil {
		sb.WriteString(err.ErrorAnnotation.File.AbsolutePath)
		sb.WriteByte(':')
	}

	if err.ErrorAnnotation.Line != 0 {
		sb.WriteString(strconv.Itoa(err.ErrorAnnotation.Line))
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(err.ErrorAnnotation.Start))
		sb.WriteString(": ")
	} else {
		sb.WriteByte(' ')
	}

	sb.WriteString(err.Message)
	return sb.String()
}

func (err *Error) Unwrap() error {
	return err.Cause
}

type PrettyOptions struct {
	FileNamePrinter func(*file.File) string
	Colored         bool
}

func (o *PrettyOptions) setDefaults() {
	if o.FileNamePrinter == nil {
		o.FileNamePrinter = func(f *file.File) string {
			if f == nil {
				return "nil file"
			}
			return f.Name
		}
	}
}

func (err *Error) Pretty(o PrettyOptions) string {
	return newPrettyPrinter(err, o).print()
}
