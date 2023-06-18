// Package voidelem provides utilities for working with HTML void elements.
package voidelem

// https://developer.mozilla.org/en-US/docs/Glossary/Empty_element
var elems = map[string]struct{}{
	"area":   {},
	"base":   {},
	"br":     {},
	"col":    {},
	"embed":  {},
	"hr":     {},
	"img":    {},
	"input":  {},
	"link":   {},
	"meta":   {},
	"param":  {},
	"source": {},
	"track":  {},
	"wbr":    {},
}

// Is reports whether the element with the passed name is an HTML void element.
func Is(name string) bool {
	_, ok := elems[name]
	return ok
}
