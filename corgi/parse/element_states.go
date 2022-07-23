package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/lex"
	"github.com/mavolin/corgi/internal/require"
)

func (p *Parser) beforeDoctype() (_ stateFn, err error) {
	if p.mode == ModeMain && p.f.Func.Name == "" {
		return nil, p.unexpectedItem(p.next(), lex.Func)
	}

	for {
		peek := p.peek()
		switch peek.Type {
		case lex.Doctype:
			return p.doctype()
		case lex.Code:
			c, err := p.code()
			if err != nil {
				return nil, err
			}

			p.f.Scope = append(p.f.Scope, *c)
		case lex.Mixin:
			m, err := p.mixin()
			if err != nil {
				return nil, err
			}

			p.f.Scope = append(p.f.Scope, *m)
		case lex.Block:
			if p.mode == ModeExtend {
				b, err := p.block(require.Always, require.Always)
				if err != nil {
					return nil, err
				}

				p.f.Scope = append(p.f.Scope, *b)
				break
			}

			fallthrough
		default:
			return p.nextHTML, nil
		}
	}
}

func (p *Parser) nextHTML() (_ stateFn, err error) {
	if p.mode == ModeMain && p.f.Func.Name == "" {
		return nil, p.unexpectedItem(p.next(), lex.Func)
	}

	if p.mode == ModeInclude {
		s, err := p.scope()
		if err != nil {
			return nil, err
		}

		p.f.Scope = append(p.f.Scope, s...)
	} else {
		s, err := p.globalScope()
		if err != nil {
			return nil, err
		}

		p.f.Scope = append(p.f.Scope, s...)
	}

	return nil, nil
}

func (p *Parser) globalScope() (file.Scope, error) {
	var s file.Scope

Loop:
	for {
		peek := p.peek()
		switch peek.Type {
		case lex.EOF:
			return s, nil
		case lex.Error:
			next := p.next()
			return nil, p.error(next, next.Err)
		case lex.Comment:
			c, err := p.comment()
			if err != nil {
				return nil, err
			}

			s = append(s, *c)
			continue Loop
		case lex.Mixin:
			m, err := p.mixin()
			if err != nil {
				return nil, err
			}

			s = append(s, *m)
			continue Loop
		}

		if p.mode != ModeUse && p.f.Extend == nil {
			switch peek.Type {
			case lex.CodeStart:
				c, err := p.code()
				if err != nil {
					return nil, err
				}

				if c != nil { // empty line of code
					s = append(s, *c)
				}
				continue Loop
			case lex.Include:
				incl, err := p.include()
				if err != nil {
					return nil, err
				}

				s = append(s, *incl)
				continue Loop
			case lex.If:
				if_, err := p.if_()
				if err != nil {
					return nil, err
				}

				s = append(s, *if_)
			case lex.Switch:
				sw, err := p.switch_()
				if err != nil {
					return nil, err
				}

				s = append(s, *sw)
			case lex.For:
				f, err := p.for_()
				if err != nil {
					return nil, err
				}

				s = append(s, *f)
			case lex.While:
				w, err := p.while()
				if err != nil {
					return nil, err
				}

				s = append(s, *w)
			case lex.MixinCall:
				c, err := p.mixinCall()
				if err != nil {
					return nil, err
				}

				s = append(s, *c)
			case lex.Assign, lex.AssignNoEscape:
				a, err := p.assign()
				if err != nil {
					return nil, err
				}

				s = append(s, *a)
			case lex.Pipe:
				pipe, err := p.pipe()
				if err != nil {
					return nil, err
				}

				s = append(s, pipe...)
			case lex.DotBlock:
				d, err := p.dotBlock()
				if err != nil {
					return nil, err
				}

				s = append(s, d...)
			case lex.Filter:
				f, err := p.filter()
				if err != nil {
					return nil, err
				}

				s = append(s, *f)
			case lex.Div, lex.Element:
				e, err := p.element()
				if err != nil {
					return nil, err
				}

				s = append(s, *e)
			case lex.Block:
				b, err := p.block(require.Always, require.Always)
				if err != nil {
					return nil, err
				}

				s = append(s, *b)
			default:
				return nil, p.unexpectedItem(p.next())
			}
		} else {
			switch peek.Type {
			case lex.Block, lex.BlockAppend, lex.BlockPrepend:
				b, err := p.block(require.Always, require.Always)
				if err != nil {
					return nil, err
				}

				s = append(s, *b)
				continue Loop
			default:
				return nil, p.unexpectedItem(p.next())
			}
		}
	}
}

func (p *Parser) scope() (file.Scope, error) {
	var s file.Scope

Loop:
	for {
		peek := p.peek()
		switch peek.Type {
		case lex.Dedent:
			p.next()
			return s, nil
		case lex.EOF:
			p.next()
			return s, nil
		case lex.Error:
			next := p.next()
			return nil, p.error(next, next.Err)
		}

		switch p.context.Peek() {
		case ContextMixinCall:
			itm, err := p.scopeItemMixinCall()
			if err != nil {
				return nil, err
			}

			s = append(s, itm)
		case ContextMixinCallConditional:
			itm, err := p.scopeItemMixinCallConditional()
			if err != nil {
				return nil, err
			}

			s = append(s, itm)
		case ContextMixinDefinition:
			switch peek.Type {
			case lex.Block:
				b, err := p.block(require.Optional, require.Optional)
				if err != nil {
					return nil, err
				}

				s = append(s, *b)
				continue Loop
			case lex.IfBlock:
				b, err := p.ifBlock(false)
				if err != nil {
					return nil, err
				}

				s = append(s, *b)
				continue Loop
			}

			fallthrough
		case ContextRegular:
			itms, err := p.scopeItemRegular()
			if err != nil {
				return nil, err
			}

			if len(itms) > 0 {
				s = append(s, itms...)
			}
		default:
			panic(fmt.Sprintf("unknown Context %d", p.context.Peek()))
		}
	}
}

func (p *Parser) scopeItemMixinCall() (file.ScopeItem, error) {
	switch p.peek().Type {
	case lex.If:
		p.context.Push(ContextMixinCallConditional)

		if_, err := p.if_()
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *if_, nil
	case lex.Switch:
		p.context.Push(ContextMixinCallConditional)

		switch_, err := p.switch_()
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *switch_, nil
	case lex.And:
		and, err := p.and()
		if err != nil {
			return nil, err
		}

		return *and, nil
	case lex.MixinCall:
		c, err := p.mixinCall()
		if err != nil {
			return nil, err
		}

		return *c, nil
	case lex.Block:
		p.context.Push(ContextRegular)

		block, err := p.block(require.Optional, require.Optional)
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *block, nil
	default:
		return nil, p.unexpectedItem(p.next())
	}
}

func (p *Parser) scopeItemMixinCallConditional() (file.ScopeItem, error) {
	switch p.peek().Type {
	case lex.If:
		if_, err := p.if_()
		if err != nil {
			return nil, err
		}

		return *if_, nil
	case lex.Switch:
		switch_, err := p.switch_()
		if err != nil {
			return nil, err
		}

		return *switch_, nil
	case lex.And:
		and, err := p.and()
		if err != nil {
			return nil, err
		}

		return *and, nil
	case lex.MixinCall:
		c, err := p.mixinCall()
		if err != nil {
			return nil, err
		}

		return *c, nil
	default:
		return nil, p.unexpectedItem(p.next())
	}
}

func (p *Parser) scopeItemRegular() ([]file.ScopeItem, error) {
	switch p.peek().Type {
	case lex.Comment:
		c, err := p.comment()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*c}, nil
	case lex.CodeStart:
		c, err := p.code()
		if err != nil {
			return nil, err
		}

		if c != nil { // empty line of code
			return []file.ScopeItem{*c}, nil
		}

		return nil, nil
	case lex.Include:
		incl, err := p.include()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*incl}, nil
	case lex.Mixin:
		m, err := p.mixin()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*m}, nil
	case lex.If:
		if_, err := p.if_()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*if_}, nil
	case lex.Switch:
		sw, err := p.switch_()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*sw}, nil
	case lex.For:
		f, err := p.for_()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*f}, nil
	case lex.While:
		w, err := p.while()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*w}, nil
	case lex.MixinCall:
		c, err := p.mixinCall()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*c}, nil
	case lex.Assign, lex.AssignNoEscape:
		a, err := p.assign()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*a}, nil
	case lex.Pipe:
		pipe, err := p.pipe()
		if err != nil {
			return nil, err
		}

		return pipe, nil
	case lex.And:
		and, err := p.and()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*and}, nil
	case lex.DotBlock:
		d, err := p.dotBlock()
		if err != nil {
			return nil, err
		}

		return d, nil
	case lex.Filter:
		f, err := p.filter()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*f}, nil
	case lex.Div, lex.Element:
		e, err := p.element()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*e}, nil
	}

	if p.mode == ModeExtend {
		switch p.peek().Type {
		case lex.Block:
			b, err := p.block(require.Always, require.Optional)
			if err != nil {
				return nil, err
			}

			return []file.ScopeItem{*b}, nil
		case lex.IfBlock:
			b, err := p.ifBlock(true)
			if err != nil {
				return nil, err
			}

			return []file.ScopeItem{*b}, nil
		}
	}

	return nil, p.unexpectedItem(p.next())
}

// ============================================================================
// Doctype
// ======================================================================================

// xhtmlRegexp is the regexp used to check if custom doctypes use XHTML.
//
// This is very much best effort, if this fails, users should manually set
// file.File.Type to whatever is correct.
var xhtmlRegexp = regexp.MustCompile(`^html PUBLIC ["'][^ ]+ XHTML `)

//goland:noinspection HttpUrlsUsage
func (p *Parser) doctype() (stateFn, error) {
	switch {
	case p.mode == ModeUse:
		return nil, p.error(p.next(), ErrUseDoctype)
	case p.f.Extend != nil:
		return nil, p.error(p.next(), ErrExtendDoctype)
	}

	p.next() // lex.Doctype

	doctypeItm := p.next()
	if doctypeItm.Type != lex.Literal {
		return nil, p.unexpectedItem(doctypeItm, lex.Literal)
	}

	doctype := doctypeItm.Val
	doctype = trimRightWhitespace(doctype)

	// special case: this is not a doctype, but a xml prolog
	if doctype == "xml" { // default prolog
		if p.f.Prolog != "" {
			return nil, p.error(doctypeItm, ErrMultipleProlog)
		}

		p.f.Type = file.TypeXML
		p.f.Prolog = `version="1.0" encoding="utf-8"`
		return p.afterDoctype, nil
	} else if strings.HasPrefix(doctype, "xml ") { // custom prolog
		if p.f.Prolog != "" {
			return nil, p.error(doctypeItm, ErrMultipleProlog)
		}

		p.f.Type = file.TypeXML
		p.f.Prolog = strings.TrimPrefix(doctype, "xml ")
		return p.afterDoctype, nil
	}

	if p.f.Doctype != "" {
		return nil, p.error(doctypeItm, ErrMultipleDoctype)
	}

	switch doctype {
	case "html":
		p.f.Type = file.TypeHTML
		p.f.Doctype = "html"
	case "transitional":
		p.f.Type = file.TypeXHTML
		p.f.Doctype = `html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" ` +
			`"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"`
	case "strict":
		p.f.Type = file.TypeXHTML
		p.f.Doctype = `html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" ` +
			`"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"`
	case "frameset":
		p.f.Type = file.TypeXHTML
		p.f.Doctype = `html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN" ` +
			`"http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd"`
	case "1.1":
		p.f.Type = file.TypeXHTML
		p.f.Doctype = `html PUBLIC "-//W3C//DTD XHTML 1.1//EN" ` +
			`"http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"`
	case "basic":
		p.f.Type = file.TypeXHTML
		p.f.Doctype = `html PUBLIC "-//W3C//DTD XHTML Basic 1.1//EN" ` +
			`"http://www.w3.org/TR/xhtml-basic/xhtml-basic11.dtd"`
	case "mobile":
		p.f.Type = file.TypeXHTML
		p.f.Doctype = `html PUBLIC "-//WAPFORUM//DTD XHTML Mobile 1.2//EN" ` +
			`"http://www.openmobilealliance.org/tech/DTD/xhtml-mobile12.dtd"`
	case "plist":
		p.f.Type = file.TypeXML
		p.f.Doctype = `plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" ` +
			`"http://www.apple.com/DTDs/PropertyList-1.0.dtd"`
	default:
		switch {
		case xhtmlRegexp.MatchString(doctype):
			p.f.Type = file.TypeXHTML
		case strings.HasPrefix(doctype, "html "):
			p.f.Type = file.TypeHTML
		default:
			p.f.Type = file.TypeXML
		}

		p.f.Doctype = doctype
	}

	return p.afterDoctype, nil
}

func (p *Parser) afterDoctype() (stateFn, error) {
	if p.peek().Type == lex.Doctype {
		return p.doctype, nil
	}

	return p.nextHTML, nil
}

// ============================================================================
// Comment
// ======================================================================================

func (p *Parser) comment() (*file.Comment, error) {
	p.next() // lex.Comment

	next := p.next()
	switch next.Type {
	case lex.Text:
		return &file.Comment{Comment: next.Val}, nil
	case lex.Indent:
		// handled below
	default:
		return nil, p.unexpectedItem(next, lex.Indent, lex.Text)
	}

	var b strings.Builder
	b.Grow(1000)

	first := true

	for {
		next = p.next()
		switch next.Type {
		case lex.EOF, lex.Dedent:
			return &file.Comment{Comment: b.String()}, nil
		case lex.Text:
			if !first {
				b.WriteString("\n")
			} else {
				first = false
			}

			b.WriteString(next.Val)
		default:
			return nil, p.unexpectedItem(next, lex.Text, lex.Dedent)
		}
	}
}

// ============================================================================
// Include
// ======================================================================================

func (p *Parser) include() (_ *file.Include, err error) {
	includeItm := p.next() // lex.Include

	pathItm := p.next()
	if pathItm.Type != lex.Literal {
		return nil, p.unexpectedItem(pathItm, lex.Literal)
	}

	incl := file.Include{
		Pos: file.Pos{Line: includeItm.Line, Col: includeItm.Col},
	}

	incl.Path, err = strconv.Unquote(pathItm.Val)
	if err != nil {
		return nil, p.error(pathItm, err)
	}

	return &incl, nil
}

// ============================================================================
// Code
// ======================================================================================

func (p *Parser) code() (*file.Code, error) {
	p.next() // lex.CodeStart

	peek := p.peek()
	switch peek.Type {
	case lex.Code:
		return &file.Code{Code: p.next().Val}, nil
	case lex.Indent:
		// handled below
	default: // empty line of code
		// add nothing because empty lines only serve the purpose of aiding
		// readability, a trait that is not required for a generated file
		return nil, nil
	}

	var b strings.Builder
	b.Grow(1000)

	p.next() // lex.Indent

	first := true

	for {
		switch p.peek().Type {
		case lex.EOF, lex.Dedent:
			p.next()
			return &file.Code{Code: b.String()}, nil
		case lex.Code:
			if !first {
				b.WriteString("\n")
			} else {
				first = false
			}

			b.WriteString(p.next().Val)
		default:
			return nil, p.unexpectedItem(p.next(), lex.Code, lex.Dedent)
		}
	}
}

func (p *Parser) block(named, body require.Required) (_ *file.Block, err error) {
	blockItm := p.next()
	b := file.Block{Pos: file.Pos{Line: blockItm.Line, Col: blockItm.Col}}

	switch blockItm.Type {
	case lex.Block:
		b.Type = file.BlockTypeBlock
	case lex.BlockAppend:
		b.Type = file.BlockTypeAppend
	case lex.BlockPrepend:
		b.Type = file.BlockTypePrepend
	default:
		return nil, p.unexpectedItem(blockItm, lex.Block, lex.BlockAppend, lex.BlockPrepend)
	}

	if named > require.Never {
		if p.peek().Type == lex.Ident {
			b.Name = file.Ident(p.next().Val)
		} else if named == require.Always {
			return nil, p.unexpectedItem(p.next(), lex.Ident)
		}
	}

	if body == require.Never {
		return &b, nil
	}

	switch p.peek().Type {
	case lex.EOF:
		p.next()
		return &b, nil
	case lex.DotBlock:
		b.Body, err = p.dotBlock()
		if err != nil {
			return nil, err
		}

		return &b, nil
	case lex.Indent:
		// handled below
	default:
		if body == require.Optional {
			return &b, nil
		}

		return nil, p.unexpectedItem(p.next(), lex.Indent, lex.DotBlock)
	}

	p.next() // lex.Indent

	b.Body, err = p.scope()
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// ============================================================================
// Mixin
// ======================================================================================

func (p *Parser) mixin() (_ *file.Mixin, err error) {
	mixinItm := p.next() // lex.Mixin

	m := file.Mixin{Pos: file.Pos{Line: mixinItm.Line, Col: mixinItm.Col}}

	nameItm := p.next()
	if nameItm.Type != lex.Ident {
		return nil, p.unexpectedItem(nameItm, lex.Ident)
	}

	m.Name = file.Ident(nameItm.Val)

	m.Params, err = p.mixinParams()
	if err != nil {
		return nil, err
	}

	indentItm := p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	p.context.Push(ContextMixinDefinition)

	m.Body, err = p.scope()
	if err != nil {
		return nil, err
	}

	p.context.Pop()

	return &m, nil
}

func (p *Parser) mixinParams() (params []file.MixinParam, err error) {
	lparenItm := p.next()
	if lparenItm.Type != lex.LParen {
		return nil, p.unexpectedItem(lparenItm, lex.LParen)
	}

	if p.peek().Type == lex.RParen {
		p.next()
		return nil, nil
	}

	for {
		var param file.MixinParam

		paramName := p.next()
		if paramName.Type != lex.Ident {
			return nil, p.unexpectedItem(paramName, lex.Ident)
		}

		param.Name = file.Ident(paramName.Val)
		param.Pos = file.Pos{Line: paramName.Line, Col: paramName.Col}

		if p.peek().Type == lex.Ident { // type
			param.Type = file.GoIdent(p.next().Val)
		}

		if p.peek().Type == lex.Assign {
			assignItm := p.next()

			defaultExpr, err := p.expression()
			if err != nil {
				return nil, err
			}

			goExpr, ok := defaultExpr.(file.GoExpression)
			if !ok {
				return nil, p.error(assignItm, ErrMixinDefaultExpression)
			}

			param.Default = &goExpr
		}

		params = append(params, param)

		next := p.next()
		switch next.Type {
		case lex.EOF:
			return nil, p.unexpectedItem(next, lex.Comma, lex.RParen)
		case lex.RParen:
			return params, nil
		case lex.Comma:
			switch p.peek().Type {
			case lex.EOF:
				return nil, p.unexpectedItem(p.next(), lex.Ident, lex.RParen)
			case lex.RParen:
				p.next()
				return params, nil
			}
		default:
			return nil, p.unexpectedItem(next, lex.Comma, lex.RParen)
		}
	}
}

// ============================================================================
// IfBlock
// ======================================================================================

func (p *Parser) ifBlock(mustBeNamed bool) (_ *file.IfBlock, err error) {
	var i file.IfBlock

	p.next() // lex.IfBlock

	nameItm := p.peek()
	if nameItm.Type == lex.Ident {
		nameItm = p.next()
		i.Name = file.Ident(nameItm.Val)
	} else if mustBeNamed {
		return nil, p.unexpectedItem(p.next(), lex.Ident)
	}

	indentItm := p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	i.Then, err = p.scope()
	if err != nil {
		return nil, err
	}

	if p.peek().Type != lex.Else {
		return &i, nil
	}

	p.next() // lex.Else

	indentItm = p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	elseScope, err := p.scope()
	if err != nil {
		return nil, err
	}

	i.Else = &file.Else{Then: elseScope}
	return &i, nil
}

// ============================================================================
// If
// ======================================================================================

func (p *Parser) if_() (_ *file.If, err error) {
	p.next() // lex.If

	var i file.If

	i.Condition, err = p.expression()
	if err != nil {
		return nil, err
	}

	switch p.peek().Type {
	case lex.DotBlock:
		i.Then, err = p.dotBlock()
		if err != nil {
			return nil, err
		}
	case lex.BlockExpansion:
		content, err := p.blockExpansion()
		if err != nil {
			return nil, err
		}

		i.Then = file.Scope{content}
	case lex.Indent:
		p.next() // lex.Indent

		i.Then, err = p.scope()
		if err != nil {
			return nil, err
		}
	case lex.EOF:
		fallthrough
	default:
		return nil, p.unexpectedItem(p.next(), lex.Indent, lex.DotBlock, lex.BlockExpansion)
	}

	for {
		var ei file.ElseIf

		if p.peek().Type != lex.ElseIf {
			break
		}

		p.next() // lex.ElseIf

		ei.Condition, err = p.expression()
		if err != nil {
			return nil, err
		}

		switch p.peek().Type {
		case lex.DotBlock:
			ei.Then, err = p.dotBlock()
			if err != nil {
				return nil, err
			}
		case lex.BlockExpansion:
			content, err := p.blockExpansion()
			if err != nil {
				return nil, err
			}

			ei.Then = file.Scope{content}
		case lex.Indent:
			p.next() // lex.Indent

			ei.Then, err = p.scope()
			if err != nil {
				return nil, err
			}
		case lex.EOF:
			fallthrough
		default:
			return nil, p.unexpectedItem(p.next(), lex.Indent, lex.DotBlock, lex.BlockExpansion)
		}

		i.ElseIfs = append(i.ElseIfs, ei)
	}

	if p.peek().Type != lex.Else {
		return &i, nil
	}

	p.next() // lex.Else

	var e file.Else

	switch p.peek().Type {
	case lex.DotBlock:
		e.Then, err = p.dotBlock()
		if err != nil {
			return nil, err
		}
	case lex.BlockExpansion:
		content, err := p.blockExpansion()
		if err != nil {
			return nil, err
		}

		e.Then = file.Scope{content}
	case lex.Indent:
		p.next() // lex.Indent

		e.Then, err = p.scope()
		if err != nil {
			return nil, err
		}
	case lex.EOF:
		fallthrough
	default:
		return nil, p.unexpectedItem(p.next(), lex.Indent, lex.DotBlock, lex.BlockExpansion)
	}

	i.Else = &e
	return &i, nil
}

// ============================================================================
// Switch
// ======================================================================================

func (p *Parser) switch_() (_ *file.Switch, err error) {
	p.next() // lex.Switch

	var s file.Switch

	if p.peek().Type != lex.Indent {
		s.Comparator, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	indentItm := p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	for {
		var c file.Case

		caseItm := p.peek()
		if caseItm.Type != lex.Case {
			break
		}

		p.next() // lex.Case

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		goExpr, ok := expr.(file.GoExpression)
		if !ok {
			return nil, p.error(caseItm, ErrCaseExpression)
		}

		c.Expression = goExpr

		switch p.peek().Type {
		case lex.Indent:
			p.next()

			c.Then, err = p.scope()
			if err != nil {
				return nil, err
			}
		case lex.DotBlock:
			c.Then, err = p.dotBlock()
			if err != nil {
				return nil, err
			}
		case lex.BlockExpansion:
			content, err := p.blockExpansion()
			if err != nil {
				return nil, err
			}

			c.Then = file.Scope{content}
		default:
			// empty case
		}

		s.Cases = append(s.Cases, c)
	}

	next := p.next()
	switch next.Type {
	case lex.EOF:
		fallthrough
	case lex.Dedent:
		return &s, nil
	case lex.DefaultCase:
		// handled below
	default:
		return nil, p.unexpectedItem(next, lex.DefaultCase, lex.EOF, lex.Dedent)
	}

	var d file.DefaultCase

	switch p.peek().Type {
	case lex.Indent:
		p.next()

		d.Then, err = p.scope()
		if err != nil {
			return nil, err
		}
	case lex.DotBlock:
		d.Then, err = p.dotBlock()
		if err != nil {
			return nil, err
		}
	case lex.BlockExpansion:
		content, err := p.blockExpansion()
		if err != nil {
			return nil, err
		}

		d.Then = file.Scope{content}
	default:
		return nil, p.unexpectedItem(p.next(), lex.Indent, lex.DotBlock, lex.BlockExpansion)
	}

	s.Default = &d

	next = p.next()
	switch next.Type {
	case lex.EOF:
		return &s, nil
	case lex.Dedent:
		return &s, nil
	default:
		return nil, p.unexpectedItem(next, lex.Dedent)
	}
}

// ============================================================================
// For
// ======================================================================================

func (p *Parser) for_() (*file.For, error) {
	forItm := p.next()

	f := file.For{
		Pos: file.Pos{Line: forItm.Line, Col: forItm.Col},
	}

	if p.peek().Type == lex.Ident {
		ident1Itm := p.next()
		if ident1Itm.Type != lex.Ident {
			return nil, p.unexpectedItem(ident1Itm, lex.Ident)
		}

		f.VarOne = file.GoIdent(ident1Itm.Val)

		if p.peek().Type == lex.Comma {
			p.next()

			ident2Itm := p.next()
			if ident2Itm.Type != lex.Ident {
				return nil, p.unexpectedItem(ident2Itm, lex.Ident)
			}

			f.VarTwo = file.GoIdent(ident2Itm.Val)
		} else if p.peek().Type != lex.Range {
			return nil, p.unexpectedItem(p.next(), lex.Comma, lex.Range)
		}
	} else if p.peek().Type != lex.Range {
		return nil, p.unexpectedItem(p.next(), lex.Ident, lex.Range)
	}

	rangeItm := p.next()
	if rangeItm.Type != lex.Range {
		return nil, p.unexpectedItem(rangeItm, lex.Range)
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	f.Range = expr

	indentItm := p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	f.Body, err = p.scope()
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// ============================================================================
// While
// ======================================================================================

func (p *Parser) while() (_ *file.While, err error) {
	whileItm := p.next() // lex.While

	w := file.While{Pos: file.Pos{Line: whileItm.Line, Col: whileItm.Col}}

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	goExpr, ok := cond.(file.GoExpression)
	if !ok {
		return nil, p.error(whileItm, ErrWhileExpression)
	}

	w.Condition = goExpr

	indentItm := p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	w.Body, err = p.scope()
	if err != nil {
		return nil, err
	}

	return &w, nil
}

// ============================================================================
// Element
// ======================================================================================

func (p *Parser) element() (*file.Element, error) {
	e, err := p.elementHeader()
	if err != nil {
		return nil, err
	}

	switch p.peek().Type {
	case lex.Assign:
		p.next() // lex.Assign

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		e.Body = file.Scope{file.Interpolation{Expression: expr}}
		return e, nil
	case lex.AssignNoEscape:
		p.next() // lex.Assign

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		e.Body = file.Scope{
			file.Interpolation{
				Expression: expr,
				NoEscape:   true,
			},
		}
		return e, nil
	case lex.Indent:
		p.next()

		e.Body, err = p.scope()
		if err != nil {
			return nil, err
		}

		return e, nil
	}

	if !e.SelfClosing {
		switch p.peek().Type {
		case lex.BlockExpansion:
			if e.SelfClosing {
				return nil, p.unexpectedItem(p.next(), lex.Assign, lex.AssignNoEscape)
			}

			exp, err := p.blockExpansion()
			if err != nil {
				return nil, err
			}

			e.Body = file.Scope{exp}
			return e, nil
		case lex.DotBlock:
			e.Body, err = p.dotBlock()
			if err != nil {
				return nil, err
			}

			return e, nil
		case lex.Text, lex.Hash:
			e.Body, err = p.text(true)
			if err != nil {
				return nil, err
			}

			return e, nil
		}
	}

	return e, nil
}

func (p *Parser) elementHeader() (*file.Element, error) {
	name := p.next()
	e := file.Element{Pos: file.Pos{Line: name.Line, Col: name.Col}}

	switch name.Type {
	case lex.Div:
		e.Name = "div"
	case lex.Element:
		e.Name = name.Val
	default:
		return nil, p.unexpectedItem(name, lex.Element, lex.Div)
	}

	for {
		switch p.peek().Type {
		case lex.Class:
			class, err := p.class()
			if err != nil {
				return nil, err
			}

			e.Classes = append(e.Classes, *class)
		case lex.ID:
			id, err := p.id()
			if err != nil {
				return nil, err
			}

			e.Attributes = append(e.Attributes, *id)
		case lex.LParen:
			as, cs, err := p.attributes()
			if err != nil {
				return nil, err
			}

			e.Attributes = append(e.Attributes, as...)
			e.Classes = append(e.Classes, cs...)
		case lex.TagVoid:
			p.next()
			e.SelfClosing = true
			return &e, nil
		default:
			return &e, nil
		}
	}
}

func (p *Parser) class() (*file.ClassLiteral, error) {
	p.next() // lex.Class

	name := p.next()
	if name.Type != lex.Literal {
		return nil, p.unexpectedItem(name, lex.Literal)
	}

	return &file.ClassLiteral{Name: name.Val}, nil
}

func (p *Parser) id() (*file.AttributeLiteral, error) {
	p.next() // lex.Id

	name := p.next()
	if name.Type != lex.Literal {
		return nil, p.unexpectedItem(name, lex.Literal)
	}

	return &file.AttributeLiteral{Name: "id", Value: name.Val}, nil
}

func (p *Parser) attributes() ([]file.Attribute, []file.Class, error) {
	p.next() // lex.LParen

	// special case: empty attribute list
	if p.peek().Type == lex.RParen {
		p.next()
		return nil, nil, nil
	}

	var attributes []file.Attribute
	var classes []file.Class

	for {
		var (
			name      string
			unescaped bool
			value     file.Expression
		)

		nameItm := p.next()
		if nameItm.Type != lex.Ident {
			return nil, nil, p.unexpectedItem(nameItm, lex.Ident)
		}

		name = trimRightWhitespace(nameItm.Val)

		switch p.peek().Type {
		case lex.AssignNoEscape:
			unescaped = true
			fallthrough
		case lex.Assign:
			p.next()

			var err error
			value, err = p.expression()
			if err != nil {
				return nil, nil, err
			}
		default:
			value = file.GoExpression{Expression: "true"}
		}

		if name == "class" {
			classes = append(classes, file.ClassExpression{
				Name:     value,
				NoEscape: unescaped,
			})
		} else {
			attributes = append(attributes, file.AttributeExpression{
				Name:     name,
				Value:    value,
				NoEscape: unescaped,
			})
		}

		next := p.next()
		switch next.Type {
		case lex.RParen:
			return attributes, classes, nil
		case lex.Comma:
			switch p.peek().Type {
			case lex.EOF:
				return nil, nil, p.unexpectedItem(p.next(), lex.Ident, lex.RParen)
			case lex.RParen:
				p.next()
				return attributes, classes, nil
			}
		case lex.EOF:
			fallthrough
		default:
			return nil, nil, p.unexpectedItem(next, lex.Comma, lex.RParen)
		}
	}
}

// ============================================================================
// Block Expansion
// ======================================================================================

func (p *Parser) blockExpansion() (file.ScopeItem, error) {
	p.next() // lex.BlockExpansion

	peek := p.peek()
	if peek.Type == lex.EOF {
		return nil, p.unexpectedItem(p.next())
	}

	if peek.Type == lex.Block {
		if p.context.Peek() == ContextMixinDefinition {
			b, err := p.block(require.Optional, require.Never)
			if err != nil {
				return nil, err
			}

			return *b, nil
		} else if p.mode == ModeExtend {
			b, err := p.block(require.Always, require.Never)
			if err != nil {
				return nil, err
			}

			return *b, nil
		}
	}

	elem, err := p.elementHeader()
	if err != nil {
		return nil, err
	}

	return *elem, nil
}

// ============================================================================
// And (&)
// ======================================================================================

func (p *Parser) and() (*file.And, error) {
	andItm := p.next()

	a := file.And{Pos: file.Pos{Line: andItm.Line, Col: andItm.Col}}

	for {
		switch p.peek().Type {
		case lex.Class:
			class, err := p.class()
			if err != nil {
				return nil, err
			}

			a.Classes = append(a.Classes, *class)
		case lex.ID:
			id, err := p.id()
			if err != nil {
				return nil, err
			}

			a.Attributes = append(a.Attributes, *id)
		case lex.LParen:
			as, cs, err := p.attributes()
			if err != nil {
				return nil, err
			}

			a.Attributes = append(a.Attributes, as...)
			a.Classes = append(a.Classes, cs...)
		default:
			return &a, nil
		}
	}
}

// ============================================================================
// Dot Block
// ======================================================================================

func (p *Parser) dotBlock() ([]file.ScopeItem, error) {
	p.next() // lex.DotBlock

	indentItm := p.next()
	if indentItm.Type != lex.Indent {
		return nil, p.unexpectedItem(indentItm, lex.Indent)
	}

	var itms []file.ScopeItem

	first := true

	for {
		if p.peek().Type != lex.DotBlockLine {
			if len(itms) == 0 {
				return nil, p.unexpectedItem(p.peek(), lex.DotBlockLine)
			}

			dedentItm := p.next()
			if dedentItm.Type != lex.Dedent && dedentItm.Type != lex.EOF {
				return nil, p.unexpectedItem(dedentItm, lex.Dedent)
			}

			return itms, nil
		}

		p.next() // lex.DotBlockLine

		if !first {
			itms = append(itms, file.Text{Text: "\n"})
		} else {
			first = false
		}

		ts, err := p.text(false)
		if err != nil {
			return nil, err
		}

		itms = append(itms, ts...)
	}
}

// ============================================================================
// Pipe
// ======================================================================================

func (p *Parser) pipe() (itms []file.ScopeItem, err error) {
	first := true

	for {
		pipeItm := p.peek()
		if pipeItm.Type != lex.Pipe {
			return itms, nil
		}
		p.next()

		if !first {
			itms = append(itms, file.Text{Text: "\n"})
		} else {
			first = false
		}

		ts, err := p.text(false)
		if err != nil {
			return nil, err
		}

		itms = append(itms, ts...)
	}
}

// ============================================================================
// Assign
// ======================================================================================

func (p *Parser) assign() (*file.Interpolation, error) {
	var interp file.Interpolation

	next := p.next()
	switch next.Type {
	case lex.Assign:
	case lex.AssignNoEscape:
		interp.NoEscape = true
	default:
		return nil, p.unexpectedItem(next, lex.Assign, lex.AssignNoEscape)
	}

	exp, err := p.expression()
	if err != nil {
		return nil, err
	}

	interp.Expression = exp

	return &interp, nil
}

// ============================================================================
// Filter
// ======================================================================================

func (p *Parser) filter() (_ *file.Filter, err error) {
	filterItm := p.next() // lex.Filter
	f := file.Filter{Pos: file.Pos{Line: filterItm.Line, Col: filterItm.Col}}

	nameItm := p.next()
	if nameItm.Type != lex.Ident {
		return nil, p.unexpectedItem(nameItm, lex.Ident)
	}

	f.Name = nameItm.Val

	for p.peek().Type == lex.Literal {
		argItm := p.next()
		arg := argItm.Val
		if strings.HasPrefix(arg, `"`) || strings.HasPrefix(arg, "`") {
			arg, err = strconv.Unquote(arg)
			if err != nil {
				return nil, p.error(argItm, err)
			}
		}

		f.Args = append(f.Args, arg)
	}

	if p.peek().Type != lex.Indent {
		return &f, nil
	}

	p.next() // lex.Indent

	var b strings.Builder

	for p.peek().Type == lex.Text {
		b.WriteString(p.next().Val + "\n")
	}

	dedentItm := p.next()
	if dedentItm.Type != lex.Dedent && dedentItm.Type != lex.EOF {
		return nil, p.unexpectedItem(dedentItm, lex.Dedent)
	}

	text := b.String()
	if len(text) > 0 {
		text = text[:len(text)-1] // rm trailing newline
	}

	f.Body = file.Text{Text: text}
	return &f, nil
}

// ============================================================================
// Mixin Call
// ======================================================================================

func (p *Parser) mixinCall() (_ *file.MixinCall, err error) {
	var c file.MixinCall

	callItm := p.next()
	c.Pos = file.Pos{Line: callItm.Line, Col: callItm.Col}

	namespaceItm := p.next()
	if namespaceItm.Type != lex.Ident {
		return nil, p.unexpectedItem(namespaceItm, lex.Ident)
	}

	if p.peek().Type != lex.Ident { // wasn't the namespace, it was the name
		c.Name = file.Ident(namespaceItm.Val)
	} else {
		c.Namespace = file.Ident(namespaceItm.Val)
		c.Name = file.Ident(p.next().Val)
	}

	c.Args, err = p.mixinArgs()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == lex.MixinBlockShortcut {
		b, err := p.mixinBlockShortcut()
		if err != nil {
			return nil, err
		}

		c.Body = file.Scope{*b}
		return &c, nil
	}

	if p.peek().Type != lex.Indent {
		return &c, nil
	}

	p.next()

	p.context.Push(ContextMixinCall)

	c.Body, err = p.scope()
	if err != nil {
		return nil, err
	}

	p.context.Pop()

	return &c, nil
}

func (p *Parser) mixinArgs() (args []file.MixinArg, err error) {
	lparenItm := p.peek()
	if lparenItm.Type != lex.LParen {
		return nil, nil
	}

	p.next() // lex.LParen

	if p.peek().Type == lex.RParen {
		p.next() // lex.RParen
		return nil, nil
	}

	for {
		var arg file.MixinArg

		name := p.next()
		if name.Type != lex.Ident {
			return nil, p.unexpectedItem(name, lex.Ident)
		}

		arg.Pos = file.Pos{Line: name.Line, Col: name.Col}

		arg.Name = file.Ident(name.Val)

		assignItm := p.next()
		if assignItm.Type != lex.Assign {
			return nil, p.unexpectedItem(assignItm, lex.Assign)
		}

		arg.Value, err = p.expression()
		if err != nil {
			return nil, err
		}

		args = append(args, arg)

		next := p.next()
		switch next.Type {
		case lex.RParen:
			return args, nil
		case lex.Comma:
			switch p.peek().Type {
			case lex.EOF:
				return nil, p.unexpectedItem(p.next(), lex.Ident, lex.RParen)
			case lex.RParen:
				p.next()
				return args, nil
			}
		case lex.EOF:
			fallthrough
		default:
			return nil, p.unexpectedItem(next, lex.Comma, lex.RParen)
		}
	}
}

func (p *Parser) mixinBlockShortcut() (_ *file.Block, err error) {
	shortCutItm := p.next() // lex.MixinBlockShortcut
	b := file.Block{
		Type: file.BlockTypeBlock,
		Pos:  file.Pos{Line: shortCutItm.Line, Col: shortCutItm.Col},
	}

	switch p.peek().Type {
	case lex.DotBlock:
		b.Body, err = p.dotBlock()
		if err != nil {
			return nil, err
		}

		return &b, nil
	case lex.Text, lex.Hash:
		b.Body, err = p.text(true)
		if err != nil {
			return nil, err
		}

		return &b, nil
	case lex.Indent:
		p.next() // lex.Indent

		b.Body, err = p.scope()
		if err != nil {
			return nil, err
		}

		return &b, nil
	default:
		return nil, p.unexpectedItem(p.next(), lex.Indent, lex.Text, lex.DotBlock)
	}
}
