package walk

import "github.com/mavolin/corgi/file"

// Scope returns the scope of itm and true, if it has one, or (nil, false), if
// it does not.
//
// For [file.If] it returns Then, and for [file.Switch] it returns
// (nil, false).
func Scope(itm file.ScopeItem) (scope file.Scope, has bool) {
	switch itm := itm.(type) {
	// body.go
	case file.BadItem:
		scope, has = itm.Body.(file.Scope)
		return scope, has

	// control_structures.go
	case file.If:
		scope, has = itm.Then.(file.Scope)
		return scope, has
	case file.For:
		scope, has = itm.Body.(file.Scope)
		return scope, has

	// element.go
	case file.Element:
		scope, has = itm.Body.(file.Scope)
		return scope, has

	// component.go
	case file.Component:
		scope, has = itm.Body.(file.Scope)
		return scope, has
	case file.ComponentCall:
		scope, has = itm.Body.(file.Scope)
		return scope, has
	case file.Block:
		scope, has = itm.Body.(file.Scope)
		return scope, has
	default:
		return file.Scope{}, false
	}
}
