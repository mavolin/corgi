package write

import (
	"strings"

	"github.com/mavolin/corgi/woof"
)

type textEscaper struct {
	name string
	f    func(string) string
}

type expressionEscaper struct {
	funcName string
}

type contextEscaper struct {
	name     string
	funcName string

	normalizer func(string) string
	safeType   string
}

func toEscaperFunc[T ~string](f func(any) (T, error)) func(s string) string {
	return func(s string) string {
		esc, err := f(s)
		if err != nil {
			panic(err)
		}
		return string(esc)
	}
}

var (
	plainBodyTextEscaper = textEscaper{
		name: "plain body",
		f:    toEscaperFunc(woof.EscapeHTMLBody),
	}
	// Browsers will interpret HTML escapes in script elements as part of js
	// i.e. ignore them.
	//
	// Therefore, our regular body escaper is not suitable, as there is nothing
	// escapable (except for `</script>`).
	//
	// Since corgi auto-escapes all text in elements, the same expectation may
	// be put forth for escaping in scripts.
	// So in case anyone writes `</script>` in script texts, (which can only
	// happen in strings) we should correctly replace it with `<\/script>` to
	// prevent premature termination of the script.
	scriptBodyTextEscaper = textEscaper{
		name: "script body",
		f:    strings.NewReplacer(`</script>`, `<\/script>`).Replace,
	}
	styleBodyTextEscaper = textEscaper{
		name: "style body",
		f:    strings.NewReplacer(`</style>`, `<\/style>`).Replace,
	}

	attrTextEscaper = textEscaper{
		name: "html attr",
		f:    toEscaperFunc(woof.EscapeHTMLAttrVal),
	}

	htmlTextEscaper = textEscaper{
		name: "html",
		f:    toEscaperFunc(woof.EscapeHTML),
	}

	plainBodyExprEscaper  = expressionEscaper{funcName: "EscapeHTMLBody"}
	scriptBodyExprEscaper = expressionEscaper{funcName: "JSify"}
	styleBodyExprEscaper  = expressionEscaper{funcName: "FilterCSSValue"}

	plainAttrExprEscaper = expressionEscaper{funcName: "EscapeHTMLAttrVal"}
	jsAttrExprEscaper    = expressionEscaper{funcName: "EscapeJSAttrVal"}
	jsStrExprEscaper     = expressionEscaper{funcName: "EscapeJSStr"}
	cssExprEscaper       = expressionEscaper{funcName: "FilterCSSValue"}
	htmlExprEscaper      = expressionEscaper{funcName: "EscapeHTML"}
	urlAttrExprEscaper   = contextEscaper{
		name:       "url attr",
		funcName:   "FilterURL",
		normalizer: func(s string) string { return string(woof.NormalizeURL(woof.URL(s))) },
		safeType:   "URL",
	}
	srcsetAttrExprEscaper = contextEscaper{
		name:     "srcset attr",
		funcName: "FilterSrcset",
		safeType: "Srcset",
	}
)
