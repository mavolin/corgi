package parse

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/load/parse/internal"
)

// Parse parses the given input file and returns a [file.File] with its AST set.
// The remaining fields of the returned file are left empty and are expected to
// be set by the caller.
//
// Parse can recover from errors and will continue parsing if it encounters any
// syntax errors.
// Therefore, Parse may return both a non-nil file and an error, indicating
// that the passed input is erroneous, but could be recovered from.
//
// Parse guarantees that the returned error is either nil or [fileerr.List].
func Parse(input []byte) (*file.File, error) {
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}

	f := &file.File{AST: &ast.AST{Raw: string(input), Lines: lines}}
	_, err := internal.Parse("bytedata", input, internal.GlobalStore("file", f))

	var errs internal.ErrList
	errors.As(err, &errs)

	for i, err := range errs {
		var parserErr *internal.ParserError
		if !errors.As(err, &parserErr) {
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

		errs[i] = cerr
	}

	return f, errors.Join(errs...)
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
		},
	}
}
