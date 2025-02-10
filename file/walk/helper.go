package walk

import (
	"github.com/mavolin/corgi/v2/file/ast"
)

// IsTopLevel indicates whether an item with the passed parents would be
// top-level in its scope.
//
// That means, it would not be nested in another element, but may very
// well be nested in a conditional or component.
//
// Always returns false for items inside component calls.
func IsTopLevel(ctx *Context) bool {
	for _, p := range ctx.Parents {
		if !isTopLevel(p.Node) {
			return false
		}
	}

	return true
}

// ChildIsTopLevel indicates whether a child of current would be top-level in
// its scope.
//
// It works under the same rules as [IsTopLevel].
func ChildIsTopLevel(ctx *Context) bool {
	return IsTopLevel(ctx) && isTopLevel(ctx.Node)
}

func isTopLevel(n ast.Node) bool {
	switch n.(type) {
	case *ast.ComponentCall:
	case *ast.Element:
	default:
		return true
	}
	return false
}

// Closest returns the closest parent of the passed type, or the zero value for
// T.
func Closest[T any](ctx *Context) T {
	for i := len(ctx.Parents) - 1; i >= 0; i-- {
		if t, ok := ctx.Parents[i].Node.(T); ok {
			return t
		}
	}

	var z T
	return z
}
