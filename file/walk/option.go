package walk

import (
	"fmt"
	"reflect"

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
//   - Component call withs that are not top-level.
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
				panic(fmt.Sprintf("walk.TopLevel called without building symbols: %s/%s:%s: file.ComponentCall not found for ast node", f.Module, f.PathInModule, cc.AST.Start()))
			}
			with := cc.WithByName(n.Name.Ident)
			if with == nil {
				panic(fmt.Sprintf("walk.TopLevel called without analyzing component calls: %s/%s:%s: file.With not found for ast node", f.Module, f.PathInModule, n.Start()))
			}
			if with.Block == nil || !with.Block.TopLevel(file.AtLeastOne) {
				return Skip
			}
			return nil
		}
		return nil
	}
}

// DontDiveAny prevents the function from diving if the current item is of
// the passed types.
func DontDiveAny(types ...ast.Node) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(wctx *Context) error {
		t := reflect.TypeOf(wctx.Node)
		for _, rType := range rTypes {
			if rType == t {
				return NoDive
			}
		}

		return nil
	}
}
