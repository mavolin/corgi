package attrtype

import "strings"

// HTML5 is a [Func] that returns the [Type] for HTML5 attributes as defined in
// the [HTML5 Attribute Spec].
//
// [HTML5 Attribute Spec]: https://www.w3.org/TR/html5/Overview.html#attributes-1
func HTML5(element, attr string) Type {
	if m, ok := html5ElementTypes[element]; ok {
		if t := m[attr]; t.IsValid() {
			return t
		}
	}

	if t := html5Types[attr]; t.IsValid() {
		return t
	}

	if strings.HasPrefix(attr, "on") {
		if _, ok := html5EventHandlers[attr[2:]]; ok {
			return JS
		}
	}

	return Unknown
}

var _ Func = HTML5

// The below maps describe the value of a given attribute.
//
// Precedence: html5ElementTypes > html5Types > html5EventHandlers
var (
	// element-specific types
	// https://html.spec.whatwg.org/multipage/indices.html#attributes-3
	html5ElementTypes = map[ /* element */ string]map[ /* attr */ string]Type{
		"base": {
			"href": ResourceURL,
		},
		"iframe": {
			"src": ResourceURL,
		},
		"link": {
			"href": ResourceURL,
		},
		"meter": {
			"value": Plain,
		},
		"script": {
			"src": ResourceURL,
		},
	}

	// HTML5 spec, excl. event handlers
	//
	// https://html.spec.whatwg.org/multipage/indices.html#attributes-3
	html5Types = map[string]Type{
		"abbr":                     Plain,
		"accept":                   Plain,
		"accept-charset":           Unsafe,
		"accesskey":                Plain,
		"action":                   URL,
		"allow":                    Unsafe,
		"allowfullscreen":          Plain,
		"alt":                      Plain,
		"as":                       Plain,
		"async":                    Unsafe,
		"autocapitalize":           Plain,
		"autocomplete":             Plain,
		"autofocus":                Plain,
		"autoplay":                 Plain,
		"blocking":                 Plain,
		"charset":                  Unsafe,
		"checked":                  Plain,
		"cite":                     URL,
		"class":                    Plain,
		"color":                    Plain,
		"cols":                     Plain,
		"colspan":                  Plain,
		"content":                  Unsafe,
		"contenteditable":          Plain,
		"controls":                 Plain,
		"coords":                   Plain,
		"crossorigin":              Unsafe,
		"data":                     URL,
		"datetime":                 Plain,
		"decoding":                 Plain,
		"default":                  Plain,
		"defer":                    Unsafe,
		"dir":                      Plain,
		"dirname":                  Plain,
		"disabled":                 Plain,
		"download":                 Plain,
		"draggable":                Plain,
		"enctype":                  Unsafe,
		"enterkeyhint":             Plain,
		"fetchpriority":            Plain,
		"for":                      Plain,
		"form":                     Unsafe,
		"formaction":               URL,
		"formenctype":              Unsafe,
		"formmethod":               Unsafe,
		"formnovalidate":           Unsafe,
		"formtarget":               Plain,
		"headers":                  Plain,
		"height":                   Plain,
		"hidden":                   Plain,
		"high":                     Plain,
		"href":                     URL, // overridden by elementAttrTypes
		"hreflang":                 Plain,
		"http-equiv":               Unsafe,
		"id":                       Plain,
		"imagesizes":               Plain,
		"imagesrcset":              Srcset,
		"inert":                    Plain,
		"inputmode":                Plain,
		"integrity":                Unsafe,
		"is":                       Plain,
		"ismap":                    Plain,
		"itemid":                   URL,
		"itemprop":                 URLList,
		"itemref":                  Plain,
		"itemscope":                Plain,
		"itemtype":                 URLList,
		"kind":                     Plain,
		"label":                    Plain,
		"lang":                     Plain,
		"list":                     Plain,
		"loading":                  Plain,
		"loop":                     Plain,
		"low":                      Plain,
		"max":                      Plain,
		"maxlength":                Plain,
		"media":                    Plain,
		"method":                   Unsafe,
		"min":                      Plain,
		"minlength":                Plain,
		"multiple":                 Plain,
		"muted":                    Plain,
		"name":                     Plain,
		"nomodule":                 Unsafe,
		"nonce":                    Unsafe,
		"novalidate":               Unsafe,
		"open":                     Plain,
		"optimum":                  Plain,
		"pattern":                  Unsafe,
		"ping":                     URLList,
		"placeholder":              Plain,
		"playsinline":              Plain,
		"popover":                  Plain,
		"popovertarget":            Plain,
		"popovertargetaction":      Plain,
		"poster":                   URL,
		"preload":                  Plain,
		"readonly":                 Plain,
		"referrerpolicy":           Unsafe,
		"rel":                      Unsafe,
		"required":                 Plain,
		"reversed":                 Plain,
		"rows":                     Plain,
		"rowspan":                  Plain,
		"sandbox":                  Unsafe,
		"scope":                    Plain,
		"selected":                 Plain,
		"shadowrootmode":           Plain,
		"shadowrootdelegatesfocus": Plain,
		"shape":                    Plain,
		"size":                     Plain,
		"sizes":                    Plain,
		"slot":                     Plain,
		"span":                     Plain,
		"spellcheck":               Plain,
		"src":                      URL,    // overridden by elementAttrTypes
		"srcdoc":                   Unsafe, // only unsafe, bc corgi can't provide context-sensitive escaping
		"srclang":                  Plain,
		"srcset":                   Srcset,
		"start":                    Plain,
		"step":                     Plain,
		"style":                    CSS,
		"tabindex":                 Plain,
		"target":                   Plain,
		"title":                    Plain,
		"translate":                Plain,
		"type":                     Unsafe,
		"usemap":                   URL,
		"value":                    Plain, // differs from html/template
		"width":                    Plain,
		"wrap":                     Plain,
	}

	// https://html.spec.whatwg.org/multipage/indices.html#attributes-3
	html5EventHandlers = map[string]struct{}{
		"auxclick":                {},
		"afterprint":              {},
		"beforematch":             {},
		"beforeprint":             {},
		"beforeunload":            {},
		"beforetoggle":            {},
		"blur":                    {},
		"cancel":                  {},
		"canplay":                 {},
		"canplaythrough":          {},
		"change":                  {},
		"click":                   {},
		"close":                   {},
		"contextlost":             {},
		"contextmenu":             {},
		"contextrestored":         {},
		"copy":                    {},
		"cuechange":               {},
		"cut":                     {},
		"dblclick":                {},
		"drag":                    {},
		"dragend":                 {},
		"dragenter":               {},
		"dragleave":               {},
		"dragover":                {},
		"dragstart":               {},
		"drop":                    {},
		"durationchange":          {},
		"emptied":                 {},
		"ended":                   {},
		"error":                   {},
		"focus":                   {},
		"formdata":                {},
		"hashchange":              {},
		"input":                   {},
		"invalid":                 {},
		"keydown":                 {},
		"keypress":                {},
		"keyup":                   {},
		"languagechange":          {},
		"load":                    {},
		"loadeddata":              {},
		"loadedmetadata":          {},
		"loadstart":               {},
		"message":                 {},
		"messageerror":            {},
		"mousedown":               {},
		"mouseenter":              {},
		"mouseleave":              {},
		"mousemove":               {},
		"mouseout":                {},
		"mouseover":               {},
		"mouseup":                 {},
		"offline":                 {},
		"online":                  {},
		"pagehide":                {},
		"pagereveal":              {},
		"pageshow":                {},
		"paste":                   {},
		"pause":                   {},
		"play":                    {},
		"playing":                 {},
		"popstate":                {},
		"progress":                {},
		"ratechange":              {},
		"reset":                   {},
		"resize":                  {},
		"rejectionhandled":        {},
		"scroll":                  {},
		"scrollend":               {},
		"securitypolicyviolation": {},
		"seeked":                  {},
		"seeking":                 {},
		"select":                  {},
		"slotchange":              {},
		"stalled":                 {},
		"storage":                 {},
		"submit":                  {},
		"suspend":                 {},
		"timeupdate":              {},
		"toggle":                  {},
		"unhandledrejection":      {},
		"unload":                  {},
		"volumechange":            {},
		"waiting":                 {},
		"wheel":                   {},
	}
)
