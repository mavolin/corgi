package parse

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/parse/internal"
)

// Parse parses the given input file and returns the generated [file.File].
//
// If it encounters any syntax errors, it attempts to recover from them and
// resume parsing.
// Therefore, Parse may return both a non-nil file and an error, indicating
// that the passed input is erroneous, but could be recovered from.
//
// If Parse returns an error, it will always be of type [fileerr.List].
//
// Callers are expected to set the Name, Module, PathInModule, and AbsolutePath
// of the returned file themselves.
//
// By default, Name is set to "bytedata", so if you print any errors without
// updating Name, this will be used as filename in the error message.
func Parse(input []byte) (*file.File, error) {
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}

	fi, err := internal.Parse("bytedata", input, internal.GlobalStore("lines", lines))

	f, _ := fi.(*file.File)
	if f == nil {
		f = new(file.File)
	}
	f.Name = "parse"
	f.Raw = string(input)
	f.Lines = lines

	var errList internal.ErrList
	if !errors.As(err, &errList) {
		return f, err
	}

	corgierrList := make(fileerr.List, len(errList))

	for i, err := range errList {
		var parserErr *internal.ParserError
		if !errors.As(err, &parserErr) {
			corgierrList[i] = &fileerr.Error{
				Message: err.Error(),
				ErrorAnnotation: fileerr.Annotation{
					ContextStart: 1,
					ContextEnd:   2,
					Start:        1,
					End:          2,
					Annotation:   "position unknown",
					Lines:        []string{""},
				},
			}
			continue
		}

		var cerr *fileerr.Error
		if !errors.As(parserErr.Inner, &cerr) {
			cerr = parserErrorToCorgiError(lines, parserErr)
		}

		cerr.ErrorAnnotation.File = f
		for j := range cerr.HintAnnotations {
			cerr.HintAnnotations[j].File = f
		}

		corgierrList[i] = cerr
	}

	return f, corgierrList
}

var parserErrorRegexp = regexp.MustCompile(`(\d+):(\d+)( \(\d+\))?: (.+)`)

func parserErrorToCorgiError(lines []string, perr *internal.ParserError) *fileerr.Error {
	matches := parserErrorRegexp.FindStringSubmatch(perr.Error())
	const ( // indexes of groups
		_ = iota // all text
		line
		col
		_ // offset
		msg
	)

	if len(matches) != 5 {
		return &fileerr.Error{
			Message: perr.Error(),
			ErrorAnnotation: fileerr.Annotation{
				ContextStart: 1,
				ContextEnd:   2,
				Start:        1,
				End:          2,
				Annotation:   "position unknown",
				Lines:        []string{""},
			},
		}
	}

	colNum, colErr := strconv.Atoi(matches[col])
	lineNum, lineErr := strconv.Atoi(matches[line])
	if colErr != nil || lineErr != nil {
		return &fileerr.Error{
			Message: perr.Error(),
			ErrorAnnotation: fileerr.Annotation{
				ContextStart: 1,
				ContextEnd:   2,
				Start:        1,
				End:          2,
				Annotation:   "position unknown",
				Lines:        []string{""},
			},
		}
	}

	if colNum == 0 {
		lineNum--
		colNum = len(lines[lineNum-1]) + 1
	}

	return &fileerr.Error{
		Message: matches[msg],
		ErrorAnnotation: fileerr.Annotation{
			ContextStart: lineNum,
			ContextEnd:   lineNum + 1,
			Line:         lineNum,
			Start:        colNum,
			End:          colNum + 1,
			Annotation:   "here",
			Lines:        []string{lines[lineNum-1]},
		},
	}
}
