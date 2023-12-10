package fileerr

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mavolin/corgi/file"
)

type (
	prettyPrinter struct {
		sb *strings.Builder
		o  PrettyOptions

		err *Error
		// fileAnnotations is an ordered list of annotations per file.
		//
		// files are sorted in the order they appear in the error list, annotations
		// are sorted by line number.
		fileAnnotations [][]annotation

		// updated per file being printed
		maxLineDigits int
	}
	annotation struct {
		Annotation
		isError bool
	}
)

func newPrettyPrinter(err *Error, o PrettyOptions) *prettyPrinter {
	fileAnnotations := [][]annotation{{{Annotation: err.ErrorAnnotation, isError: true}}}

	for _, ha := range err.HintAnnotations {
		for j, fas := range fileAnnotations {
			if equalFile(ha.File, fas[0].File) {
				fileAnnotations[j] = append(fas, annotation{Annotation: ha})
			}
		}
		fileAnnotations = append(fileAnnotations, []annotation{{Annotation: ha}})
	}

	for _, fas := range fileAnnotations {
		sort.Slice(fas, func(i, j int) bool {
			return fas[i].Line < fas[j].Line ||
				(fas[i].Line == fas[j].Line && fas[i].Start < fas[j].Start)
		})
	}

	return &prettyPrinter{
		sb:              new(strings.Builder),
		o:               o,
		err:             err,
		fileAnnotations: fileAnnotations,
	}
}

func (p *prettyPrinter) print() string {
	// This error is 'lineless', so no need to print anything else.
	if p.printMessage() {
		return p.sb.String()
	}

	for i, fas := range p.fileAnnotations {
		if i > 0 {
			p.sb.WriteByte('\n')
		}
		p.printFile(fas)
	}

	if p.err.ShouldBe != "" {
		p.sb.WriteByte('\n')
		p.printHelpText("should be", p.err.ShouldBe)
	}
	if p.err.Example != "" {
		p.sb.WriteByte('\n')
		p.printHelpText("example", p.err.Example)
	}
	for i, sug := range p.err.Suggestions {
		p.sb.WriteByte('\n')
		title := "suggestion"
		if len(p.err.Suggestions) > 1 {
			title += " " + strconv.Itoa(i+1)
		}

		p.printHelpText(title, sug.Suggestion)

		if sug.Example != "" {
			p.sb.WriteByte('\n')
			p.printHelpText("  -> example", sug.Example)
		}
		if sug.ShouldBe != "" {
			p.sb.WriteByte('\n')
			p.printHelpText("  -> should be", sug.ShouldBe)
		}
		if sug.Code != "" {
			p.sb.WriteByte('\n')
			p.printHelpText("  -> code", sug.Code)
		}
	}

	return p.sb.String()
}

// Prints the header with the error message.
// If the err has no line information, the message is prefixed by the file name
// and printMessage returns true.
//
// Otherwise, printMessage returns false.
func (p *prettyPrinter) printMessage() bool {
	p.colored("error: ", color.Bold, color.FgRed)

	if p.err.ErrorAnnotation.Line == 0 {
		p.colored(p.o.FileNamePrinter(p.err.ErrorAnnotation.File), color.Bold)
	}
	p.printText(p.err.Message, color.Bold)

	if p.err.ErrorAnnotation.Line != 0 {
		p.sb.WriteByte('\n')
		return false
	}

	return true
}

func (p *prettyPrinter) printFile(annotations []annotation) {
	lineRanges := lineRanges(annotations)

	// digits in the biggest line number
	p.maxLineDigits = numDigits(lineRanges[len(lineRanges)-1].end)

	// if the file being printed contains the error annotation
	var errFile bool
	for _, a := range annotations {
		if a.isError {
			errFile = true
			break
		}
	}

	p.sb.WriteString(strings.Repeat(" ", p.maxLineDigits))

	p.colored(" > ", color.Faint)
	p.coloredFunc(func() {
		p.sb.WriteString(p.o.FileNamePrinter(annotations[0].File))
		if errFile {
			p.sb.WriteByte(':')
			p.sb.WriteString(strconv.Itoa(p.err.ErrorAnnotation.Line))
			p.sb.WriteByte(':')
			p.sb.WriteString(strconv.Itoa(p.err.ErrorAnnotation.Start))
			p.sb.WriteByte('\n')
		}
	}, color.Bold, color.Faint)

	for i, lr := range lineRanges {
		if i > 0 {
			p.sb.WriteByte('\n')

			// check if the end of the previous annotation and the start of the
			// current annotation are only a single line apart
			// if so, and we have access to all lines, print the line in between
			// instead of `...`
			if lineRanges[i-1].end+1 == lr.start && annotations[0].File.Lines != nil {
				lineNo := lr.start - 1
				p.printLineStart(lineNo)
				p.sb.WriteString(annotations[0].File.Lines[lineNo-1])
			} else {
				p.sb.WriteString(strings.Repeat(" ", p.maxLineDigits))
				p.colored(" ...", color.Faint)
			}

			p.sb.WriteByte('\n')
		}

		p.printLineRange(annotations, lr)
	}
}

func (p *prettyPrinter) printLineRange(annotations []annotation, lr lineRange) {
	// for each context line
	for lineNo := lr.start; lineNo < lr.end; lineNo++ {
		if lineNo > lr.start {
			p.sb.WriteByte('\n')
		}

		p.printLineStart(lineNo)

		line := lr.lines[lineNo-lr.start]
		p.sb.WriteString(line)

		var lineAnnotations []annotation
		for _, a := range annotations {
			if a.Line == lineNo {
				lineAnnotations = append(lineAnnotations, a)
			}
		}

		if len(lineAnnotations) == 0 {
			continue
		}

		p.sb.WriteByte('\n')
		p.printLineStart(-1)

		p.printAnnotationMarkers(lineAnnotations)

		last := lineAnnotations[len(lineAnnotations)-1]

		// if we can comfortably fit the entire annotation behind the markers,
		// render the annotation in a single line to save space
		lastInline := p.shouldInline(last)
		if lastInline {
			p.sb.WriteByte(' ')
			p.printText(last.Annotation.Annotation, color.Bold, p.annoColor(last))
		}

		if lastInline {
			p.printAnnotations(lineAnnotations[:len(lineAnnotations)-1])
		} else {
			p.printAnnotations(lineAnnotations)
		}

	}
}

func (p *prettyPrinter) printAnnotationMarkers(lineAnnotations []annotation) {
	// mark segments relevant in current line
	var offset int
	for _, la := range lineAnnotations {
		numSpaces := la.Start - 1 - offset
		p.sb.WriteString(strings.Repeat(" ", numSpaces))
		offset += numSpaces

		offset += p.printAnnotationMarker(la)
	}
}

func (p *prettyPrinter) printAnnotations(as []annotation) {
	// start writing the rightmost annotation
	for i := len(as) - 1; i >= 0; i-- {
		a := as[i]

		lines := strings.Split(a.Annotation.Annotation, "\n")
		for _, textLine := range lines {
			p.sb.WriteByte('\n')
			p.printLineStart(-1)

			offset := 0
			for _, otherLA := range as[:i] {
				p.sb.WriteString(strings.Repeat(" ", otherLA.Start-1-offset))
				p.colored("|", color.Bold, p.annoColor(otherLA))

				offset += otherLA.Start - 1 - offset + len("|")
			}
			p.sb.WriteString(strings.Repeat(" ", a.Start-1-offset))

			p.colored("| ", color.Bold, p.annoColor(a))
			p.printText(textLine, color.Bold, p.annoColor(a))
		}
	}
}

// ============================================================================
// Helpers
// ======================================================================================

func (p *prettyPrinter) printText(text string, style ...color.Attribute) {
	if !p.o.Colored {
		p.sb.WriteString(text)
		return
	}

	normal := color.New(style...)
	normal.EnableColor()
	code := color.New(style...).Add(color.Italic)
	code.EnableColor()

	normal.SetWriter(p.sb)
	defer normal.UnsetWriter(p.sb)

	var inCode bool
	var lastBacktick bool
	for _, r := range []byte(text) {
		if r == '`' && !lastBacktick {
			if inCode {
				code.UnsetWriter(p.sb)
				normal.SetWriter(p.sb)
			} else {
				normal.UnsetWriter(p.sb)
				code.SetWriter(p.sb)
			}
			inCode = !inCode
			lastBacktick = true
		} else {
			p.sb.WriteByte(r)
			lastBacktick = false
		}
	}
}

func (p *prettyPrinter) renderedTextLength(text string) int {
	if !p.o.Colored {
		return len(text)
	}

	var length int
	var inCode bool
	var lastBacktick bool
	for _, r := range []byte(text) {
		if r == '`' && !lastBacktick {
			inCode = !inCode
			lastBacktick = true
		} else {
			length++
			lastBacktick = false
		}
	}

	return length
}

func (p *prettyPrinter) printHelpText(title, text string) {
	p.coloredFunc(func() {
		p.sb.WriteString(title)
		p.sb.WriteString(": ")
	})

	split := strings.Split(text, "\n")
	p.printText(split[0])

	padding := strings.Repeat(" ", len(title)+len(": "))
	for _, s := range split[1:] {
		p.sb.WriteByte('\n')
		p.sb.WriteString(padding)
		p.printText(s)
	}
}

func (p *prettyPrinter) shouldInline(a annotation) bool {
	if strings.Contains(a.Annotation.Annotation, "\n") {
		return false
	}

	markerEnd := p.maxLineDigits + len(" | ") + a.End
	return markerEnd+len(" ")+p.renderedTextLength(a.Annotation.Annotation) < 100
}

func (p *prettyPrinter) printLineStart(lineNo int) {
	if lineNo <= 0 {
		p.colored(fmt.Sprintf("%*s | ", p.maxLineDigits, ""), color.Faint)
		return
	}

	p.colored(fmt.Sprintf("%*d | ", p.maxLineDigits, lineNo), color.Faint)
}

func (p *prettyPrinter) printAnnotationMarker(a annotation) int {
	endCol := a.End
	if endCol <= 0 {
		endCol = len(a.Lines[a.Line-a.ContextStart]) + 1
	} else if endCol < a.Start {
		endCol = a.Start + 1
	}

	repeatCount := endCol - a.Start

	p.colored(strings.Repeat("^", repeatCount), color.Bold, p.annoColor(a))

	return repeatCount
}

func (p *prettyPrinter) annoColor(a annotation) color.Attribute {
	if a.isError {
		return color.FgRed
	}
	return color.FgCyan
}

func (p *prettyPrinter) colored(text string, attrs ...color.Attribute) {
	if p.o.Colored && len(attrs) > 0 {
		c := color.New(attrs...)
		c.EnableColor()
		c.Fprint(p.sb, text)
		return
	}

	p.sb.WriteString(text)
}

func (p *prettyPrinter) coloredFunc(f func(), attrs ...color.Attribute) {
	if p.o.Colored && len(attrs) > 0 {
		color.New(attrs...).SetWriter(p.sb)
		defer color.New(attrs...).UnsetWriter(p.sb)
	}

	f()
}

type lineRange struct {
	start, end int
	lines      []string
}

// Reports the ranges of lines that are to be printed in the error.
func lineRanges(as []annotation) []lineRange {
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

func equalFile(a, b *file.File) bool {
	return a.Module == b.Module && a.PathInModule == b.PathInModule
}

func numDigits(n int) int {
	return int(math.Log10(float64(n)) + 1)
}
