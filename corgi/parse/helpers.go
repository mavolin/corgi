package parse

import (
	"strings"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/lex"
)

// ============================================================================
// Whitespace
// ======================================================================================

func trimRightWhitespace(s string) string {
	return strings.TrimRight(s, " \t\n")
}

// ============================================================================
// Expression
// ======================================================================================

func (p *Parser) expression() (file.Expression, error) {
	peek := p.peek()
	switch peek.Type {
	case lex.Ternary:
		texpr, err := p.ternary()
		if err != nil {
			return nil, err
		}

		return *texpr, nil
	case lex.Expression:
		// handled below
	default:
		return nil, p.unexpectedItem(p.next(), lex.Expression, lex.Ternary)
	}

	exprItm := p.next()
	expr := file.GoExpression{
		Expression: exprItm.Val,
		Pos:        file.Pos{Line: exprItm.Line, Col: exprItm.Col},
	}

	if p.peek().Type == lex.NilCheck {
		ncExpr, err := p.nilCheck(expr)
		if err != nil {
			return nil, err
		}

		return *ncExpr, nil
	}

	return expr, nil
}

func (p *Parser) ternary() (*file.TernaryExpression, error) {
	p.next() // lex.Ternary

	condStartItm := p.peek()

	var texpr file.TernaryExpression

	var err error

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	if cond, ok := cond.(file.GoExpression); ok {
		texpr.Condition = cond
	} else {
		return nil, p.error(condStartItm, ErrTernaryCondition)
	}

	next := p.next()
	if next.Type != lex.LParen {
		return nil, p.unexpectedItem(next, lex.RParen)
	}

	texpr.IfTrue, err = p.expression()
	if err != nil {
		return nil, err
	}

	next = p.next()
	if next.Type != lex.TernaryElse {
		return nil, p.unexpectedItem(next, lex.TernaryElse)
	}

	texpr.IfTrue, err = p.expression()
	if err != nil {
		return nil, err
	}

	next = p.next()
	if next.Type != lex.RParen {
		return nil, p.unexpectedItem(next, lex.RParen)
	}

	return &texpr, nil
}

func (p *Parser) nilCheck(checkExpr file.GoExpression) (_ *file.NilCheckExpression, err error) {
	var ncExpr file.NilCheckExpression

	ncExpr.Root, ncExpr.Chain, err = p.parseValueExpression(checkExpr)
	if err != nil {
		return nil, err
	}

	p.next() // lex.NilCheck

	if p.peek().Type != lex.LParen {
		return &ncExpr, nil
	}

	// we have a default value

	p.next() // lex.LParen

	ncExpr.Default, err = p.expression()
	if err != nil {
		return nil, err
	}

	next := p.next()
	if next.Type != lex.RParen {
		return nil, p.unexpectedItem(next, lex.LParen)
	}

	return &ncExpr, nil
}

func (p *Parser) parseValueExpression(
	raw file.GoExpression,
) (root file.GoExpression, chain []file.ValueExpression, err error) {
	rawRunes := []rune(trimRightWhitespace(raw.Expression))

	var parenCount int

	var offset int

Root:
	for i, r := range rawRunes {
		switch r {
		case '(':
			parenCount++
		case ')':
			parenCount++
		case '.', '[':
			if parenCount == 0 {
				offset += i
				root.Expression = string(rawRunes[:i])
				rawRunes = rawRunes[i:]
				break Root
			}
		}
	}

	for len(rawRunes) > 0 {
		switch rawRunes[0] {
		case '.':
			rawRunes = rawRunes[1:]
			exprString := nextChainExpr(rawRunes)
			chain = append(chain, file.FieldFuncExpression{
				Expression: exprString,
				Pos:        file.Pos{Line: raw.Line, Col: raw.Col + offset},
			})

			offset += len(exprString) + 1
			rawRunes = rawRunes[len(exprString):]
		case '[':
			rawRunes = rawRunes[1:]
			exprString := nextIndexExpr(rawRunes)
			ffExpr := file.IndexExpression{
				Expression: exprString,
				Pos:        file.Pos{Line: raw.Line, Col: raw.Col + offset},
			}
			chain = append(chain, ffExpr)

			if len(rawRunes) < len(exprString)+1 {
				return root, chain, p.error(lex.Item{
					Line: ffExpr.Line,
					Col:  ffExpr.Col,
				}, ErrIndexExpression)
			}

			rawRunes = rawRunes[len(exprString)+1:] // strip the ']'
			offset += len(exprString) + 2
		default:
			panic("stopped at invalid indicator: " + string(rawRunes[0]))
		}
	}

	return root, chain, nil
}

func nextChainExpr(rawRunes []rune) string {
	var parenCount int

ChainExpr:
	for i, r := range rawRunes {
		switch r {
		case '(', '{':
			parenCount++
		case ')', '}':
			parenCount--
		case '[', '.':
			if parenCount > 0 {
				continue ChainExpr
			}

			return string(rawRunes[:i])
		}
	}

	return string(rawRunes)
}

func nextIndexExpr(rawRunes []rune) string {
	var parenCount int

IndexExpr:
	for i, r := range rawRunes {
		switch r {
		case '(', '{', '[':
			parenCount++
		case ')', '}':
			parenCount--
		case ']':
			if parenCount > 0 {
				continue IndexExpr
			}

			return string(rawRunes[:i])
		}
	}

	return string(rawRunes)
}

func (p *Parser) text(required bool) (itms []file.ScopeItem, err error) {
	for {
		peek := p.peek()
		switch peek.Type {
		case lex.Text:
			s := p.next().Val
			if p.peek().Type != lex.Text && p.peek().Type != lex.Hash {
				s = trimRightWhitespace(s)
			}

			s = strings.ReplaceAll(s, "##", "#")

			itms = append(itms, file.Text{Text: s})
		case lex.Hash:
			h, err := p.hash()
			if err != nil {
				return nil, err
			}

			itms = append(itms, h)
		default:
			if len(itms) == 0 && required {
				return nil, p.unexpectedItem(peek, lex.Text, lex.Hash)
			}

			return itms, nil
		}
	}
}

func (p *Parser) hash() (file.ScopeItem, error) {
	p.next() // lex.Hash

	var noEscape bool

	switch p.peek().Type {
	case lex.MixinCall:
		c, err := p.mixinCall()
		if err != nil {
			return nil, err
		}

		return *c, nil
	case lex.NoEscape:
		p.next()
		noEscape = true
	}

	if p.peek().Type == lex.LBracket {
		p.next()

		textItm := p.next()
		if textItm.Type != lex.Text {
			return nil, p.unexpectedItem(textItm, lex.Text)
		}

		if p.next().Type != lex.RBracket {
			return nil, p.unexpectedItem(p.next(), lex.RBracket)
		}

		return file.InlineText{Text: textItm.Val, NoEscape: noEscape}, nil
	}

	if p.peek().Type != lex.LBrace {
		e, err := p.inlineElement(noEscape)
		if err != nil {
			return nil, err
		}

		return *e, nil
	}

	p.next() // lex.LBrace

	e, err := p.expression()
	if err != nil {
		return nil, err
	}

	if p.next().Type != lex.RBrace {
		return nil, p.unexpectedItem(p.next(), lex.RBrace)
	}

	return file.Interpolation{Expression: e, NoEscape: noEscape}, nil
}

func (p *Parser) inlineElement(noEscape bool) (*file.InlineElement, error) {
	ie := file.InlineElement{NoEscape: noEscape}

	elem, err := p.elementHeader()
	if err != nil {
		return nil, err
	}

	ie.Name = elem.Name
	ie.Attributes = elem.Attributes
	ie.Classes = elem.Classes
	ie.SelfClosing = elem.SelfClosing

	switch p.peek().Type {
	case lex.LBracket:
		p.next()

		textItm := p.next()
		if textItm.Type != lex.Text {
			return nil, p.unexpectedItem(textItm, lex.Text)
		}

		ie.Value = file.Text{Text: textItm.Val}

		rBracketItm := p.next()
		if rBracketItm.Type != lex.RBracket {
			return nil, p.unexpectedItem(rBracketItm, lex.RBracket)
		}

		return &ie, nil
	case lex.LBrace:
		p.next()

		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		ie.Value = expr

		rBraceItm := p.next()
		if rBraceItm.Type != lex.RBrace {
			return nil, p.unexpectedItem(rBraceItm, lex.RBrace)
		}

		return &ie, nil
	default:
		return nil, p.unexpectedItem(p.next(), lex.LBracket, lex.LBrace)
	}
}
