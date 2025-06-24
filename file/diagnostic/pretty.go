package diagnostic

import (
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/mavolin/corgi/v2/file"
)

type PrettyOptions struct {
	// Color specifies whether the output should be colored.
	//
	// Default: false
	Color bool
	// Width is the maximum width of the output.
	//
	// Default: 80
	Width int
	// TypeColors is a map of error types to
	//
	// Default:
	//  Error:   color.FgRed
	//  Warning: color.FgYellow
	//  Lint:    color.FgBlue
	TypeColors map[Type]color.Attribute

	// FileNamePrinter is a function that returns the name of the file.
	//
	// Default: f.PathInModule
	FileNamePrinter func(f *file.File) string

	// DocsBaseURL is the base URL to the documentation.
	// No documentation link is added if this is empty.
	//
	// Default: https://corgi.mavolin.co
	DocsBaseURL string
}

var defaultTypeColors = map[Type]color.Attribute{
	Error:   color.FgRed,
	Warning: color.FgYellow,
	Lint:    color.FgBlue,
}

func (o *PrettyOptions) applyDefaults() {
	if o.Width == 0 {
		o.Width = 80
	}
	if o.TypeColors == nil {
		o.TypeColors = defaultTypeColors
	}
	if o.FileNamePrinter == nil {
		o.FileNamePrinter = func(f *file.File) string {
			return filepath.FromSlash("./" + f.PathInModule)
		}
	}
	if o.DocsBaseURL == "" {
		o.DocsBaseURL = "https://corgi.mavolin.co"
	}
}

func (o *PrettyOptions) typeColor(t Type) color.Attribute {
	if t == "" {
		t = Error
	}
	if c, ok := o.TypeColors[t]; ok {
		return c
	}

	return color.FgRed
}

type (
	prettyPrinter struct {
		sb *strings.Builder
		o  PrettyOptions

		diagnostic *Diagnostic
		// files is an ordered list of annotations per file.
		//
		// files are sorted in the order they appear in the error list, annotations
		// are sorted by line number.
		files []*fileAnnos

		nDigits int
	}
	annotation struct {
		Annotation
		primary bool
	}
	fileAnnos struct {
		file  *file.File
		annos []annotation
	}
)

// Pretty returns a pretty printed version of the error, formatted according
// to the provided options.
func (d *Diagnostic) Pretty(o PrettyOptions) string {
	o.applyDefaults()

	var sb strings.Builder
	sb.Grow(2048)
	d.pretty(&sb, o)
	return sb.String()
}

func (d *Diagnostic) pretty(sb *strings.Builder, o PrettyOptions) {
	p := &prettyPrinter{
		sb:         sb,
		o:          o,
		diagnostic: d,
		files:      make([]*fileAnnos, 0, len(d.Primary)+len(d.Secondary)),
	}
	for _, a := range d.Primary {
		p.insertAnno(a, true, len(d.Primary))
		if n := numDigits(a.ContextEnd - 1); n > p.nDigits {
			p.nDigits = n
		}
	}
	for _, a := range d.Secondary {
		p.insertAnno(a, false, len(d.Primary))
		if n := numDigits(a.ContextEnd - 1); n > p.nDigits {
			p.nDigits = n
		}
	}

	p.print()
}

func numDigits(n int) int { // essentially log10
	switch {
	case n < 10:
		return 1
	case n < 100:
		return 2
	case n < 1000:
		return 3
	case n < 10_000:
		return 4
	default:
		return 5
	}
}

func (p *prettyPrinter) insertAnno(a Annotation, primary bool, n int) {
	for _, f := range p.files {
		if f.file == a.File {
			f.annos = append(f.annos, annotation{Annotation: a, primary: primary})
			return
		}
	}

	fa := fileAnnos{
		file:  a.File,
		annos: make([]annotation, 1, n),
	}
	fa.annos[0] = annotation{Annotation: a, primary: true}
	p.files = append(p.files, &fa)
}

func (p *prettyPrinter) print() {
	p.printMessage()
	p.printFiles()
	p.printCause()
	p.printExplanation()
	p.printExamples()
	p.printHints()
	p.printDocs()
}

func (p *prettyPrinter) printMessage() {
	typ := p.diagnostic.Type
	if typ == "" {
		typ = Error
	}
	p.colored(string(typ)+":", color.Bold, p.o.typeColor(typ))
	p.uncolored(" ")
	p.printText(p.diagnostic.Message, len(typ)+len(": "), false, color.Bold)
}

func (p *prettyPrinter) printFiles() {
	if len(p.files) == 0 {
		return
	}

	p.uncolored("\n")

	for i, f := range p.files {
		p.skip(p.nDigits + 1)
		if i == 0 {
			p.box("╭─ ")
		} else {
			p.skip(p.nDigits + 1)
			p.box("├─ ")
		}
		p.colored(p.o.FileNamePrinter(f.file)+":"+f.annos[0].Start.String(), color.Bold)
		p.printFile(f)
	}
}

func (p *prettyPrinter) printFile(f *fileAnnos) {
	for i, lr := range lineRanges(f) {
		if i > 0 {
			p.uncolored("\n")
			p.skip(p.nDigits + 1)
			p.box("┆ ")
			p.colored("...", color.FgWhite, color.Faint)
		}

		p.printLineRange(f, lr)
	}
}

func (p *prettyPrinter) printLineRange(f *fileAnnos, lr lineRange) {
	lineAnnotations := make([]annotation, 0, len(f.annos))

	for lnNo := lr.start; lnNo < lr.end; lnNo++ {
		p.uncolored("\n")
		p.printLineStart(lnNo)

		ln := lr.lines[lnNo-lr.start]
		p.uncolored(ln)

		lineAnnotations = lineAnnotations[:0]
		for _, a := range f.annos {
			if a.Start.Line <= lnNo && lnNo <= a.End.Line {
				lineAnnotations = append(lineAnnotations, a)
			}
		}
		if len(lineAnnotations) == 0 {
			continue
		}
		slices.SortFunc(lineAnnotations, func(a, b annotation) int {
			if a.Start.Line == b.Start.Line {
				return a.Start.Col - b.End.Col
			}
			return a.Start.Line - b.End.Line
		})

		p.uncolored("\n")
		p.printLineStart(-1)

		p.printAnnotationMarkers(lnNo, ln, lineAnnotations)

		last := lineAnnotations[len(lineAnnotations)-1]

		// if we can comfortably fit the entire annotation behind the markers,
		// render the annotation in a single line to save space
		if p.shouldInline(lnNo, last) {
			p.uncolored(" ")
			p.printText(last.Annotation.Annotation, 0, false, color.Bold, p.annoColor(last))
			lineAnnotations = lineAnnotations[:len(lineAnnotations)-1]
		} else if last.End.Line != lnNo {
			lineAnnotations = lineAnnotations[:len(lineAnnotations)-1]
		}

		p.printAnnotations(lineAnnotations)
	}
}

func (p *prettyPrinter) printAnnotationMarkers(lnNo line, ln string, lineAnnotations []annotation) {
	var offset int
	for _, la := range lineAnnotations {
		var numSpaces int
		if lnNo == la.Start.Line {
			numSpaces = la.Start.Col - 1 - offset
		} else {
			numSpaces = 0
		}
		p.skip(numSpaces)
		offset += numSpaces
		offset += p.printAnnotationMarker(lnNo, ln, la)
	}
}

func (p *prettyPrinter) printAnnotationMarker(lnNo line, ln string, a annotation) int {
	var start, end col
	if a.Start.Line == lnNo {
		start = a.Start.Col
	} else {
		start = 1
	}
	if a.End.Line == lnNo {
		end = a.End.Col
	} else {
		end = len(ln) + 1
	}
	repeatCount := end - start
	if a.primary {
		p.colored(strings.Repeat("^", repeatCount), p.annoColor(a))
	} else {
		p.colored(strings.Repeat("~", repeatCount), p.annoColor(a))
	}
	return repeatCount
}

func (p *prettyPrinter) printAnnotations(as []annotation) {
	// start writing the rightmost annotation
	for currentI, current := range slices.Backward(as) {
		p.uncolored("\n")

		offset := 0
		for _, preceding := range as[:currentI] {
			var start col
			if preceding.Start.Line == preceding.End.Line {
				start = preceding.Start.Col
			} else {
				start = 1 // multi-line annotation
			}

			p.skip(start - 1 - offset)
			p.colored("│", p.annoColor(preceding))
			offset += start - 1 - offset + len("|")
		}

		var start col
		if current.Start.Line == current.End.Line {
			start = current.Start.Col
		} else {
			start = 1 // multi-line annotation
		}

		p.printLineStart(-1)
		p.skip(start - 1 - offset)
		p.colored("╰ ", p.annoColor(current))
		p.printText(current.Annotation.Annotation, (start-1-offset)+len("| "), true, color.Bold, p.annoColor(current))
	}
}

func (p *prettyPrinter) shouldInline(lnNo line, a annotation) bool {
	if strings.Contains(a.Annotation.Annotation, "\n") || lnNo != a.End.Line {
		return false
	}

	markerEnd := p.nDigits + len(" | ") + a.End.Col
	return markerEnd+len(" ")+p.renderedTextLength(a.Annotation.Annotation) <= p.o.Width
}

func (p *prettyPrinter) printCause() {
	if p.diagnostic.Cause != nil {
		p.uncolored("\n\n")
		p.colored("Cause: ", color.Bold)
		p.uncolored(p.diagnostic.Cause.Error())
	}
}

func (p *prettyPrinter) printExplanation() {
	if p.diagnostic.Explanation == "" {
		return
	}

	p.uncolored("\n\n")
	p.printText(p.diagnostic.Explanation, 0, false)
}

func (p *prettyPrinter) printExamples() {
	var indent int
	switch len(p.diagnostic.Examples) {
	case 0:
		return
	case 1:
		indent = len("Example: ")
		p.colored("\n\nExample: ", color.Bold)
	default:
		indent = len("Examples: ")
		p.colored("\n\nExamples: ", color.Bold)
	}

	var titleIndent int
	var hasTitle, needHeadline bool
	for _, example := range p.diagnostic.Examples {
		if example.Title == "" {
			continue
		}
		hasTitle = true
		if strings.Contains(example.Example, "\n") {
			needHeadline = true
		} else {
			titleIndent = max(titleIndent, indent+p.renderedTextLength(example.Example)+len(" "))
		}
	}
	if titleIndent+len("(")+len(")") > p.o.Width {
		needHeadline = true
	}
	if !needHeadline && hasTitle {
		for _, example := range p.diagnostic.Examples {
			if titleIndent+len("(")+len(example.Title)+len(")") > p.o.Width {
				needHeadline = true
				break
			}
		}
	}

	for i, example := range p.diagnostic.Examples {
		if i > 0 {
			p.uncolored("\n")
			p.skip(indent)
		}
		if needHeadline {
			title := example.Title
			if title == "" {
				title = fmt.Sprint("example ", i+1)
			}
			if p.o.Color {
				p.colored(title, color.Bold)
			} else {
				p.uncolored("*")
				p.uncolored(title)
				p.uncolored("*")
			}
			p.uncolored("\n")
			p.skip(indent)
		}
		p.printText(example.Example, indent, false)
		if !needHeadline && hasTitle {
			p.skip(titleIndent - indent - p.renderedTextLength(example.Example))
			p.colored("("+example.Title+")", color.Bold)
		}
	}
}

func (p *prettyPrinter) printHints() {
	var indent int
	switch len(p.diagnostic.Hints) {
	case 0:
		return
	case 1:
		indent = len("Hint: ")
		p.colored("\n\nHint: ", color.Bold)
	case 2:
		indent = len("Hints: ")
		p.colored("\n\nHints: ", color.Bold)
	}

	for i, hint := range p.diagnostic.Hints {
		if i > 0 {
			p.uncolored("\n")
			p.skip(indent)
		}

		p.printText(hint.Hint, indent, false)
		if hint.Example != "" {
			p.uncolored("\n")
			p.skip(indent)
			p.printText("|> "+hint.Example, indent+len("|> "), false)
		}
	}
}

func (p *prettyPrinter) printDocs() {
	if p.diagnostic.Docs == "" {
		return
	}
	p.uncolored("\n\n")
	p.colored("Docs: ", color.Bold)
	link, err := url.JoinPath(p.o.DocsBaseURL, "!"+p.diagnostic.Docs)
	if err != nil {
		link = p.o.DocsBaseURL + "/!" + p.diagnostic.Docs
	}
	p.uncolored(link)
}

// ============================================================================
// Helpers
// ======================================================================================

func (p *prettyPrinter) uncolored(text string) {
	p.sb.WriteString(text)
}

func (p *prettyPrinter) colored(text string, attrs ...color.Attribute) {
	if p.o.Color {
		c := color.New(attrs...)
		c.EnableColor()
		c.Fprint(p.sb, text)
		return
	}

	p.uncolored(text)
}

func (p *prettyPrinter) box(b string) {
	p.colored(b, color.FgWhite, color.Faint)
}

func (p *prettyPrinter) printLineStart(ln line) {
	if ln <= 0 {
		p.skip(p.nDigits + 1)
		p.box("│ ")
		return
	}

	n := numDigits(ln)
	p.skip(p.nDigits - n)
	p.colored(fmt.Sprint(ln), color.FgWhite, color.Faint)
	p.box(" │ ")
}

func (p *prettyPrinter) skip(n int) {
	p.uncolored(strings.Repeat(" ", n))
}

func (p *prettyPrinter) annoColor(a annotation) color.Attribute {
	tc := p.o.typeColor(p.diagnostic.Type)
	if a.primary {
		return tc
	}
	if tc == color.FgCyan {
		return color.FgHiCyan
	}
	return color.FgCyan
}

func (p *prettyPrinter) printText(text string, indent int, needLineStart bool, style ...color.Attribute) {
	p.sb.Grow(len(text))

	regular := color.New(style...)
	code := color.New(style...).Add(color.Italic)
	if p.o.Color {
		regular.EnableColor()
		code.EnableColor()
	} else {
		regular.DisableColor()
		code.DisableColor()
	}

	regular.SetWriter(p.sb)
	defer regular.UnsetWriter(p.sb)

	var inCode bool
	var forceWrite bool
	baseCol := indent + 1
	if needLineStart {
		baseCol += p.nDigits + len(" | ")
	}
	col := baseCol
	for i, r := range text {
		wordEnd := strings.IndexAny(text[i:], " \n")
		if wordEnd < 0 {
			wordEnd = len(text) - i
		}
		if col+wordEnd > p.o.Width && !forceWrite || r == '\n' {
			if inCode {
				code.UnsetWriter(p.sb)
			} else {
				regular.UnsetWriter(p.sb)
			}
			p.uncolored("\n")
			if needLineStart {
				p.printLineStart(-1)
			}
			p.skip(indent)
			if inCode {
				code.SetWriter(p.sb)
			} else {
				regular.SetWriter(p.sb)
			}
			col = baseCol
			if r == '\n' {
				continue
			}
			forceWrite = true
		} else {
			forceWrite = false
		}

		if r == '`' {
			if inCode {
				code.UnsetWriter(p.sb)
				regular.SetWriter(p.sb)
			} else {
				regular.UnsetWriter(p.sb)
				code.SetWriter(p.sb)
			}
			inCode = !inCode
		} else {
			p.sb.WriteRune(r)
			col++
		}
	}
}

func (p *prettyPrinter) renderedTextLength(text string) int {
	if !p.o.Color {
		return len(text)
	}
	return len(text) - strings.Count(text, "`") // backticks are not printed
}

type lineRange struct {
	start, end line
	lines      []string
}

// Reports the ranges of lines that are to be printed in the error.
func lineRanges(f *fileAnnos) []lineRange {
	lines := make([]lineRange, len(f.annos))
	for i, a := range f.annos {
		lines[i] = lineRange{start: a.ContextStart, end: a.ContextEnd}
	}

	slices.SortFunc(lines, func(a, b lineRange) int {
		return a.start - b.start
	})

	merged := make([]lineRange, 0, len(lines))

	for ai := 0; ai < len(lines); ai++ { // merge time
		a := lines[ai]

		bi := ai + 1
		for ; bi < len(lines); bi++ {
			b := lines[bi]
			if b.start-1 > a.end { // collapse gaps of a single line
				break
			}
			a.end = b.end
		}
		ai = bi - 1

		merged = append(merged, a)
	}

	for i, l := range merged {
		merged[i].lines = f.file.AST.Lines[l.start-1 : l.end-1]
	}

	return merged
}
