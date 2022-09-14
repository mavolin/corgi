package element

import (
	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/pkg/stack"
)

// AndChecker checks that &s are used inside elements and only before their
// body has been written.
//
// It must be run after mixins are resolved.
type AndChecker struct {
	f file.File

	// mixinCalls are the mixin calls that are currently being processed.
	mixinCalls stack.Stack[file.MixinCall]
}

// NewAndChecker creates a new AndChecker for the given file.
func NewAndChecker(f file.File) *AndChecker {
	return &AndChecker{f: f}
}

// Check checks the file and reports any errors it encounters.
func (c *AndChecker) Check() error {
	return file.WalkError(c.f.Scope, func(itm *file.ScopeItem) (bool, error) {
		switch itm := (*itm).(type) {
		case file.And:
			return false, &NoAndElementError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   itm.Line,
				Col:    itm.Col,
			}
		case file.Element:
			_, err := c.checkInElement(itm.Body, true)
			return false, err
		case file.MixinCall:
			_, err := c.checkMixinCall(itm, false)
			return false, err
		case file.Mixin:
			// do nothing, mixins are checked when they are called
			return false, nil
		default:
			return true, nil
		}
	})
}

// checkInElement walks the given scope and checks that &s are only written if
// andAllowed and they are placed before the body of the containing element is
// filled.
//
// It must only be called on file.Scopes inside elements.
func (c *AndChecker) checkInElement(s file.Scope, andAllowed bool) (_ bool, err error) {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Element:
			if _, err := c.checkInElement(itm.Body, true); err != nil {
				return false, err
			}

			andAllowed = false
		case file.And:
			if !andAllowed {
				return false, c.andPlacementError(itm)
			}
		case file.Block:
			// Only check if this block belongs to a mixin call.
			// Extend blocks don't allow &.
			andAllowed, err = c.checkBlock(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.Include:
			andAllowed, err = c.checkInclude(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.If:
			andAllowed, err = c.checkIf(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.IfBlock:
			andAllowed, err = c.checkIfBlock(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.Switch:
			andAllowed, err = c.checkSwitch(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.For:
			andAllowed, err = c.checkFor(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.While:
			andAllowed, err = c.checkWhile(itm, andAllowed)
			if err != nil {
				return false, err
			}
		case file.MixinCall:
			andAllowed, err = c.checkMixinCall(itm, andAllowed)
			if err != nil {
				return false, err
			}

		case file.Mixin, file.Code:
			// do nothing
		default:
			andAllowed = false
		}
	}

	return andAllowed, nil
}

func (c *AndChecker) andPlacementError(and file.And) error {
	if c.mixinCalls.Len() == 0 {
		return &AndPlacementError{
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   and.Line,
			Col:    and.Col,
		}
	}

	mcs := c.mixinCalls.Clone()
	mc := mcs.Pop()

	var err error = &AndPlacementError{
		Source: mc.MixinSource,
		File:   mc.MixinFile,
		Line:   and.Line,
		Col:    and.Col,
	}

	for mcs.Len() > 0 {
		mc = mcs.Pop()
		err = errors.Wrapf(err, "%s/%s:%d:%d", mc.MixinSource, mc.MixinFile, mc.Line, mc.Col)
	}

	return errors.Wrapf(err, "%s/%s:%d:%d", c.f.Source, c.f.Name, mc.Line, mc.Col)
}

// ============================================================================
// Checkers for Specific Item Types
// ======================================================================================

func (c *AndChecker) checkBlock(b file.Block, andAllowed bool) (bool, error) {
	if c.mixinCalls.Len() == 0 {
		// We're not in a mixin call.
		// This means this block belongs either to a main or to an extend
		// file, or to a mixin definition.
		// Regardless of which, they may not use top-level &s, neither for
		// defaults nor when actually filling the block.
		if _, err := c.checkInElement(b.Body, false); err != nil {
			return false, err
		}

		return false, nil
	}

	mc := c.mixinCalls.Pop()
	defer c.mixinCalls.Push(mc)

	// check that the default for b is also valid, even if we don't use it
	defaultAndAllowed, err := c.checkInElement(b.Body, andAllowed)
	if err != nil {
		return false, err
	}

	// find the content of the block
	for _, itm := range mc.Body {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name == b.Name {
			return c.checkInElement(filledBlock.Body, andAllowed)
		}
	}

	// The block has not been filled.
	// Return whatever the default allows.
	return defaultAndAllowed, nil
}

func (c *AndChecker) checkInclude(incl file.Include, andAllowed bool) (bool, error) {
	ci, ok := incl.Include.(file.CorgiInclude)
	if !ok {
		return false, nil
	}

	return c.checkInElement(ci.File.Scope, andAllowed)
}

func (c *AndChecker) checkIf(if_ file.If, andAllowed bool) (bool, error) {
	andAfter, err := c.checkInElement(if_.Then, andAllowed)
	if err != nil {
		return false, err
	}

	for _, ei := range if_.ElseIfs {
		andAfterElseIf, err := c.checkInElement(ei.Then, andAllowed)
		if err != nil {
			return false, err
		}

		if !andAfterElseIf {
			andAfter = false
		}
	}

	if if_.Else != nil {
		andAfterElse, err := c.checkInElement(if_.Else.Then, andAllowed)
		if err != nil {
			return false, err
		}

		if !andAfterElse {
			andAfter = false
		}
	}

	return andAfter, nil
}

func (c *AndChecker) checkIfBlock(ifBlock file.IfBlock, andAllowed bool) (bool, error) {
	andAfter, err := c.checkInElement(ifBlock.Then, andAllowed)
	if err != nil {
		return false, err
	}

	if ifBlock.Else != nil {
		andAfterElse, err := c.checkInElement(ifBlock.Else.Then, andAllowed)
		if err != nil {
			return false, err
		}

		if !andAfterElse {
			andAfter = false
		}
	}

	return andAfter, nil
}

func (c *AndChecker) checkSwitch(sw file.Switch, andAllowed bool) (bool, error) {
	andAfter := andAllowed

	for _, case_ := range sw.Cases {
		andAfterCase, err := c.checkInElement(case_.Then, andAllowed)
		if err != nil {
			return false, err
		}

		if !andAfterCase {
			andAfter = false
		}
	}

	if sw.Default != nil {
		andAfterDefault, err := c.checkInElement(sw.Default.Then, andAllowed)
		if err != nil {
			return false, err
		}

		if !andAfterDefault {
			andAfter = false
		}
	}

	return andAfter, nil
}

func (c *AndChecker) checkFor(f file.For, andAllowed bool) (_ bool, err error) {
	if len(f.Body) == 0 {
		return andAllowed, nil
	}

	// If & was allowed before the loop, but not after that means we wrote to
	// the elements body in there somewhere.
	// There is only one edge case to consider:
	// If before the loop &s were allowed, but the loop itself only writes
	// text.
	andAllowedBefore := andAllowed

	// So let's check if the first item is an &.
	firstIsAnd := file.IsFirstAnd(f.Body)

	// Now we just need to check if the loop writes to the body of the element.
	andAllowed, err = c.checkInElement(f.Body, andAllowed)
	if err != nil {
		return false, err
	}

	// check if &s were allowed before, but not after, i.e. if we wrote to the
	// body of the element
	if andAllowedBefore != andAllowed {
		// We did.
		// In this case the first item must not be an &, because if the first
		// isn't, then so can't be the others.
		if firstIsAnd {
			return false, c.loopAndError(f)
		}
	}

	return andAllowed, nil
}

func (c *AndChecker) loopAndError(f file.For) error {
	if c.mixinCalls.Len() == 0 {
		return &LoopAndError{
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   f.Line,
			Col:    f.Col,
		}
	}

	mcs := c.mixinCalls.Clone()
	mc := mcs.Pop()

	var err error = &LoopAndError{
		Source: mc.MixinSource,
		File:   mc.MixinFile,
		Line:   f.Line,
		Col:    f.Col,
	}

	for mcs.Len() > 0 {
		mc = mcs.Pop()
		err = errors.Wrapf(err, "%s/%s:%d:%d", mc.MixinSource, mc.MixinFile, mc.Line, mc.Col)
	}

	return errors.Wrapf(err, "%s/%s:%d:%d", c.f.Source, c.f.Name, mc.Line, mc.Col)
}

func (c *AndChecker) checkWhile(f file.While, andAllowed bool) (_ bool, err error) {
	if len(f.Body) == 0 {
		return andAllowed, nil
	}

	// If & was allowed before the loop, but not after that means we wrote to
	// the elements body in there somewhere.
	// There is only one edge case to consider:
	// If before the loop &s were allowed, but the loop itself only writes
	// text.
	andAllowedBefore := andAllowed

	// So let's check if the first item is an &.
	firstIsAnd := file.IsFirstAnd(f.Body)

	// Now we just need to check if the loop writes to the body of the element.
	andAllowed, err = c.checkInElement(f.Body, andAllowed)
	if err != nil {
		return false, err
	}

	// check if &s were allowed before, but not after, i.e. if we wrote to the
	// body of the element
	if andAllowedBefore != andAllowed {
		// We did.
		// In this case the first item must not be an &, because if the first
		// isn't, then so can't be the others.
		if firstIsAnd {
			return false, &LoopAndError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   f.Line,
				Col:    f.Col,
			}
		}
	}

	return andAllowed, nil
}

func (c *AndChecker) checkMixinCall(mc file.MixinCall, andAllowed bool) (bool, error) {
	c.mixinCalls.Push(mc)
	defer c.mixinCalls.Pop()

	return c.checkInElement(mc.Mixin.Body, andAllowed)
}
