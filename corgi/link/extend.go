package link

import (
	"github.com/mavolin/corgi/corgi/file"
)

func (l *Linker) checkExtendBlocks() error {
	if l.f.Extend == nil {
		return nil
	}

	for _, itm := range l.f.Scope {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if err := l.checkBlockScope(block.Body); err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkBlockScope(s file.Scope) error {
	for _, itm := range s {
		err := l.checkBlockScopeItem(itm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkBlockScopeItem(itm file.ScopeItem) error {
	switch itm := itm.(type) {
	case file.And:
		return &IllegalAndError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   itm.Line,
			Col:    itm.Col,
		}
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
		if err := l.checkBlockScope(itm.Body); err != nil {
			return err
		}
	case file.If:
		if err := l.checkBlockScope(itm.Then); err != nil {
			return err
		}

		for _, ei := range itm.ElseIfs {
			if err := l.checkBlockScope(ei.Then); err != nil {
				return err
			}
		}

		if itm.Else != nil {
			if err := l.checkBlockScope(itm.Else.Then); err != nil {
				return err
			}
		}
	case file.IfBlock:
		if err := l.checkBlockScope(itm.Then); err != nil {
			return err
		}

		if itm.Else != nil {
			if err := l.checkBlockScope(itm.Else.Then); err != nil {
				return err
			}
		}
	case file.Switch:
		for _, c := range itm.Cases {
			if err := l.checkBlockScope(c.Then); err != nil {
				return err
			}
		}

		if itm.Default != nil {
			if err := l.checkBlockScope(itm.Default.Then); err != nil {
				return err
			}
		}
	case file.For:
		if err := l.checkBlockScope(itm.Body); err != nil {
			return err
		}
	case file.While:
		if err := l.checkBlockScope(itm.Body); err != nil {
			return err
		}
	case file.MixinCall:
		if err := l.checkBlockMixinCall(itm); err != nil {
			return err
		}
	}

	return nil
}

func (l *Linker) checkBlockMixinCall(m file.MixinCall) error {
	for _, itm := range m.Mixin.Body {
		block, ok := itm.(file.Block)
		if !ok {
			if err := l.checkBlockScopeItem(itm); err != nil {
				return err
			}

			continue
		}

		if err := l.checkBlockMixinCallBlock(block, m); err != nil {
			return err
		}
	}

	return nil
}

// checkAndsMixinCallBlock checks if the block with the given name allows ands.
//
// For that it employs the same algorithm as checkAnds for blocks that
// are static, i.e. not dependent on a condition.
func (l *Linker) checkBlockMixinCallBlock(block file.Block, m file.MixinCall) error {
	for _, itm := range m.Body {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name == block.Name {
			return l.checkBlockScope(filledBlock.Body)
		}
	}

	if len(block.Body) > 0 {
		return l.checkBlockScope(block.Body)
	}

	// empty
	return nil
}
