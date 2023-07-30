package woof

import "strings"

type ContentType uint8

const (
	ContentTypePlain ContentType = iota
	ContentTypeCSS
	ContentTypeHTML
	ContentTypeJS
	ContentTypeURL
	ContentTypeSrcset
	// // ContentTypeUnsafe is used in attr.go for values that affect how
	// // embedded content and network messages are formed, vetted,
	// // or interpreted; or which credentials network messages carry.
	//
	// html/template doesn't seem to treat ContentTypeUnsafe any differently
	// (at least the constant is never used outside the map and some test with
	// unsafe attrs showed no difference in treatment to plain attrs) so I'm
	// gonna ignore them fully and place trust on the dev to think about what
	// they're doing.
	// Maybe a lint rule isn't a half bad solution for this.
	// ContentTypeUnsafe
)

// AttrTypes describes the value of the given attribute.
// If an attribute affects (or can mask) the encoding or interpretation of
// other content, or affects the contents, idempotency, or credentials of a
// network message, then the value in this map is ContentTypeUnsafe.
// This map is derived from HTML5, specifically
// https://www.w3.org/TR/html5/Overview.html#attributes-1
// as well as "%URI"-typed attributes from
// https://www.w3.org/TR/html4/index/attributes.html
var AttrTypes = map[string]ContentType{
	"accept": ContentTypePlain,
	// "accept-charset":  ContentTypeUnsafe,
	"action":  ContentTypeURL,
	"alt":     ContentTypePlain,
	"archive": ContentTypeURL,
	// "async":           ContentTypeUnsafe,
	"autocomplete": ContentTypePlain,
	"autofocus":    ContentTypePlain,
	"autoplay":     ContentTypePlain,
	"background":   ContentTypeURL,
	"border":       ContentTypePlain,
	"checked":      ContentTypePlain,
	"cite":         ContentTypeURL,
	// "challenge":       ContentTypeUnsafe,
	// "charset":         ContentTypeUnsafe,
	"class":    ContentTypePlain,
	"classid":  ContentTypeURL,
	"codebase": ContentTypeURL,
	"cols":     ContentTypePlain,
	"colspan":  ContentTypePlain,
	// "content":         ContentTypeUnsafe,
	"contenteditable": ContentTypePlain,
	"contextmenu":     ContentTypePlain,
	"controls":        ContentTypePlain,
	"coords":          ContentTypePlain,
	// "crossorigin":     ContentTypeUnsafe,
	"data":     ContentTypeURL,
	"datetime": ContentTypePlain,
	"default":  ContentTypePlain,
	// "defer":           ContentTypeUnsafe,
	"dir":       ContentTypePlain,
	"dirname":   ContentTypePlain,
	"disabled":  ContentTypePlain,
	"draggable": ContentTypePlain,
	"dropzone":  ContentTypePlain,
	// "enctype":         ContentTypeUnsafe,
	"for": ContentTypePlain,
	// "form":            ContentTypeUnsafe,
	"formaction": ContentTypeURL,
	// "formenctype":     ContentTypeUnsafe,
	// "formmethod":      ContentTypeUnsafe,
	// "formnovalidate":  ContentTypeUnsafe,
	"formtarget": ContentTypePlain,
	"headers":    ContentTypePlain,
	"height":     ContentTypePlain,
	"hidden":     ContentTypePlain,
	"high":       ContentTypePlain,
	"href":       ContentTypeURL,
	"hreflang":   ContentTypePlain,
	// "http-equiv":      ContentTypeUnsafe,
	"icon":  ContentTypeURL,
	"id":    ContentTypePlain,
	"ismap": ContentTypePlain,
	// "keytype":         ContentTypeUnsafe,
	"kind":  ContentTypePlain,
	"label": ContentTypePlain,
	"lang":  ContentTypePlain,
	// "language":        ContentTypeUnsafe,
	"list":       ContentTypePlain,
	"longdesc":   ContentTypeURL,
	"loop":       ContentTypePlain,
	"low":        ContentTypePlain,
	"manifest":   ContentTypeURL,
	"max":        ContentTypePlain,
	"maxlength":  ContentTypePlain,
	"media":      ContentTypePlain,
	"mediagroup": ContentTypePlain,
	// "method":          ContentTypeUnsafe,
	"min":      ContentTypePlain,
	"multiple": ContentTypePlain,
	"name":     ContentTypePlain,
	// "novalidate":      ContentTypeUnsafe,
	// Skip handler names from
	// https://www.w3.org/TR/html5/webappapis.html#event-handlers-on-elements,-document-objects,-and-window-objects
	// since we have special handling in AttrType.
	"open":    ContentTypePlain,
	"optimum": ContentTypePlain,
	// "pattern":     ContentTypeUnsafe,
	"placeholder": ContentTypePlain,
	"poster":      ContentTypeURL,
	"profile":     ContentTypeURL,
	"preload":     ContentTypePlain,
	"pubdate":     ContentTypePlain,
	"radiogroup":  ContentTypePlain,
	"readonly":    ContentTypePlain,
	// "rel":         ContentTypeUnsafe,
	"required": ContentTypePlain,
	"reversed": ContentTypePlain,
	"rows":     ContentTypePlain,
	"rowspan":  ContentTypePlain,
	// "sandbox":     ContentTypeUnsafe,
	"spellcheck": ContentTypePlain,
	"scope":      ContentTypePlain,
	"scoped":     ContentTypePlain,
	"seamless":   ContentTypePlain,
	"selected":   ContentTypePlain,
	"shape":      ContentTypePlain,
	"size":       ContentTypePlain,
	"sizes":      ContentTypePlain,
	"span":       ContentTypePlain,
	"src":        ContentTypeURL,
	"srcdoc":     ContentTypeHTML,
	"srclang":    ContentTypePlain,
	"srcset":     ContentTypeSrcset,
	"start":      ContentTypePlain,
	"step":       ContentTypePlain,
	"style":      ContentTypeCSS,
	"tabindex":   ContentTypePlain,
	"target":     ContentTypePlain,
	"title":      ContentTypePlain,
	// "type":        ContentTypeUnsafe,
	"usemap": ContentTypeURL,
	// "value":       ContentTypeUnsafe,
	"width": ContentTypePlain,
	"wrap":  ContentTypePlain,
	"xmlns": ContentTypeURL,
}

// AttrType returns a conservative (upper-bound on authority) guess at the
// type of the lowercase named attribute.
func AttrType(name string) ContentType {
	if strings.HasPrefix(name, "data-") {
		// Strip data- so that custom attribute heuristics below are
		// widely applied.
		// Treat data-action as URL below.
		name = name[5:]
	} else if prefix, short, ok := strings.Cut(name, ":"); ok {
		if prefix == "xmlns" {
			return ContentTypeURL
		}
		// Treat svg:href and xlink:href as href below.
		name = short
	}
	if t, ok := AttrTypes[name]; ok {
		return t
	}
	// Treat partial event handler names as script.
	if strings.HasPrefix(name, "on") {
		return ContentTypeJS
	}

	// Heuristics to prevent "javascript:..." injection in custom
	// data attributes and custom attributes like g:tweetUrl.
	// https://www.w3.org/TR/html5/dom.html#embedding-custom-non-visible-data-with-the-data-*-attributes
	// "Custom data attributes are intended to store custom data
	//  private to the page or application, for which there are no
	//  more appropriate attributes or elements."
	// Developers seem to store URL content in data URLs that start
	// or end with "URI" or "URL".
	if strings.Contains(name, "src") ||
		strings.Contains(name, "uri") ||
		strings.Contains(name, "url") {
		return ContentTypeURL
	}
	return ContentTypePlain
}
