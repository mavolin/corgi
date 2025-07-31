//go:build test_stubs

package attribute

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

//nolint:gochecknoinits
func init() {
	SetElementReference(elementReferenceStub)
}

func elementReferenceStub(p *parser.Parser) (*ast.ElementReference, *diagnostic.Diagnostic) {
	var n ast.ElementName
	n.Position = p.PosPtr()
	n.Name = parser.Try(p, html.TagName())
	if n.Name == "" {
		return nil, &diagnostic.Diagnostic{
			Message: "missing element name",
			Primary: quickanno.Expected(p, *n.Position, "an html element name"),
		}
	}

	return &ast.ElementReference{Name: &n}, nil
}
