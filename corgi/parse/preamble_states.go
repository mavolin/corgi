package parse

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/lex/token"
)

func (p *Parser) start() (stateFn, error) {
	switch p.peek().Type {
	case token.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case token.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case token.Extend:
		return p.extend, nil
	case token.Import:
		return p.import_, nil
	case token.Use:
		return p.use, nil
	case token.CodeStart:
		// all code in included files is regular code
		if p.mode == ModeInclude {
			return p.beforeDoctype, nil
		}

		return p.globalCode, nil
	case token.Func:
		return p.func_, nil
	default:
		return p.beforeDoctype, nil
	}
}

// ============================================================================
// Extend
// ======================================================================================

func (p *Parser) extend() (_ stateFn, err error) {
	extend := p.next()
	switch p.mode {
	case ModeUse:
		return nil, p.error(extend, ErrUseExtends)
	case ModeInclude:
		return nil, p.error(extend, ErrIncludeExtends)
	}

	lit := p.next()
	if lit.Type != token.Literal {
		return nil, p.unexpectedItem(lit, token.Literal)
	}

	p.f.Extend = &file.Extend{
		Pos: file.Pos{Line: extend.Line, Col: extend.Col},
	}

	p.f.Extend.Path, err = strconv.Unquote(lit.Val)
	if err != nil {
		return nil, p.error(lit, errors.Wrap(err, "invalid extend path"))
	}

	return p.afterExtend, nil
}

func (p *Parser) afterExtend() (stateFn, error) {
	switch p.peek().Type {
	case token.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case token.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case token.Extend:
		return nil, p.error(p.next(), ErrMultipleExtend)
	case token.Import:
		return p.import_, nil
	case token.Use:
		return p.use, nil
	case token.CodeStart: // all code in included files is regular code
		if p.mode == ModeInclude {
			return p.beforeDoctype, nil
		}

		return p.globalCode, nil
	case token.Func:
		return p.func_, nil
	default:
		return p.beforeDoctype, nil
	}
}

// ============================================================================
// Import
// ======================================================================================

func (p *Parser) import_() (stateFn, error) {
	p.next() // token.Import

	peek := p.peek()
	if peek.Type == token.Ident || peek.Type == token.Literal {
		return p.afterImport, p.singleImport()
	}

	if p.next().Type != token.Indent {
		return nil, p.unexpectedItem(p.next(), token.Indent, token.Ident, token.Literal)
	}

	for {
		switch p.peek().Type {
		case token.EOF:
			p.next()
			return nil, nil
		case token.Dedent:
			p.next()
			return p.afterImport, nil
		}

		if err := p.singleImport(); err != nil {
			return nil, err
		}
	}
}

func (p *Parser) singleImport() (err error) {
	var alias file.GoIdent

	pathItm := p.next()
	first := pathItm
	if pathItm.Type == token.Ident { // not the path, but the alias
		first = pathItm
		alias = file.GoIdent(pathItm.Val)
		pathItm = p.next()
	}

	if pathItm.Type != token.Literal {
		return p.unexpectedItem(pathItm, token.Literal)
	}

	imp := file.Import{
		Alias:  alias,
		Pos:    file.Pos{Line: first.Line, Col: first.Col},
		Source: p.f.Source,
		File:   p.f.Name,
	}

	imp.Path, err = strconv.Unquote(pathItm.Val)
	if err != nil {
		return p.error(pathItm, errors.Wrap(err, "invalid use string"))
	}

	p.f.Imports = append(p.f.Imports, imp)

	return nil
}

func (p *Parser) afterImport() (stateFn, error) {
	switch p.peek().Type {
	case token.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case token.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case token.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case token.Import:
		return p.import_, nil
	case token.Use:
		return p.use, nil
	case token.CodeStart:
		// all code in included files is regular code
		if p.mode == ModeInclude {
			return p.beforeDoctype, nil
		}

		return p.globalCode, nil
	case token.Func:
		return p.func_, nil
	default:
		return p.beforeDoctype, nil
	}
}

// ============================================================================
// Use
// ======================================================================================

func (p *Parser) use() (stateFn, error) {
	p.next() // token.Use

	peek := p.peek()
	if peek.Type == token.Literal || peek.Type == token.Ident {
		return p.afterUse, p.singleUse()
	}

	if p.next().Type != token.Indent {
		return nil, p.unexpectedItem(p.next(), token.Ident, token.Literal)
	}

	for {
		switch p.peek().Type {
		case token.EOF:
			p.next()
			return nil, nil
		case token.Dedent:
			p.next()
			return p.afterUse, nil
		}

		if err := p.singleUse(); err != nil {
			return nil, err
		}
	}
}

func (p *Parser) singleUse() (err error) {
	var u file.Use

	next := p.next()
	u.Pos = file.Pos{Line: next.Line, Col: next.Col}

	if next.Type == token.Ident {
		u.Namespace = file.Ident(next.Val)
		next = p.next()
	}

	if next.Type != token.Literal {
		return p.unexpectedItem(next, token.Literal)
	}

	u.Path, err = strconv.Unquote(next.Val)
	if err != nil {
		return p.error(next, errors.Wrap(err, "invalid use string"))
	}

	if u.Namespace == "" {
		u.Namespace = file.Ident(filepath.Base(u.Path))
	}

	p.f.Uses = append(p.f.Uses, u)

	return nil
}

func (p *Parser) afterUse() (stateFn, error) {
	switch p.peek().Type {
	case token.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case token.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case token.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case token.Import:
		return nil, p.error(p.next(), ErrImportPlacement)
	case token.Use:
		return p.use, nil
	case token.CodeStart:
		// all code in included files is regular code
		if p.mode == ModeInclude {
			return p.beforeDoctype, nil
		}

		return p.globalCode, nil
	case token.Func:
		return p.func_, nil
	default:
		return p.beforeDoctype, nil
	}
}

// ============================================================================
// Global Code
// ======================================================================================

func (p *Parser) globalCode() (stateFn, error) {
	p.next() // token.CodeStart

	peek := p.peek()
	switch peek.Type {
	case token.Code:
		p.f.GlobalCode = append(p.f.GlobalCode, file.Code{Code: p.next().Val})
		return p.afterGlobalCode, nil
	case token.Indent:
		// handled below
	default: // empty line of code
		// add nothing because empty lines only serve the purpose of aiding
		// readability, a trait that is not required for a generated file
		return p.afterGlobalCode, nil
	}

	var b strings.Builder
	b.Grow(1000)

	p.next() // token.Indent

	first := true

	for {
		switch p.peek().Type {
		case token.EOF:
			p.next()
			p.f.GlobalCode = append(p.f.GlobalCode, file.Code{Code: b.String()})
			return nil, nil
		case token.Dedent:
			p.next()
			p.f.GlobalCode = append(p.f.GlobalCode, file.Code{Code: b.String()})
			return p.afterGlobalCode, nil
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

func (p *Parser) afterGlobalCode() (stateFn, error) {
	switch p.peek().Type {
	case token.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case token.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case token.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case token.Import:
		return nil, p.error(p.next(), ErrImportPlacement)
	case token.Use:
		return p.use, p.error(p.next(), ErrUsePlacement)
	case token.CodeStart:
		// all code in included files is regular code
		if p.mode == ModeInclude {
			return p.beforeDoctype, nil
		}

		return p.globalCode, nil
	case token.Func:
		return p.func_, nil
	default:
		return p.beforeDoctype, nil
	}
}

// ============================================================================
// Func
// ======================================================================================

func (p *Parser) func_() (stateFn, error) {
	funcItm := p.next()
	switch p.mode {
	case ModeUse:
		return nil, p.error(funcItm, ErrUseFunc)
	case ModeExtend:
		return nil, p.error(funcItm, ErrExtendFunc)
	}

	ident := p.next()
	if ident.Type != token.Ident {
		return nil, p.unexpectedItem(ident, token.Ident)
	}

	params := p.next()
	if params.Type != token.Literal {
		// LParen is not actually what we expect here, but it's certainly more
		// descriptive than Literal
		return nil, p.unexpectedItem(params, token.LParen)
	}

	p.f.Func = file.Func{
		Name:   file.GoIdent(ident.Val),
		Params: file.GoExpression{Expression: params.Val},
	}

	return p.afterFunc, nil
}

func (p *Parser) afterFunc() (stateFn, error) {
	switch p.peek().Type {
	case token.EOF:
		return nil, nil
	case token.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case token.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case token.Import:
		return nil, p.error(p.next(), ErrImportPlacement)
	case token.Use:
		return nil, p.error(p.next(), ErrUsePlacement)
	case token.Func:
		return nil, p.error(p.next(), ErrMultipleFunc)
	default:
		return p.beforeDoctype, nil
	}
}
