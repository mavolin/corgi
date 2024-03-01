package walk

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
)

// IsTopLevel indicates whether an item with the passed parents would be
// top-level in its scope.
//
// That means, it would not be nested in another element, but may very
// well be nested in a conditional or component.
//
// When judging blocks inside component calls, IsTopLevel relies on the File's
// link information.
// If the block is not linked, IsTopLevel always returns false.
func IsTopLevel(parents []Context) bool {
	p, _ := childIsTopLevel(parents, Context{})
	return p
}

// ChildIsTopLevel indicates whether a child of current would be top-level in
// its scope.
//
// It works under the same rules as [IsTopLevel].
func ChildIsTopLevel(parents []Context, current Context) bool {
	_, t := childIsTopLevel(parents, current)
	return t
}

func childIsTopLevel(parents []Context, current Context) (bool, bool) {
	var comp *file.Component

parents:
	for _, p := range parents {
		switch itm := p.Node.(type) {
		case *ast.If:
		case *ast.For:
		case *ast.Switch:
		case *ast.ComponentCall:
			if p.File == nil {
				return false, false
			}
			cc := p.File.Package.ComponentCallByPtr(itm)
			if cc == nil {
				return false, false
			}
			comp = cc.Component
		case *ast.Block:
			if comp != nil {
				for _, b := range comp.Blocks {
					if b.Name == itm.Name.Ident {
						if !b.TopLevel {
							return false, false
						}
						break parents
					}
				}
				return false, false
			}
		default:
			return false, false
		}
	}

	if current.Node == nil {
		return true, false
	}

	switch itm := current.Node.(type) {
	case *ast.If:
	case *ast.For:
	case *ast.Switch:
	case *ast.ComponentCall:
		if comp == nil {
			return true, false
		}
	case *ast.Block:
		if comp != nil {
			for _, b := range comp.Blocks {
				if b.Name == itm.Name.Ident {
					return true, b.TopLevel
				}
			}
			return true, false
		}
	default:
		return true, false
	}

	return true, true
}

// Closest returns the closest parent of the passed type, or the zero value for
// T.
func Closest[T any](parents []Context) T {
	t, _ := ClosestItem[T](parents)
	return t
}

func ClosestItem[T any](parents []Context) (T, int) {
	for i := len(parents) - 1; i >= 0; i-- {
		if t, ok := parents[i].Node.(T); ok {
			return t, i
		}
	}

	var zero T
	return zero, -1
}
