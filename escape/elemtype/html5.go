package elemtype

// HTML5 is a [Func] that returns the [Type] for HTML5 elements, as defined
// in chapter 4 "[The elements of HTML]" of the HTML specification.
//
// [The elements of HTML]: https://html.spec.whatwg.org/multipage/indices.html#elements-3
func HTML5(element string) Type {
	switch element {
	//
	// 4.1 The document element

	case "html":
		return HTML

	//
	// 4.2 Document metadata

	case "head":
		return HTML
	case "title":
		return Text
	case "base":
		return Void
	case "link":
		return Void
	case "meta":
		return Void
	case "style":
		return CSS

	//
	// 4.3 Sections

	case "body":
		return HTML
	case "article":
		return HTML
	case "section":
		return HTML
	case "nav":
		return HTML
	case "aside":
		return HTML
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return HTML
	case "hgroup":
		return HTML
	case "header":
		return HTML
	case "footer":
		return HTML
	case "address":
		return HTML

	//
	// 4.4 Grouping content

	case "p":
		return HTML
	case "hr":
		return Void
	case "pre":
		return HTML
	case "blockquote":
		return HTML
	case "ol":
		return HTML
	case "ul":
		return HTML
	case "menu":
		return HTML
	case "li":
		return HTML
	case "dl":
		return HTML
	case "dt":
		return HTML
	case "dd":
		return HTML
	case "figure":
		return HTML
	case "figcaption":
		return HTML
	case "main":
		return HTML
	case "search":
		return HTML
	case "div":
		return HTML

	//
	// 4.5 Text-level semantics

	case "a":
		return HTML
	case "em":
		return HTML
	case "strong":
		return HTML
	case "small":
		return HTML
	case "s":
		return HTML
	case "cite":
		return HTML
	case "q":
		return HTML
	case "dfn":
		return HTML
	case "abbr":
		return HTML
	case "ruby":
		return HTML
	case "rt":
		return HTML
	case "rp":
		return Text
	case "data":
		return HTML
	case "time":
		return HTML
	case "code":
		return HTML
	case "var":
		return HTML
	case "samp":
		return HTML
	case "kbd":
		return HTML
	case "sub", "sup":
		return HTML
	case "i":
		return HTML
	case "b":
		return HTML
	case "u":
		return HTML
	case "mark":
		return HTML
	case "bdi":
		return HTML
	case "bdo":
		return HTML
	case "span":
		return HTML
	case "br":
		return Void
	case "wbr":
		return Void

	//
	// 4.6 Links

	// irrelevant

	//
	// 4.7 Edits

	case "ins":
		return HTML
	case "del":
		return HTML

	//
	// 4.8 Embedded content

	case "picture":
		return HTML
	case "source":
		return Void
	case "img":
		return Void
	case "iframe":
		return Nothing
	case "embed":
		return Void
	case "object":
		return HTML
	case "video":
		return HTML
	case "audio":
		return HTML
	case "track":
		return Void
	case "map":
		return HTML
	case "area":
		return Void
	// Excluding support for math and svg elements, simply because they add a
	// plethora of elements (sometimes even in conflict with HTML5 elements,
	// e.g. "title" has a different content model in SVG).
	// While those alone are still manageable, the additional amount of
	// attributes is not.
	// I expect most SVG images to be static, usually already in XML form,
	// so !raw is a better choice for those anyway.
	// As for MathML, I have never used it myself, but I'm going to speculate
	// that most people wouldn't bother writing MathML by hand (which seems
	// rather tedious), but use something like MathJax or a server-side
	// converter from TeX to MathML instead, so excluding MathML is probably
	// not a big deal either.
	// case "math":
	// case "svg":

	//
	// 4.9 Tabular data

	case "table":
		return HTML
	case "caption":
		return HTML
	case "colgroup":
		return HTML
	case "col":
		return Void
	case "tbody":
		return HTML
	case "thead":
		return HTML
	case "tfoot":
		return HTML
	case "tr":
		return HTML
	case "td":
		return HTML
	case "th":
		return HTML

	//
	// 4.10 Forms

	case "form":
		return HTML
	case "label":
		return HTML
	case "input":
		return Void
	case "button":
		return HTML
	case "select":
		return HTML
	case "datalist":
		return HTML
	case "optgroup":
		return HTML
	case "option":
		return Text
	case "textarea":
		return Text
	case "output":
		return HTML
	case "progress":
		return HTML
	case "meter":
		return HTML
	case "fieldset":
		return HTML
	case "legend":
		return HTML

	//
	// 4.11 Interactive elements

	case "details":
		return HTML
	case "summary":
		return HTML
	case "dialog":
		return HTML

	//
	// 4.12 Scripting

	case "script":
		return JS
	case "noscript":
		return HTML
	case "template":
		return HTML
	case "slot":
		return HTML
	case "canvas":
		return HTML

	default:
		return Unknown
	}
}

var _ Func = HTML5
