package typeinfer

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
)

// MixinParams attempts to infer the type of m's params without an explicitly
// set type but a default expression.
//
// When it succeeds, it stores the inferred type as
// [file.MixinParam.InferredType].
func MixinParams(m *file.Mixin) {
	for i, param := range m.Params {
		if param.Type == nil && param.Default != nil {
			m.Params[i].InferredType = Infer(*param.Default)
		}
	}
}

// Scope runs [MixinParams] on all mixins in the passed scope.
func Scope(s file.Scope) {
	fileutil.Walk(s, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		m, ok := (*ctx.Item).(file.Mixin)
		if !ok {
			return true, nil
		}

		MixinParams(&m)
		// *ctx.Item = m // symbolic: MixinParams modifies m.Params, so this is not needed
		return true, nil
	})
}
