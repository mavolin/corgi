package walk

import (
	"fmt"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
)

// todo: test

// An Option is a function that can influence how/if a node is walked.
//
// If at least one Option returns [Ignore], the item will be ignored, i.e. the
// walk function with that option will not be called.
//
// If at least one Option returns [NoDive], the item will not be dived into,
// even if the walk function returns true.
//
// Note that an Option is called on every item, even if the walk function
// is typed.
type Option func(*Context) error

// TopLevel prevents the function from diving into:
//   - Elements
//   - Component call block setters where no instance of the
//     referenced block top-level.
//
// Requires component calls to be analyzed and therefore, transitively, the
// file's symbols to be built.
//
// Ignores withs not linked to a block.
func TopLevel(f *file.File) Option {
	return func(ctx *Context) error {
		switch n := ctx.Node.(type) {
		case *ast.Element:
			return Skip
		case *ast.With:
			astCC := Closest[*ast.ComponentCall](ctx.Parents)
			cc := f.ComponentCallByNode(astCC)
			if cc == nil {
				panic(fmt.Sprintf("walk.TopLevel called without building symbols: %s:%s: file.ComponentCall not found for ast node", f.ModulePath(), cc.AST.Start()))
			}
			blockSetter := cc.BlockSetterByNode(n)
			if blockSetter == nil {
				panic(fmt.Sprintf("walk.TopLevel called without analyzing component calls: %s:%s: file.BlockSetter not found for ast node", f.ModulePath(), n.Start()))
			}
			if blockSetter.Block == nil || blockSetter.Block.TopLevel(file.AtLeastOne).Equal(false) {
				return Skip
			}
			return nil
		}
		return nil
	}
}

func DontDive[N ast.Node]() Option {
	return func(ctx *Context) error {
		if _, ok := ctx.Node.(N); ok {
			return NoDive
		}
		return nil
	}
}
