package parse

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/lex/token"
	"github.com/mavolin/corgi/internal/require"
)

func (p *Parser) beforeDoctype() (_ stateFn, err error) {
	if p.mode == ModeMain && p.f.Func.Name == "" {
		return nil, p.unexpectedItem(p.next(), token.Func)
	}

	for {
		peek := p.peek()
		switch peek.Type {
		case token.Code:
			c, err := p.code()
			if err != nil {
				return nil, err
			}

			p.f.Scope = append(p.f.Scope, *c)
		case token.Mixin:
			m, err := p.mixin()
			if err != nil {
				return nil, err
			}

			p.f.Scope = append(p.f.Scope, *m)
		case token.Block:
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
		return nil, p.unexpectedItem(p.next(), token.Func)
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
		case token.EOF:
			return s, nil
		case token.Error:
			next := p.next()
			return nil, p.error(next, next.Err)
		case token.Comment:
			c, err := p.comment()
			if err != nil {
				return nil, err
			}

			s = append(s, *c)
			continue Loop
		case token.Mixin:
			m, err := p.mixin()
			if err != nil {
				return nil, err
			}

			s = append(s, *m)
			continue Loop
		}

		if p.mode != ModeUse && p.f.Extend == nil {
			switch peek.Type {
			case token.CodeStart:
				c, err := p.code()
				if err != nil {
					return nil, err
				}

				if c != nil { // empty line of code
					s = append(s, *c)
				}
				continue Loop
			case token.Include:
				incl, err := p.include()
				if err != nil {
					return nil, err
				}

				s = append(s, *incl)
				continue Loop
			case token.If:
				if_, err := p.if_()
				if err != nil {
					return nil, err
				}

				s = append(s, *if_)
			case token.Switch:
				sw, err := p.switch_()
				if err != nil {
					return nil, err
				}

				s = append(s, *sw)
			case token.For:
				f, err := p.for_()
				if err != nil {
					return nil, err
				}

				s = append(s, *f)
			case token.While:
				w, err := p.while()
				if err != nil {
					return nil, err
				}

				s = append(s, *w)
			case token.MixinCall:
				c, err := p.mixinCall()
				if err != nil {
					return nil, err
				}

				s = append(s, *c)
			case token.Assign, token.AssignNoEscape:
				a, err := p.assign()
				if err != nil {
					return nil, err
				}

				s = append(s, *a)
			case token.Pipe:
				pipe, err := p.pipe()
				if err != nil {
					return nil, err
				}

				s = append(s, pipe...)
			case token.DotBlock:
				d, err := p.dotBlock()
				if err != nil {
					return nil, err
				}

				s = append(s, d...)
			case token.Filter:
				f, err := p.filter()
				if err != nil {
					return nil, err
				}

				s = append(s, *f)
			case token.Div, token.Element:
				e, err := p.element()
				if err != nil {
					return nil, err
				}

				s = append(s, *e)
			case token.Block:
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
			case token.Block, token.Append, token.Prepend:
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
		case token.Dedent:
			p.next()
			return s, nil
		case token.EOF:
			p.next()
			return s, nil
		case token.Error:
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
			case token.Block:
				b, err := p.block(require.Optional, require.Optional)
				if err != nil {
					return nil, err
				}

				s = append(s, *b)
				continue Loop
			case token.IfBlock:
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
	case token.If:
		p.context.Push(ContextMixinCallConditional)

		if_, err := p.if_()
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *if_, nil
	case token.Switch:
		p.context.Push(ContextMixinCallConditional)

		switch_, err := p.switch_()
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *switch_, nil
	case token.For:
		p.context.Push(ContextMixinCallConditional)

		for_, err := p.for_()
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *for_, nil
	case token.While:
		p.context.Push(ContextMixinCallConditional)

		while, err := p.while()
		if err != nil {
			return nil, err
		}

		p.context.Pop()

		return *while, nil
	case token.CodeStart:
		code, err := p.code()
		if err != nil {
			return nil, err
		}

		return *code, nil
	case token.And:
		and, err := p.and()
		if err != nil {
			return nil, err
		}

		return *and, nil
	case token.MixinCall:
		c, err := p.mixinCall()
		if err != nil {
			return nil, err
		}

		return *c, nil
	case token.Block:
		prev := p.context.Pop()

		block, err := p.block(require.Optional, require.Optional)
		if err != nil {
			return nil, err
		}

		p.context.Push(prev)

		return *block, nil
	default:
		return nil, p.unexpectedItem(p.next())
	}
}

func (p *Parser) scopeItemMixinCallConditional() (file.ScopeItem, error) {
	switch p.peek().Type {
	case token.If:
		if_, err := p.if_()
		if err != nil {
			return nil, err
		}

		return *if_, nil
	case token.Switch:
		switch_, err := p.switch_()
		if err != nil {
			return nil, err
		}

		return *switch_, nil
	case token.For:
		for_, err := p.for_()
		if err != nil {
			return nil, err
		}

		return *for_, nil
	case token.While:
		while, err := p.while()
		if err != nil {
			return nil, err
		}

		return *while, nil
	case token.CodeStart:
		code, err := p.code()
		if err != nil {
			return nil, err
		}

		return *code, nil
	case token.And:
		and, err := p.and()
		if err != nil {
			return nil, err
		}

		return *and, nil
	case token.MixinCall:
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
	case token.Comment:
		c, err := p.comment()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*c}, nil
	case token.CodeStart:
		c, err := p.code()
		if err != nil {
			return nil, err
		}

		if c != nil { // empty line of code
			return []file.ScopeItem{*c}, nil
		}

		return nil, nil
	case token.Include:
		incl, err := p.include()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*incl}, nil
	case token.Mixin:
		m, err := p.mixin()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*m}, nil
	case token.If:
		if_, err := p.if_()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*if_}, nil
	case token.Switch:
		sw, err := p.switch_()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*sw}, nil
	case token.For:
		f, err := p.for_()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*f}, nil
	case token.While:
		w, err := p.while()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*w}, nil
	case token.MixinCall:
		c, err := p.mixinCall()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*c}, nil
	case token.Assign, token.AssignNoEscape:
		a, err := p.assign()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*a}, nil
	case token.Pipe:
		pipe, err := p.pipe()
		if err != nil {
			return nil, err
		}

		return pipe, nil
	case token.And:
		and, err := p.and()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*and}, nil
	case token.DotBlock:
		d, err := p.dotBlock()
		if err != nil {
			return nil, err
		}

		return d, nil
	case token.Filter:
		f, err := p.filter()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*f}, nil
	case token.Div, token.Element:
		e, err := p.element()
		if err != nil {
			return nil, err
		}

		return []file.ScopeItem{*e}, nil
	}

	if p.mode == ModeExtend {
		switch p.peek().Type {
		case token.Block:
			b, err := p.block(require.Always, require.Optional)
			if err != nil {
				return nil, err
			}

			return []file.ScopeItem{*b}, nil
		case token.IfBlock:
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
// Comment
// ======================================================================================

func (p *Parser) comment() (*file.Comment, error) {
	p.next() // token.Comment

	next := p.next()
	switch next.Type {
	case token.Text:
		return &file.Comment{Comment: next.Val}, nil
	case token.Indent:
		// handled below
	default:
		return nil, p.unexpectedItem(next, token.Indent, token.Text)
	}

	var b strings.Builder
	b.Grow(1000)

	first := true

	for {
		next = p.next()
		switch next.Type {
		case token.EOF, token.Dedent:
			return &file.Comment{Comment: b.String()}, nil
		case token.Text:
			if !first {
				b.WriteString("\n")
			} else {
				first = false
			}

			b.WriteString(next.Val)
		default:
			return nil, p.unexpectedItem(next, token.Text, token.Dedent)
		}
	}
}

// ============================================================================
// Include
// ======================================================================================

func (p *Parser) include() (_ *file.Include, err error) {
	includeItm := p.next() // token.Include

	pathItm := p.next()
	if pathItm.Type != token.Literal {
		return nil, p.unexpectedItem(pathItm, token.Literal)
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
	p.next() // token.CodeStart

	peek := p.peek()
	switch peek.Type {
	case token.Code:
		return &file.Code{Code: p.next().Val}, nil
	case token.Indent:
		// handled below
	default: // empty line of code
		// add nothing because empty lines only serve the purpose of aiding
		// readability, a trait that is not required for a generated file
		return nil, nil
	}

	var b strings.Builder
	b.Grow(1000)

	p.next() // token.Indent

	first := true

	for {
		switch p.peek().Type {
		case token.EOF, token.Dedent:
			p.next()
			return &file.Code{Code: b.String()}, nil
		case token.Code:
			if !first {
				b.WriteString("\n")
			} else {
				first = false
			}

			b.WriteString(p.next().Val)
		default:
			return nil, p.unexpectedItem(p.next(), token.Code, token.Dedent)
		}
	}
}

func (p *Parser) block(named, body require.Required) (_ *file.Block, err error) {
	blockItm := p.next()
	b := file.Block{Pos: file.Pos{Line: blockItm.Line, Col: blockItm.Col}}

	switch blockItm.Type {
	case token.Block:
		b.Type = file.BlockTypeBlock
	case token.Append:
		b.Type = file.BlockTypeAppend
	case token.Prepend:
		b.Type = file.BlockTypePrepend
	default:
		return nil, p.unexpectedItem(blockItm, token.Block, token.Append, token.Prepend)
	}

	if named > require.Never {
		if p.peek().Type == token.Ident {
			b.Name = file.Ident(p.next().Val)
		} else if named == require.Always {
			return nil, p.unexpectedItem(p.next(), token.Ident)
		}
	}

	if body == require.Never {
		return &b, nil
	}

	switch p.peek().Type {
	case token.EOF:
		p.next()
		return &b, nil
	case token.DotBlock:
		b.Body, err = p.dotBlock()
		if err != nil {
			return nil, err
		}

		return &b, nil
	case token.Indent:
		// handled below
	default:
		if body == require.Optional {
			return &b, nil
		}

		return nil, p.unexpectedItem(p.next(), token.Indent, token.DotBlock)
	}

	p.next() // token.Indent

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
	mixinItm := p.next() // token.Mixin

	m := file.Mixin{Pos: file.Pos{Line: mixinItm.Line, Col: mixinItm.Col}}

	nameItm := p.next()
	if nameItm.Type != token.Ident {
		return nil, p.unexpectedItem(nameItm, token.Ident)
	}

	m.Name = file.Ident(nameItm.Val)

	m.Params, err = p.mixinParams()
	if err != nil {
		return nil, err
	}

	indentItm := p.next()
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
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
	if lparenItm.Type != token.LParen {
		return nil, p.unexpectedItem(lparenItm, token.LParen)
	}

	if p.peek().Type == token.RParen {
		p.next()
		return nil, nil
	}

	for {
		var param file.MixinParam

		paramName := p.next()
		if paramName.Type != token.Ident {
			return nil, p.unexpectedItem(paramName, token.Ident)
		}

		param.Name = file.Ident(paramName.Val)
		param.Pos = file.Pos{Line: paramName.Line, Col: paramName.Col}

		if p.peek().Type == token.Ident { // type
			param.Type = file.GoIdent(p.next().Val)
		}

		if p.peek().Type == token.Assign {
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
		case token.EOF:
			return nil, p.unexpectedItem(next, token.Comma, token.RParen)
		case token.RParen:
			return params, nil
		case token.Comma:
			switch p.peek().Type {
			case token.EOF:
				return nil, p.unexpectedItem(p.next(), token.Ident, token.RParen)
			case token.RParen:
				p.next()
				return params, nil
			}
		default:
			return nil, p.unexpectedItem(next, token.Comma, token.RParen)
		}
	}
}

// ============================================================================
// IfBlock
// ======================================================================================

func (p *Parser) ifBlock(mustBeNamed bool) (_ *file.IfBlock, err error) {
	var i file.IfBlock

	p.next() // token.IfBlock

	nameItm := p.peek()
	if nameItm.Type == token.Ident {
		nameItm = p.next()
		i.Name = file.Ident(nameItm.Val)
	} else if mustBeNamed {
		return nil, p.unexpectedItem(p.next(), token.Ident)
	}

	indentItm := p.next()
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
	}

	i.Then, err = p.scope()
	if err != nil {
		return nil, err
	}

	if p.peek().Type != token.Else {
		return &i, nil
	}

	p.next() // token.Else

	indentItm = p.next()
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
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
	p.next() // token.If

	var i file.If

	i.Condition, err = p.expression()
	if err != nil {
		return nil, err
	}

	switch p.peek().Type {
	case token.DotBlock:
		i.Then, err = p.dotBlock()
		if err != nil {
			return nil, err
		}
	case token.BlockExpansion:
		content, err := p.blockExpansion()
		if err != nil {
			return nil, err
		}

		i.Then = file.Scope{content}
	case token.Indent:
		p.next() // token.Indent

		i.Then, err = p.scope()
		if err != nil {
			return nil, err
		}
	case token.EOF:
		fallthrough
	default:
		return nil, p.unexpectedItem(p.next(), token.Indent, token.DotBlock, token.BlockExpansion)
	}

	for {
		var ei file.ElseIf

		if p.peek().Type != token.ElseIf {
			break
		}

		p.next() // token.ElseIf

		ei.Condition, err = p.expression()
		if err != nil {
			return nil, err
		}

		switch p.peek().Type {
		case token.DotBlock:
			ei.Then, err = p.dotBlock()
			if err != nil {
				return nil, err
			}
		case token.BlockExpansion:
			content, err := p.blockExpansion()
			if err != nil {
				return nil, err
			}

			ei.Then = file.Scope{content}
		case token.Indent:
			p.next() // token.Indent

			ei.Then, err = p.scope()
			if err != nil {
				return nil, err
			}
		case token.EOF:
			fallthrough
		default:
			return nil, p.unexpectedItem(p.next(), token.Indent, token.DotBlock, token.BlockExpansion)
		}

		i.ElseIfs = append(i.ElseIfs, ei)
	}

	if p.peek().Type != token.Else {
		return &i, nil
	}

	p.next() // token.Else

	var e file.Else

	switch p.peek().Type {
	case token.DotBlock:
		e.Then, err = p.dotBlock()
		if err != nil {
			return nil, err
		}
	case token.BlockExpansion:
		content, err := p.blockExpansion()
		if err != nil {
			return nil, err
		}

		e.Then = file.Scope{content}
	case token.Indent:
		p.next() // token.Indent

		e.Then, err = p.scope()
		if err != nil {
			return nil, err
		}
	case token.EOF:
		fallthrough
	default:
		return nil, p.unexpectedItem(p.next(), token.Indent, token.DotBlock, token.BlockExpansion)
	}

	i.Else = &e
	return &i, nil
}

// ============================================================================
// Switch
// ======================================================================================

func (p *Parser) switch_() (_ *file.Switch, err error) {
	p.next() // token.Switch

	var s file.Switch

	if p.peek().Type != token.Indent {
		s.Comparator, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	indentItm := p.next()
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
	}

	for {
		var c file.Case

		caseItm := p.peek()
		if caseItm.Type != token.Case {
			break
		}

		p.next() // token.Case

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
		case token.Indent:
			p.next()

			c.Then, err = p.scope()
			if err != nil {
				return nil, err
			}
		case token.DotBlock:
			c.Then, err = p.dotBlock()
			if err != nil {
				return nil, err
			}
		case token.BlockExpansion:
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
	case token.EOF:
		fallthrough
	case token.Dedent:
		return &s, nil
	case token.Default:
		// handled below
	default:
		return nil, p.unexpectedItem(next, token.Default, token.EOF, token.Dedent)
	}

	var d file.DefaultCase

	switch p.peek().Type {
	case token.Indent:
		p.next()

		d.Then, err = p.scope()
		if err != nil {
			return nil, err
		}
	case token.DotBlock:
		d.Then, err = p.dotBlock()
		if err != nil {
			return nil, err
		}
	case token.BlockExpansion:
		content, err := p.blockExpansion()
		if err != nil {
			return nil, err
		}

		d.Then = file.Scope{content}
	default:
		return nil, p.unexpectedItem(p.next(), token.Indent, token.DotBlock, token.BlockExpansion)
	}

	s.Default = &d

	next = p.next()
	switch next.Type {
	case token.EOF:
		return &s, nil
	case token.Dedent:
		return &s, nil
	default:
		return nil, p.unexpectedItem(next, token.Dedent)
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

	if p.peek().Type == token.Ident {
		ident1Itm := p.next()
		if ident1Itm.Type != token.Ident {
			return nil, p.unexpectedItem(ident1Itm, token.Ident)
		}

		f.VarOne = file.GoIdent(ident1Itm.Val)

		if p.peek().Type == token.Comma {
			p.next()

			ident2Itm := p.next()
			if ident2Itm.Type != token.Ident {
				return nil, p.unexpectedItem(ident2Itm, token.Ident)
			}

			f.VarTwo = file.GoIdent(ident2Itm.Val)
		} else if p.peek().Type != token.Range {
			return nil, p.unexpectedItem(p.next(), token.Comma, token.Range)
		}
	} else if p.peek().Type != token.Range {
		return nil, p.unexpectedItem(p.next(), token.Ident, token.Range)
	}

	rangeItm := p.next()
	if rangeItm.Type != token.Range {
		return nil, p.unexpectedItem(rangeItm, token.Range)
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	f.Range = expr

	indentItm := p.next()
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
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
	whileItm := p.next() // token.While

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
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
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
	case token.Assign:
		p.next() // token.Assign

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		e.Body = file.Scope{file.Interpolation{Expression: expr}}
		return e, nil
	case token.AssignNoEscape:
		p.next() // token.Assign

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
	case token.Indent:
		p.next()

		e.Body, err = p.scope()
		if err != nil {
			return nil, err
		}

		return e, nil
	}

	if !e.SelfClosing {
		switch p.peek().Type {
		case token.BlockExpansion:
			if e.SelfClosing {
				return nil, p.unexpectedItem(p.next(), token.Assign, token.AssignNoEscape)
			}

			exp, err := p.blockExpansion()
			if err != nil {
				return nil, err
			}

			e.Body = file.Scope{exp}
			return e, nil
		case token.DotBlock:
			e.Body, err = p.dotBlock()
			if err != nil {
				return nil, err
			}

			return e, nil
		case token.Text, token.Interpolation:
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
	case token.Div:
		e.Name = "div"
	case token.Element:
		e.Name = name.Val
	default:
		return nil, p.unexpectedItem(name, token.Element, token.Div)
	}

	for {
		switch p.peek().Type {
		case token.Class:
			class, err := p.class()
			if err != nil {
				return nil, err
			}

			e.Classes = append(e.Classes, *class)
		case token.ID:
			id, err := p.id()
			if err != nil {
				return nil, err
			}

			e.Attributes = append(e.Attributes, *id)
		case token.LParen:
			as, cs, err := p.attributes()
			if err != nil {
				return nil, err
			}

			e.Attributes = append(e.Attributes, as...)
			e.Classes = append(e.Classes, cs...)
		case token.TagVoid:
			p.next()
			e.SelfClosing = true
			return &e, nil
		default:
			return &e, nil
		}
	}
}

func (p *Parser) class() (*file.ClassLiteral, error) {
	p.next() // token.Class

	name := p.next()
	if name.Type != token.Literal {
		return nil, p.unexpectedItem(name, token.Literal)
	}

	return &file.ClassLiteral{Name: name.Val}, nil
}

func (p *Parser) id() (*file.AttributeLiteral, error) {
	p.next() // token.Id

	name := p.next()
	if name.Type != token.Literal {
		return nil, p.unexpectedItem(name, token.Literal)
	}

	return &file.AttributeLiteral{Name: "id", Value: name.Val}, nil
}

func (p *Parser) attributes() ([]file.Attribute, []file.Class, error) {
	p.next() // token.LParen

	// special case: empty attribute list
	if p.peek().Type == token.RParen {
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
		if nameItm.Type != token.Ident {
			return nil, nil, p.unexpectedItem(nameItm, token.Ident)
		}

		name = trimRightWhitespace(nameItm.Val)

		switch p.peek().Type {
		case token.AssignNoEscape:
			unescaped = true
			fallthrough
		case token.Assign:
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
		case token.RParen:
			return attributes, classes, nil
		case token.Comma:
			switch p.peek().Type {
			case token.EOF:
				return nil, nil, p.unexpectedItem(p.next(), token.Ident, token.RParen)
			case token.RParen:
				p.next()
				return attributes, classes, nil
			}
		case token.EOF:
			fallthrough
		default:
			return nil, nil, p.unexpectedItem(next, token.Comma, token.RParen)
		}
	}
}

// ============================================================================
// Block Expansion
// ======================================================================================

func (p *Parser) blockExpansion() (file.ScopeItem, error) {
	p.next() // token.BlockExpansion

	peek := p.peek()
	if peek.Type == token.EOF {
		return nil, p.unexpectedItem(p.next())
	}

	if peek.Type == token.Block {
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

	// todo: make sure we get at least one attr

	for {
		switch p.peek().Type {
		case token.Class:
			class, err := p.class()
			if err != nil {
				return nil, err
			}

			a.Classes = append(a.Classes, *class)
		case token.ID:
			id, err := p.id()
			if err != nil {
				return nil, err
			}

			a.Attributes = append(a.Attributes, *id)
		case token.LParen:
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
	p.next() // token.DotBlock

	indentItm := p.next()
	if indentItm.Type != token.Indent {
		return nil, p.unexpectedItem(indentItm, token.Indent)
	}

	var itms []file.ScopeItem

	first := true

	for {
		if p.peek().Type != token.DotBlockLine {
			if len(itms) == 0 {
				return nil, p.unexpectedItem(p.peek(), token.DotBlockLine)
			}

			dedentItm := p.next()
			if dedentItm.Type != token.Dedent && dedentItm.Type != token.EOF {
				return nil, p.unexpectedItem(dedentItm, token.DotBlock, token.Dedent)
			}

			return itms, nil
		}

		p.next() // token.DotBlockLine

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
		if pipeItm.Type != token.Pipe {
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
	case token.Assign:
	case token.AssignNoEscape:
		interp.NoEscape = true
	default:
		return nil, p.unexpectedItem(next, token.Assign, token.AssignNoEscape)
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
	filterItm := p.next() // token.Filter
	f := file.Filter{Pos: file.Pos{Line: filterItm.Line, Col: filterItm.Col}}

	nameItm := p.next()
	if nameItm.Type != token.Ident {
		return nil, p.unexpectedItem(nameItm, token.Ident)
	}

	f.Name = nameItm.Val

	for p.peek().Type == token.Literal {
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

	if p.peek().Type != token.Indent {
		return &f, nil
	}

	p.next() // token.Indent

	var b strings.Builder

	for p.peek().Type == token.Text {
		b.WriteString(p.next().Val + "\n")
	}

	dedentItm := p.next()
	if dedentItm.Type != token.Dedent && dedentItm.Type != token.EOF {
		return nil, p.unexpectedItem(dedentItm, token.Dedent)
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
	if namespaceItm.Type != token.Ident {
		return nil, p.unexpectedItem(namespaceItm, token.Ident)
	}

	if p.peek().Type != token.Ident { // wasn't the namespace, it was the name
		c.Name = file.Ident(namespaceItm.Val)
	} else {
		c.Namespace = file.Ident(namespaceItm.Val)
		c.Name = file.Ident(p.next().Val)
	}

	c.Args, err = p.mixinArgs()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == token.MixinMainBlockShorthand {
		b, err := p.mixinBlockShortcut()
		if err != nil {
			return nil, err
		}

		c.Body = file.Scope{*b}
		return &c, nil
	}

	if p.peek().Type != token.Indent {
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
	if lparenItm.Type != token.LParen {
		return nil, nil
	}

	p.next() // token.LParen

	if p.peek().Type == token.RParen {
		p.next() // token.RParen
		return nil, nil
	}

	for {
		var arg file.MixinArg

		name := p.next()
		if name.Type != token.Ident {
			return nil, p.unexpectedItem(name, token.Ident)
		}

		arg.Pos = file.Pos{Line: name.Line, Col: name.Col}

		arg.Name = file.Ident(name.Val)

		assignItm := p.next()
		if assignItm.Type != token.Assign {
			return nil, p.unexpectedItem(assignItm, token.Assign)
		}

		arg.Value, err = p.expression()
		if err != nil {
			return nil, err
		}

		args = append(args, arg)

		next := p.next()
		switch next.Type {
		case token.RParen:
			return args, nil
		case token.Comma:
			switch p.peek().Type {
			case token.EOF:
				return nil, p.unexpectedItem(p.next(), token.Ident, token.RParen)
			case token.RParen:
				p.next()
				return args, nil
			}
		case token.EOF:
			fallthrough
		default:
			return nil, p.unexpectedItem(next, token.Comma, token.RParen)
		}
	}
}

func (p *Parser) mixinBlockShortcut() (_ *file.Block, err error) {
	shortCutItm := p.next() // token.MixinBlockShortcut
	b := file.Block{
		Type: file.BlockTypeBlock,
		Pos:  file.Pos{Line: shortCutItm.Line, Col: shortCutItm.Col},
	}

	switch p.peek().Type {
	case token.DotBlock:
		b.Body, err = p.dotBlock()
		if err != nil {
			return nil, err
		}

		return &b, nil
	case token.Text, token.Interpolation:
		b.Body, err = p.text(true)
		if err != nil {
			return nil, err
		}

		return &b, nil
	case token.Indent:
		p.next() // token.Indent

		b.Body, err = p.scope()
		if err != nil {
			return nil, err
		}

		return &b, nil
	default:
		return nil, p.unexpectedItem(p.next(), token.Indent, token.Text, token.DotBlock)
	}
}
