package writer

import "github.com/mavolin/corgi/corgi/file"

// isFirstAnd reports whether the next written item is an &.
// If it encounters a conditional, it reports true if any of 'thens' have an &
// as first item.
func isFirstAnd(s file.Scope) bool {
	if len(s) == 0 {
		return false
	}

	switch itm := s[0].(type) {
	case file.And:
		return true
	case file.IfBlock:
		if isFirstAnd(itm.Then) {
			return true
		}

		if itm.Else == nil {
			return false
		}

		return isFirstAnd(itm.Else.Then)
	case file.If:
		if isFirstAnd(itm.Then) {
			return true
		}

		for _, ei := range itm.ElseIfs {
			if isFirstAnd(ei.Then) {
				return true
			}
		}

		if itm.Else == nil {
			return false
		}

		return isFirstAnd(itm.Else.Then)
	case file.Switch:
		for _, c := range itm.Cases {
			if isFirstAnd(c.Then) {
				return true
			}
		}

		if itm.Default == nil {
			return false
		}

		return isFirstAnd(itm.Default.Then)
	case file.For:
		return isFirstAnd(itm.Body)
	case file.While:
		return isFirstAnd(itm.Body)
	case file.MixinCall:
		return isFirstAnd(itm.Mixin.Body)
	case file.Mixin:
		return isFirstAnd(s[1:])
	case file.Code:
		return isFirstAnd(s[1:])
	default:
		return false
	}
}

// firstAlwaysWritesBody reports whether the first written element in the scope
// writes to the body of the element (as opposed to its attributes).
//
// If it encounters a conditional, it only reports true if all 'thens' write to
// the body including the default/else.
// If no default/else is present, it will return false.
func firstAlwaysWritesBody(s file.Scope) bool {
	if len(s) == 0 {
		return false
	}

	switch itm := s[0].(type) {
	case file.And:
		return false
	case file.IfBlock:
		if !firstAlwaysWritesBody(itm.Then) {
			return false
		}

		if itm.Else == nil {
			return false
		}

		return firstAlwaysWritesBody(itm.Else.Then)
	case file.If:
		if !firstAlwaysWritesBody(itm.Then) {
			return false
		}

		for _, ei := range itm.ElseIfs {
			if !firstAlwaysWritesBody(ei.Then) {
				return false
			}
		}

		if itm.Else == nil {
			return false
		}

		return firstAlwaysWritesBody(itm.Else.Then)
	case file.Switch:
		for _, c := range itm.Cases {
			if !firstAlwaysWritesBody(c.Then) {
				return false
			}
		}

		if itm.Default == nil {
			return false
		}

		return firstAlwaysWritesBody(itm.Default.Then)
	case file.For:
		return firstAlwaysWritesBody(itm.Body)
	case file.While:
		return firstAlwaysWritesBody(itm.Body)
	case file.MixinCall:
		return firstAlwaysWritesBody(itm.Mixin.Body)
	case file.Mixin:
		return firstAlwaysWritesBody(s[1:])
	case file.Code:
		return firstAlwaysWritesBody(s[1:])
	default:
		return true
	}
}
