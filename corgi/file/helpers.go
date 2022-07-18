package file

// IsFirstAnd reports whether the next written item is an &.
// If it encounters a conditional, it reports true if any of 'thens' have an &
// as first item.
//
// If it is unknown, it will return true.
func IsFirstAnd(s Scope) bool {
	if len(s) == 0 {
		return false
	}

	switch itm := s[0].(type) {
	case And:
		return true
	case IfBlock:
		if IsFirstAnd(itm.Then) {
			return true
		}

		if itm.Else == nil {
			return false
		}

		return IsFirstAnd(itm.Else.Then)
	case If:
		if IsFirstAnd(itm.Then) {
			return true
		}

		for _, ei := range itm.ElseIfs {
			if IsFirstAnd(ei.Then) {
				return true
			}
		}

		if itm.Else == nil {
			return false
		}

		return IsFirstAnd(itm.Else.Then)
	case Switch:
		for _, c := range itm.Cases {
			if IsFirstAnd(c.Then) {
				return true
			}
		}

		if itm.Default == nil {
			return false
		}

		return IsFirstAnd(itm.Default.Then)
	case For:
		return IsFirstAnd(itm.Body)
	case While:
		return IsFirstAnd(itm.Body)
	case MixinCall:
		return IsFirstAnd(itm.Mixin.Body)
	case Mixin:
		return IsFirstAnd(s[1:])
	case Code:
		return IsFirstAnd(s[1:])
	default:
		return false
	}
}
