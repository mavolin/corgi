// Package walk provides utilities for walking through a corgi AST.1
package walk

import (
	"errors"
	"fmt"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
)

var (
	// Stop is a sentinel error used to signal that Walk should return without
	// an error.
	//nolint:revive,errname
	Stop = errors.New("stop walk")
	// NoDive is a sentinel error used to signal that Walk should not dive into
	// the current node's body.
	//nolint:revive,errname
	NoDive = errors.New("no dive")
	// Ignore is a sentinel error available to [Option] functions to signal
	// that Walk should not call the [Func] for the current node, but it should
	// still dive it, if possible.
	//
	// It has no effect if returned by a [Func].
	//
	// It may still be dived.
	//nolint:revive,errname
	Ignore = errors.New("ignore")
	// Skip is a sentinel error available to [Option] functions to signal
	// to skip over the current node, i.e. ignore it and don't dive into it.
	//
	// Skip is essentially the combination of [Ignore] and [NoDive].
	//
	// If returned by a [Func], it behaves like [NoDive].
	Skip = errors.New("ignore no dive")
	// SkipIf is like [Skip], but skips over an [ast.If] and its else ifs and
	// else, in contrast to [Skip], which only skips the [ast.If] but still
	// visits its else ifs and else.
	//
	// If returned for a non-if, it behaves like [Skip].
	SkipIf = errors.New("ignore no dive")
)

func isSentinel(err error) bool {
	return errors.Is(err, Stop) || errors.Is(err, NoDive) || errors.Is(err, Ignore) || errors.Is(err, Skip)
}

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
		// If records the if statement this else if or else belongs to.
		If *ast.If

		// Comments are the corgi comments preceding the node.
		//
		// Functions must not take ownership after the function returns, or
		// alter it in any way, as Walk may reuse and alter it.
		Comments []*ast.DevComment
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

type Walkable interface {
	*ast.Scope | *ast.BracketText | *ast.Extend |
		*ast.UnderscoreBlockShorthand |
		*ast.If | *ast.Switch |
		ast.TextBlock | ast.TextLine
}

// Walk walks the passed [Walkable] in depth-first order, calling f for each
// node it encounters.
//
// Walk defines a node as a [ast.ScopeNode], [ast.TextNode], [ast.TexLine],
// [ast.ElseIf], [ast.Else], [ast.Case].
//
// If walking an [ast.ElseIf] or [ast.Else], the context will contain the
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
func Walk[W Walkable](w W, f Func, opts ...Option) error {
	if w == nil {
		return nil
	}

	err := walk(nil, w, f, opts)
	if isSentinel(err) {
		return nil
	}
	return err
}

// WalkT is the same as [Walk] but only calls f for nodes of type T.
func WalkT[T ast.Node, W Walkable](w W, f FuncT[T], opts ...Option) error {
	return Walk(w, func(wctx *Context) error {
		t, ok := wctx.Node.(T)
		if !ok {
			return nil
		}

		return f(&ContextT[T]{Node: t, Context: wctx})
	}, opts...)
}

func walk[W Walkable](parent *Context, w W, f Func, opts []Option) error {
	switch w := any(w).(type) {
	case ast.Body:
		return walkBody(parent, w, f, opts)
	case *ast.If:
		return walkIf(appendContext(parent, w, false), f, opts)
	case *ast.Switch:
		return walkSwitchCases(appendContext(parent, w, false), f, opts)
	case ast.TextBlock:
		return walkText(parent, w, f, opts)
	case ast.TextLine:
		return walkTextLine(parent, w, f, opts)
	default:
		panic(fmt.Sprintf("walk: called with type %T, not recognized as a Walkable", w))
	}

	return nil
}

func walkBody(parent *Context, b ast.Body, f Func, opts []Option) error {
	for {
		switch t := b.(type) {
		case *ast.Scope:
			return walkScope(parent, t, f, opts)
		case *ast.BracketText:
			return walkText(parent, t.Lines, f, opts)
		case *ast.Extend:
			return walkExtend(parent, t, f, opts)
		case *ast.UnderscoreBlockShorthand:
			b = t.Body
		default:
			return fmt.Errorf("walk: unrecognized body type %T", b)
		}
	}
}

func walkScope(parent *Context, s *ast.Scope, f Func, opts []Option) error {
	ctx := appendContext(parent, nil, false)
	numInheritCommentDirectives := len(ctx.CommentDirectives)

	for _, n := range s.Nodes {
		ctx.Node = n
		if err := walkScopeNode(ctx, f, opts); err != nil {
			return err
		}

		if c, ok := n.(*ast.DevComment); ok {
			if ctx.Comments == nil {
				ctx.Comments = make([]*ast.DevComment, 1, 48)
				ctx.Comments[0] = c
			} else {
				ctx.Comments = append(ctx.Comments, c)
			}

			if cd := file.ParseCommentDirective(c); cd != nil {
				if ctx.CommentDirectives == nil {
					ctx.CommentDirectives = make([]file.CommentDirective, 1, 24)
					ctx.CommentDirectives[0] = *cd
				} else {
					ctx.CommentDirectives = append(ctx.CommentDirectives, *cd)
				}
			}
		} else {
			if ctx.Comments != nil {
				ctx.Comments = ctx.Comments[:0]
			}
			if ctx.CommentDirectives != nil {
				ctx.CommentDirectives = ctx.CommentDirectives[:numInheritCommentDirectives]
			}
		}
	}

	return nil
}

func walkScopeNode(ctx *Context, f Func, opts []Option) error {
	// there are two special scope nodes that need to be handled differently
	if _, ok := ctx.Node.(*ast.If); ok {
		return walkIf(ctx, f, opts)
	}

	sen, err := applyOptions(ctx, opts)
	if err != nil {
		return err
	}

	if !sen.ignore {
		err = handleError(f(ctx), &sen)
		if err != nil {
			return err
		}
	}

	if sen.noDive {
		return nil
	}

	switch n := ctx.Node.(type) {
	case *ast.Switch:
		return walkSwitchCases(ctx, f, opts)
	case *ast.ArrowBlock:
		return walkText(ctx, n.Lines, f, opts)
	default:
		if b, _ := file.Body(n); b != nil {
			return walkBody(ctx, b, f, opts)
		}
		return nil
	}
}

func walkIf(ctx *Context, f Func, opts []Option) error {
	if_ := ctx.Node.(*ast.If)

	sen, err := applyOptions(ctx, opts)
	if err != nil {
		return err
	}

	if !sen.ignore {
		if err = handleError(f(ctx), &sen); err != nil {
			return err
		}
	}

	if sen.skipIf {
		return nil
	} else if !sen.noDive {
		if err = walkBody(ctx, if_.Then, f, opts); err != nil {
			return err
		}
	}

	ctx.If = if_
	for _, elseIf := range if_.ElseIfs {
		ctx.Node = elseIf
		sen, err = applyOptions(ctx, opts)
		if err != nil {
			return err
		}

		if !sen.ignore {
			if err = handleError(f(ctx), &sen); err != nil {
				return err
			}
		}

		if sen.skipIf {
			return nil
		} else if !sen.noDive {
			if err = walkBody(ctx, elseIf.Then, f, opts); err != nil {
				return err
			}
		}
	}

	if if_.Else == nil {
		return nil
	}

	ctx.Node = if_.Else
	sen, err = applyOptions(ctx, opts)
	if err != nil {
		return err
	}

	if !sen.ignore {
		if err = handleError(f(ctx), &sen); err != nil {
			return err
		}
	}

	if sen.skipIf {
		return nil
	} else if !sen.noDive {
		if err = walkBody(ctx, if_.Else.Then, f, opts); err != nil {
			return err
		}
	}

	return nil
}

func walkSwitchCases(switchCtx *Context, f Func, opts []Option) error {
	s := switchCtx.Node.(*ast.Switch)

	ctx := appendContext(switchCtx, nil, false)
	for _, c := range s.Cases {
		ctx.Node = c
		sen, err := applyOptions(ctx, opts)
		if err != nil {
			return err
		}

		if !sen.ignore {
			if err = handleError(f(ctx), &sen); err != nil {
				return err
			}
		}

		if sen.skipIf {
			return nil
		} else if !sen.noDive {
			if err = walkScope(ctx, c.Then, f, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func walkText(parent *Context, b ast.TextBlock, f Func, opts []Option) error {
	ctx := appendContext(parent, nil, true)

	for _, ln := range b {
		ctx.Node = ln

		sen, err := applyOptions(ctx, opts)
		if err != nil {
			return err
		}

		if !sen.ignore {
			err = handleError(f(ctx), &sen)
			if err != nil {
				return err
			}
		}

		if !sen.noDive {
			if err = walkTextLine(ctx, ln, f, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func walkTextLine(parent *Context, ln ast.TextLine, f Func, opts []Option) error {
	ctx := appendContext(parent, nil, true)
	for _, n := range ln {
		ctx.Node = n
		if err := walkTextNode(ctx, f, opts); err != nil {
			return err
		}
	}

	return nil
}

func walkTextNode(ctx *Context, f Func, opts []Option) error {
	sen, err := applyOptions(ctx, opts)
	if err != nil {
		return err
	}

	if !sen.ignore {
		err = handleError(f(ctx), &sen)
		if err != nil {
			return err
		}
	}

	if sen.noDive {
		return nil
	}

	switch n := ctx.Node.(type) {
	case *ast.ComponentCallInterpolation:
		return walkTextLine(ctx, n.Value.Text, f, opts)
	case *ast.ElementInterpolation:
		return walkTextLine(ctx, n.Value.Text, f, opts)
	}
	return nil
}

func walkExtend(parent *Context, e *ast.Extend, f Func, opts []Option) error {
	if e == nil || e.ComponentCall == nil {
		return nil
	}
	return walkScopeNode(appendContext(parent, e.ComponentCall, false), f, opts)
}

func appendContext(parent *Context, n ast.Node, inheritAllCommentDirectives bool) *Context {
	if parent == nil {
		return &Context{
			Parents: make([]*Context, 0, 32),
			Node:    n,
		}
	}

	ctx := &Context{Parents: append(parent.Parents, parent), Node: n}
	if inheritAllCommentDirectives {
		ctx.CommentDirectives = parent.CommentDirectives
	} else {
		for i, c := range parent.CommentDirectives {
			if !c.Inherited {
				continue
			}
			if ctx.CommentDirectives == nil { // first inherited comment directive
				ctx.CommentDirectives = make([]file.CommentDirective, 1, max(len(parent.CommentDirectives)-i, 24))
				ctx.CommentDirectives[0] = c
			} else {
				ctx.CommentDirectives = append(ctx.CommentDirectives, c)
			}
		}
	}
	return ctx
}

type sentinel struct {
	ignore, noDive, skipIf bool
}

func applyOptions(ctx *Context, opts []Option) (sentinel, error) {
	var sen sentinel
	for _, opt := range opts {
		err := handleError(opt(ctx), &sen)
		if err != nil {
			return sen, err
		}
	}
	return sen, nil
}

func handleError(err error, sen *sentinel) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, Stop) {
		return Stop
	} else if errors.Is(err, Skip) {
		sen.ignore, sen.noDive = true, true
		return nil
	} else if errors.Is(err, SkipIf) {
		sen.ignore, sen.noDive, sen.skipIf = true, true, true
		return nil
	}

	// err might impl interface { Unwrap() []error } and thus return Ignore
	// and NoDive
	if errors.Is(err, Ignore) {
		sen.ignore = true
		if errors.Is(err, NoDive) {
			sen.noDive = true
		}
		return nil
	} else if errors.Is(err, NoDive) {
		sen.noDive = true
		return nil
	}
	return err
}
