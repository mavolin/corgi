package file

import (
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

// Block provides information about a block used in a
// Component.
type Block struct {
	//
	// BUILD SYMBOLS

	// Component is the component this block belongs to.
	Component *Component

	// Name is the name of the block.
	Name Identifier

	Instances []*BlockInstance
}

func (b *Block) InstanceByNode(n *ast.Block) *BlockInstance {
	for _, instance := range b.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

// Required indicates whether this block is required to be set by component
// calls.
func (b *Block) Required() (a Analysis[bool]) {
	a.SetResult(b.Instances[0].AST.Default == nil)
	for _, instance := range b.Instances[1:] {
		if a.Equal(true) && instance.AST.Default != nil {
			a.SetFailed()
			return a
		} else if a.Equal(false) && instance.AST.Default == nil {
			a.SetFailed()
			return a
		}
	}
	return a
}

// Forwarded indicates at least one instance of this block is placed
// outside any element.
//
// The reason is that block instance.
func (b *Block) Forwarded() (a AnalysisWithReason[*BlockInstance]) {
	a.SetFalse()
	for _, instance := range b.Instances {
		if instance.Forwarded.Failed() {
			a.SetFailed()
			return a
		} else if instance.Forwarded.Equal(true) {
			a.SetReason(instance)
			return a
		}
	}
	return a
}

// ForwardsAttributes indicates that all instances of this block can forward
// attributes.
func (b *Block) ForwardsAttributes() (a Analysis[bool]) {
	cannotForwardAttrs := b.CannotForwardAttributes()
	if cannotForwardAttrs.Failed() {
		a.SetFailed()
	} else {
		a.SetResult(cannotForwardAttrs.False())
	}
	return a
}

// CannotForwardAttributes indicates that at least one instance of this
// block cannot forward attributes.
//
// The reason is the first instance that cannot forward attributes.
func (b *Block) CannotForwardAttributes() (a AnalysisWithReason[*BlockInstance]) {
	a.SetFalse()
	for _, instance := range b.Instances {
		if instance.CannotForwardAttributes.Failed() {
			a.SetFailed()
			return a
		} else if instance.CannotForwardAttributes.True() {
			a.SetReason(instance)
			return a
		}
	}
	return a
}

// ElementType is the element type that this block assumes.
//
// If the block instance is fully forwarded, i.e. has no containing
// elements, ElementType is set to Normal.
//
// For all other element types it is the minimum of all containing
// elements.
func (b *Block) ElementType() (a Analysis[elemtype.Type]) {
	t := elemtype.Normal
	for _, instance := range b.Instances {
		if instance.ContainingElementSpecs.Failed() {
			a.SetFailed()
			return a
		}

		for _, spec := range instance.ContainingElementSpecs.Result().Get() {
			if spec.Type.Failed() {
				a.SetFailed()
				return a
			}

			specType := spec.Type.Result()
			switch specType {
			case elemtype.Unknown:
				a.SetFailed()
				return a
			case elemtype.JS, elemtype.CSS:
				a.SetFailed()
				return a
			case elemtype.Void:
				specType = elemtype.Nothing
			case elemtype.Nothing, elemtype.Normal, elemtype.Text:
			}
			t = min(t, specType)
		}
	}
	a.SetResult(t)
	return a
}

// MostRestrictiveElement is one (of the possibly multiple) element with the
// most restrictive (smallest) element type.
//
// It follows the same rules as [ElementType].
func (b *Block) MostRestrictiveElement() (a AnalysisWithReason[ast.ContainingElement]) {
	f := b.Component.File

	var resultCEl ast.ContainingElement
	var resultTyp elemtype.Type
	for _, instance := range b.Instances {
		if instance.ContainingElements.Failed() {
			a.SetFailed()
			return a
		}

		for _, cEl := range instance.ContainingElements.Result().Get() {
			switch cEl := cEl.(type) {
			case *ast.BlockSetterContainingElement:
				cc := f.ComponentCallByNode(cEl.ComponentCall)
				s := cc.BlockSetterByNode(cEl.BlockSetter)
				if s == nil || s.Block == nil {
					a.SetFailed()
					return a
				}

				for _, instance := range s.Block.Instances {
					if instance.ContainingElementSpecs.Failed() {
						a.SetFailed()
						return a
					}
					for _, spec := range instance.ContainingElementSpecs.Result().Get() {
						if spec.Type.Failed() {
							a.SetFailed()
							return a
						}

						if resultCEl == nil || spec.Type.Result() < resultTyp {
							resultCEl = cEl
							resultTyp = spec.Type.Result()
						}
					}
				}
			case *ast.Element:
				ref := f.ElementReferenceByNode(cEl.Header.Name)
				if ref.Spec == nil {
					a.SetFailed()
					return a
				}

				// Always prefer elements to block setters, so use <=
				if resultCEl == nil || ref.Spec.Type.Result() <= resultTyp {
					resultCEl = cEl
					resultTyp = ref.Spec.Type.Result()
				}
			default:
				panic("unknown containing element type")
			}

			switch resultTyp {
			case elemtype.Unknown:
				a.SetFailed()
				return a
			case elemtype.JS, elemtype.CSS:
				a.SetFailed()
				return a
			case elemtype.Void:
				resultTyp = elemtype.Nothing
			case elemtype.Nothing, elemtype.Normal, elemtype.Text:
			}
		}
	}

	a.SetReason(resultCEl)
	return a
}
