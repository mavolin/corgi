package anno

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
)

type Annotation struct {
	ContextStart      file.Position
	ContextStartDelta int
	ContextEnd        file.Position
	ContextEndDelta   int
	ContextLen        int

	Start       file.Position
	StartOffset int
	End         file.Position
	EndOffset   int
	Len         int
	EOLDelta    int
	ToEOL       bool

	Annotation string
}

func Anno(f *file.File, aw Annotation) corgierr.Annotation {
	a := Lines(f.Lines, aw)
	a.File = f
	return a
}

func Lines(lines []string, aw Annotation) corgierr.Annotation {
	var a corgierr.Annotation
	a.Annotation = aw.Annotation

	if aw.ContextStart.Line > 0 {
		a.ContextStart = aw.ContextStart.Line
		if aw.ContextStart.Col == 0 {
			a.ContextStart--
		}
	} else {
		a.ContextStart = aw.Start.Line
		if aw.Start.Col == 0 {
			a.ContextStart--
		}
	}
	if aw.ContextStartDelta != 0 {
		a.ContextStart += aw.ContextStartDelta
	}
	if a.ContextStart <= 0 {
		a.ContextStart = 1
	}
	if a.ContextStart >= len(lines) {
		a.ContextStart = len(lines)
	}

	switch {
	case aw.ContextLen >= 1:
		a.ContextEnd = a.ContextStart + aw.ContextLen
	case aw.ContextEnd.Line > 0:
		a.ContextEnd = aw.ContextEnd.Line
	case aw.End.Line > 0:
		a.ContextEnd = aw.End.Line + 1
		if aw.End.Col == 0 {
			a.ContextEnd--
		}
	default:
		a.ContextEnd = aw.Start.Line + 1
		if aw.Start.Col == 0 {
			a.ContextEnd--
		}
	}
	if aw.ContextEndDelta != 0 {
		a.ContextEnd += aw.ContextEndDelta
	}
	if a.ContextEnd <= a.ContextStart {
		a.ContextEnd = a.ContextStart + 1
	}
	if a.ContextEnd > len(lines) {
		a.ContextEnd = len(lines) + 1
	}

	a.Lines = lines[a.ContextStart-1 : a.ContextEnd-1] // lines are 1-indexed

	a.Line = aw.Start.Line
	if aw.Start.Col == 0 {
		a.Line--
	}

	a.Start = aw.Start.Col
	if a.Start == 0 {
		a.Start = len(lines[a.Line-1]) + 1
	} else {
		a.Start += aw.StartOffset
	}

	if aw.End.Line > 0 && aw.End.Col == 0 {
		aw.End.Line--
		aw.End.Col = len(lines[aw.End.Line-1]) + 1
	}

	switch {
	case aw.End.Line > 0:
		if aw.End.Line != aw.Start.Line {
			a.End = len(lines[a.Line-1]) + 1
			break
		}

		a.End = aw.End.Col
	case aw.Len > 0:
		a.End = a.Start + aw.Len
	case aw.EOLDelta != 0:
		a.End = len(lines[a.Line-1]) + 1 + aw.EOLDelta
	case aw.ToEOL:
		a.End = len(lines[a.Line-1]) + 1
	default:
		a.End = a.Start + 1
	}
	if a.End <= a.Start {
		a.End = a.Start + 1
	}
	a.End += aw.EndOffset

	return a
}
