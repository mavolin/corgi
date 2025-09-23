package analyze

import (
	"github.com/mavolin/corgi/v2/file"
)

// AnalyzeComponentCall_ComponentArguments analyzes all component arguments of the given
// component call.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponentCall_ComponentArguments(cc *file.ComponentCall) {
	for _, arg := range cc.ComponentArguments {
		z.AnalyzeComponentArgument(cc, arg)
	}
}

// AnalyzeComponentArgument analyzes the passed component argument.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponentArgument(cc *file.ComponentCall, arg *file.ComponentArgument) {
	z.AnalyzeComponentArgument_Value(cc, arg)
}

// ============================================================================
// Value
// ======================================================================================

// AnalyzeComponentArgument_Value analyzes the value of the given component
// argument.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentArguments.Value
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponentArgument_Value(cc *file.ComponentCall, arg *file.ComponentArgument) {
	arg.Value = z.expressionToResolvedValue(cc.File, arg.AST.Value)
}
