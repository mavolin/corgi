// Package walk provides utilities for walking through a corgi AST.
package walk

import (
	"errors"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
)

var (
	// Stop is a sentinel error used to signal that Walk should return without
	// an error.
	Stop = errors.New("stop walk") //nolint:staticcheck,revive,errname
	// NoDive is a sentinel error used to signal that Walk should not dive into
	// the current node's body.
	NoDive = errors.New("no dive") //nolint:staticcheck,revive,errname
	// Ignore is a sentinel error available to [Option] functions to signal
	// that Walk should not call the [Func] for the current node, but it should
	// still dive it, if possible.
	//
	// It has no effect if returned by a [Func].
	//
	// It may still be dived.
	Ignore = errors.New("ignore") //nolint:staticcheck,revive,errname
	// Skip is a sentinel error available to [Option] functions to signal
	// to skip over the current node, i.e. ignore it and don't dive into it.
	//
	// Skip is essentially the combination of [Ignore] and [NoDive].
	//
	// If returned by a [Func], it behaves like [NoDive].
	Skip = errors.New("ignore no dive") //nolint:staticcheck,revive,errname
)

type (
	// Func is the function called by Walk for each node it encounters.
	//
	// It must not take ownership of the Context or any of its fields except
	// Node, as Walk may reuse Context and its fields.
	Func func(*Context) error
	// FuncT is to [WalkT], as [Func] is to [Walk].
	// Read the documentation of [Func] for more information.
	FuncT[T ast.Node] func(*ContextT[T]) error

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
// f will only be called with the containing [ast.TextLine] if that text line
// was embedded in a [ast.TextBlock] and thus inserts newlines.
// This applies namely to arrow blocks and bracket text.
// It will not be called for the body of [ast.ElementInterpolation] and
// [ast.ComponentCallInterpolation].
// Instead, if diven, the TextNodes will be walked directly.
// This makes Walk more predictable
//
// If a node has a body and f doesn't return [NoDive], Walk will dive into it,
// walking it as well.
// Returning [NoDive] for an [ast.If] has no effect on the if's else ifs and
// else.
//
// Walk calls f with a slice of ctx.Node's parents.
// That slice is reused for each call to f, and should not be retained after f
// returns.
//
// You may return [Stop] from f to stop the walk without an error.
//
// Walk's file parameter is optional, but should always be supplied if using
// options or a helper like [IsTopLevel].
func Walk(n ast.Node, f Func, opts ...Option) error {
	if n == nil {
		return nil
	}

	ctx := &Context{
		Parents: make([]*Context, 0, 32),
		Node:    n,
	}

	err := walk(ctx, f, opts)
	if            //goland:noinspection GoDirectComparisonOfErrors
	err == Stop { //nolint:errorlint
		return nil
	}
	return err
}

func walk(ctx *Context, f Func, opts []Option) error {
	var noDive, ignore bool

	for i := 0; i < len(opts) && (!noDive || !ignore); i++ {
		opt := opts[i]

		err := opt(ctx)
		if err != nil {
			//nolint:errorlint
			switch //goland:noinspection GoDirectComparisonOfErrors
			err {
			case NoDive:
				noDive = true
			case Ignore:
				ignore = true
			case Skip:
				noDive, ignore = true, true
			default:
				return err
			}
		}
	}

	if !ignore {
		err := f(ctx)
		if err != nil {
			if              //goland:noinspection GoDirectComparisonOfErrors
			err == NoDive { //nolint:errorlint
				noDive = true
			} else {
				return err
			}
		}
	}
	if !noDive {
		parents := append(ctx.Parents, ctx) //nolint:gocritic

		var err error
		ctx.Node.Walk(func(n ast.Node) {
			if err != nil {
				return
			}

			ctx := &Context{Parents: parents, Node: n}
			err = walk(ctx, f, opts)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// WalkT is the same as [Walk] but only calls f for nodes of type T.
//
//goland:noinspection GoNameStartsWithPackageName
func WalkT[T ast.Node](n ast.Node, f FuncT[T], opts ...Option) error { //nolint:revive
	return Walk(n, func(wctx *Context) error {
		t, ok := wctx.Node.(T)
		if !ok {
			return nil
		}

		return f(&ContextT[T]{Node: t, Context: wctx})
	}, opts...)
}
