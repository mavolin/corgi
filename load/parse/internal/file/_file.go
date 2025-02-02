package file

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/text"
)

func init() {
	interpolation.SetExpression(code.Expression())
	interpolation.SetElementHeader(element.Header())
	interpolation.SetComponentCallHeader(component.CallHeader())

	body.SetTextLine(text.Line)
	body.SetScopeNode(scopeNode)
}

func scopeNode(p *parser.Parser) (ast.ScopeNode, *fancyerr.Error) {
	panic("implement me")

}

func File() parser.Func[*ast.File] {

}
