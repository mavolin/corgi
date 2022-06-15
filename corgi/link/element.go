package link

import (
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/pkg/voidelem"
)

func (l *Linker) checkElements() error {
	return l.checkElementsScope(l.f.Scope)
}

func (l *Linker) checkElementsScope(s file.Scope) error {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Include:
			ci, ok := itm.Include.(file.CorgiInclude)
			if !ok {
				break
			}

			for _, imp := range ci.File.Imports {
				if err := l.addImport(imp); err != nil {
					return err
				}
			}
		case file.Block:
			if err := l.checkElementsScope(itm.Body); err != nil {
				return err
			}
		case file.Element:
			if _, err := l.checkAnds(itm.Body, true); err != nil {
				return err
			}

			if itm.SelfClosing || (l.f.Type == file.TypeHTML && voidelem.Is(itm.Name)) {
				if err := l.checkSelfClosingElement(itm, itm.Body); err != nil {
					return err
				}
			}
		case file.If:
			if err := l.checkElementsScope(itm.Then); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.checkElementsScope(ei.Then); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.checkElementsScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.checkElementsScope(itm.Then); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.checkElementsScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.checkElementsScope(c.Then); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.checkElementsScope(itm.Default.Then); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.checkElementsScope(itm.Body); err != nil {
				return err
			}
		case file.While:
			if err := l.checkElementsScope(itm.Body); err != nil {
				return err
			}
		case file.Mixin:
			if err := l.checkElementsScope(itm.Body); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.checkElementsScope(itm.Body); err != nil {
				return err
			}
		}
	}

	return nil
}

// ============================================================================
// Ands
// ======================================================================================

func (l *Linker) checkAnds(s file.Scope, andAllowed bool) (bool, error) {
	for _, itm := range s {
		after, err := l.checkAndsScopeItem(itm, andAllowed)
		if err != nil {
			return false, err
		}

		if !after {
			andAllowed = false
		}
	}

	return andAllowed, nil
}

func (l *Linker) checkAndsScopeItem(itm file.ScopeItem, andAllowed bool) (bool, error) {
	switch itm := itm.(type) {
	case file.Element:
		if _, err := l.checkAnds(itm.Body, true); err != nil {
			return false, err
		}

		return false, nil
	case file.And:
		if !andAllowed {
			return false, &IllegalAndError{
				Source: l.f.Source,
				File:   l.f.Name,
				Line:   itm.Line,
				Col:    itm.Col,
			}
		}

		return true, nil
	case file.Text:
		return false, nil
	case file.Interpolation:
		return false, nil
	case file.InlineElement:
		return false, nil
	case file.InlineText:
		return false, nil
	case file.Filter:
		return false, nil
	case file.Comment:
		return false, nil
	case file.Block:
		// we could check if the resulting block has allows &s after it,
		// however, it's not worth building the checking logic, since I
		// couldn't imagine much use for an & after an extends block
		return false, nil

	case file.Include:
		ci, ok := itm.Include.(file.CorgiInclude)
		if !ok {
			return false, nil
		}

		afterInclude, err := l.checkAnds(ci.File.Scope, andAllowed)
		if err != nil {
			return false, err
		}

		if !afterInclude {
			return false, nil
		}
	case file.If:
		afterIf, err := l.checkAnds(itm.Then, andAllowed)
		if err != nil {
			return false, nil
		}

		if !afterIf {
			return false, nil
		}

		for _, ei := range itm.ElseIfs {
			afterElseIf, err := l.checkAnds(ei.Then, andAllowed)
			if err != nil {
				return false, err
			}

			if !afterElseIf {
				return false, nil
			}
		}

		if itm.Else != nil {
			afterElse, err := l.checkAnds(itm.Else.Then, andAllowed)
			if err != nil {
				return false, err
			}

			if !afterElse {
				return false, nil
			}
		}
	case file.IfBlock:
		afterIf, err := l.checkAnds(itm.Then, andAllowed)
		if err != nil {
			return false, err
		}

		if !afterIf {
			return false, nil
		}

		if itm.Else != nil {
			afterElse, err := l.checkAnds(itm.Else.Then, andAllowed)
			if err != nil {
				return false, err
			}

			if !afterElse {
				return false, nil
			}
		}
	case file.Switch:
		for _, c := range itm.Cases {
			afterCase, err := l.checkAnds(c.Then, andAllowed)
			if err != nil {
				return false, err
			}

			if !afterCase {
				return false, nil
			}
		}

		if itm.Default != nil {
			afterCase, err := l.checkAnds(itm.Default.Then, andAllowed)
			if err != nil {
				return false, err
			}

			if !afterCase {
				return false, nil
			}
		}
	case file.For:
		afterFor, err := l.checkAnds(itm.Body, andAllowed)
		if err != nil {
			return false, err
		}

		if !afterFor {
			return false, nil
		}
	case file.While:
		afterWhile, err := l.checkAnds(itm.Body, andAllowed)
		if err != nil {
			return false, err
		}

		if !afterWhile {
			return false, nil
		}
	case file.MixinCall:
		afterCall, err := l.checkAndsMixinCall(itm, andAllowed)
		if err != nil {
			return false, err
		}

		if !afterCall {
			return false, nil
		}
	}

	return false, nil
}

func (l *Linker) checkAndsMixinCall(m file.MixinCall, andAllowed bool) (bool, error) {
	for _, itm := range m.Mixin.Body {
		block, ok := itm.(file.Block)
		if !ok {
			after, err := l.checkAndsScopeItem(itm, andAllowed)
			if err != nil {
				return false, err
			}

			if !after {
				return false, nil
			}

			continue
		}

		after, err := l.checkAndsMixinCallBlock(block, m, andAllowed)
		if err != nil {
			return false, err
		}

		if !after {
			return false, nil
		}
	}

	return true, nil
}

// checkAndsMixinCallBlock checks if the block with the given name allows ands.
//
// For that it employs the same algorithm as checkAnds for blocks that
// are static, i.e. not dependent on a condition.
func (l *Linker) checkAndsMixinCallBlock(block file.Block, m file.MixinCall, andAllowed bool) (bool, error) {
	for _, itm := range m.Body {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name == block.Name {
			return l.checkAnds(filledBlock.Body, andAllowed)
		}
	}

	if len(block.Body) > 0 {
		return l.checkAnds(block.Body, andAllowed)
	}

	// empty
	return false, nil
}

// ============================================================================
// Self Closing Elements
// ======================================================================================

func (l *Linker) checkSelfClosingElement(e file.Element, s file.Scope) error {
	for _, itm := range s {
		if err := l.checkSelfClosingElementScopeItem(e, itm); err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkSelfClosingElementScopeItem(e file.Element, itm file.ScopeItem) error {
	switch itm := itm.(type) {
	case file.Element, file.Text, file.Interpolation, file.InlineElement,
		file.InlineText, file.Filter, file.Comment, file.Block:
		return &SelfClosingContentError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   e.Line,
			Col:    e.Col,
		}

	case file.And:
		return nil

	case file.Include:
		ci, ok := itm.Include.(file.CorgiInclude)
		if !ok {
			return &SelfClosingContentError{
				Source: l.f.Source,
				File:   l.f.Name,
				Line:   e.Line,
				Col:    e.Col,
			}
		}

		if err := l.checkSelfClosingElement(e, ci.File.Scope); err != nil {
			return err
		}
	case file.If:
		if err := l.checkSelfClosingElement(e, itm.Then); err != nil {
			return err
		}

		for _, ei := range itm.ElseIfs {
			if err := l.checkSelfClosingElement(e, ei.Then); err != nil {
				return err
			}
		}

		if itm.Else != nil {
			if err := l.checkSelfClosingElement(e, itm.Else.Then); err != nil {
				return err
			}
		}
	case file.IfBlock:
		if err := l.checkSelfClosingElement(e, itm.Then); err != nil {
			return err
		}

		if itm.Else != nil {
			if err := l.checkSelfClosingElement(e, itm.Else.Then); err != nil {
				return err
			}
		}
	case file.Switch:
		for _, c := range itm.Cases {
			if err := l.checkSelfClosingElement(e, c.Then); err != nil {
				return err
			}
		}

		if itm.Default != nil {
			if err := l.checkSelfClosingElement(e, itm.Default.Then); err != nil {
				return err
			}
		}
	case file.For:
		if err := l.checkSelfClosingElement(e, itm.Body); err != nil {
			return err
		}
	case file.While:
		if err := l.checkSelfClosingElement(e, itm.Body); err != nil {
			return err
		}
	case file.MixinCall:
		if err := l.checkSelfClosingElementMixinCall(e, itm); err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkSelfClosingElementMixinCall(e file.Element, m file.MixinCall) error {
	for _, itm := range m.Mixin.Body {
		block, ok := itm.(file.Block)
		if !ok {
			if err := l.checkSelfClosingElementScopeItem(e, itm); err != nil {
				return err
			}

			continue
		}

		if err := l.checkSelfClosingElementMixinCallBlock(e, block, m); err != nil {
			return err
		}
	}

	return nil
}

// checkAndsMixinCallBlock checks if the block with the given name allows ands.
//
// For that it employs the same algorithm as checkAnds for blocks that
// are static, i.e. not dependent on a condition.
func (l *Linker) checkSelfClosingElementMixinCallBlock(e file.Element, block file.Block, m file.MixinCall) error {
	for _, itm := range m.Body {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name == block.Name {
			return l.checkSelfClosingElement(e, filledBlock.Body)
		}
	}

	if len(block.Body) > 0 {
		return l.checkSelfClosingElement(e, block.Body)
	}

	// empty
	return nil
}
