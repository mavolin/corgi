package parse

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/lex"
)

func (p *Parser) start() (stateFn, error) {
	switch p.peek().Type {
	case lex.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case lex.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case lex.Extend:
		return p.extend, nil
	case lex.Import:
		return p.import_, nil
	case lex.Use:
		return p.use, nil
	case lex.CodeStart:
		return p.globalCode, nil
	case lex.Func:
		return p.func_, nil
	case lex.Doctype:
		return p.doctype, nil
	default:
		return p.nextHTML, nil
	}
}

// ============================================================================
// Extend
// ======================================================================================

func (p *Parser) extend() (_ stateFn, err error) {
	extend := p.next()
	if p.mode == ModeUse {
		return nil, p.error(extend, ErrUseExtends)
	}

	lit := p.next()
	if lit.Type != lex.Literal {
		return nil, p.unexpectedItem(lit, lex.Literal)
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
	case lex.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case lex.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case lex.Extend:
		return nil, p.error(p.next(), ErrMultipleExtend)
	case lex.Import:
		return p.import_, nil
	case lex.Use:
		return p.use, nil
	case lex.CodeStart:
		return p.globalCode, nil
	case lex.Func:
		return p.func_, nil
	case lex.Doctype:
		return p.doctype, nil
	default:
		return p.nextHTML, nil
	}
}

// ============================================================================
// Import
// ======================================================================================

func (p *Parser) import_() (stateFn, error) {
	p.next() // lex.Import

	peek := p.peek()
	if peek.Type == lex.Ident || peek.Type == lex.Literal {
		return p.afterImport, p.singleImport()
	}

	if p.next().Type != lex.Indent {
		return nil, p.unexpectedItem(p.next(), lex.Indent, lex.Ident, lex.Literal)
	}

	for {
		switch p.peek().Type {
		case lex.EOF:
			p.next()
			return nil, nil
		case lex.Dedent:
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
	if pathItm.Type == lex.Ident { // not the path, but the alias
		first = pathItm
		alias = file.GoIdent(pathItm.Val)
		pathItm = p.next()
	}

	if pathItm.Type != lex.Literal {
		return p.unexpectedItem(pathItm, lex.Literal)
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
	case lex.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case lex.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case lex.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case lex.Import:
		return p.import_, nil
	case lex.Use:
		return p.use, nil
	case lex.CodeStart:
		return p.globalCode, nil
	case lex.Func:
		return p.func_, nil
	case lex.Doctype:
		return p.doctype, nil
	default:
		return p.nextHTML, nil
	}
}

// ============================================================================
// Use
// ======================================================================================

func (p *Parser) use() (stateFn, error) {
	p.next() // lex.Use

	peek := p.peek()
	if peek.Type == lex.Literal || peek.Type == lex.Ident {
		return p.afterUse, p.singleUse()
	}

	if p.next().Type != lex.Indent {
		return nil, p.unexpectedItem(p.next(), lex.Ident, lex.Literal)
	}

	for {
		switch p.peek().Type {
		case lex.EOF:
			p.next()
			return nil, nil
		case lex.Dedent:
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

	if next.Type == lex.Ident {
		u.Namespace = file.Ident(next.Val)
		next = p.next()
	}

	if next.Type != lex.Literal {
		return p.unexpectedItem(next, lex.Literal)
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
	case lex.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case lex.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case lex.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case lex.Import:
		return nil, p.error(p.next(), ErrImportPlacement)
	case lex.Use:
		return p.use, nil
	case lex.CodeStart:
		return p.globalCode, nil
	case lex.Func:
		return p.func_, nil
	case lex.Doctype:
		return p.doctype, nil
	default:
		return p.nextHTML, nil
	}
}

// ============================================================================
// Global Code
// ======================================================================================

func (p *Parser) globalCode() (stateFn, error) {
	p.next() // lex.CodeStart

	peek := p.peek()
	switch peek.Type {
	case lex.Code:
		p.f.GlobalCode = append(p.f.GlobalCode, file.Code{Code: p.next().Val})
		return p.afterGlobalCode, nil
	case lex.Indent:
		// handled below
	default: // empty line of code
		// add nothing because empty lines only serve the purpose of aiding
		// readability, a trait that is not required for a generated file
		return p.afterGlobalCode, nil
	}

	var b strings.Builder
	b.Grow(1000)

	p.next() // lex.Indent

	first := true

	for {
		switch p.peek().Type {
		case lex.EOF:
			p.next()
			p.f.GlobalCode = append(p.f.GlobalCode, file.Code{Code: b.String()})
			return nil, nil
		case lex.Dedent:
			p.next()
			p.f.GlobalCode = append(p.f.GlobalCode, file.Code{Code: b.String()})
			return p.afterGlobalCode, nil
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

func (p *Parser) afterGlobalCode() (stateFn, error) {
	switch p.peek().Type {
	case lex.EOF:
		if p.mode == ModeMain {
			return nil, p.error(p.next(), ErrNoFunc)
		}

		return nil, nil
	case lex.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case lex.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case lex.Import:
		return nil, p.error(p.next(), ErrImportPlacement)
	case lex.Use:
		return p.use, p.error(p.next(), ErrUsePlacement)
	case lex.CodeStart:
		return p.globalCode, nil
	case lex.Func:
		return p.func_, nil
	case lex.Doctype:
		return p.doctype, nil
	default:
		return p.nextHTML, nil
	}
}

// ============================================================================
// Func
// ======================================================================================

func (p *Parser) func_() (stateFn, error) {
	funcItm := p.next()
	if p.mode == ModeUse {
		return nil, p.error(funcItm, ErrUseFunc)
	}

	ident := p.next()
	if ident.Type != lex.Ident {
		return nil, p.unexpectedItem(ident, lex.Ident)
	}

	params := p.next()
	if params.Type != lex.Literal {
		// LParen is not actually what we expect here, but it's certainly more
		// descriptive than Literal
		return nil, p.unexpectedItem(params, lex.LParen)
	}

	p.f.Func = file.Func{
		Name:   file.GoIdent(ident.Val),
		Params: file.GoExpression{Expression: params.Val},
	}

	return p.afterFunc, nil
}

func (p *Parser) afterFunc() (stateFn, error) {
	switch p.peek().Type {
	case lex.EOF:
		return nil, nil
	case lex.Error:
		next := p.next()
		return nil, p.error(next, next.Err)
	case lex.Extend:
		return nil, p.error(p.next(), ErrExtendPlacement)
	case lex.Import:
		return nil, p.error(p.next(), ErrImportPlacement)
	case lex.Use:
		return nil, p.error(p.next(), ErrUsePlacement)
	case lex.Func:
		return nil, p.error(p.next(), ErrMultipleFunc)
	case lex.Doctype:
		return p.doctype, nil
	default:
		return p.nextHTML, nil
	}
}
