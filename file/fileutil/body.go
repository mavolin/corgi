package fileutil

import "github.com/mavolin/corgi/file"

// Body returns the body of itm and true, if it has one, or nil and false, if
// it does not.
//
// For [file.If] and [file.IfBlock] it returns Then, and for [file.Switch] it
// returns (nil, false).
func Body(itm file.ScopeItem) (body file.Scope, has bool) {
	switch itm := itm.(type) {
	// bad_item.go
	case file.BadItem:
		return itm.Body, true

	// block.go
	case file.Block:
		return itm.Body, true
	case file.BlockExpansion:
		return file.Scope{itm.Item}, true

	// control_structures.go
	case file.If:
		return itm.Then, true
	case file.IfBlock:
		return itm.Then, true
	case file.For:
		return itm.Body, true

	// element.go
	case file.Element:
		return itm.Body, true
	case file.DivShorthand:
		return itm.Body, true

	// include.go
	case file.Include:
		cincl, ok := itm.Include.(file.CorgiInclude)
		if ok {
			return cincl.File.Scope, true
		}
		return nil, false

	// mixin.go
	case file.Mixin:
		return itm.Body, true
	case file.MixinCall:
		return itm.Body, true
	case file.MixinMainBlockShorthand:
		return itm.Body, true
	default:
		return nil, false
	}
}
