package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

var componentCall parser.Func[*ast.ComponentCall]

func SetComponentCall(f parser.Func[*ast.ComponentCall]) {
	componentCall = f
}
