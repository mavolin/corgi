package parse

import (
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	fileparser "github.com/mavolin/corgi/v2/load/parse/internal/file"
)

type Options struct{}

// Parse parses the given input and returns a [file.File] with its AST set.
// The remaining fields of the returned file are left empty and are expected to
// be set by the caller.
//
// Both LF and CRLF line endings are supported.
//
// Parse can recover from errors and will continue parsing if it encounters any
// syntax errors.
// Therefore, Parse may return both a non-nil file and an error, indicating
// that the passed input is erroneous, but could be recovered from.
// The number of errors is capped at 255, additional errors are discarded.
func Parse(input string, _ Options) (*file.File, diagnostic.List) {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}

	f := &file.File{
		Name:  "<string input>",
		Raw:   input,
		Lines: lines,
	}

	p := parser.New(f, input, ast.Position{Line: 1, Col: 1})

	f.AST = parser.Try(p, fileparser.File())
	return f, p.Errors()
}
