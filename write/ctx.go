package write

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/internal/stack"
	"github.com/mavolin/corgi/internal/voidelem"
	"github.com/mavolin/corgi/woof"
)

const (
	ctxVar = "ctx"
)

// A ctx is used to convert a corgi file or scope item to Go code.
//
// It must not be used concurrently.
type ctx struct {
	out io.Writer

	destPackage string

	identPrefix  string
	gotoCounter  int
	mixinCounter int

	hasNonce bool

	// _stack are the files we are currently in.
	//
	// If len(stack) > 0, then _stack[0] is the base template,
	// _stack[1:len(stack)-1] are other templates extending each other
	// (_stack[1] extends _stack[0], _stack[2] extends _stack[1], etc.), and
	// _stack[len(_stack)-1] is the main file.
	_stack []*file.File
	// stackStart is the start of the file stack for the current file.
	//
	// For example, if we are generating from the main file, stackStart will be
	// len(_stack)-1.
	stackStart int

	// mixin is the mixin we are currently generating.
	mixin *file.Mixin

	scopes         *stack.Stack[*nesting]
	classBuf       strings.Builder
	calledUnclosed bool

	mixinFuncNames mixinFuncMap

	allowedFilters  []string
	allowAllFilters bool

	cli            bool
	corgierrPretty corgierr.PrettyOptions

	debugEnabled bool

	generateBuf bytes.Buffer
}

type nesting struct {
	elemName       string
	customVoidElem bool
	exprEscaper    expressionEscaper
	txtEscaper     textEscaper
	startClosed    closeState
	haveBufClasses bool
	shallow        bool
}

type mixinFuncMap struct {
	m     map[string]map[string]string
	scope map[*file.File]*list.List[map[string]string]
}

func (mfm *mixinFuncMap) mixin(ctx *ctx, mc file.MixinCall) string {
	var module, pathInModule string
	if lib := mc.Mixin.File.Library; lib != nil {
		module, pathInModule = lib.Module, lib.PathInModule
	} else {
		module, pathInModule = mc.Mixin.File.Module, mc.Mixin.File.PathInModule
	}

	return mfm.mixinByName(ctx, module, pathInModule, mc.Name.Ident)
}

func (mfm *mixinFuncMap) mixinByName(ctx *ctx, module, pathInModule, name string) string {
	if module == ctx.currentFile().Module && pathInModule == ctx.currentFile().PathInModule {
		scopeMixins := mfm.scope[ctx.currentFile()]
		for e := scopeMixins.Back(); e != nil; e = e.Prev() {
			if funcName := e.V()[name]; funcName != "" {
				return funcName
			}
		}
	}

	p := module + "/" + pathInModule
	return mfm.m[p][name]
}

func (mfm *mixinFuncMap) startScope(ctx *ctx) {
	scopeMixins := mfm.scope[ctx.currentFile()]
	if scopeMixins == nil {
		scopeMixins = new(list.List[map[string]string])
		mfm.scope[ctx.currentFile()] = scopeMixins
	}
	scopeMixins.PushBack(make(map[string]string))
}

func (mfm *mixinFuncMap) addScope(ctx *ctx, m *file.Mixin) string {
	scopeMixins := mfm.scope[ctx.currentFile()]
	currScope := scopeMixins.Back()

	varName := ctx.nextMixinIdent()
	currScope.V()[m.Name.Ident] = varName
	return varName
}

func (mfm *mixinFuncMap) endScope(ctx *ctx) {
	scopeMixins := mfm.scope[ctx.currentFile()]
	scopeMixins.Remove(scopeMixins.Back())
}

type closeState uint8

const (
	unclosed closeState = iota
	maybeClosed
	closed
)

func (s closeState) String() string {
	switch s {
	case unclosed:
		return "unclosed"
	case maybeClosed:
		return "maybe closed"
	case closed:
		return "closed"
	default:
		return ""
	}
}

func newCtx(o Options) *ctx {
	return &ctx{
		identPrefix: o.IdentPrefix,
		scopes: stack.New1(&nesting{
			exprEscaper: plainBodyExprEscaper,
			txtEscaper:  plainBodyTextEscaper,
			startClosed: closed,
		}),
		mixinFuncNames: mixinFuncMap{
			m:     make(map[string]map[string]string),
			scope: make(map[*file.File]*list.List[map[string]string]),
		},
		allowedFilters:  o.AllowedFilters,
		allowAllFilters: o.AllowAllFilters,
		cli:             o.CLI,
		corgierrPretty:  o.CorgierrPretty,
		debugEnabled:    o.Debug,
	}
}

// start scope starts a new scope.
//
// If shallow is set to true, the scope is considered shallow and as such, if
// the scope is ended, its startClosed and haveBufClasses values will be copied
// into the current state.
// Effectively, shallow should be set to true, if the scope started does not
// reflect a new scope (i.e. '{}') in Go and does not start a new element.
func (ctx *ctx) startScope(shallow bool) *nesting {
	ctx.mixinFuncNames.startScope(ctx)

	old := ctx.scope()
	newScope := &nesting{
		elemName:       old.elemName,
		customVoidElem: old.customVoidElem,
		exprEscaper:    old.exprEscaper,
		txtEscaper:     old.txtEscaper,
		startClosed:    old.startClosed,
		haveBufClasses: old.haveBufClasses,
		shallow:        shallow,
	}
	ctx.scopes.Push(newScope)
	return newScope
}

func (ctx *ctx) scope() *nesting {
	return ctx.scopes.Peek()
}

func (ctx *ctx) endScope() *nesting {
	ctx.mixinFuncNames.endScope(ctx)
	old := ctx.scopes.Pop()
	if old.shallow {
		ctx.scope().startClosed = old.startClosed
		ctx.scope().haveBufClasses = old.haveBufClasses
	}
	return old
}

func (ctx *ctx) youShouldntSeeThisError(err error) {
	if ctx.cli {
		panic(fmt.Errorf("%w (you shouldn't see this error, please open an issue)", err))
	}

	panic(err)
}

func (ctx *ctx) ident(s string) string {
	return ctx.identPrefix + s
}

func (ctx *ctx) contextFunc(name string, args ...string) string {
	var sb strings.Builder

	sb.WriteString(ctx.identPrefix)
	sb.WriteString("ctx")
	sb.WriteByte('.')
	sb.WriteString(name)
	sb.WriteByte('(')

	for i, arg := range args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg)
	}

	sb.WriteByte(')')
	return sb.String()
}

func (ctx *ctx) woofQual(ident string) string {
	return ctx.identPrefix + "woof" + "." + ident
}

func (ctx *ctx) woofFunc(name string, args ...string) string {
	var sb strings.Builder

	sb.WriteString(ctx.identPrefix)
	sb.WriteString("woof")
	sb.WriteByte('.')
	sb.WriteString(name)
	sb.WriteByte('(')

	for i, arg := range args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg)
	}

	sb.WriteByte(')')
	return sb.String()
}

func (ctx *ctx) ioQual(ident string) string {
	return ctx.identPrefix + "io" + "." + ident
}

func (ctx *ctx) fmtFunc(name string, args ...string) string {
	var sb strings.Builder

	sb.WriteString(ctx.identPrefix)
	sb.WriteString("fmt")
	sb.WriteByte('.')
	sb.WriteString(name)
	sb.WriteByte('(')

	for i, arg := range args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg)
	}

	sb.WriteByte(')')
	return sb.String()
}

func (ctx *ctx) nextGotoIdent() string {
	ctx.gotoCounter++
	return fmt.Sprint(ctx.identPrefix, "goto", ctx.gotoCounter-1)
}

func (ctx *ctx) nextMixinIdent() string {
	ctx.mixinCounter++
	return fmt.Sprint(ctx.identPrefix, "mixin", ctx.mixinCounter-1)
}

func (ctx *ctx) stack() []*file.File {
	return ctx._stack[ctx.stackStart:]
}

func (ctx *ctx) baseFile() *file.File {
	return ctx._stack[0]
}

func (ctx *ctx) currentFile() *file.File {
	return ctx._stack[ctx.stackStart]
}

func (ctx *ctx) mainFile() *file.File {
	return ctx._stack[len(ctx._stack)-1]
}

func (ctx *ctx) startElem(name string, void bool) {
	ctx.closeStartTag()

	ctx.generate("<"+name, nil)
	ctx.scopes.Push(&nesting{
		elemName:       name,
		customVoidElem: void,
		startClosed:    unclosed,
	})
	ctx.calledUnclosed = false

	switch name {
	case "script":
		ctx.scope().exprEscaper = scriptBodyExprEscaper
		ctx.scope().txtEscaper = scriptBodyTextEscaper

		if ctx.hasNonce {
			ctx.flushGenerate()
			ctx.writeln(ctx.contextFunc("InjectNonce"))
		}
	case "style":
		ctx.scope().exprEscaper = styleBodyExprEscaper
		ctx.scope().txtEscaper = styleBodyTextEscaper
	default:
		ctx.scope().exprEscaper = plainBodyExprEscaper
		ctx.scope().txtEscaper = plainBodyTextEscaper
	}
}

func (ctx *ctx) closeElem() {
	ctx.closeStartTag()
	nest := ctx.scopes.Pop()

	if nest.customVoidElem || voidelem.Is(nest.elemName) {
		return
	}

	ctx.generate("</"+nest.elemName+">", nil)
}

func (ctx *ctx) debug(typ, s string) {
	if !ctx.debugEnabled {
		return
	}

	ctx.writeln(fmt.Sprintf("// [%s] %s", typ, s))
}

func (ctx *ctx) debugInline(typ, s string) {
	if !ctx.debugEnabled {
		return
	}

	ctx.write(fmt.Sprintf(" /* [%s] %s */ ", typ, s))
}

func (ctx *ctx) debugItem(itm file.Poser, s string) {
	if !ctx.debugEnabled {
		return
	}

	ctx.debug("item", fmt.Sprintf("%T (%d:%d): %s", itm, itm.Pos().Line, itm.Pos().Col, s))
}

func (ctx *ctx) debugItemInline(itm file.Poser, s string) {
	if !ctx.debugEnabled {
		return
	}

	ctx.debugInline("item", fmt.Sprintf("%T [%d:%d]: %s", itm, itm.Pos().Line, itm.Pos().Col, s))
}

func (ctx *ctx) write(s string) {
	_, err := io.WriteString(ctx.out, s)
	if err != nil {
		panic(err)
	}
}

func (ctx *ctx) writeBytes(p []byte) {
	_, err := ctx.out.Write(p)
	if err != nil {
		panic(err)
	}
}

func (ctx *ctx) writeln(s string) {
	ctx.write(s + "\n")
}

func (ctx *ctx) writeString(s string) {
	ctx.write(strconv.Quote(s))
}

func (ctx *ctx) inContext(f func()) {
	ctx.writeln("{")
	f()
	ctx.writeln("}")
}

func (ctx *ctx) generate(s string, esc *textEscaper) {
	if esc != nil {
		if ctx.debugEnabled {
			old := s
			s = esc.f(s)
			ctx.debug("escape/"+esc.name, strconv.Quote(old)+" -> "+strconv.Quote(s))
		} else {
			s = esc.f(s)
		}
	} else {
		ctx.debug("escape/none", strconv.Quote(s)+" -> [equal]")
	}
	ctx.debug("write buffer", strconv.Quote(s))
	ctx.generateBuf.WriteString(s)
}

func (ctx *ctx) flushGenerate() {
	if ctx.generateBuf.Len() > 0 {
		ctx.debug("flush generate buffer", "(see below)")
		ctx.writeln(ctx.contextFunc("Write", strconv.Quote(ctx.generateBuf.String())))
		ctx.generateBuf.Reset()
	} else {
		ctx.debug("flush generate buffer", "[empty buffer]")
	}
}

func (ctx *ctx) flushClasses() {
	if ctx.classBuf.Len() > 0 {
		ctx.debug("flush class buffer", "(see below)")
		ctx.writeln(ctx.contextFunc("BufferClassAttr", strconv.Quote(ctx.classBuf.String())))
		ctx.scope().haveBufClasses = true
		ctx.classBuf.Reset()

		ctx.callUnclosedIfUnclosed()
	} else {
		ctx.debug("flush class buffer", "[empty buffer]")
	}
}

var unquotedAttrValEscaper = strings.NewReplacer(`&`, `&amp;`)

func (ctx *ctx) generateStringAttr(name, val string) {
	ctx.generate(" "+name+"=", nil)

	// see if we even need to quote val
	// https://html.spec.whatwg.org/#syntax-attribute-value
	if val == "" || strings.ContainsAny(val, "\t\n\f\r \"'=<>`") {
		ctx.generate(`"`, nil)
		ctx.generate(val, &attrTextEscaper)
		ctx.generate(`"`, nil)
		return
	}

	ctx.generate(unquotedAttrValEscaper.Replace(val), nil)
}

func (ctx *ctx) generateExpr(expr string, esc *expressionEscaper) {
	ctx.flushGenerate()

	escName := ctx.woofQual(esc.funcName)
	ctx.writeln(ctx.woofFunc("WriteAny", ctx.ident(ctxVar), escName, expr))
}

func (ctx *ctx) bufClass(class string) {
	ctx.debug("class buffer", strconv.Quote(class))

	if ctx.classBuf.Len() == 0 {
		ctx.classBuf.WriteString(class)
	} else {
		ctx.classBuf.Grow(len(" ") + len(class))
		ctx.classBuf.WriteByte(' ')
		ctx.classBuf.WriteString(class)
	}
}

func (ctx *ctx) closeStartTag() {
	nest := ctx.scope()
	if nest.startClosed == closed {
		return
	}

	if nest.startClosed == unclosed {
		ctx.debug("close start tag", "[prev state: unclosed, class buf: "+ctx.classBuf.String()+"]")
	} else {
		ctx.debug("close start tag", "[prev state: maybe closed, class buf: "+ctx.classBuf.String()+"]")
	}

	defer func() { nest.startClosed = closed }()
	defer ctx.classBuf.Reset()

	if nest.startClosed == unclosed && !nest.haveBufClasses {
		classes := ctx.classBuf.String()
		if classes != "" {
			ctx.generateStringAttr("class", classes)
		}

		if nest.customVoidElem {
			ctx.generate("/>", nil)
			return
		}

		ctx.generate(">", nil)
		return
	}

	ctx.flushGenerate()

	ctx.write(ctx.ident(ctxVar) + ".CloseStartTag(")
	ctx.writeString(ctx.classBuf.String())
	ctx.write(", ")
	if nest.customVoidElem {
		ctx.writeln("true)")
	} else {
		ctx.writeln("false)")
	}
}

// horrible name, but a ton less confusion than just reading callUnclosed.
func (ctx *ctx) callUnclosedIfUnclosed() {
	ctx.debug("call unclosed if unclosed",
		"close state: "+ctx.scope().startClosed.String()+", called before: "+fmt.Sprint(ctx.calledUnclosed))

	if ctx.calledUnclosed {
		return
	}
	ctx.calledUnclosed = true
	if ctx.scope().startClosed == unclosed {
		ctx.writeln(ctx.contextFunc("Unclosed"))
	}
}

// horrible name, but a ton less confusion than just reading callUnclosed.
func (ctx *ctx) callClosedIfClosed() {
	ctx.debug("call closed if closed", "close state: "+ctx.scope().startClosed.String())

	if ctx.scope().startClosed == closed {
		ctx.writeln(ctx.contextFunc("Closed"))
	}
}

func (ctx *ctx) stringify(val any) string {
	s, err := woof.Stringify(val)
	if err != nil {
		panic(err)
	}

	return s
}
