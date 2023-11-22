package corgierr

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/mavolin/corgi/file"
)

// As returns a [List] if [errors.As] yields a List or an [Error].
func As(err error) List {
	var errs []error
	if uw, ok := err.(interface{ Unwrap() []error }); ok {
		errs = uw.Unwrap()
	} else {
		errs = []error{err}
	}

	var list List

	for _, err := range errs {
		var lerr List
		if errors.As(err, &lerr) {
			list = append(list, lerr...)
			continue
		}

		var cerr *Error
		if errors.As(err, &cerr) {
			list = append(list, cerr)
		}
	}

	return list
}

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
	// For example a mixin argument error may choose to include the point when
	// the mixin was called.
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

	isError bool
}

type Suggestion struct {
	Suggestion string
	Example    string
	ShouldBe   string
	Code       string
}

func (err *Error) Error() string {
	if err.ErrorAnnotation.File == nil {
		if len(err.HintAnnotations) == 0 {
			return fmt.Sprintf("%d:%d: %s: %s",
				err.ErrorAnnotation.Line, err.ErrorAnnotation.Start, err.Message, err.ErrorAnnotation.Annotation)
		}
		return fmt.Sprintf("%d:%d: %s", err.ErrorAnnotation.Line, err.ErrorAnnotation.Start, err.Message)
	}

	if len(err.HintAnnotations) == 0 {
		return fmt.Sprintf("%s:%d:%d: %s: %s",
			err.ErrorAnnotation.File.Name, err.ErrorAnnotation.Line, err.ErrorAnnotation.Start,
			err.Message, err.ErrorAnnotation.Annotation)
	}
	return fmt.Sprintf("%s:%d:%d: %s",
		err.ErrorAnnotation.File.Name, err.ErrorAnnotation.Line, err.ErrorAnnotation.Start, err.Message)
}

func (err *Error) Unwrap() error {
	return err.Cause
}

type PrettyOptions struct {
	FileNamePrinter func(*file.File) string
	Colored         bool
}

func colored(sb *strings.Builder, o PrettyOptions, text string, attrs ...color.Attribute) {
	if o.Colored && len(attrs) > 0 {
		c := color.New(attrs...)
		c.EnableColor()
		c.Fprint(sb, text)
		return
	}

	sb.WriteString(text)
}

func coloredFunc(sb *strings.Builder, o PrettyOptions, f func(sb *strings.Builder), attrs ...color.Attribute) {
	if o.Colored && len(attrs) > 0 {
		color.New(attrs...).SetWriter(sb)

		defer color.New(attrs...).UnsetWriter(sb)
	}

	f(sb)
}

func (err *Error) Pretty(o PrettyOptions) string {
	if o.FileNamePrinter == nil {
		o.FileNamePrinter = func(f *file.File) string { return f.Name }
	}

	var sb strings.Builder

	err.prettyMessage(&sb, o)

	err.ErrorAnnotation.isError = true

	fileAnnotations := [][]Annotation{{err.ErrorAnnotation}}

	for _, ha := range err.HintAnnotations {
		for j, fas := range fileAnnotations {
			if equalFile(ha.File, fas[0].File) {
				fileAnnotations[j] = append(fas, ha)
			}
		}
	}

	for i, fas := range fileAnnotations {
		// yes this is slower, but also shorter (more readable) than in-place sorting
		// shouldn't matter when usually len(err.HintAnnotations) is no greater than one or two
		sort.Slice(fas, func(i, j int) bool {
			return fas[i].Line < fas[j].Line ||
				(fas[i].Line == fas[j].Line && fas[i].Start < fas[j].Start)
		})

		if i > 0 {
			sb.WriteByte('\n')
		}
		err.prettyFile(&sb, fas, i == 0, o)
	}

	if err.ShouldBe != "" {
		sb.WriteByte('\n')
		err.prettyHelpText(&sb, "should be", err.ShouldBe, o)
	}
	if err.Example != "" {
		sb.WriteByte('\n')
		err.prettyHelpText(&sb, "example", err.Example, o)
	}
	for i, sug := range err.Suggestions {
		sb.WriteByte('\n')
		title := "suggestion"
		if len(err.Suggestions) > 1 {
			title += " " + strconv.Itoa(i+1)
		}

		err.prettyHelpText(&sb, title, sug.Suggestion, o)

		if sug.Example != "" {
			sb.WriteByte('\n')
			err.prettyHelpText(&sb, "  -> example", sug.Example, o)
		}
		if sug.ShouldBe != "" {
			sb.WriteByte('\n')
			err.prettyHelpText(&sb, "  -> should be", sug.ShouldBe, o)
		}
		if sug.Code != "" {
			sb.WriteByte('\n')
			err.prettyHelpText(&sb, "  -> code", sug.Code, o)
		}
	}

	return sb.String()
}

func equalFile(a, b *file.File) bool {
	return a.Module == b.Module && a.PathInModule == b.PathInModule
}

func (err *Error) prettyMessage(sb *strings.Builder, o PrettyOptions) {
	colored(sb, o, "error: ", color.Bold, color.FgRed)

	err.prettyText(o, sb, err.Message, color.Bold)
	sb.WriteByte('\n')
}

func (err *Error) prettyFile(sb *strings.Builder, annotations []Annotation, errFile bool, o PrettyOptions) {
	lineRanges := printedLines(annotations)

	lineNoWidth := int(math.Log10(float64(lineRanges[len(lineRanges)-1].end)) + 1)

	noLineNoPad := strings.Repeat(" ", lineNoWidth)

	sb.WriteString(noLineNoPad)

	colored(sb, o, " > ", color.Faint)
	coloredFunc(sb, o, func(sb *strings.Builder) {
		sb.WriteString(o.FileNamePrinter(annotations[0].File))
		if errFile {
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(err.ErrorAnnotation.Line))
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(err.ErrorAnnotation.Start))
			sb.WriteByte('\n')
		}
	}, color.Bold, color.Faint)

	for i, lr := range lineRanges {
		if i > 0 {
			sb.WriteByte('\n')

			if lineRanges[i-1].end+1 == lr.start && annotations[0].File.Lines != nil {
				lineNo := lr.start - 1
				colored(sb, o, fmt.Sprintf("%*d | ", lineNoWidth, lineNo), color.Faint)
				sb.WriteString(annotations[0].File.Lines[lineNo-1])
			} else {
				sb.WriteString(noLineNoPad)
				colored(sb, o, " ...", color.Faint)
			}

			sb.WriteByte('\n')
		}

		err.prettyLineRange(sb, annotations, lr, lineNoWidth, o)
	}
}

func (err *Error) prettyLineRange(sb *strings.Builder, annotations []Annotation, lr lineRange, lineNoWidth int, o PrettyOptions) {
	noLinePad := strings.Repeat(" ", lineNoWidth)

	// for each context line
	for lineNo := lr.start; lineNo < lr.end; lineNo++ {
		if lineNo > lr.start {
			sb.WriteByte('\n')
		}

		colored(sb, o, fmt.Sprintf("%*d | ", lineNoWidth, lineNo), color.Faint)

		line := lr.lines[lineNo-lr.start]
		sb.WriteString(line)

		var lineAnnotations []Annotation
		for _, a := range annotations {
			if a.Line == lineNo {
				lineAnnotations = append(lineAnnotations, a)
			}
		}

		if len(lineAnnotations) == 0 {
			continue
		}

		sb.WriteByte('\n')
		sb.WriteString(noLinePad)
		colored(sb, o, " | ", color.Faint)

		// mark segments relevant in current line
		var offset int
		for _, la := range lineAnnotations {
			sb.WriteString(strings.Repeat(" ", la.Start-1-offset))
			offset += la.Start - 1 - offset

			endCol := la.End
			if endCol <= 0 {
				endCol = len(line) + 1
			} else if endCol < la.Start {
				endCol = la.Start + 1
			}

			repeatCount := endCol - la.Start
			offset += repeatCount

			if la.isError {
				colored(sb, o, strings.Repeat("^", repeatCount), color.Bold, color.FgRed)
			} else {
				colored(sb, o, strings.Repeat("~", repeatCount), color.Bold, color.FgCyan)
			}
		}

		var lastInline bool
		last := lineAnnotations[len(lineAnnotations)-1]
		renderedAnnotationLen := len(last.Annotation)
		if o.Colored {
			renderedAnnotationLen -= strings.Count(last.Annotation, "`")
		}
		// if we can comfortably fit the entire annotation behind the markers,
		// render the annotation in a single line to save space
		if !strings.Contains(last.Annotation, "\n") && len(noLinePad)+len(" | ")+last.End+renderedAnnotationLen < 100 {
			sb.WriteByte(' ')
			if last.isError {
				err.prettyText(o, sb, last.Annotation, color.Bold, color.FgRed)
			} else {
				err.prettyText(o, sb, last.Annotation, color.Bold, color.FgCyan)
			}
			lastInline = true
		}

		// start writing the rightmost annotation
		start := len(lineAnnotations) - 1
		if lastInline {
			start--
		}
		for i := start; i >= 0; i-- {
			la := lineAnnotations[i]

			annotationText := strings.Split(la.Annotation, "\n")
			for _, textLine := range annotationText {
				sb.WriteByte('\n')

				coloredFunc(sb, o, func(sb *strings.Builder) {
					sb.WriteString(noLinePad)
					sb.WriteString(" | ")
				}, color.Faint)

				offset := 0
				for _, otherLA := range lineAnnotations[:i] {
					sb.WriteString(strings.Repeat(" ", otherLA.Start-1-offset))

					if otherLA.isError {
						colored(sb, o, "|", color.Bold, color.FgRed)
					} else {
						colored(sb, o, "|", color.Bold, color.FgCyan)
					}

					offset += otherLA.Start - 1 - offset + len("|")
				}
				sb.WriteString(strings.Repeat(" ", la.Start-1-offset))

				if la.isError {
					colored(sb, o, "| ", color.Bold, color.FgRed)
					err.prettyText(o, sb, textLine, color.Bold, color.FgRed)
				} else {
					colored(sb, o, "| ", color.Bold, color.FgCyan)
					err.prettyText(o, sb, textLine, color.Bold, color.FgCyan)
				}
			}
		}
	}
}

func (err *Error) prettyHelpText(sb *strings.Builder, title, text string, o PrettyOptions) {
	colored(sb, o, title+": ", color.Bold, color.FgCyan)

	split := strings.Split(text, "\n")
	err.prettyText(o, sb, split[0])

	padding := strings.Repeat(" ", len(title+": "))
	for _, s := range split[1:] {
		sb.WriteByte('\n')
		sb.WriteString(padding)
		err.prettyText(o, sb, s)
	}
}

func (err *Error) prettyText(o PrettyOptions, sb *strings.Builder, text string, style ...color.Attribute) {
	if !o.Colored {
		sb.WriteString(text)
		return
	}

	normal := color.New(style...)
	normal.EnableColor()
	code := color.New(style...).Add(color.Italic)
	code.EnableColor()

	normal.SetWriter(sb)
	defer normal.UnsetWriter(sb)

	var inCode bool
	var lastBacktick bool
	for _, r := range []byte(text) {
		if r == '`' && !lastBacktick {
			if inCode {
				code.UnsetWriter(sb)
				normal.SetWriter(sb)
			} else {
				normal.UnsetWriter(sb)
				code.SetWriter(sb)
			}
			inCode = !inCode
			lastBacktick = true
		} else {
			sb.WriteByte(r)
			lastBacktick = false
		}
	}
}

type lineRange struct {
	start, end int
	lines      []string
}

// Reports the ranges of lines that are to be printed in the error.
func printedLines(as []Annotation) []lineRange {
	lines := make([]lineRange, len(as))
	for i, a := range as {
		lines[i] = lineRange{a.ContextStart, a.ContextEnd, a.Lines}
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].start < lines[j].start
	})

	for i := 0; i < len(lines)-1; i++ { // merge time
		a := lines[i]
		b := lines[i+1]

		if b.start <= a.end {
			lines[i] = lineRange{
				start: a.start,
				end:   b.end,
				lines: append(a.lines, b.lines[a.end-b.start:]...),
			} // merge a and b
			copy(lines[i+1:], lines[i+2:]) // remove b
			lines = lines[:len(lines)-1]
			i-- // new a is now possibly able to merge w new b
		}
	}

	return lines
}
