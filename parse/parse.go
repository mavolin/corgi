package parse

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/parse/internal"
)

// Parse parses the given input file and returns the generated [file.File].
//
// If it encounters any syntax errors, it attempts to recover from them and
// resume parsing.
// Therefore, Parse may return both a non-nil file and an error, indicating
// that the passed input is erroneous, but could be recovered from.
//
// If Parse returns an error, it will always be of type [corgierr.List].
//
// Callers are expected to set the Name, Module, ModulePath, and AbsolutePath
// of the returned file themselves.
//
// By default, Name is set to "parse", so if you print any errors without
// updating Name, this will be used as filename in the error message.
func Parse(input []byte) (*file.File, error) {
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if len(line) > 0 && line[len(line)-1] == '\r' {
			lines[i] = line[:len(line)-1]
		}
	}

	fi, err := internal.Parse("test.corgi", input, internal.GlobalStore("lines", lines))

	f, ok := fi.(*file.File)
	if ok {
		f.Name = "parse"
		f.Raw = string(input)
		f.Lines = lines
	} else {
		f = &file.File{
			Name:  "parse",
			Raw:   string(input),
			Lines: lines,
		}
	}

	errList, ok := err.(internal.ErrList) //nolint: errorlint
	if !ok {
		return f, err
	}

	corgierrList := make(corgierr.List, len(errList))

	for i, err := range errList {
		parserErr, ok := err.(*internal.ParserError) //nolint: errorlint
		if !ok {
			corgierrList[i] = &corgierr.Error{
				Message: err.Error(),
				ErrorAnnotation: corgierr.Annotation{
					ContextStart: 1,
					ContextEnd:   2,
					Start:        1,
					End:          2,
					Annotation:   "position unknown",
					Lines:        []string{""},
				},
			}
		}

		cerr, ok := parserErr.Inner.(*corgierr.Error) //nolint: errorlint
		if !ok {
			cerr = parserErrorToCorgiError(lines, parserErr)
		}

		cerr.ErrorAnnotation.File = f
		for j := range cerr.HintAnnotations {
			cerr.HintAnnotations[j].File = f
		}

		corgierrList[i] = cerr
	}

	sort.Sort(corgierrList)
	return f, corgierrList
}

var parserErrorRegexp = regexp.MustCompile(`(\d+):(\d+)( \(\d+\))?: (.+)`)

func parserErrorToCorgiError(lines []string, perr *internal.ParserError) *corgierr.Error {
	matches := parserErrorRegexp.FindStringSubmatch(perr.Error())
	const ( // indexes of groups
		_ = iota // all text
		col
		line
		_ // offset
		msg
	)

	if len(matches) != 5 {
		return &corgierr.Error{
			Message: perr.Error(),
			ErrorAnnotation: corgierr.Annotation{
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
		return &corgierr.Error{
			Message: perr.Error(),
			ErrorAnnotation: corgierr.Annotation{
				ContextStart: 1,
				ContextEnd:   2,
				Start:        1,
				End:          2,
				Annotation:   "position unknown",
				Lines:        []string{""},
			},
		}
	}

	return &corgierr.Error{
		Message: matches[msg],
		ErrorAnnotation: corgierr.Annotation{
			File:         nil,
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
