// Package walk provides utilities for walking through a corgi AST.1
package walk

import "github.com/mavolin/corgi/file"

type Context struct {
	// Scope is the scope in which the item was found.
	Scope file.Scope
	// Index is the index of the item.
	Index int
	// Item is a pointer to the item, allowing it to be edited.
	Item *file.ScopeItem

	// Case is the case of [file.Switch] that we are walking.
	Case *file.Case
	// ElseIf is the else if of the [file.If] that we are walking.
	ElseIf *file.ElseIf
	// Else is the [file.Else] of the [file.If] that we are walking.
	Else *file.Else

	// Comments are the corgi comments preceding the item.
	Comments []file.CorgiComment
}

type MachineComment struct {
	Namespace string
	Directive string
	Args      string
}

func (ctx *Context) MachineComments() []MachineComment {
	mcs := make([]MachineComment, 0, len(ctx.Comments))
	for _, c := range ctx.Comments {
		if c.IsMachineComment() {
			var mc MachineComment
			mc.Namespace, mc.Directive, mc.Args = c.MachineComment()
			mcs = append(mcs, mc)
		}
	}

	if len(mcs) == 0 {
		return nil
	}
	return mcs[:len(mcs):len(mcs)]
}
