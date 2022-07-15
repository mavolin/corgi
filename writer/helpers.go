package writer

import "github.com/mavolin/corgi/corgi/file"

// isFirstContent reports whether the first written item in the passed scope
// writes content (as opposed to an attribute).
func isFirstContent(s file.Scope) bool {
	if len(s) == 0 {
		return false
	}

	switch itm := s[0].(type) {
	case file.And:
		return false
	case file.IfBlock:
		if isFirstContent(itm.Then) {
			return true
		}

		if itm.Else == nil {
			return false
		}

		return isFirstContent(itm.Else.Then)
	case file.If:
		if isFirstContent(itm.Then) {
			return true
		}

		for _, ei := range itm.ElseIfs {
			if isFirstContent(ei.Then) {
				return true
			}
		}

		if itm.Else == nil {
			return false
		}

		return isFirstContent(itm.Else.Then)
	case file.Switch:
		for _, c := range itm.Cases {
			if isFirstContent(c.Then) {
				return true
			}
		}

		if itm.Default == nil {
			return false
		}

		return isFirstContent(itm.Default.Then)
	case file.For:
		return isFirstContent(itm.Body)
	case file.While:
		return isFirstContent(itm.Body)
	case file.MixinCall:
		return isFirstContent(itm.Mixin.Body)
	case file.Mixin:
		return isFirstContent(s[1:])
	default:
		return true
	}
}
