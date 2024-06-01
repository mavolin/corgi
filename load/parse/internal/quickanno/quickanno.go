// Package quickanno provides utilities for creating annotations.
package quickanno

import (
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/hintfmt"
	"github.com/mavolin/corgi/file/hintfmt/anno"
	parser "github.com/mavolin/corgi/load/parse/internal"
)

func Expected(p *parser.Parser, pos ast.Position, expected string) hintfmt.Annotation {
	return anno.NChars(p.File, pos, 1, "expected "+expected)
}
