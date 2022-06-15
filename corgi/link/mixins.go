package link

import (
	"unicode"

	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/pkg/stack"
)

func (l *Linker) linkMixinCalls() error {
	scopes := stack.New[file.Scope](200)
	return l.linkMixinsCallsScope(l.f.Scope, &scopes)
}

func (l *Linker) linkMixinsCallsScope(s file.Scope, scopes *stack.Stack[file.Scope]) error {
	scopes.Push(s)

	for i, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if err := l.linkMixinsCallsScope(itm.Body, scopes); err != nil {
				return err
			}
		case file.Element:
			if err := l.linkMixinsCallsScope(itm.Body, scopes); err != nil {
				return err
			}
		case file.If:
			if err := l.linkMixinsCallsScope(itm.Then, scopes); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.linkMixinsCallsScope(ei.Then, scopes); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.linkMixinsCallsScope(itm.Else.Then, scopes); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.linkMixinsCallsScope(itm.Then, scopes); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.linkMixinsCallsScope(itm.Else.Then, scopes); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.linkMixinsCallsScope(c.Then, scopes); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.linkMixinsCallsScope(itm.Default.Then, scopes); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.linkMixinsCallsScope(itm.Body, scopes); err != nil {
				return err
			}
		case file.While:
			if err := l.linkMixinsCallsScope(itm.Body, scopes); err != nil {
				return err
			}
		case file.Mixin:
			if err := l.linkMixinsCallsScope(itm.Body, scopes); err != nil {
				return err
			}
		case file.MixinCall:
			m, err := l.resolveMixinCall(&itm, scopes.Clone())
			if err != nil {
				return err
			}

			itm.Mixin = m
			s[i] = itm

			if err = l.checkMixinParams(*m, itm); err != nil {
				return err
			}

			if err := l.checkMixinBlocks(itm); err != nil {
				return err
			}

			if err = l.linkMixinsCallsScope(itm.Body, scopes); err != nil {
				return err
			}

			if err := l.checkMixinContent(itm); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Linker) resolveMixinCall(c *file.MixinCall, scopes stack.Stack[file.Scope]) (*file.Mixin, error) {
	if c.Namespace != "" && unicode.IsLower(rune(c.Name[0])) {
		return nil, &UnexportedExternalMixinError{
			Source:    l.f.Source,
			File:      l.f.Name,
			Line:      c.Line,
			Col:       c.Col,
			Namespace: string(c.Namespace),
			Name:      string(c.Name),
		}
	}

	if c.Namespace == "" {
		for scopes.Len() > 0 {
			s := scopes.Pop()

			for _, itm := range s {
				mixin, ok := itm.(file.Mixin)
				if !ok {
					continue
				}

				if mixin.Name == c.Name {
					return &mixin, nil
				}
			}
		}
	}

	for _, rfile := range l.rFiles {
		for _, itm := range rfile.Scope {
			mixin, ok := itm.(file.Mixin)
			if !ok {
				continue
			}

			if mixin.Name == c.Name {
				return &mixin, nil
			}
		}
	}

	if unicode.IsLower(rune(c.Name[0])) {
		return nil, &MixinNotFoundError{
			Source:    l.f.Source,
			File:      l.f.Name,
			Line:      c.Line,
			Col:       c.Col,
			Namespace: string(c.Namespace),
			Name:      string(c.Name),
		}
	}

Uses:
	for _, use := range l.f.Uses {
		switch {
		case use.Namespace == "." && c.Namespace == "":
		case use.Namespace == c.Namespace:
		case use.Namespace == "_":
			continue Uses
		default:
			continue Uses
		}

		for _, uf := range use.Files {
			for _, itm := range uf.Scope {
				mixin, ok := itm.(file.Mixin)
				if !ok {
					continue
				}

				if mixin.Name == c.Name {
					return &mixin, nil
				}
			}
		}
	}

	return nil, &MixinNotFoundError{
		Source:    l.f.Source,
		File:      l.f.Name,
		Line:      c.Line,
		Col:       c.Col,
		Namespace: string(c.Namespace),
		Name:      string(c.Name),
	}
}

func (l *Linker) checkMixinParams(m file.Mixin, c file.MixinCall) error {
	args := make(map[string]filePos, len(c.Args))

Args:
	for _, arg := range c.Args {
		first, ok := args[string(arg.Name)]
		if ok {
			return &DuplicateParamError{
				Source:    l.f.Source,
				File:      l.f.Name,
				Line:      first.line,
				Col:       first.col,
				OtherLine: arg.Line,
				OtherCol:  arg.Col,
				Name:      string(arg.Name),
			}
		}

		args[string(arg.Name)] = filePos{
			line: arg.Line,
			col:  arg.Col,
		}

		for _, param := range m.Params {
			if param.Name == arg.Name {
				continue Args
			}
		}

		return &UnknownParamError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   arg.Line,
			Col:    arg.Col,
			Name:   string(arg.Name),
		}
	}

	// check if required args are filled

Params:
	for _, param := range m.Params {
		if param.Default != nil {
			continue
		}

		for _, arg := range c.Args {
			if arg.Name == param.Name {
				nilCheckArg, ok := arg.Value.(file.NilCheckExpression)
				if !ok {
					continue Params
				}

				if nilCheckArg.Default == nil {
					continue Params
				}
			}
		}

		return &MissingParamError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   c.Line,
			Col:    c.Col,
			Name:   string(param.Name),
		}
	}

	return nil
}

func (l *Linker) checkMixinBlocks(c file.MixinCall) error {
	blocks := make(map[string]filePos, len(c.Body))

	for _, itm := range c.Body {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		pos, ok := blocks[string(block.Name)]
		if ok {
			return &DuplicateBlockError{
				Source:    l.f.Source,
				File:      l.f.Name,
				Line:      pos.line,
				Col:       pos.col,
				OtherLine: block.Line,
				OtherCol:  block.Col,
			}
		}
	}

	return nil
}

func (l *Linker) checkMixinContent(c file.MixinCall) error {
	if err := l.checkMixinContentScope(c.Body, c.Line, c.Col); err != nil {
		return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, c.Line, c.Col)
	}

	return nil
}

func (l *Linker) checkMixinContentScope(s file.Scope, line, col int) error {
	for _, itm := range s {
		if err := l.checkMixinContentScopeItem(itm, line, col); err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkMixinContentScopeItem(itm file.ScopeItem, line, col int) error {
	switch itm := itm.(type) {
	case file.Element, file.Text, file.Interpolation, file.InlineElement,
		file.InlineText, file.Filter, file.Comment:
		return &MixinContentError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   line,
			Col:    col,
		}

	case file.And:
		return nil

	case file.Include:
		ci, ok := itm.Include.(file.CorgiInclude)
		if !ok {
			return &MixinContentError{
				Source: l.f.Source,
				File:   l.f.Name,
				Line:   line,
				Col:    col,
			}
		}

		if err := l.checkMixinContentScope(ci.File.Scope, line, col); err != nil {
			return err
		}
	case file.If:
		if err := l.checkMixinContentScope(itm.Then, line, col); err != nil {
			return err
		}

		for _, ei := range itm.ElseIfs {
			if err := l.checkMixinContentScope(ei.Then, line, col); err != nil {
				return err
			}
		}

		if itm.Else != nil {
			if err := l.checkMixinContentScope(itm.Else.Then, line, col); err != nil {
				return err
			}
		}
	case file.IfBlock:
		if err := l.checkMixinContentScope(itm.Then, line, col); err != nil {
			return err
		}

		if itm.Else != nil {
			if err := l.checkMixinContentScope(itm.Else.Then, line, col); err != nil {
				return err
			}
		}
	case file.Switch:
		for _, c := range itm.Cases {
			if err := l.checkMixinContentScope(c.Then, line, col); err != nil {
				return err
			}
		}

		if itm.Default != nil {
			if err := l.checkMixinContentScope(itm.Default.Then, line, col); err != nil {
				return err
			}
		}
	case file.For:
		if err := l.checkMixinContentScope(itm.Body, line, col); err != nil {
			return err
		}
	case file.While:
		if err := l.checkMixinContentScope(itm.Body, line, col); err != nil {
			return err
		}
	case file.MixinCall:
		if err := l.checkMixinContentMixinCall(itm); err != nil {
			return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, itm.Line, itm.Col)
		}
	}

	return nil
}

func (l *Linker) checkMixinContentMixinCall(m file.MixinCall) error {
	for _, itm := range m.Mixin.Body {
		block, ok := itm.(file.Block)
		if !ok {
			if err := l.checkMixinContentScopeItem(itm, m.Mixin.Line, m.Mixin.Col); err != nil {
				return err
			}

			continue
		}

		if err := l.checkMixinContentMixinCallBlock(block, m); err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkMixinContentMixinCallBlock(block file.Block, m file.MixinCall) error {
	for _, itm := range m.Body {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name == block.Name {
			return l.checkMixinContentScope(filledBlock.Body, m.Line, m.Col)
		}
	}

	if len(block.Body) > 0 {
		return l.checkMixinContentScope(block.Body, m.Line, m.Col)
	}

	// empty
	return nil
}
