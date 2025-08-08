// Package walk provides utilities for walking through a corgi AST.
package walk

import (
	"fmt"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
)

// Action is the action to take after calling a [Func].
//
// Actions should only be referred to by their constants.
// Their numerical values are not guaranteed to be stable across versions.
type Action uint8

const (
	Continue   Action = 0b0    // Continue walking the node's children.
	Break             = 0b1    // Break the walk, i.e. stop walking the node's children.
	singlesEnd        = Break  // end of non-combined actions
	NoDive            = 1 << 1 // Don't dive the node's body
	// Ignore is an Action available to [Option] functions to signal that Walk
	// should not call the [Func] for the current node.
	//
	// Unless combined (i.e. or-ed) with [NoDive], it will still be diven into.
	//
	// It has no effect if returned by a [Func].
	Ignore  = 1 << 2
	invalid = 1 << 3
)

func (a Action) noDive() bool {
	return a&NoDive != 0
}

func (a Action) ignore() bool {
	return a&Ignore != 0
}

func (a Action) valid() bool {
	return a <= invalid && (a <= singlesEnd || (a&singlesEnd) == 0)
}

type (
	// Func is the function called by Walk for each node it encounters.
	//
	// It must not take ownership of the Context or any of its fields except
	// Node, as Walk may reuse Context and its fields.
	Func func(w *Context) Action
	// FuncT is to [WalkT], as [Func] is to [Walk].
	// Read the documentation of [Func] for more information.
	FuncT[T ast.Node] func(w *ContextT[T]) Action

	Context struct {
		// Parents are the parents of this node.
		//
		// Functions must not take ownership after the function returns, or
		// alter it in any way, as Walk may reuse and alter it.
		Parents []*Context

		Node ast.Node

		// Comments are the corgi comments preceding the node.
		//
		// Functions must not take ownership after the function returns, or
		// alter it in any way, as Walk may reuse and alter it.
		Comments []*ast.Comment
		// CommentDirectives are the comment directives that apply to the node.
		// It accounts for inherited directives from parents.
		//
		// Functions must not take ownership after the function returns, or
		// alter it in any way, as Walk may reuse and alter it.
		CommentDirectives []file.CommentDirective
	}
	ContextT[T ast.Node] struct {
		Node T
		*Context
	}
)

// Walk walks the passed [ast.Node] in depth-first order, calling f first for
// the node itself, then for each of its children.
//
// When walking an [ast.ElseIf] or [ast.Else], the context will contain the
// [ast.If] it belongs to.
// Cases, being children of a switch, can access the switch through the parents
// slice.
//
// If a node has a body and f doesn't return [NoDive], Walk will dive into it,
// walking it as well.
//
// Walk calls f with a slice of ctx.Node's parents.
// That slice is reused for each call to f, and should not be retained after f
// returns.
//
// You may return [Break] from f to stop the walk.
func Walk(n ast.Node, f Func, opts ...Option) {
	ctx := &Context{
		Parents: make([]*Context, 0, 32),
		Node:    n,
	}
	walk(ctx, f, opts)
}

func walk(w *Context, f Func, opts []Option) (cont bool) {
	var combined Action

	for _, opt := range opts {
		a := opt(w)
		switch {
		case !a.valid():
			panic(fmt.Sprintf("invalid action returned by option: %b", a))
		case a == Break:
			return false
		default:
			combined |= a
		}
		if combined.noDive() && combined.ignore() {
			return true
		}
	}

	if !combined.ignore() {
		a := f(w)
		switch {
		case !a.valid():
			panic(fmt.Sprintf("invalid action returned by function: %b", a))
		case a.ignore():
			panic("function returned Ignore, reserved for options")
		case a == Break:
			return false
		case a.noDive():
			combined |= NoDive
		}
	}
	if !combined.noDive() {
		parents := append(w.Parents, w) //nolint:gocritic

		cont = true
		w.Node.Walk(func(n ast.Node) {
			if !cont {
				return
			}

			ctx := &Context{Parents: parents, Node: n}
			cont = walk(ctx, f, opts)
		})
		if !cont {
			return false
		}
	}

	return true
}

// WalkT is the same as [Walk] but only calls f for nodes of type T.
//
//goland:noinspection GoNameStartsWithPackageName
func WalkT[T ast.Node](n ast.Node, f FuncT[T], opts ...Option) { //nolint:revive
	Walk(n, func(w *Context) Action {
		t, ok := w.Node.(T)
		if !ok {
			return Continue
		}

		return f(&ContextT[T]{Node: t, Context: w})
	}, opts...)
}
