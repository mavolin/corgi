package writer

import "github.com/mavolin/corgi/corgi/file"

// firstAlwaysWritesBody reports whether the first written element in the scope
// writes to the body of the element (as opposed to its attributes).
//
// If it encounters a conditional, it only reports true if all 'thens' write to
// the body including the default/else.
// If no default/else is present, it will return false.
//
// If it encounters a file.Block, it reports true if the block is an extend
// block, as those may never write attributes.
func (w *Writer) firstAlwaysWritesBody(s file.Scope) bool {
	if len(s) == 0 {
		return false
	}

	switch itm := s[0].(type) {
	case file.And:
		return false
	case file.IfBlock:
		if !w.firstAlwaysWritesBody(itm.Then) {
			return false
		}

		if itm.Else == nil {
			return false
		}

		return w.firstAlwaysWritesBody(itm.Else.Then)
	case file.If:
		if !w.firstAlwaysWritesBody(itm.Then) {
			return false
		}

		for _, ei := range itm.ElseIfs {
			if !w.firstAlwaysWritesBody(ei.Then) {
				return false
			}
		}

		if itm.Else == nil {
			return false
		}

		return w.firstAlwaysWritesBody(itm.Else.Then)
	case file.Switch:
		for _, c := range itm.Cases {
			if !w.firstAlwaysWritesBody(c.Then) {
				return false
			}
		}

		if itm.Default == nil {
			return false
		}

		return w.firstAlwaysWritesBody(itm.Default.Then)
	case file.For:
		return w.firstAlwaysWritesBody(itm.Body)
	case file.While:
		return w.firstAlwaysWritesBody(itm.Body)
	case file.MixinCall:
		return w.firstAlwaysWritesBody(itm.Mixin.Body)
	case file.Mixin:
		return w.firstAlwaysWritesBody(s[1:])
	case file.Code:
		return w.firstAlwaysWritesBody(s[1:])
	case file.Block:
		return false
	default:
		return true
	}
}
