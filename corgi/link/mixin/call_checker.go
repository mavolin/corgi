package mixin

import "github.com/mavolin/corgi/corgi/file"

// CallChecker checks that mixin calls are in order.
//
// It may only be run after mixin calls have been linked.
type CallChecker struct {
	f file.File
}

func NewCallChecker(f file.File) *CallChecker {
	return &CallChecker{f: f}
}

// Check checks that mixin calls ...
//
// 1. don't specify any arguments twice.
//
// 2. specify all required arguments and no inexistent arguments.
//
// 3. only use blocks that are defined by the mixin.
//
// 4. define blocks only at the top-level of their body.
//
// 5. don't specify any blocks twice.
//
// 6. besides blocks, only define & attributes in their body.
func (c *CallChecker) Check() error {
	return file.WalkError(c.f.Scope, func(itmPtr *file.ScopeItem) (bool, error) {
		switch itm := (*itmPtr).(type) {
		case file.MixinCall:
			if err := c.checkUnknownParams(itm); err != nil {
				return false, err
			}

			if err := c.checkArgDuplicates(itm); err != nil {
				return false, err
			}

			if err := c.checkRequiredParams(itm); err != nil {
				return false, err
			}

			if err := c.checkBodyItems(itm); err != nil {
				return false, err
			}

			if err := c.checkNestedBlocks(itm); err != nil {
				return false, err
			}

			if err := c.checkDuplicateBlocks(itm); err != nil {
				return false, err
			}

			return false, nil
		default:
			return true, nil
		}
	})
}

// ============================================================================
// Args and Params
// ======================================================================================

func (c *CallChecker) checkUnknownParams(mc file.MixinCall) error {
Args:
	for _, args := range mc.Args {
		for _, param := range mc.Mixin.Params {
			if args.Name == param.Name {
				continue Args
			}
		}

		return &UnknownParamError{
			Name:   string(args.Name),
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   args.Line,
			Col:    args.Col,
		}
	}

	return nil
}

func (c *CallChecker) checkArgDuplicates(mc file.MixinCall) error {
	if len(mc.Args) <= 1 {
		return nil
	}

	for i, arg := range mc.Args[:len(mc.Args)-1] {
		for _, cmp := range mc.Args[i+1:] {
			if arg.Name != cmp.Name {
				continue
			}

			return &DuplicateArgError{
				Name:      string(arg.Name),
				Source:    c.f.Source,
				File:      c.f.Name,
				Line:      arg.Line,
				Col:       arg.Col,
				OtherLine: cmp.Line,
				OtherCol:  cmp.Col,
			}
		}
	}

	return nil
}

func (c *CallChecker) checkRequiredParams(mc file.MixinCall) error {
Params:
	for _, param := range mc.Mixin.Params {
		if param.Default != nil {
			continue
		}

		for _, arg := range mc.Args {
			if arg.Name == param.Name {
				// make sure someone isn't real sneaky and used a nil check
				// without a default
				nce, ok := arg.Value.(*file.NilCheckExpression)
				if !ok {
					continue Params
				}

				if nce.Default == nil {
					return &MissingNilCheckDefaultError{
						Source: c.f.Source,
						File:   c.f.Name,
						Line:   arg.Line,
						Col:    arg.Col,
						Name:   string(param.Name),
					}
				}

				continue Params
			}
		}

		return &MissingParamError{
			Name:   string(param.Name),
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   mc.Line,
			Col:    mc.Col,
		}
	}

	return nil
}

// ============================================================================
// Mixin Call Body
// ======================================================================================

// checkBodyItems checks that the mixin call body only contains blocks, &s, and
// other mixin calls.
func (c *CallChecker) checkBodyItems(mc file.MixinCall) error {
	return file.WalkError(mc.Body, func(itmPtr *file.ScopeItem) (bool, error) {
		switch itm := (*itmPtr).(type) {
		case file.Element, file.Text, file.Interpolation, file.InlineElement,
			file.InlineText, file.Filter, file.Comment, file.Mixin:
			return false, &CallBodyError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   mc.Line,
				Col:    mc.Col,
			}
		case file.Include:
			_, ok := itm.Include.(*file.CorgiInclude)
			if !ok {
				return false, &CallBodyError{
					Source: c.f.Source,
					File:   c.f.Name,
					Line:   mc.Line,
					Col:    mc.Col,
				}
			}

			return true, nil
		case file.MixinCall:
			return false, c.checkBodyMixinCall(mc, itm)
		case file.Block:
			return false, nil
		default:
			return true, nil
		}
	})
}

func (c *CallChecker) checkBodyMixinCall(orig file.MixinCall, mc file.MixinCall) error {
	// make sure that no blocks write content
	err := file.WalkError(mc.Body, func(itmPtr *file.ScopeItem) (bool, error) {
		switch itm := (*itmPtr).(type) {
		case file.If, file.IfBlock, file.Switch, file.For, file.While:
			return true, nil
		case file.Block:
			return true, nil
		case file.MixinCall:
			if err := c.checkBodyMixinCall(mc, itm); err != nil {
				return false, err
			}

			return true, nil
		case file.And:
			return false, nil
		default:
			return false, &CallBodyError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   orig.Line,
				Col:    orig.Col,
			}
		}
	})
	if err != nil {
		return err
	}

	return file.WalkError(mc.Mixin.Body, func(itmPtr *file.ScopeItem) (bool, error) {
		switch itm := (*itmPtr).(type) {
		case file.If, file.IfBlock, file.Switch, file.For, file.While:
			return true, nil
		case file.Block:
			return true, nil
		case file.MixinCall:
			if err := c.checkBodyMixinCall(mc, itm); err != nil {
				return false, err
			}

			return true, nil
		case file.And:
			return false, nil
		default:
			return false, &CallBodyError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   orig.Line,
				Col:    orig.Col,
			}
		}
	})
}

// checkNestedBlocks checks that blocks are only at the top-level of the mixin
// call body.
func (c *CallChecker) checkNestedBlocks(mc file.MixinCall) error {
	return c.checkNestedBlocksScope(mc.Body, true)
}

func (c *CallChecker) checkNestedBlocksScope(s file.Scope, topLevel bool) error {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if !topLevel {
				return &NestedBlockError{
					Source: c.f.Source,
					File:   c.f.Name,
					Line:   itm.Line,
					Col:    itm.Col,
				}
			}
		case file.Element:
			if err := c.checkNestedBlocksScope(itm.Body, false); err != nil {
				return err
			}
		case file.Include:
			corgiIncl, ok := itm.Include.(file.CorgiInclude)
			if ok {
				err := c.checkNestedBlocksScope(corgiIncl.File.Scope, false)
				if err != nil {
					return err
				}
			}
		case file.If:
			err := c.checkNestedBlocksScope(itm.Then, false)
			if err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				err := c.checkNestedBlocksScope(ei.Then, false)
				if err != nil {
					return err
				}
			}

			if itm.Else != nil {
				err := c.checkNestedBlocksScope(itm.Else.Then, false)
				if err != nil {
					return err
				}
			}
		case file.Switch:
			for _, case_ := range itm.Cases {
				err := c.checkNestedBlocksScope(case_.Then, false)
				if err != nil {
					return err
				}
			}

			if itm.Default != nil {
				err := c.checkNestedBlocksScope(itm.Default.Then, false)
				if err != nil {
					return err
				}
			}
		case file.For:
			err := c.checkNestedBlocksScope(itm.Body, false)
			if err != nil {
				return err
			}
		case file.While:
			err := c.checkNestedBlocksScope(itm.Body, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *CallChecker) checkDuplicateBlocks(mc file.MixinCall) error {
	if len(mc.Body) <= 1 {
		return nil
	}

	for i, itm := range mc.Body[:len(mc.Body)-1] {
		b, ok := itm.(file.Block)
		if !ok {
			continue
		}

		for _, cmpItm := range mc.Body[i+1:] {
			cmp, ok := cmpItm.(file.Block)
			if !ok {
				continue
			}

			if b.Name == cmp.Name {
				return &DuplicateBlockError{
					Name:      string(b.Name),
					Source:    c.f.Source,
					File:      c.f.Name,
					Line:      b.Line,
					Col:       b.Col,
					OtherLine: cmp.Line,
					OtherCol:  cmp.Col,
				}
			}
		}
	}

	return nil
}
