// Package quickanno provides utilities for creating annotations.
package quickanno

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func Expected(p *parser.Parser, pos ast.Position, expected string) []diagnostic.Annotation {
	return []diagnostic.Annotation{anno.Position(p.File, pos, "expected "+expected)}
}

func DeltaPos(p ast.Position, dLine, dCol int) ast.Position {
	return ast.Position{
		Line: ast.Line(int(p.Line) + dLine), //nolint:gosec
		Col:  ast.Col(int(p.Col) + dCol),    //nolint:gosec
	}
}
