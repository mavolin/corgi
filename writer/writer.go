// Package writer provides a writer that allows converting a file.File to Go
// code.
package writer

import (
	"io"
	"strconv"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/pkg/stack"
	"github.com/mavolin/corgi/pkg/writeutil"
)

type Writer struct {

	// mixin calls were in right now
	mixins stack.Stack[file.MixinCall]
	// files available in the current block.
	// Starts with the current file, up till the main file.
	files stack.Stack[[]file.File]

	main        *file.File // convenience
	packageName string
	out         io.Writer
}

type elem struct {
	e        file.Element
	isClosed bool
	needBuf  bool
}

func New(f *file.File, packageName string) *Writer {
	files := []file.File{*f}

	next := f.Extend

	for next != nil {
		files = append([]file.File{next.File}, files...)
		next = next.File.Extend
	}

	fileStack := stack.New[[]file.File](50)
	fileStack.Push(files)

	return &Writer{
		files:       fileStack,
		main:        f,
		packageName: packageName,
		mixins:      stack.New[file.MixinCall](50),
	}
}

func (w *Writer) Write(out io.Writer) error {
	w.out = out

	if err := w.writePackage(); err != nil {
		return err
	}

	if err := w.writeDoNotEdit(); err != nil {
		return err
	}

	if err := w.writeImports(); err != nil {
		return err
	}

	if err := w.writeGlobalCode(); err != nil {
		return err
	}

	if err := w.writeFunc(); err != nil {
		return err
	}

	return w.writeFile()
}

// ============================================================================
// Helpers
// ======================================================================================

func (w *Writer) writeToFile(s string) error {
	_, err := io.WriteString(w.out, s)
	return err
}

func (w *Writer) writeUnescaped(s string) error {
	return w.writeUnescapedStringExpression(strconv.Quote(s))
}

func (w *Writer) writeUnescapedStringExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.Write(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeEscapedHTMLStringExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.Write(_w, string(_writeutil.EscapeHTML(" + s + ")))\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeEscapedCSSExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.WriteEscapedCSS(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeEscapedJSStrExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.WriteEscapedJSStr(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writePreEscapedHTML(s string) error {
	s = string(writeutil.EscapeHTML(s))

	return w.writeToFile(
		"err = _writeutil.Write(_w, " + strconv.Quote(s) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writePreEscapedCSS(s string) error {
	s = string(writeutil.EscapeCSS(s))

	return w.writeToFile(
		"err = _writeutil.Write(_w, " + strconv.Quote(s) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writePreEscapedJSStr(s string) error {
	s = string(writeutil.EscapeJSStr(s))

	return w.writeToFile(
		"err = _writeutil.Write(_w, " + strconv.Quote(s) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeCSSExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.WriteCSS(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeHTMLExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.WriteHTML(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeJSExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.WriteJS(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeUnescapedExpression(s string) error {
	return w.writeToFile(
		"err = _writeutil.WriteAnyUnescaped(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeToBuf(s string) error {
	return w.writeToFile(
		"if _buf.Len() == 0 {\n" +
			"	_buf.WriteString(string(_writeutil.EscapeHTML(" + s + ")))\n" +
			"} else {\n" +
			"    _buf.WriteString(\" \" + string(_writeutil.EscapeHTML(" + s + ")))\n" +
			"}\n")
}

func (w *Writer) writeToBufUnescaped(s string) error {
	return w.writeToFile(
		"if _buf.Len() == 0 {\n" +
			"	_buf.WriteString(" + s + ")\n" +
			"} else {" +
			"    _buf.WriteString(\" \" + " + s + ")\n" +
			"}\n")
}

func (w *Writer) writeBuf() error {
	return w.writeToFile(
		"err = _writeutil.WriteBytes(_w, _buf.Bytes())\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeAttr(name string, val string, mirror bool) error {
	return w.writeToFile(
		"err = _writeutil.WriteAttr(_w, " + strconv.Quote(name) + ", " + val + ", " +
			strconv.FormatBool(mirror) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeAttrUnescaped(name string, val string, mirror bool) error {
	return w.writeToFile(
		"err = _writeutil.WriteAttrUnescaped(_w, " + strconv.Quote(name) + ", " + val + ", " +
			strconv.FormatBool(mirror) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeUnsafeAttr(name string, val string) error {
	return w.writeToFile(
		"err = _writeutil.WriteUnsafeAttr(_w, " + strconv.Quote(name) + ", " + val + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeCSSAttr(name string, val string) error {
	return w.writeToFile(
		"err = _writeutil.WriteCSSAttr(_w, " + strconv.Quote(name) + ", " + val + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeHTMLAttr(name string, val string) error {
	return w.writeToFile(
		"err = _writeutil.WriteHTMLAttr(_w, " + strconv.Quote(name) + ", " + val + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeJSAttr(name string, val string) error {
	return w.writeToFile(
		"err = _writeutil.WriteJSAttr(_w, " + strconv.Quote(name) + ", " + val + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeURLAttr(name string, val string) error {
	return w.writeToFile(
		"err = _writeutil.WriteURLAttr(_w, " + strconv.Quote(name) + ", " + val + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeSrcsetAttr(name string, val string) error {
	return w.writeToFile(
		"err = _writeutil.WriteSrcsetAttr(_w, " + strconv.Quote(name) + ", " + val + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}
