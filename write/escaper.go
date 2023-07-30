package write

import "github.com/mavolin/corgi/woof"

type escaper struct {
	f    func(s string) string
	name string
}

func (esc *escaper) escape(ctx *ctx, s string) string {
	if esc == nil {
		return s
	}
	escaped := esc.f(s)
	ctx.debug(esc.name, s+" -> "+escaped)
	return escaped
}
func (esc *escaper) qualName(ctx *ctx) string { return ctx.woofQual(esc.name) }

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
	bodyEscaper = &escaper{
		f:    toEscaperFunc(woof.EscapeHTMLBody),
		name: "EscapeHTMLBody",
	}

	attrEscaper = &escaper{
		f:    toEscaperFunc(woof.EscapeHTMLAttrVal),
		name: "EscapeHTMLAttrVal",
	}

	cssEscaper = &escaper{
		f:    toEscaperFunc(woof.FilterCSSValue),
		name: "FilterCSSValue",
	}
	htmlEscaper = &escaper{
		f:    toEscaperFunc(woof.EscapeHTML),
		name: "EscapeHTML",
	}
	jsEscaper = &escaper{
		f:    toEscaperFunc(woof.JSify),
		name: "JSify",
	}
)
