package walk

import (
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
type Option func(*Context) Action

func DontDive[N ast.Node]() Option {
	return func(ctx *Context) Action {
		if _, ok := ctx.Node.(N); ok {
			return NoDive
		}
		return Continue
	}
}
