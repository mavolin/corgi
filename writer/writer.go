// Package writer provides a writer that allows converting a file.File to Go
// code.
package writer

import (
	"bytes"
	"io"
	"strconv"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/internal/stack"
	"github.com/mavolin/corgi/writeutil"
)

type Writer struct {
	// mixin calls we're in right now
	mixins stack.Stack[file.MixinCall]
	// files available in the current block.
	// Starts with the current file, up till the main file.
	files stack.Stack[[]file.File]

	main        *file.File // convenience
	packageName string
	out         io.Writer

	// rawBuf holds static content to be written in a single write operation.
	// It is populated whenever successive calls to writeRawUnescaped and
	// writePreEscapedHTML are made.
	// Any call to another write method except writeToFile will flush rawBuf.
	//
	// Code must ensure that rawBuf is flushed before writing the body of a
	// conditional or loop.
	rawBuf bytes.Buffer

	// setClosed is a utility to prevent unnecessary adjacent _closed
	// assignments when writing to rawBuf.
	//
	// Instead of functions setting _closed directly, they should set setClosed
	// to either "true" or "false".
	// In doing so, _closed assignments are only written when writeToFile is
	// called, as only then will the _close state be relevant.
	setClosed string
}

type elem struct {
	e        file.Element
	isClosed bool
	needBuf  bool
}

func (e *elem) clone() *elem {
	if e == nil {
		return nil
	}

	eCp := *e
	return &eCp
}

func New(f *file.File, packageName string) *Writer {
	files := []file.File{*f}

	next := f.Extend

	for next != nil {
		files = append([]file.File{next.File}, files...)
		next = next.File.Extend
	}

	w := &Writer{
		main:        f,
		packageName: packageName,
		mixins:      stack.New[file.MixinCall](50),
	}

	w.files = stack.New[[]file.File](50)
	w.files.Push(files)

	return w
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
	if w.setClosed != "" {
		_, err := io.WriteString(w.out, "_closed = "+w.setClosed+"\n")
		if err != nil {
			return err
		}

		w.setClosed = ""
	}

	_, err := io.WriteString(w.out, s)
	return err
}

// writeRawUnescaped writes raw HTML to the output.
// It does not escape s.
func (w *Writer) writeRawUnescaped(s string) {
	w.rawBuf.WriteString(s)
}

func (w *Writer) flushRawBuf() error {
	if w.rawBuf.Len() == 0 {
		return nil
	}

	defer w.rawBuf.Reset()

	return w.writeToFile("err = _writeutil.Write(_w, " + strconv.Quote(w.rawBuf.String()) + ")\n" +
		"if err != nil {\n" +
		"    return err\n" +
		"}\n")
}

// ================================ Pre Escaped Strings =================================

// writePreEscapedHTML escaped the raw HTML s and writes it to the output.
func (w *Writer) writePreEscapedHTML(s string) {
	w.rawBuf.WriteString(string(writeutil.EscapeHTML(s)))
}

// ==================================== Expressions =====================================

func (w *Writer) writeUnescapedStringExpression(s string) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.Write(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeCSSExpression(s string) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteCSS(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeHTMLExpression(s string) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteHTML(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeEscapedHTMLStringExpression(s string) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.Write(_w, string(_writeutil.EscapeHTML(" + s + ")))\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeJSExpression(s string) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteJS(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeUnescapedExpression(s string) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteAnyUnescaped(_w, " + s + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

// ======================================= Buffer =======================================

func (w *Writer) writeToBufExpression(s string) error {
	return w.writeToFile(
		"if _buf.Len() == 0 {\n" +
			"	_buf.WriteString(string(_writeutil.EscapeHTML(" + s + ")))\n" +
			"} else {\n" +
			"    _buf.WriteString(\" \" + string(_writeutil.EscapeHTML(" + s + ")))\n" +
			"}\n")
}

func (w *Writer) writeToBufPreEscaped(s string) error {
	s = string(writeutil.EscapeHTML(s))

	return w.writeToFile(
		"if _buf.Len() == 0 {\n" +
			"    _buf.WriteString(" + strconv.Quote(s) + ")\n" +
			"} else {\n" +
			"    _buf.WriteString(" + strconv.Quote(" "+s) + ")\n" +
			"}\n")
}

func (w *Writer) writeToBufExpressionUnescaped(s string) error {
	return w.writeToFile(
		"if _buf.Len() == 0 {\n" +
			"	_buf.WriteString(" + s + ")\n" +
			"} else {" +
			"    _buf.WriteString(\" \" + " + s + ")\n" +
			"}\n")
}

func (w *Writer) writeBuf() error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteBytes(_w, _buf.Bytes())\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

// ===================================== Attributes =====================================

func (w *Writer) writeAttrExpression(name string, val string, mirror bool) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteAttr(_w, " + strconv.Quote(name) + ", " + val + ", " +
			strconv.FormatBool(mirror) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}

func (w *Writer) writeAttrUnescapedExpression(name string, val string, mirror bool) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"err = _writeutil.WriteAttrUnescaped(_w, " + strconv.Quote(name) + ", " + val + ", " +
			strconv.FormatBool(mirror) + ")\n" +
			"if err != nil {\n" +
			"    return err\n" +
			"}\n")
}
