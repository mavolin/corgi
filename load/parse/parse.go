package parse

import (
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	fileparser "github.com/mavolin/corgi/v2/load/parse/internal/file"
)

type Options struct {
	// Preloader injects a preloader into the parser.
	//
	// The preloader is called everytime an import statement is encountered.
	// Since import statements appear early in the file, this can give the
	// loader a head start on loading the imported package before it is
	// actually needed by the linker.
	//
	// A preloader must not block.
	Preloader func(importPath string)
}

// Parse parses the given input file and returns a [file.File] with its AST set.
// The remaining fields of the returned file are left empty and are expected to
// be set by the caller.
//
// Both LF and CRLF line endings are supported.
//
// Parse can recover from errors and will continue parsing if it encounters any
// syntax errors.
// Therefore, Parse may return both a non-nil file and an error, indicating
// that the passed input is erroneous, but could be recovered from.
func Parse(input string, o Options) (*file.File, diagnostic.List) {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}

	f := &file.File{
		Name: "<string_input>",
		AST: &ast.File{
			Raw:   input,
			Lines: lines,
		},
	}

	p := parser.New(f)
	if o.Preloader != nil {
		p.Preload = o.Preloader
	}

	parser.Try(p, fileparser.File())
	return f, p.Errors()
}
