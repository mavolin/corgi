package writer

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/internal/voidelem"
	"github.com/mavolin/corgi/woof"
)

// This file contains code that produces the body of the generated function.

func (w *Writer) writeFile() error {
	if err := w.writeToFile("{\n"); err != nil {
		return err
	}

	err := w.writeToFile("var (\n" +
		"    _buf _bytes.Buffer\n" +
		"    _closed bool\n" +
		"    // in case we never use them\n" +
		"    _ = _buf\n" +
		"    _ = _closed\n" +
		")\n\n")
	if err != nil {
		return err
	}

	if err = w.writeInit(); err != nil {
		return err
	}

	if err = w.writeDoctype(); err != nil {
		return err
	}

	if err = w.writeScope(w.files.Peek()[0].Scope, nil, nil); err != nil {
		return err
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(
		"return nil\n" +
			"}\n")
}

// ============================================================================
// Init
// ======================================================================================

func (w *Writer) writeInit() error {
	return w.writeInitFile(*w.main, make(map[string]struct{}))
}

func (w *Writer) writeInitFile(f file.File, alreadyProcessed map[string]struct{}) error {
	_, ok := alreadyProcessed[f.Source+"/"+f.Name]
	if ok {
		return nil
	}

	if err := w.writeInitScope(f.Scope); err != nil {
		return err
	}

	alreadyProcessed[f.Source+"/"+f.Name] = struct{}{}

	if f.Extend != nil {
		if err := w.writeInitFile(f.Extend.File, alreadyProcessed); err != nil {
			return err
		}
	}

	for _, use := range f.Uses {
		for _, uf := range use.Files {
			if err := w.writeInitFile(uf, alreadyProcessed); err != nil {
				return err
			}
		}
	}

	return w.writeInitIncludes(f.Scope, alreadyProcessed)
}

func (w *Writer) writeInitIncludes(s file.Scope, alreadyProcessed map[string]struct{}) error {
	return file.WalkError(s, func(itmPtr *file.ScopeItem) (bool, error) {
		incl, ok := (*itmPtr).(file.Include)
		if !ok {
			return true, nil
		}

		corgiIncl, ok := incl.Include.(file.CorgiInclude)
		if !ok {
			return true, nil
		}

		if err := w.writeInitFile(corgiIncl.File, alreadyProcessed); err != nil {
			return false, err
		}

		return true, nil
	})
}

func (w *Writer) writeInitScope(s file.Scope) error {
	for _, itm := range s {
		m, ok := itm.(file.Mixin)
		if !ok {
			continue
		}

		if m.Name != "init" {
			continue
		}

		return w.writeScope(m.Body, nil, nil)
	}

	return nil
}

// ============================================================================
// Doctype
// ======================================================================================

func (w *Writer) writeDoctype() error {
	base := w.files.Peek()[0]

	if base.Prolog != "" {
		w.writeRawUnescaped("<?xml " + base.Prolog + "?>")
	}

	if base.Doctype != "" {
		w.writeRawUnescaped("<!doctype " + base.Doctype + ">")
	}

	return nil
}

// ============================================================================
// Content
// ======================================================================================

func (w *Writer) writeScope(s file.Scope, e *elem, extraAttributes func(*elem) error) error {
	for _, itm := range s {
		if err := w.writeScopeItem(itm, e, extraAttributes); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) writeScopeItem(itm file.ScopeItem, e *elem, extraAttributes func(*elem) error) error {
	switch itm := itm.(type) {
	case file.Code:
		return w.writeCode(itm)
	case file.Include:
		return w.writeInclude(itm, e)
	case file.And:
		return w.writeAnd(itm, e)
	case file.If:
		return w.writeIf(itm, e)
	case file.Switch:
		return w.writeSwitch(itm, e)
	case file.For:
		return w.writeFor(itm, e)
	case file.While:
		return w.writeWhile(itm, e)
	case file.MixinCall:
		return w.writeMixinCall(itm, e)
	case file.Block:
		return w.writeBlock(itm, e)
	case file.IfBlock:
		return w.writeIfBlock(itm, e)
	}

	if e != nil && !e.isClosed {
		if err := w.closeTag(e); err != nil {
			return err
		}
	}

	switch itm := itm.(type) {
	case file.Comment:
		w.writeComment(itm)
		return nil
	case file.Text:
		w.writeText(itm, e)
		return nil
	case file.ExpressionInterpolation:
		return w.writeInterpolation(itm, e)
	case file.ElementInterpolation:
		return w.writeInlineElement(itm)
	case file.TextInterpolation:
		w.writeInlineText(itm, e)
		return nil
	case file.Filter:
		return w.writeFilter(itm)
	case file.Element:
		return w.writeElement(itm, extraAttributes)
	case file.Mixin:
		return nil
	default:
		panic(fmt.Sprintf("unknown scope item %T", itm))
	}
}

// ============================================================================
// Comment
// ======================================================================================

// this is obviously not a great solution, but still way better than writing
// broken HTML/XML because someone used '--'/'-->' in their comment.
var (
	htmlCommentEscaper = strings.NewReplacer("-->", "-- >")
	xmlCommentEscaper  = strings.NewReplacer("--", "- -") // also covers "-->"
)

func (w *Writer) writeComment(c file.Comment) {
	var comment string

	if w.main.Type == file.TypeHTML {
		comment = htmlCommentEscaper.Replace(c.Comment)
	} else {
		comment = xmlCommentEscaper.Replace(c.Comment)
	}

	w.writeRawUnescaped("<!--" + comment + "-->")
}

// ============================================================================
// Block
// ======================================================================================

func (w *Writer) writeBlock(b file.Block, e *elem) error {
	if w.mixins.Len() > 0 {
		c := w.mixins.Peek()
		for _, itm := range c.Body {
			filledBlock, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if filledBlock.Name == b.Name {
				tmp := w.mixins.Pop()

				if err := w.writeScope(filledBlock.Body, e, nil); err != nil {
					w.mixins.Push(tmp)
					return err
				}

				w.mixins.Push(tmp)
				return nil
			}
		}

		// use the default
		if len(b.Body) > 0 {
			return w.writeScope(b.Body, e, nil)
		}

		// we have no default
		return nil
	}

	bs := []block{{scope: b.Body, files: w.files.Peek()}}

	otherFiles := w.files.Peek()[1:]

	alreadyProcessed := make(map[string]struct{})

	for i, f := range otherFiles {
		for _, use := range f.Uses {
			for _, uf := range use.Files {
				if _, ok := alreadyProcessed[uf.Source+"/"+uf.Name]; ok {
					continue
				}

				alreadyProcessed[uf.Source+"/"+uf.Name] = struct{}{}
				bs = w.resolveBlock(uf.Scope, b.Name, bs, otherFiles[i:])
			}
		}

		bs = w.resolveBlock(f.Scope, b.Name, bs, otherFiles[i:])
	}

	for _, b := range bs {
		w.files.Push(b.files)

		if err := w.writeScope(b.scope, e, nil); err != nil {
			return err
		}

		w.files.Pop()
	}

	return nil
}

type block struct {
	scope []file.ScopeItem

	// files are the files starting at the file providing this block to the
	// main file
	files []file.File
}

func (w *Writer) resolveBlock(s file.Scope, name file.Ident, bs []block, otherFiles []file.File) []block {
	for _, itm := range s {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name != name {
			continue
		}

		wrapper := block{scope: filledBlock.Body, files: otherFiles}

		switch filledBlock.Type {
		case file.BlockTypeBlock:
			bs = []block{wrapper}
		case file.BlockTypeAppend:
			bs = append(bs, wrapper)
		case file.BlockTypePrepend:
			bs = append([]block{wrapper}, bs...)
		}
	}

	return bs
}

// ============================================================================
// Element
// ======================================================================================

func (w *Writer) writeElement(e file.Element, extraAttrs func(*elem) error) error {
	var inContext bool

	for _, itm := range e.Body {
		_, ok := itm.(file.Code)
		if ok {
			inContext = true
			break
		}
	}

	if inContext {
		if err := w.writeToFile("{\n"); err != nil {
			return err
		}
	}

	w.setClosed = "false"

	ew := elem{e: e}

	w.writeRawUnescaped("<" + e.Name)

	for _, attr := range e.Attributes {
		if err := w.writeAttribute(attr); err != nil {
			return err
		}
	}

	if extraAttrs != nil {
		if err := extraAttrs(&ew); err != nil {
			return err
		}
	}

	if err := w.writeScope(e.Body, &ew, nil); err != nil {
		return err
	}

	if err := w.closeElement(&ew); err != nil {
		return err
	}

	if inContext {
		if err := w.writeToFile("}\n"); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) writeAttribute(attr file.Attribute) error {
	switch attr := attr.(type) {
	case file.LiteralAttribute:
		w.writeRawUnescaped(fmt.Sprintf(` %s="%s"`, attr.Name, woof.EscapeHTML(attr.Value)))
	case file.ExpresssionAttribute:
		iexp := w.inlineExpression(attr.Value)
		if iexp != "" {
			if iexp == "true" {
				if w.main.Type == file.TypeHTML {
					w.writeRawUnescaped(" " + attr.Name)
					return nil
				}

				w.writeRawUnescaped(fmt.Sprintf(` %s="%s"`, attr.Name, attr.Name))
				return nil
			} else if iexp == "false" {
				return nil
			}

			// check if this is a string literal
			unq, err := strconv.Unquote(iexp)
			if err == nil {
				if !attr.NoEscape {
					unq = string(woof.EscapeHTML(unq))
				}

				w.writeRawUnescaped(fmt.Sprintf(` %s="%s"`, attr.Name, unq))
				return nil
			}
		}

		return w.expression(attr.Value, func(val string) error {
			if attr.NoEscape {
				return w.writeAttrUnescapedExpression(attr.Name, val, w.main.Type != file.TypeHTML)
			}

			return w.writeAttrExpression(attr.Name, val, w.main.Type != file.TypeHTML)
		}, nil)
	}

	return nil
}

func (w *Writer) closeTag(e *elem) error {
	if e.isClosed {
		return nil
	}

	e.isClosed = true

	// Check if the first item fills the elements body or the element has an
	// empty body.
	// If so, we can save ourselves the if not closed check.
	//
	// It is safe to check this even if closeTag is called after the
	// first element has been written, as in that case isClosed will always be
	// true.
	noIf := w.firstAlwaysWritesBody(e.e.Body) || len(e.e.Body) == 0

	if !noIf {
		if err := w.flushRawBuf(); err != nil {
			return err
		}

		if err := w.writeToFile("if !_closed {\n"); err != nil {
			return err
		}

		if err := w.writeToFile("_closed = true\n"); err != nil {
			return err
		}
	} else {
		w.setClosed = "true"
	}

	if err := w.writeClasses(e); err != nil {
		return err
	}

	if e.e.SelfClosing {
		w.writeRawUnescaped("/>")
	} else {
		w.writeRawUnescaped(">")
	}

	if !noIf {
		if err := w.flushRawBuf(); err != nil {
			return err
		}

		return w.writeToFile("}\n")
	}

	return nil
}

func (w *Writer) closeElement(e *elem) error {
	if w.main.Type == file.TypeHTML {
		if err := w.closeTag(e); err != nil {
			return err
		}

		if !e.e.SelfClosing && !voidelem.Is(e.e.Name) {
			w.writeRawUnescaped("</" + e.e.Name + ">")
		}

		return nil
	}

	// we've written content already
	if e.isClosed {
		w.writeRawUnescaped("</" + e.e.Name + ">")
		return nil
	}

	// elem must be self-closing if it has no body
	if len(e.e.Body) == 0 {
		w.writeRawUnescaped("/>")
		return nil
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	if err := w.writeToFile("if _closed {\n"); err != nil {
		return err
	}

	w.writeRawUnescaped("</" + e.e.Name + ">")
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	if err := w.writeToFile("} else {\n"); err != nil {
		return err
	}

	if err := w.writeClasses(e); err != nil {
		return err
	}

	w.writeRawUnescaped("/>")
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile("}\n")
}

func (w *Writer) writeClasses(e *elem) error {
	if e.needBuf || len(e.e.Classes) != 0 {
		w.writeRawUnescaped(` class="`)

		for i, c := range e.e.Classes {
			switch c := c.(type) {
			case file.ClassLiteral:
				if i > 0 {
					w.writePreEscapedHTML(" " + c.Name)
				} else {
					w.writePreEscapedHTML(c.Name)
				}
			case file.ClassExpression:
				err := w.expression(c.Names, func(val string) error {
					unq, unqErr := strconv.Unquote(val)

					if c.NoEscape {
						if i > 0 {
							if unqErr == nil {
								w.writeRawUnescaped(" " + unq)
								return nil
							}

							return w.writeUnescapedStringExpression(`" "+` + val)
						}

						if unqErr == nil {
							w.writeRawUnescaped(unq)
							return nil
						}

						return w.writeUnescapedStringExpression(val)
					}

					if i > 0 {
						if unqErr == nil {
							w.writePreEscapedHTML(" " + unq)
							return nil
						}

						return w.writeEscapedHTMLStringExpression(`" "+` + val)
					}

					if unqErr == nil {
						w.writePreEscapedHTML(unq)
						return nil
					}

					return w.writeEscapedHTMLStringExpression(val)
				}, nil)
				if err != nil {
					return err
				}
			}
		}

		if e.needBuf {
			if err := w.flushRawBuf(); err != nil {
				return err
			}

			if err := w.writeToFile("if _buf.Len() > 0 {\n"); err != nil {
				return err
			}

			if len(e.e.Classes) > 0 {
				w.writeRawUnescaped(" ")
				if err := w.flushRawBuf(); err != nil {
					return err
				}
			}

			if err := w.writeBuf(); err != nil {
				return err
			}

			if err := w.writeToFile("_buf.Reset()\n"); err != nil {
				return err
			}

			if err := w.writeToFile("}\n"); err != nil {
				return err
			}
		}

		w.writeRawUnescaped(`"`)
	}

	return nil
}

// ============================================================================
// &
// ======================================================================================

func (w *Writer) writeAnd(and file.And, e *elem) error {
	for _, attr := range and.Attributes {
		if err := w.writeAttribute(attr); err != nil {
			return err
		}
	}

	if len(and.Classes) == 0 {
		return nil
	}

	e.needBuf = true

	for _, c := range and.Classes {
		switch c := c.(type) {
		case file.ClassLiteral:
			if err := w.writeToBufPreEscaped(c.Name); err != nil {
				return err
			}
		case file.ClassExpression:
			err := w.expression(c.Names, func(val string) error {
				if c.NoEscape {
					return w.writeToBufExpressionUnescaped(val)
				}

				// save ourselves escaping at runtime
				unq, err := strconv.Unquote(val)
				if err == nil {
					return w.writeToBufPreEscaped(unq)
				}

				return w.writeToBufExpression(val)
			}, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ============================================================================
// Include
// ======================================================================================

func (w *Writer) writeInclude(incl file.Include, e *elem) error {
	switch incl := incl.Include.(type) {
	case file.CorgiInclude:
		return w.writeScope(incl.File.Scope, e, nil)
	case file.OtherInclude:
		if e != nil {
			if err := w.closeTag(e); err != nil {
				return err
			}
		}

		w.writeRawUnescaped(incl.Contents)
	}

	return nil
}

// ============================================================================
// Code
// ======================================================================================

func (w *Writer) writeCode(c file.Code) error {
	// the code could conditionalize execution, start a loop or the like, which
	// could potentially mess with the generated HTML
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile(c.Code + "\n")
}

// ============================================================================
// If
// ======================================================================================

func (w *Writer) writeIf(if_ file.If, e *elem) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	if err := w.writeToFile("if "); err != nil {
		return err
	}

	closed := true

	if err := w.ifExpression(if_.Condition); err != nil {
		return err
	}

	if err := w.writeToFile(" {\n"); err != nil {
		return err
	}

	eCp := e.clone()

	if err := w.writeScope(if_.Then, eCp, nil); err != nil {
		return err
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	if e != nil {
		e.needBuf = eCp.needBuf
		if !eCp.isClosed {
			closed = false
		}
	}

	if err := w.writeToFile("}"); err != nil {
		return err
	}

	for _, ei := range if_.ElseIfs {
		if err := w.writeToFile(" else if "); err != nil {
			return err
		}

		if err := w.ifExpression(ei.Condition); err != nil {
			return err
		}

		if err := w.writeToFile(" {\n"); err != nil {
			return err
		}

		eCp := e.clone()

		if err := w.writeScope(ei.Then, eCp, nil); err != nil {
			return err
		}

		if err := w.flushRawBuf(); err != nil {
			return err
		}

		if e != nil {
			e.needBuf = eCp.needBuf
			if !eCp.isClosed {
				closed = false
			}
		}

		if err := w.writeToFile("}"); err != nil {
			return err
		}
	}

	if if_.Else != nil {
		if err := w.writeToFile(" else {\n"); err != nil {
			return err
		}

		eCp := e.clone()

		if err := w.writeScope(if_.Else.Then, eCp, nil); err != nil {
			return err
		}

		if err := w.flushRawBuf(); err != nil {
			return err
		}

		if e != nil {
			e.needBuf = eCp.needBuf
			if !eCp.isClosed {
				closed = false
			}
		}

		if err := w.writeToFile("}"); err != nil {
			return err
		}
	}

	// only if all branches closed the tag and if we have an else can we be
	// sure that the tag is always closed
	if e != nil && closed && if_.Else != nil {
		e.isClosed = true
	}

	return w.writeToFile("\n")
}

// ============================================================================
// IfBlock
// ======================================================================================

func (w *Writer) writeIfBlock(ib file.IfBlock, e *elem) error {
	if w.mixins.Len() > 0 {
		c := w.mixins.Peek()
		for _, itm := range c.Body {
			filledBlock, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if filledBlock.Name == ib.Name {
				return w.writeScope(ib.Then, e, nil)
			}
		}

		if ib.Else != nil {
			return w.writeScope(ib.Else.Then, e, nil)
		}

		return nil
	}

	otherFiles := w.files.Peek()[1:]

	for _, f := range otherFiles {
		for _, use := range f.Uses {
			for _, uf := range use.Files {
				if w.hasBlock(uf.Scope, ib.Name) {
					return w.writeScope(ib.Then, e, nil)
				}
			}
		}

		if w.hasBlock(f.Scope, ib.Name) {
			return w.writeScope(ib.Then, e, nil)
		}
	}

	if ib.Else != nil {
		return w.writeScope(ib.Else.Then, e, nil)
	}

	return nil
}

func (w *Writer) hasBlock(s file.Scope, name file.Ident) bool {
	for _, itm := range s {
		b, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if b.Name == name {
			return true
		}
	}

	return false
}

// ============================================================================
// Switch
// ======================================================================================

func (w *Writer) writeSwitch(sw file.Switch, e *elem) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	if sw.Comparator == nil {
		if err := w.writeToFile("switch {\n"); err != nil {
			return err
		}

		if err := w.writeSwitchCases(sw, e); err != nil {
			return err
		}

		return w.writeToFile("}\n")
	}

	return w.expression(sw.Comparator, func(val string) error {
		if err := w.writeToFile("switch " + val + " {\n"); err != nil {
			return err
		}

		if err := w.writeSwitchCases(sw, e); err != nil {
			return err
		}

		return w.writeToFile("}\n")
	}, nil)
}

func (w *Writer) writeSwitchCases(sw file.Switch, e *elem) error {
	closed := true

	for _, c := range sw.Cases {
		if err := w.writeToFile("case " + c.Expression.Expression + ":\n"); err != nil {
			return err
		}

		eCp := e.clone()

		if err := w.writeScope(c.Then, eCp, nil); err != nil {
			return err
		}

		if err := w.flushRawBuf(); err != nil {
			return err
		}

		if e != nil {
			e.needBuf = eCp.needBuf
			if !eCp.isClosed {
				closed = false
			}
		}
	}

	if sw.Default != nil {
		if err := w.writeToFile("default:\n"); err != nil {
			return err
		}

		eCp := e.clone()

		if err := w.writeScope(sw.Default.Then, eCp, nil); err != nil {
			return err
		}

		if err := w.flushRawBuf(); err != nil {
			return err
		}

		if e != nil {
			e.needBuf = eCp.needBuf
			if !eCp.isClosed {
				closed = false
			}
		}
	}

	// only if all branches closed the tag and if we have a default case can we
	// be sure that the tag is always closed
	if e != nil && closed && sw.Default != nil {
		e.isClosed = true
	}

	return nil
}

// ============================================================================
// For
// ======================================================================================

func (w *Writer) writeFor(f file.For, e *elem) error {
	if e != nil && !file.IsFirstAnd(f.Body) {
		if err := w.closeTag(e); err != nil {
			return err
		}
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.expression(f.Range, func(val string) error {
		if err := w.writeToFile("for "); err != nil {
			return err
		}

		if f.VarOne != "" {
			if err := w.writeToFile(string(f.VarOne)); err != nil {
				return err
			}

			if f.VarTwo != "" {
				if err := w.writeToFile(", " + string(f.VarTwo)); err != nil {
					return err
				}
			}

			if err := w.writeToFile(" := "); err != nil {
				return err
			}
		}

		if err := w.writeToFile("range " + val + " {\n"); err != nil {
			return err
		}

		if err := w.writeScope(f.Body, e, nil); err != nil {
			return err
		}

		if err := w.flushRawBuf(); err != nil {
			return err
		}

		return w.writeToFile("}\n")
	}, nil)
}

// ============================================================================
// While
// ======================================================================================

func (w *Writer) writeWhile(wh file.While, e *elem) error {
	if e != nil && !file.IsFirstAnd(wh.Body) {
		if err := w.closeTag(e); err != nil {
			return err
		}
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	err := w.writeToFile("for " + wh.Condition.Expression + " {\n")
	if err != nil {
		return err
	}

	if err = w.writeScope(wh.Body, e, nil); err != nil {
		return err
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile("}\n")
}

// ============================================================================
// Text
// ======================================================================================

func (w *Writer) writeText(t file.Text, e *elem) {
	switch {
	case w.main.Type == file.TypeXML:
		w.writePreEscapedHTML(t.Text)
	case e.e.Name == "style" || e.e.Name == "script":
		w.writeRawUnescaped(t.Text)
	default:
		w.writePreEscapedHTML(t.Text)
	}
}

// ============================================================================
// ExpressionInterpolation
// ======================================================================================

func (w *Writer) writeInterpolation(i file.ExpressionInterpolation, e *elem) error {
	return w.expression(i.Expression, func(val string) error {
		if i.NoEscape {
			return w.writeUnescapedExpression(val)
		}

		if w.main.Type == file.TypeXML {
			return w.writeHTMLExpression(val)
		}

		switch e.e.Name {
		case "style":
			return w.writeCSSExpression(val)
		case "script":
			return w.writeJSExpression(val)
		default:
			return w.writeHTMLExpression(val)
		}
	}, nil)
}

// ============================================================================
// ElementInterpolation
// ======================================================================================

func (w *Writer) writeInlineElement(ie file.ElementInterpolation) error {
	w.writeRawUnescaped("<" + ie.Name)

	for _, attr := range ie.Attributes {
		if err := w.writeAttribute(attr); err != nil {
			return err
		}
	}

	if len(ie.Classes) > 0 {
		w.writeRawUnescaped(` class="`)

		for i, c := range ie.Classes {
			switch c := c.(type) {
			case file.ClassLiteral:
				if i > 0 {
					w.writeRawUnescaped(" " + c.Name)
				} else {
					w.writePreEscapedHTML(c.Name)
				}
			case file.ClassExpression:
				err := w.expression(c.Names, func(val string) error {
					unq, unqErr := strconv.Unquote(val)

					if c.NoEscape {
						if i > 0 {
							if unqErr == nil {
								w.writeRawUnescaped(" " + unq)
								return nil
							}

							return w.writeUnescapedStringExpression(`" "+` + val)
						}

						if unqErr == nil {
							w.writeRawUnescaped(unq)
							return nil
						}

						return w.writeUnescapedStringExpression(val)
					}

					if i > 0 {
						if unqErr == nil {
							w.writePreEscapedHTML(" " + unq)
							return nil
						}

						return w.writeEscapedHTMLStringExpression(`" "+` + val)
					}

					if unqErr == nil {
						w.writePreEscapedHTML(unq)
						return nil
					}

					return w.writeEscapedHTMLStringExpression(val)
				}, nil)
				if err != nil {
					return err
				}
			}
		}

		w.writeRawUnescaped(`"`)
	}

	if ie.SelfClosing {
		w.writeRawUnescaped("/>")
		return nil
	}

	w.writeRawUnescaped(">")

	if voidelem.Is(ie.Name) || ie.Body == nil {
		return nil
	}

	switch val := ie.Body.(type) {
	case file.Text:
		switch {
		case ie.NoEscape:
			w.writeRawUnescaped(val.Text)
		case w.main.Type == file.TypeXML:
			w.writePreEscapedHTML(val.Text)
		default:
			switch ie.Name {
			case "style", "script":
				w.writeRawUnescaped(val.Text)
			default:
				w.writePreEscapedHTML(val.Text)
			}
		}
	case file.Expression:
		err := w.expression(val, func(val string) error {
			if ie.NoEscape {
				return w.writeUnescapedExpression(val)
			}

			if w.main.Type == file.TypeXML {
				return w.writeHTMLExpression(val)
			}

			switch ie.Name {
			case "style":
				return w.writeCSSExpression(val)
			case "script":
				return w.writeJSExpression(val)
			default:
				return w.writeHTMLExpression(val)
			}
		}, nil)
		if err != nil {
			return err
		}
	}

	w.writeRawUnescaped("</" + ie.Name + ">")
	return nil
}

// ============================================================================
// TextInterpolation
// ======================================================================================

func (w *Writer) writeInlineText(it file.TextInterpolation, e *elem) {
	switch {
	case it.NoEscape:
		w.writeRawUnescaped(it.Text)
	case w.main.Type == file.TypeXML:
		w.writePreEscapedHTML(it.Text)
	case e.e.Name == "style" || e.e.Name == "script":
		w.writeRawUnescaped(it.Text)
	default:
		w.writePreEscapedHTML(it.Text)
	}
}

// ============================================================================
// Mixin
// ======================================================================================

func (w *Writer) writeMixinCall(c file.MixinCall, e *elem) error {
	w.mixins.Push(c)
	defer w.mixins.Pop()

	if err := w.writeToFile("{\n"); err != nil {
		return err
	}

Params:
	for _, param := range c.Mixin.Params {
		// Go doesn't allow 'foo := nil', hence manually set the type.
		// This seems sensible, as 'nil' will probably be a common default.
		if param.Type == "" && param.Default.Expression == "nil" {
			param.Type = "any"
		}

		for _, arg := range c.Args {
			if arg.Name != param.Name {
				continue
			}

			shadows, err := doesExpressionShadow(arg.Value, string(param.Name))
			if err != nil {
				return err
			}

			paramName := string(param.Name)
			if shadows {
				// double underscore to not clash with '_buf' and '_closed'
				paramName = "__" + paramName
			}

			if param.Type != "" {
				err = w.writeToFile("var " + paramName + " " + string(param.Type) + "\n")
				if err != nil {
					return err
				}
			} else {
				err = w.writeToFile(paramName + ":=" + param.Default.Expression + "\n")
				if err != nil {
					return err
				}
			}

			err = w.expression(arg.Value, func(val string) error {
				return w.writeToFile(paramName + "=" + val + "\n")
			}, func() error {
				// if there is no type, the default will have already been written
				if param.Type != "" {
					return w.writeToFile(string(param.Name) + "=" + param.Default.Expression + "\n")
				}

				return nil
			})
			if err != nil {
				return err
			}

			if shadows {
				err = w.writeToFile(string(param.Name) + " = " + paramName + "\n")
				if err != nil {
					return err
				}
			}

			continue Params
		}

		// use the default

		err := w.writeToFile(fmt.Sprintf("var %s %s = %s\n", param.Name, param.Type, param.Default.Expression))
		if err != nil {
			return err
		}
	}

	err := w.writeScope(c.Mixin.Body, e, w.writeMixinAnds(c.Body))
	if err != nil {
		return err
	}

	return w.writeToFile("}\n")
}

func (w *Writer) writeMixinAnds(s file.Scope) func(e *elem) error {
	return func(e *elem) error {
		for _, itm := range s {
			_, skip := itm.(file.Block)
			if skip {
				continue
			}

			if err := w.writeScopeItem(itm, e, nil); err != nil {
				return err
			}
		}

		return nil
	}
}

// doesExpressionShadow reports whether assigning expr to varName would be the
// same as `foo := foo`, i.e. the varName is the same as the value it is
// assigned.
func doesExpressionShadow(e file.Expression, varName string) (bool, error) {
	switch e := e.(type) {
	case file.GoExpression:
		return e.Expression == varName, nil
	case file.TernaryExpression:
		shadows, err := doesExpressionShadow(e.IfFalse, varName)
		if err != nil {
			return false, err
		}

		if shadows {
			return true, nil
		}

		return doesExpressionShadow(e.IfTrue, varName)
	case file.SetCheckExpression:
		if e.Default != nil {
			shadows, err := doesExpressionShadow(*e.Default, varName)
			if err != nil {
				return false, err
			}

			if shadows {
				return true, nil
			}
		}

		if len(e.Chain) == 0 && e.Deref == "" {
			return e.Root.Expression == varName, nil
		}

		return false, nil
	default:
		return false, fmt.Errorf("unsupported expression type %T", e)
	}
}

// ============================================================================
// Filter
// ======================================================================================

func (w *Writer) writeFilter(f file.Filter) error {
	cmd := exec.Command(f.Name, f.Args...) //nolint:gosec
	cmd.Stdin = strings.NewReader(f.Body.Text)

	var stdout, stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		cf := w.files.Peek()[0]

		return errors.Wrapf(err, "%s/%s:%d:%d failed to run filter %s: %s: %s",
			cf.Source, cf.Name, f.Line, f.Col, f.Name, err.Error(), stderr.String())
	}

	w.writeRawUnescaped(stdout.String())
	return nil
}

// ============================================================================
// Expression
// ======================================================================================

func (w *Writer) inlineExpression(exp file.Expression) string {
	goExp, ok := exp.(file.GoExpression)
	if !ok {
		return ""
	}

	return goExp.Expression
}

// expression writes the expression to the file and calls ifVal in the block
// where the resolved value is available.
// val will be set to an expression that yields the resolved value.
//
// If noVal is set, it will be called in all else blocks.
func (w *Writer) expression(e file.Expression, ifVal func(val string) error, noVal func() error) error {
	switch e := e.(type) {
	case file.GoExpression:
		return ifVal(e.Expression)
	case file.TernaryExpression:
		if err := w.flushRawBuf(); err != nil {
			return err
		}

		return w.ifElse(e.Condition, func() error {
			if err := w.expression(e.IfTrue, ifVal, noVal); err != nil {
				return err
			}

			return w.flushRawBuf()
		}, func() error {
			if err := w.expression(e.IfFalse, ifVal, noVal); err != nil {
				return err
			}

			return w.flushRawBuf()
		})
	case file.SetCheckExpression:
		var ifNil func() error

		if e.Default != nil {
			ifNil = func() error {
				return ifVal(e.Default.Expression)
			}
		} else if noVal != nil {
			ifNil = noVal
		}

		return w.nilCheckExpr(e, ifVal, ifNil)
	default:
		return fmt.Errorf("unsupported expression type %T", e)
	}
}

func (w *Writer) ifElse(cond file.GoExpression, ifTrue, ifFalse func() error) error {
	if err := w.writeToFile("if " + cond.Expression + " {\n"); err != nil {
		return err
	}

	if err := ifTrue(); err != nil {
		return err
	}

	if ifFalse == nil {
		if err := w.writeToFile("}"); err != nil {
			return err
		}

		return nil
	}

	if err := w.writeToFile("} else {\n"); err != nil {
		return err
	}

	if err := ifFalse(); err != nil {
		return err
	}

	return w.writeToFile("}\n")
}

// nilCheckExpr writes a nil check expression that processes the resolved
// value of the given expression.
func (w *Writer) nilCheckExpr(
	e file.SetCheckExpression, notNil func(val string) error, isNil func() error,
) error {
	if err := w.flushRawBuf(); err != nil {
		return err
	}

	err := w.writeToFile("if ")
	if err != nil {
		return err
	}

	if err := w.nilCheckIfCondition(e); err != nil {
		return err
	}

	if err := w.writeToFile(" {\n"); err != nil {
		return err
	}

	if err := notNil(nilCheckToGoExpression(e)); err != nil {
		return err
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	if isNil == nil {
		return w.writeToFile("}\n")
	}

	if err := w.writeToFile("} else {\n"); err != nil {
		return err
	}

	if err := isNil(); err != nil {
		return err
	}

	if err := w.flushRawBuf(); err != nil {
		return err
	}

	return w.writeToFile("}\n")
}

func nilCheckToGoExpression(e file.SetCheckExpression) string {
	var b strings.Builder

	b.WriteString(e.Deref)
	b.WriteString(e.Root.Expression)

	for _, chainExpr := range e.Chain {
		switch chainExpr := chainExpr.(type) {
		case file.IndexExpression:
			b.WriteString("[")
			b.WriteString(chainExpr.Expression)
			b.WriteString("]")
		case file.FieldMethodExpression:
			b.WriteString(".")
			b.WriteString(chainExpr.Expression)
		case file.FuncCallExpression:
			b.WriteString("(")
			b.WriteString(chainExpr.Expression)
			b.WriteString(")")
		}
	}

	return b.String()
}

// ifExpression writes a condition for an if statement.
func (w *Writer) ifExpression(e file.Expression) error {
	switch e := e.(type) {
	case file.GoExpression:
		return w.writeToFile(e.Expression)
	case file.SetCheckExpression:
		return w.nilCheckIfCondition(e)
	default:
		return fmt.Errorf("unsupported expression type %T", e)
	}
}

func (w *Writer) nilCheckIfCondition(e file.SetCheckExpression) error {
	if err := w.writeToFile("woof.IsSet("); err != nil {
		return err
	}

	err := w.writeToFile(e.Root.Expression)
	if err != nil {
		return err
	}

	for _, expr := range e.Chain {
		if err = w.writeToFile(", "); err != nil {
			return err
		}

		switch expr := expr.(type) {
		case file.IndexExpression:
			err = w.writeToFile("woof.IndexChainItm{Index: " + expr.Expression + "}")
			if err != nil {
				return err
			}
		case file.FieldMethodExpression:
			err = w.writeToFile("woof.FieldMethodChainItem{Names: \"" + expr.Expression + "\"}")
			if err != nil {
				return err
			}
		case file.FuncCallExpression:
			err = w.writeToFile("woof.FuncCallChainItem{Args: []any{" + expr.Expression + "}}")
			if err != nil {
				return err
			}
		}
	}

	return w.writeToFile(")")
}
