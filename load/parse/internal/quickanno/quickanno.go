// Package quickanno provides utilities for creating annotations.
package quickanno

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func Expected(p *parser.Parser, pos ast.Position, expected string) []fancyerr.Annotation {
	return []fancyerr.Annotation{anno.Position(p.File, pos, "expected "+expected)}
}

func DeltaPos(p ast.Position, dLine, dCol int) ast.Position {
	return ast.Position{Line: p.Line + dLine, Col: p.Col + dCol}
}
