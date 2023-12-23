package template

import (
	"bytes"
	"io"

	"github.com/mavolin/corgi/escape"
	"github.com/mavolin/corgi/escape/safe"
)

type Context struct {
	err      error
	w        io.Writer
	nonce    string
	classBuf bytes.Buffer
	closed   bool
}

func NewContext(w io.Writer) *Context {
	return &Context{w: w, closed: true}
}

func (ctx *Context) SetScriptNonce(nonce any) {
	s, err := escape.PlainAttr(nonce)
	if err != nil {
		ctx.Panic(err)
	}
	ctx.nonce = s.Escaped()
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
	safeClass, err := escape.PlainAttr(class)
	if err != nil {
		ctx.Panic(err)
	}

	ctx.BufferClassAttr(safeClass)
}

// BufferClassAttr buffers the passed class(es), to be written to the class
// attribute when the start tag is closed.
//
// If this is the first call to BufferClassAttr after the start tag has been
// opened, BufferClassAttr implicitly calls [Context.Unclosed].
func (ctx *Context) BufferClassAttr(class safe.PlainAttr) {
	esc := class.Escaped()

	if ctx.classBuf.Len() > 0 {
		ctx.classBuf.Grow(len(esc) + 1)
		ctx.classBuf.WriteByte(' ')
		ctx.classBuf.WriteString(esc)
		return
	}

	ctx.classBuf.WriteString(esc)
	ctx.closed = false
}

// Unclosed signals that an element tag has been opened but not yet closed.
//
// Unclosed only needs to be called, if a call to CloseStartTag is expected.
func (ctx *Context) Unclosed() {
	ctx.closed = false
}

// Closed signals that an element tag has been closed.
//
// Closed only needs to be called, if a call to CloseStartTag is expected.
func (ctx *Context) Closed() {
	ctx.closed = true
}

// CloseStartTag writes the buffered classes, if any, plus, optionally, the
// passed pre-escaped extra classes and closes the start tag.
//
// If the tag has already been closed, CloseStartTag is a no-op.
//
// CloseStartTag needn't always be called to close a tag.
// If the tag is known to not have any buffered classes and the close state is
// definite, callers may choose to directly write the closing tag.
func (ctx *Context) CloseStartTag(extraClasses safe.PlainAttr, void bool) {
	if ctx.closed {
		return
	}
	ctx.closed = true

	extraClassesStr := extraClasses.Escaped()

	if ctx.classBuf.Len() > 0 {
		if len(extraClassesStr) > 0 {
			ctx.Write(` class="` + extraClassesStr + ` `)
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

	if len(extraClassesStr) > 0 {
		if void {
			ctx.Write(` class="` + extraClassesStr + `"/>`)
			return
		}
		ctx.Write(` class="` + extraClassesStr + `">`)
		return
	}

	if void {
		ctx.Write("/>")
		return
	}
	ctx.Write(">")
}

func WriteAny[T safe.BodyFragment](ctx *Context, escaper escape.Func[T], val any) {
	if escaper != nil {
		s, err := escaper(val)
		if err != nil {
			ctx.Panic(err)
		}
		ctx.Write(s.Escaped())
		return
	}

	s, err := escape.Stringify(val)
	if err != nil {
		ctx.Panic(err)
	}

	ctx.Write(s)
}

func WriteAnys[T safe.BodyFragment](ctx *Context, escaper escape.ContextFunc[T], vals ...any) {
	s, err := escaper(vals...)
	if err != nil {
		ctx.Panic(err)
	}
	ctx.Write(s.Escaped())
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
func WriteAttr[T safe.AttrFragment](ctx *Context, name string, val any, escaper escape.Func[T]) {
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

		ctx.Write(` ` + name + `="` + s.Escaped() + `"`)
		return
	}

	s, err := escape.Stringify(val)
	if err != nil {
		ctx.Panic(err)
	}

	ctx.Write(` ` + name + `="` + s + `"`)
}

func Must[T safe.Fragment](ctx *Context, f escape.Func[T], val any) T {
	t, err := f(val)
	if err != nil {
		ctx.Panic(err)
	}

	return t
}

func MustContext[T safe.Fragment](ctx *Context, f escape.ContextFunc[T], vals ...any) T {
	t, err := f(vals...)
	if err != nil {
		ctx.Panic(err)
	}

	return t
}
