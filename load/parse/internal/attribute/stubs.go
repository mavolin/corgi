//go:build test_stubs

package attribute

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
)

//nolint:gochecknoinits
func init() {
	SetElementReference(elementReferenceStub)
}

func elementReferenceStub(p *parser.Parser) *ast.ElementReference {
	pos := p.Pos()
	name := parser.TokenWhile(p, func() bool {
		return parser.MatchesRunePredicate(p, html.TagNameStartRune)
	})
	if name == "" {
		return nil
	}

	return &ast.ElementReference{Name: &ast.ElementName{Name: name, Position: &pos}}
}
