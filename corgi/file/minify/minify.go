// Package minify implements a utility to reduce the number of items in the
// file.
package minify

import "github.com/mavolin/corgi/corgi/file"

// Minify minifies the file:
//
// It collapses adjacent Text items into a single one.
//
// Furthermore, it directly adds & directives to the parent element, if
// possible.
//
// It does not minify extended, used, or included files.
func Minify(f *file.File) {
	f.Scope = minifyText(f.Scope)
	minifyAndScope(f.Scope)
}

func minifyText(s file.Scope) file.Scope {
Items:
	for i := 1; i < len(s); i++ {
		switch itm := s[i].(type) {
		case file.InlineText:
			if itm.NoEscape {
				continue Items
			}

			// check if we can collapse it with the previous item
			prevText, ok := s[i-1].(file.Text)
			if ok {
				prevText.Text += itm.Text
				s[i-1] = prevText

				copy(s[i:], s[i+1:])
				s = s[:len(s)-1]

				i--
			} else {
				// so we can collapse it next iteration, if possible
				s[i] = file.Text{Text: itm.Text}
			}
		case file.Text:
			// check if we can collapse it with the previous item
			prevText, ok := s[i-1].(file.Text)
			if !ok {
				continue Items
			}

			prevText.Text += itm.Text
			s[i-1] = prevText

			copy(s[i:], s[i+1:])
			s = s[:len(s)-1]

			i--
		case file.Block:
			itm.Body = minifyText(itm.Body)
			s[i] = itm
		case file.Element:
			itm.Body = minifyText(itm.Body)
			s[i] = itm
		case file.If:
			itm.Then = minifyText(itm.Then)

			for i, ei := range itm.ElseIfs {
				itm.ElseIfs[i].Then = minifyText(ei.Then)
			}

			if itm.Else != nil {
				itm.Else.Then = minifyText(itm.Else.Then)
			}

			s[i] = itm
		case file.IfBlock:
			itm.Then = minifyText(itm.Then)

			if itm.Else != nil {
				itm.Else.Then = minifyText(itm.Else.Then)
			}

			s[i] = itm
		case file.Switch:
			for _, c := range itm.Cases {
				c.Then = minifyText(c.Then)
			}

			if itm.Default != nil {
				itm.Default.Then = minifyText(itm.Default.Then)
			}

			s[i] = itm
		case file.For:
			itm.Body = minifyText(itm.Body)
			s[i] = itm
		case file.While:
			itm.Body = minifyText(itm.Body)
			s[i] = itm
		case file.MixinCall:
			itm.Body = minifyText(itm.Body)
			s[i] = itm
		}
	}

	return s
}

// minifyAndScope searches the scope for file.Elements and calls
// minifyAndElement on them.
func minifyAndScope(s file.Scope) {
	for i, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			minifyAndScope(itm.Body)
		case file.Element:
			itm = minifyAndElement(itm)
			s[i] = itm
		case file.If:
			minifyAndScope(itm.Then)

			for _, ei := range itm.ElseIfs {
				minifyAndScope(ei.Then)
			}

			if itm.Else != nil {
				minifyAndScope(itm.Else.Then)
			}
		case file.IfBlock:
			minifyAndScope(itm.Then)

			if itm.Else != nil {
				minifyAndScope(itm.Else.Then)
			}
		case file.Switch:
			for _, c := range itm.Cases {
				minifyAndScope(c.Then)
			}

			if itm.Default != nil {
				minifyAndScope(itm.Default.Then)
			}
		case file.For:
			minifyAndScope(itm.Body)
		case file.While:
			minifyAndScope(itm.Body)
		case file.MixinCall:
			minifyAndScope(itm.Body)
		}
	}
}

func minifyAndElement(e file.Element) file.Element {
	for i, itm := range e.Body {
		switch itm := itm.(type) {
		case file.And:
			e.Classes = append(e.Classes, itm.Classes...)
			e.Attributes = append(e.Attributes, itm.Attributes...)

			copy(e.Body[i:], e.Body[i+1:])
			e.Body = e.Body[:len(e.Body)-1]

		case file.Block:
			minifyAndScope(itm.Body)
		case file.Element:
			itm = minifyAndElement(itm)
			e.Body[i] = itm
		case file.If:
			minifyAndScope(itm.Then)

			for _, ei := range itm.ElseIfs {
				minifyAndScope(ei.Then)
			}

			if itm.Else != nil {
				minifyAndScope(itm.Else.Then)
			}
		case file.IfBlock:
			minifyAndScope(itm.Then)

			if itm.Else != nil {
				minifyAndScope(itm.Else.Then)
			}
		case file.Switch:
			for _, c := range itm.Cases {
				minifyAndScope(c.Then)
			}

			if itm.Default != nil {
				minifyAndScope(itm.Default.Then)
			}
		case file.For:
			minifyAndScope(itm.Body)
		case file.While:
			minifyAndScope(itm.Body)
		case file.MixinCall:
			minifyAndScope(itm.Body)
		}
	}

	return e
}
