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
	var n ast.ElementName
	n.Position = p.PosPtr()
	n.Name = parser.Try(p, html.TagName())
	if n.Name == "" {
		return nil
	}

	return &ast.ElementReference{Name: &n}
}
