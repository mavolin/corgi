package woof

import (
	"bytes"
	"io"
)

type Context struct {
	err      error
	w        io.Writer
	nonce    string
	classBuf bytes.Buffer
	closed   bool
	inAttr   bool
}

func NewContext(w io.Writer) *Context {
	return &Context{w: w, closed: true}
}

func (ctx *Context) SetScriptNonce(nonce any) {
	s, err := Stringify(nonce)
	if err != nil {
		ctx.Panic(err)
	}
	ctx.nonce = htmlAttrValEscaper.Replace(s)
}

func (ctx *Context) InjectNonce() {
	if ctx.nonce == "" {
		return
	}
	ctx.Write(` nonce="` + ctx.nonce + `"`)
}

func (ctx *Context) Panic(err error) {
	ctx.err = err
	panic(err)
}

func (ctx *Context) Recover() error {
	if ctx.err != nil {
		_ = recover()
		return ctx.err
	}

	return nil
}

func (ctx *Context) Write(s string) {
	if _, err := io.WriteString(ctx.w, s); err != nil {
		ctx.Panic(err)
	}
}

func (ctx *Context) WriteBytes(data []byte) {
	if _, err := ctx.w.Write(data); err != nil {
		ctx.Panic(err)
	}
}

func (ctx *Context) BufferClass(class any) {
	classStr, err := EscapeHTMLAttrVal(class)
	if err != nil {
		ctx.Panic(err)
	}

	ctx.BufferClassAttr(classStr)
}

func (ctx *Context) BufferClassAttr(class HTMLAttrVal) {
	if ctx.classBuf.Len() > 0 {
		ctx.classBuf.WriteString(string(" " + class))
		return
	}

	ctx.classBuf.WriteString(string(class))
}

// StartAttribute is a helper signalling that a mixin is about to be called that
// is being used as an attribute value.
//
// It causes calls to [Context.Unclosed], [Context.Closed], and
// [Context.CloseStartTag] made by the mixin to be ignored.
func (ctx *Context) StartAttribute() {
	ctx.inAttr = true
}

// EndAttribute reverts the effects of [Context.StartAttribute].
func (ctx *Context) EndAttribute() {
	ctx.inAttr = false
}

// Unclosed signals that an element tag has been opened but not yet closed.
func (ctx *Context) Unclosed() {
	if !ctx.inAttr {
		ctx.closed = false
	}
}

// Closed signals that an element tag has been closed.
func (ctx *Context) Closed() {
	if !ctx.inAttr {
		ctx.closed = true
	}
}

// CloseStartTag writes the buffered classes, if any, plus, optionally, the
// passed pre-escaped extra classes and closes the start tag.
//
// If the tag has already been closed, CloseStartTag is a no-op.
func (ctx *Context) CloseStartTag(extraClasses HTMLAttrVal, void bool) {
	if ctx.closed || ctx.inAttr {
		return
	}
	ctx.closed = true

	if ctx.classBuf.Len() > 0 {
		if len(extraClasses) > 0 {
			ctx.Write(` class="` + string(extraClasses) + ` `)
			ctx.WriteBytes(ctx.classBuf.Bytes())
		} else {
			ctx.Write(` class="`)
			ctx.WriteBytes(ctx.classBuf.Bytes())
		}
		if void {
			ctx.Write(`"/>`)
			return
		}
		ctx.Write(`">`)

		ctx.classBuf.Reset()
		return
	}

	if len(extraClasses) > 0 {
		if void {
			ctx.Write(` class="` + string(extraClasses) + `"/>`)
			return
		}
		ctx.Write(` class="` + string(extraClasses) + `">`)
		return
	}

	if void {
		ctx.Write("/>")
		return
	}
	ctx.Write(">")
}

func WriteAny[T ~string](ctx *Context, escaper func(val any) (T, error), val any) {
	if escaper != nil {
		s, err := escaper(val)
		if err != nil {
			ctx.Panic(err)
		}
		ctx.Write(string(s))
		return
	}

	s, err := Stringify(val)
	if err != nil {
		ctx.Panic(err)
	}

	ctx.Write(s)
}

func WriteAnys[T ~string](ctx *Context, escaper func(vals ...any) (T, error), vals ...any) {
	s, err := escaper(vals...)
	if err != nil {
		ctx.Panic(err)
	}
	ctx.Write(string(s))
}

// WriteAttr is a utility for writing attributes, that correctly handles bool
// values.
//
// If val is a bool and true, WriteAttr writes a space followed by name.
//
// If val is a bool and false, WriteAttr writes nothing.
//
// In any other case, WriteAttr writes a space, followed by the name, '="', and
// then the escaped attributes.
func WriteAttr[T ~string](ctx *Context, name string, val any, escaper func(val any) (T, error)) {
	if b, ok := val.(bool); ok {
		if !b {
			return
		}

		ctx.Write(" " + name)
		return
	}

	if escaper != nil {
		s, err := escaper(val)
		if err != nil {
			ctx.Panic(err)
		}

		ctx.Write(` ` + name + `="` + string(s) + `"`)
		return
	}

	s, err := Stringify(val)
	if err != nil {
		ctx.Panic(err)
	}

	ctx.Write(` ` + name + `="` + s + `"`)
}

func Must[T ~string](ctx *Context, f func(val any) (T, error), val any) string {
	t, err := f(val)
	if err != nil {
		ctx.Panic(err)
	}

	return string(t)
}

func MustContext[T ~string](ctx *Context, f func(vals ...any) (T, error), vals ...any) string {
	t, err := f(vals...)
	if err != nil {
		ctx.Panic(err)
	}

	return string(t)
}
