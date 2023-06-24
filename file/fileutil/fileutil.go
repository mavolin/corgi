// Package fileutil provides utilities for interacting with a corgi AST and its
// contents.
package fileutil

import (
	"github.com/mavolin/corgi/file"
)

// FirstNonControl returns the first non-if, if block, switch and for item in
// the given scope.
//
// This means it either returns nil for an empty scope, s[0] for a scope that
// doesn't start with a control directive, or the first non-control item nested
// inside the control item at s[0].
func FirstNonControl(s file.Scope) file.ScopeItem {
	var itm file.ScopeItem
	_ = Walk(s, func(_ []WalkContext, ctx WalkContext) (dive bool, err error) {
		switch (*ctx.Item).(type) {
		case file.If:
			return true, nil
		case file.IfBlock:
			return true, nil
		case file.Switch:
			return true, nil
		case file.For:
			return true, nil
		case file.Code:
			return false, nil
		case file.CorgiComment:
			return false, nil
		default:
			return false, StopWalk
		}
	})

	return itm
}

func IsFirstNonControlAttr(s file.Scope) (file.ScopeItem, bool) {
	var ret file.ScopeItem
	var ret2 bool
	_ = Walk(s, func(parents []WalkContext, ctx WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.If:
			return true, nil
		case file.IfBlock:
			return true, nil
		case file.Switch:
			return true, nil
		case file.For:
			return true, nil
		case file.Code:
			return false, nil
		case file.CorgiComment:
			return false, nil
		case file.And:
			ret, ret2 = itm, true
			return false, StopWalk
		case file.MixinCall:
			if itm.Mixin.WritesTopLevelAttributes || IsAttrMixin(*itm.Mixin) {
				ret, ret2 = itm, true
				return false, StopWalk
			} else if itm.Mixin.WritesBody || IsElementMixin(*itm.Mixin) {
				return false, StopWalk
			}

			firstAttr := firstActualAttr(itm.Body)
			if itm.Mixin.TopLevelAndPlaceholder && firstAttr != nil {
				ret, ret2 = firstAttr, true
				return false, StopWalk
			}

			for _, block := range itm.Mixin.Blocks {
				if !block.TopLevel {
					continue
				}

				if len(itm.Body) == 1 {
					sh, ok := itm.Body[0].(file.MixinMainBlockShorthand)
					if ok && block.Name == "_" {
						if subCtrl, ok := IsFirstNonControlAttr(sh.Body); ok {
							ret, ret2 = subCtrl, true
							return false, StopWalk
						}
					}
				}

				for _, bodyItm := range itm.Body {
					mcBlock, ok := bodyItm.(file.Block)
					if !ok || mcBlock.Name.Ident != block.Name {
						continue
					}

					if subCtrl, ok := IsFirstNonControlAttr(mcBlock.Body); ok {
						ret, ret2 = subCtrl, true
						return false, StopWalk
					}

					break
				}

				if block.DefaultWritesBody || block.DefaultWritesTopLevelAttributes {
					ret = itm
					return false, StopWalk
				} else if block.DefaultTopLevelAndPlaceholder && firstAttr != nil {
					ret = firstAttr
					return false, StopWalk
				}
			}

			return false, nil
		default:
			return false, StopWalk
		}
	})

	return ret, ret2
}

func firstActualAttr(s file.Scope) file.ScopeItem {
	var ret file.ScopeItem
	_ = Walk(s, func(parents []WalkContext, ctx WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.If:
			return true, nil
		case file.IfBlock:
			return true, nil
		case file.Switch:
			return true, nil
		case file.For:
			return true, nil
		case file.Code:
			return false, nil
		case file.CorgiComment:
			return false, nil
		case file.Block:
			return false, nil
		case file.MixinCall:
			if itm.Mixin.WritesTopLevelAttributes || itm.Mixin.WritesBody || IsAttrMixin(*itm.Mixin) || IsElementMixin(*itm.Mixin) {
				ret = itm
				return false, StopWalk
			}

			firstAttr := firstActualAttr(itm.Body)
			if itm.Mixin.TopLevelAndPlaceholder && firstAttr != nil {
				ret = firstAttr
				return false, StopWalk
			}

			for _, block := range itm.Mixin.Blocks {
				if !block.TopLevel {
					continue
				}

				if len(itm.Body) == 1 {
					sh, ok := itm.Body[0].(file.MixinMainBlockShorthand)
					if ok && block.Name == "_" && block.TopLevel {
						if subCtrl, ok := IsFirstNonControlAttr(sh.Body); ok {
							ret = subCtrl
							return false, StopWalk
						}
					}
				}

				for _, bodyItm := range itm.Body {
					mcBlock, ok := bodyItm.(file.Block)
					if !ok || mcBlock.Name.Ident != block.Name {
						continue
					}

					if subCtrl, ok := IsFirstNonControlAttr(mcBlock.Body); ok {
						ret = subCtrl
						return false, StopWalk
					}

					break
				}

				if block.DefaultWritesBody || block.DefaultWritesTopLevelAttributes {
					ret = itm
					return false, StopWalk
				} else if block.DefaultTopLevelAndPlaceholder && firstAttr != nil {
					ret = firstAttr
					return false, StopWalk
				}
			}

			return false, nil
		default:
			return false, StopWalk
		}
	})

	return ret
}

func IsStdLibFile(f *file.File) bool {
	return f.Module == "github.com/mavolin/corgi" && f.ModulePath == "std"
}

func IsAttrMixin(lm file.LinkedMixin) bool {
	return IsStdLibFile(lm.File) && lm.File.ModulePath == "std/html" && lm.Mixin.Name.Ident == "Attr"
}

func IsElementMixin(lm file.LinkedMixin) bool {
	return IsStdLibFile(lm.File) && lm.File.ModulePath == "std/html" && lm.Mixin.Name.Ident == "Attr"
}
