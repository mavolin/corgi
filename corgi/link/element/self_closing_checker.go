package element

import (
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/internal/stack"
	"github.com/mavolin/corgi/internal/voidelem"
)

// SelfClosingChecker checks that self-closing and void elements have no body.
type SelfClosingChecker struct {
	f file.File

	mixinCalls stack.Stack[file.MixinCall]
}

// NewSelfClosingChecker creates a new SelfClosingChecker that checks the
// passed file.
func NewSelfClosingChecker(f file.File) *SelfClosingChecker {
	return &SelfClosingChecker{f: f, mixinCalls: stack.New[file.MixinCall](50)}
}

func (c *SelfClosingChecker) Check() error {
	return c.checkScope(c.f.Scope)
}

func (c *SelfClosingChecker) checkScope(s file.Scope) error {
	return file.WalkError(s, func(itm *file.ScopeItem) (bool, error) {
		switch itm := (*itm).(type) {
		case file.Element:
			return false, c.checkElement(itm)
		default:
			return true, nil
		}
	})
}

func (c *SelfClosingChecker) checkElement(e file.Element) error {
	isVoidElem := (c.f.Type == file.TypeHTML || c.f.Type == file.TypeXHTML) &&
		voidelem.Is(e.Name)

	if !e.SelfClosing && !isVoidElem {
		return c.checkScope(e.Body)
	}

	return c.checkElementScope(e, e.Body)
}

// checkElementScope checks that the passed file.Scope does not fill the body
// of the passed file.Element.
//
// e must be a void element or self-closing.
func (c *SelfClosingChecker) checkElementScope(e file.Element, s file.Scope) error {
	return file.WalkError(s, func(itm *file.ScopeItem) (bool, error) {
		switch itm := (*itm).(type) {
		case file.Element, file.Text, file.Interpolation, file.InlineElement,
			file.InlineText, file.Filter, file.Comment:
			return false, &SelfClosingBodyError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   e.Line,
				Col:    e.Col,
			}
		case file.Include:
			_, isRaw := itm.Include.(file.RawInclude)
			if isRaw {
				return false, &SelfClosingBodyError{
					Source: c.f.Source,
					File:   c.f.Name,
					Line:   e.Line,
					Col:    e.Col,
				}
			}

			return true, nil
		case file.MixinCall:
			return false, c.checkMixinCall(e, itm)
		case file.Block:
			return false, c.checkBlock(e, itm)
		case file.Mixin:
			return false, nil
		default:
			return true, nil
		}
	})
}

func (c *SelfClosingChecker) checkMixinCall(e file.Element, mc file.MixinCall) error {
	c.mixinCalls.Push(mc)
	defer c.mixinCalls.Pop()

	return c.checkElementScope(e, mc.Mixin.Body)
}

func (c *SelfClosingChecker) checkBlock(e file.Element, b file.Block) error {
	if c.mixinCalls.Len() == 0 {
		// We're not in a mixin call.
		// This means this block belongs either to a main, use, or to an extend
		// file.
		// Regardless of which, they may not be placed inside self-closing or
		// void elements.

		return &SelfClosingBodyError{
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   e.Line,
			Col:    e.Col,
		}
	}

	// check that the default for b is also valid, even if we don't use it
	if err := c.checkElementScope(e, b.Body); err != nil {
		return err
	}

	mc := c.mixinCalls.Peek()

	// find the content of the block
	for _, itm := range mc.Body {
		filledBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if filledBlock.Name == b.Name {
			return c.checkElementScope(e, filledBlock.Body)
		}
	}

	// The block has not been filled and therefore no content can be written to
	// e.
	return nil
}
