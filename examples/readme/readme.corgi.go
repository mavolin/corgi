package corgi

import (
	__corgi_io "io"
	__corgi_std_reflect "reflect"
	"strings"

	__corgi_woof "github.com/mavolin/corgi/woof"
)

func LearnCorgi(__corgi_w __corgi_io.Writer, name string, knowsPug bool, friends []string) error {
	__corgi_ctx := __corgi_woof.NewContext(__corgi_w)
	defer __corgi_ctx.Recover()
	var __corgi_mixin0 func(any, *string, *string)
	{
		listSep := ", "
		listLastSep := ", and "
		var __corgi_preMixin5 func(any, *string, *string)
		__corgi_preMixin5 = func(val any, __corgi_mixinParam_sep *string, __corgi_mixinParam_lastSep *string) {
			sep := __corgi_woof.ResolveDefault(__corgi_mixinParam_sep, listSep)
			lastSep := __corgi_woof.ResolveDefault(__corgi_mixinParam_lastSep, listLastSep)
			if val == nil {
				return
			}
			rval := __corgi_std_reflect.ValueOf(val)
			switch rval.Len() {
			case 0:
				return
			case 1:
				__corgi_ctx.CloseStartTag("", false)
				__corgi_woof.WriteAny(__corgi_ctx, rval.Index(0).Interface(), __corgi_woof.EscapeHTMLBody)
				__corgi_ctx.Closed()
				return
				__corgi_ctx.Closed()
			}
			__corgi_ctx.CloseStartTag("", false)
			__corgi_woof.WriteAny(__corgi_ctx, rval.Index(0).Interface(), __corgi_woof.EscapeHTMLBody)
			for i := 1; i < rval.Len()-1; i++ {
				__corgi_woof.WriteAny(__corgi_ctx, sep, __corgi_woof.EscapeHTMLBody)
				__corgi_woof.WriteAny(__corgi_ctx, rval.Index(i).Interface(), __corgi_woof.EscapeHTMLBody)
			}
			__corgi_woof.WriteAny(__corgi_ctx, lastSep, __corgi_woof.EscapeHTMLBody)
			__corgi_woof.WriteAny(__corgi_ctx, rval.Index(rval.Len()-1).Interface(), __corgi_woof.EscapeHTMLBody)
			__corgi_ctx.Closed()
		}

		__corgi_mixin0 = __corgi_preMixin5
		_ = __corgi_mixin0 // in case this is only a dependency of another mixin in this lib
	}
	__corgi_mixin1 := func(name string) {
		__corgi_ctx.CloseStartTag("", false)
		__corgi_ctx.Write("Hello, ")
		__corgi_woof.WriteAny(__corgi_ctx, name, __corgi_woof.EscapeHTMLBody)
		__corgi_ctx.Write("!")
		__corgi_ctx.Closed()
	}
	__corgi_ctx.Write("<!doctype html><html lang=en><head><title>Learn Corgi</title></head><body><h1>Learn Corgi</h1><p id=greeting")
	__corgi_ctx.BufferClassAttr("greeting")
	__corgi_ctx.Unclosed()
	if strings.HasPrefix(name, "M") {
		__corgi_ctx.BufferClassAttr("font-size--big")
	}
	__corgi_mixin1(name)
	__corgi_ctx.CloseStartTag("", false)
	__corgi_ctx.Write("</p><p")
	__corgi_ctx.Unclosed()
	if knowsPug {
		__corgi_ctx.Write(">")
		__corgi_woof.WriteAny(__corgi_ctx, name, __corgi_woof.EscapeHTMLBody)
		__corgi_ctx.Write(", since you already know pug,\nlearning corgi will be even more of <strong>a breeze</strong> for you! ")
		__corgi_ctx.Closed()
	}
	__corgi_ctx.CloseStartTag("", false)
	__corgi_ctx.Write("Head over to <a href=https://mavolin.gitbook.io/corgi>GitBook</a>\nto learn it.</p><p>And make sure to tell ")
	__corgi_mixin0(friends, nil, nil)
	__corgi_ctx.Write(" about corgi too!")
	__corgi_ctx.CloseStartTag("", false)
	__corgi_ctx.Write("</p></body></html>")
	return __corgi_ctx.Err()
}
