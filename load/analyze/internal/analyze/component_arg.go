package analyze

import (
	"github.com/mavolin/corgi/v2/file"
)

func (z *analyzer) AnalyzeComponentCall_ComponentArguments(cc *file.ComponentCall) {
	for _, arg := range cc.ComponentArguments {
		z.AnalyzeComponentArgument(cc, arg)
	}
}

func (z *analyzer) AnalyzeComponentArgument(cc *file.ComponentCall, arg *file.ComponentArgument) {
	z.AnalyzeComponentArgument_Value(cc, arg)
}

// ============================================================================
// Value
// ======================================================================================

type componentArgument_Value struct{}

func (z *analyzer) AnalyzeComponentArgument_Value(cc *file.ComponentCall, arg *file.ComponentArgument) {
	defer z.Ran(arg, componentArgument_Value{})
	arg.Value = z.expressionToResolvedValue(cc.File, arg.AST.Value)
}
