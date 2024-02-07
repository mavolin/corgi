// Package walk provides utilities for walking through a corgi AST.1
package walk

import (
	"errors"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
)

var (
	// Stop is a sentinel error used to signal that Walk should return without
	// an error.
	//nolint:revive,errname
	Stop = errors.New("stop walk")
	// NoDive is a sentinel error used to signal that Walk should not dive into
	// the current item's body.
	//nolint:revive,errname
	NoDive = errors.New("no dive")
	// Ignore is a sentinel error available to [Option] functions to signal that
	// the current item should be ignored.
	//nolint:revive,errname
	Ignore = errors.New("ignore")
	// IgnoreNoDive is a sentinel error available to [Option] functions to signal
	// that the current item should be ignored and not dived into.
	// It is effectively the combination of [Ignore] and [NoDive].
	IgnoreNoDive = errors.New("ignore no dive")
)

func isSentinel(err error) bool {
	return errors.Is(err, Stop) || errors.Is(err, NoDive) || errors.Is(err, Ignore) || errors.Is(err, IgnoreNoDive)
}

type (
	Func                       func(parents []Context, wctx Context) error
	TypedFunc[T ast.ScopeItem] func(parents []Context, wctx TypedContext[T]) error

	Context struct {
		// File is the file we are walking.
		File *file.File
		// Scope is the scope in which the item was found.
		Scope *ast.Scope
		// Index is the index of the item in the scope.
		Index int
		Item  ast.ScopeItem

		// Case is the case of [file.Switch] that we are walking.
		Case *ast.Case
		// ElseIf is the else if of the [file.If] that we are walking.
		ElseIf *ast.ElseIf
		// Else is the [file.Else] of the [file.If] that we are walking.
		Else *ast.Else

		// Comments are the corgi comments preceding the item.
		Comments []*ast.DevComment
	}
	TypedContext[T ast.ScopeItem] struct {
		Item T
		Context
	}
)

func (ctx *Context) CommentDirectives() []file.CommentDirective {
	mcs := make([]file.CommentDirective, 0, len(ctx.Comments))
	for _, c := range ctx.Comments {
		if d := file.ParseCommentDirective(c); d != nil {
			mcs = append(mcs, *d)
		}
	}

	if len(mcs) == 0 {
		return nil
	}
	return mcs[:len(mcs):len(mcs)]
}

// Walk walks the passed body in depth-first order, calling f for each item it
// encounters.
//
// If an item has a body and f doesn't return [NoDive], Walk will dive into it,
// walking it as well.
//
// If it dives conditionals, it walks all branches.
// Branches may be distinguished by the appropriate fields of the passed
// Context.
//
// Walk calls f with a slice of ctx.Item's parents.
// That slice is reused for each call to f, and should not be retained after f
// returns.
//
// You may return [Stop] from f to stop the walk without an error.
//
// Walk's file parameter is optional, but should always be supplied if using
// options or a helper like [IsTopLevel].
func Walk(fil *file.File, b ast.Body, f Func, opts ...Option) error {
	if b == nil {
		return nil
	}

	s, _ := b.(*ast.Scope)
	if s == nil {
		return nil
	}

	err := walk(fil, make([]Context, 0, 50), s, f, opts)
	if isSentinel(err) {
		return nil
	}

	return err
}

// WalkT is the same as [Walk] but only calls f for items of type T.
func WalkT[T ast.ScopeItem](fil *file.File, b ast.Body, f TypedFunc[T], opts ...Option) error {
	return Walk(fil, b, func(parents []Context, wctx Context) error {
		t, ok := wctx.Item.(T)
		if !ok {
			return nil
		}

		return f(parents, TypedContext[T]{Item: t, Context: wctx})
	}, opts...)
}

func walk(fil *file.File, parents []Context, s *ast.Scope, f Func, opts []Option) error {
	var comments []*ast.DevComment

	for i, itm := range s.Items {
		ctx := Context{
			File:     fil,
			Scope:    s,
			Index:    i,
			Item:     s.Items[i],
			Comments: comments,
		}

		var ignore, noDive bool
		for _, opt := range opts {
			err := opt(parents, ctx)
			if errors.Is(err, IgnoreNoDive) {
				ignore, noDive = true, true
			} else if errors.Is(err, Ignore) {
				ignore = true
			} else if errors.Is(err, NoDive) {
				noDive = true
			}
		}

		if !ignore {
			err := f(parents, ctx)
			if !isSentinel(err) || errors.Is(err, Stop) {
				return err
			}
			if !noDive {
				noDive = errors.Is(err, NoDive)
			}
		}

		if !noDive {
			switch itm := itm.(type) {
			case *ast.If:
				then, _ := itm.Then.(*ast.Scope)
				if then != nil {
					parents = append(parents, ctx)
					err := walk(fil, parents, then, f, opts)

					if !isSentinel(err) || errors.Is(err, Stop) {
						return err
					}
					if !noDive {
						noDive = errors.Is(err, NoDive)

						parents = parents[:len(parents)-1]
					}

					for i, elseIf := range itm.ElseIfs {
						then, _ = elseIf.Then.(*ast.Scope)
						if then == nil {
							continue
						}

						ctx.ElseIf = itm.ElseIfs[i]
						parents = append(parents, ctx)

						err := walk(fil, parents, then, f, opts)
						if !isSentinel(err) || errors.Is(err, Stop) {
							return err
						}
						if !noDive {
							noDive = errors.Is(err, NoDive)
						}

						parents = parents[:len(parents)-1]
					}
					ctx.ElseIf = nil

					if itm.Else != nil {
						then, _ = itm.Else.Then.(*ast.Scope)
						if then == nil {
							continue
						}

						ctx.Else = itm.Else
						parents = append(parents, ctx)

						err := walk(fil, parents, then, f, opts)
						if !isSentinel(err) || errors.Is(err, Stop) {
							return err
						}
						if !noDive {
							noDive = errors.Is(err, NoDive)
						}

						parents = parents[:len(parents)-1]
						ctx.Else = nil
					}
				}
			case *ast.Switch:
				for i, c := range itm.Cases {
					if c.Then == nil {
						continue
					}

					ctx.Case = itm.Cases[i]
					parents = append(parents, ctx)

					err := walk(fil, parents, c.Then, f, opts)
					if !isSentinel(err) || errors.Is(err, Stop) {
						return err
					}
					if !noDive {
						noDive = errors.Is(err, NoDive)
					}

					parents = parents[:len(parents)-1]
				}
				ctx.Case = nil
			default:
				s, ok := file.Scope(itm)
				if !ok {
					break
				}
				parents = append(parents, ctx)
				comments = nil

				err := walk(fil, parents, s, f, opts)
				if !isSentinel(err) || errors.Is(err, Stop) {
					return err
				}
				if !noDive {
					noDive = errors.Is(err, NoDive)
				}

				parents = parents[:len(parents)-1]
			}
		}

		if cc, ok := itm.(*ast.DevComment); ok {
			comments = append(comments, cc)
		} else {
			comments = nil
		}
	}

	return nil
}
