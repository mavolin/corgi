package analyze

import (
	"github.com/mavolin/corgi/v2/file"
)

func (z *analyzer) AnalyzeComponentArguments(cc *file.ComponentCall) {
	for _, arg := range cc.ComponentArguments {
		z.AnalyzeComponentArgumentValue(cc, arg)
	}
}

// ============================================================================
// Value
// ======================================================================================

func (z *analyzer) AnalyzeComponentArgumentValue(cc *file.ComponentCall, arg *file.ComponentArgument) {
	arg.Value = z.expressionToAttributeValue(cc.File, arg.AST.Value)
}
