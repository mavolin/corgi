package elemtype

// HTML5 is a [Func] that returns the [Type] for HTML5 elements, as defined
// in chapter 4 "[The elements of HTML]" of the HTML specification.
//
// [The elements of HTML]: https://html.spec.whatwg.org/multipage/semantics.html#semantics
func HTML5(element string) Type {
	switch element {
	//
	// 4.1 The document element

	case "html":
		return HTML // https://html.spec.whatwg.org/multipage/semantics.html#the-html-element

	//
	// 4.2 Document metadata

	case "head":
		return HTML // https://html.spec.whatwg.org/multipage/semantics.html#the-head-element
	case "title":
		return Text // https://html.spec.whatwg.org/multipage/semantics.html#the-title-element
	case "base":
		return Void // https://html.spec.whatwg.org/multipage/semantics.html#the-base-element
	case "link":
		return Void // https://html.spec.whatwg.org/multipage/semantics.html#the-link-element
	case "meta":
		return Void // https://html.spec.whatwg.org/multipage/semantics.html#the-meta-element
	case "style":
		return CSS // https://html.spec.whatwg.org/multipage/semantics.html#the-style-element

	//
	// 4.3 Sections

	case "body":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-body-element
	case "article":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-article-element
	case "section":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-section-element
	case "nav":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-nav-element
	case "aside":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-aside-element
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-h1,-h2,-h3,-h4,-h5,-and-h6-elements
	case "hgroup":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-hgroup-element
	case "header":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-header-element
	case "footer":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-footer-element
	case "address":
		return HTML // https://html.spec.whatwg.org/multipage/sections.html#the-address-element

	//
	// 4.4 Grouping content

	case "p":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-p-element
	case "hr":
		return Void // https://html.spec.whatwg.org/multipage/grouping-content.html#the-hr-element
	case "pre":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-pre-element
	case "blockquote":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-blockquote-element
	case "ol":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-ol-element
	case "ul":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-ul-element
	case "menu":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-menu-element
	case "li":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-li-element
	case "dl":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-dl-element
	case "dt":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-dt-element
	case "dd":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-dd-element
	case "figure":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-figure-element
	case "figcaption":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-figcaption-element
	case "main":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-main-element
	case "search":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-search-element
	case "div":
		return HTML // https://html.spec.whatwg.org/multipage/grouping-content.html#the-div-element

	//
	// 4.5 Text-level semantics

	case "a":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-a-element
	case "em":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-em-element
	case "strong":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-strong-element
	case "small":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-small-element
	case "s":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-s-element
	case "cite":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-cite-element
	case "q":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-q-element
	case "dfn":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-dfn-element
	case "abbr":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-abbr-element
	case "ruby":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-ruby-element
	case "rt":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-rt-element
	case "rp":
		return Text // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-rp-element
	case "data":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-data-element
	case "time":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-time-element
	case "code":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-code-element
	case "var":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-var-element
	case "samp":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-samp-element
	case "kbd":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-kbd-element
	case "sub", "sup":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-sub-and-sup-elements
	case "i":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-i-element
	case "b":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-b-element
	case "u":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-u-element
	case "mark":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-mark-element
	case "bdi":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-bdi-element
	case "bdo":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-bdo-element
	case "span":
		return HTML // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-span-element
	case "br":
		return Void // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-br-element
	case "wbr":
		return Void // https://html.spec.whatwg.org/multipage/text-level-semantics.html#the-wbr-element

	//
	// 4.6 Links

	// irrelevant

	//
	// 4.7 Edits

	case "ins":
		return HTML // https://html.spec.whatwg.org/multipage/edits.html#the-ins-element
	case "del":
		return HTML // https://html.spec.whatwg.org/multipage/edits.html#the-del-element

	//
	// 4.8 Embedded content

	case "picture":
		return HTML // https://html.spec.whatwg.org/multipage/embedded-content.html#the-picture-element
	case "source":
		return Void // https://html.spec.whatwg.org/multipage/embedded-content.html#the-source-element
	case "img":
		return Void // https://html.spec.whatwg.org/multipage/embedded-content.html#the-img-element
	case "iframe":
		return Nothing // https://html.spec.whatwg.org/multipage/iframe-embed-object.html#the-iframe-element
	case "embed":
		return Void // https://html.spec.whatwg.org/multipage/iframe-embed-object.html#the-embed-element
	case "object":
		return HTML // https://html.spec.whatwg.org/multipage/iframe-embed-object.html#the-object-element
	case "video":
		return HTML // https://html.spec.whatwg.org/multipage/media.html#the-video-element
	case "audio":
		return HTML // https://html.spec.whatwg.org/multipage/media.html#the-audio-element
	case "track":
		return Void // https://html.spec.whatwg.org/multipage/media.html#the-track-element
	case "map":
		return HTML // https://html.spec.whatwg.org/multipage/image-maps.html#the-map-element
	case "area":
		return Void // https://html.spec.whatwg.org/multipage/image-maps.html#the-area-element
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
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-table-element
	case "caption":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-caption-element
	case "colgroup":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-colgroup-element
	case "col":
		return Void // https://html.spec.whatwg.org/multipage/tabular-data.html#the-col-element
	case "tbody":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-tbody-element
	case "thead":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-thead-element
	case "tfoot":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-tfoot-element
	case "tr":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-tr-element
	case "td":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-td-element
	case "th":
		return HTML // https://html.spec.whatwg.org/multipage/tabular-data.html#the-th-element

	//
	// 4.10 Forms

	case "form":
		return HTML // https://html.spec.whatwg.org/multipage/forms.html#the-form-element
	case "label":
		return HTML // https://html.spec.whatwg.org/multipage/forms.html#the-label-element
	case "input":
		return Void // https://html.spec.whatwg.org/multipage/input.html
	case "button":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-button-element
	case "select":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-select-element
	case "datalist":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-datalist-element
	case "optgroup":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-optgroup-element
	case "option":
		return Text // https://html.spec.whatwg.org/multipage/form-elements.html#the-option-element
	case "textarea":
		return Text // https://html.spec.whatwg.org/multipage/form-elements.html#the-textarea-element
	case "output":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-output-element
	case "progress":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-progress-element
	case "meter":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-meter-element
	case "fieldset":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-fieldset-element
	case "legend":
		return HTML // https://html.spec.whatwg.org/multipage/form-elements.html#the-legend-element

	//
	// 4.11 Interactive elements

	case "details":
		return HTML // https://html.spec.whatwg.org/multipage/interactive-elements.html#the-details-element
	case "summary":
		return HTML // https://html.spec.whatwg.org/multipage/interactive-elements.html#the-summary-element
	case "dialog":
		return HTML // https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element

	//
	// 4.12 Scripting

	case "script":
		return JS // https://html.spec.whatwg.org/multipage/scripting.html#the-script-element
	case "noscript":
		return HTML // https://html.spec.whatwg.org/multipage/scripting.html#the-noscript-element
	case "template":
		return HTML // https://html.spec.whatwg.org/multipage/scripting.html#the-template-element
	case "slot":
		return HTML // https://html.spec.whatwg.org/multipage/scripting.html#the-slot-element
	case "canvas":
		return HTML // https://html.spec.whatwg.org/multipage/scripting.html#the-canvas-element

	default:
		return Unknown
	}
}

var _ Func = HTML5
