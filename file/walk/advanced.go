package walk

import (
	"errors"
	"reflect"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/internal/cheat"
)

type Walker struct {
	CollectErrors bool
	walkers       []walker

	parents []Context

	errs []error
}

// Walk walks through the given scope, calling the registered functions.
//
// If CollectErrors is true, it will collect all [fileerr.Error] pointers
// returned by the registered functions and return them as a [fileerr.list].
// Errors of other types will be immediately returned.
//
// Walk must not be called concurrently.
func (w *Walker) Walk(s file.Scope) error {
	w.errs = make([]error, 0, 50)
	w.parents = make([]Context, 0, 50)

	err := w.walk(s)
	if err != nil {
		w.errs = append(w.errs, err)
	}

	return errors.Join(w.errs...)
}

func (w *Walker) walk(s file.Scope) error {
	var comments []file.CorgiComment

	resetI := make([]int, 0, len(w.walkers))

	for i, itm := range s.Items {
		resetI = resetI[:0]
		i := i
		ctx := Context{
			Scope:    s,
			Index:    i,
			Item:     &s.Items[i],
			Comments: comments,
		}
		var dive bool
		for i, wf := range w.walkers {
			if !wf.shouldCall {
				continue
			}

			d, err := wf.f(w.parents, &ctx)
			if err = w.handleErr(err); err != nil {
				return err
			}

			if !d {
				resetI = append(resetI, i)
				w.walkers[i].shouldCall = false
			} else {
				dive = true
			}
		}

		if dive {
			switch itm := itm.(type) {
			case file.If:
				if then, ok := itm.Then.(file.Scope); ok {
					w.parents = append(w.parents, ctx)
					if err := w.walk(then); err != nil {
						return err
					}
					w.parents = w.parents[:len(w.parents)-1]
				}

				for iElseIf, elseIf := range itm.ElseIfs {
					if then, ok := elseIf.Then.(file.Scope); ok {
						ctx.ElseIf = &itm.ElseIfs[iElseIf]
						w.parents = append(w.parents, ctx)
						if err := w.walk(then); err != nil {
							return err
						}
						w.parents = w.parents[:len(w.parents)-1]
					}
				}
				ctx.ElseIf = nil

				if itm.Else != nil {
					if then, ok := itm.Else.Then.(file.Scope); ok {
						ctx.Else = itm.Else
						w.parents = append(w.parents, ctx)
						if err := w.walk(then); err != nil {
							return err
						}
						w.parents = w.parents[:len(w.parents)-1]
						ctx.Else = nil
					}
				}
			case file.Switch:
				for caseI, c := range itm.Cases {
					caseI := caseI
					ctx.Case = &itm.Cases[caseI]
					w.parents = append(w.parents, ctx)
					if err := w.walk(c.Then); err != nil {
						return err
					}
					w.parents = w.parents[:len(w.parents)-1]
				}
				ctx.Case = nil
			default:
				if body, hasBody := Scope(itm); hasBody {
					comments = nil
					w.parents = append(w.parents, ctx)
					if err := w.walk(body); err != nil {
						return err
					}
					w.parents = w.parents[:len(w.parents)-1]
				}
			}
		}

		for _, i := range resetI {
			w.walkers[i].shouldCall = true
		}

		if cc, ok := itm.(file.CorgiComment); ok {
			comments = append(comments, cc)
		} else {
			comments = nil
		}
	}

	return nil
}

func (w *Walker) handleErr(err error) error {
	if err == nil {
		return nil
	}

	if !w.CollectErrors {
		return err
	}

	var ferr *fileerr.Error
	if errors.As(err, &ferr) {
		w.errs = append(w.errs, ferr)
		return nil
	}

	return err
}

type walker struct {
	f          func(parents []Context, wctx *Context) (dive bool, err error)
	shouldCall bool
}

func (w *Walker) Register(f func(parents []Context, wctx *Context) (dive bool, err error)) {
	w.walkers = append(w.walkers, walker{f: f, shouldCall: true})
}

type TypedContext[T file.ScopeItem] struct {
	Item *T
	Context
}

type Option func(parents []Context, wctx *Context) bool

func ChildOf(types ...any) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(parents []Context, wctx *Context) bool {
		var typI int
		for _, p := range parents {
			pt := reflect.TypeOf(*p.Item)
			if pt != rTypes[typI] {
				continue
			}

			typI++
			if typI == len(rTypes) {
				return true
			}
		}

		return false
	}
}

func NotChildOf(types ...any) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(parents []Context, wctx *Context) bool {
		var typI int
		for _, p := range parents {
			pt := reflect.TypeOf(*p.Item)
			if pt != rTypes[typI] {
				continue
			}

			typI++
			if typI == len(rTypes) {
				return false
			}
		}

		return true
	}
}

func Register[T file.ScopeItem](w *Walker, f func(parents []Context, wctx *TypedContext[T]) (dive bool, err error), opts ...Option) {
	w.Register(func(p []Context, wctx *Context) (bool, error) {
		_, ok := (*wctx.Item).(T)
		if !ok {
			return true, nil
		}

		for _, opt := range opts {
			if !opt(p, wctx) {
				return true, nil
			}
		}

		return f(p, &TypedContext[T]{
			Item: cheat.PtrToSliceElem[file.ScopeItem, T](wctx.Scope.Items, wctx.Index), Context: *wctx,
		})
	})
}
