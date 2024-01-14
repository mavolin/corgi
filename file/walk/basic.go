package walk

import (
	"errors"

	"github.com/mavolin/corgi/file"
)

// Stop can be returned from a Func to indicate that Walk should return
// immediately, but without actually returning an error.
//
//nolint:revive,errname
var Stop = errors.New("stop walk")

type Func func(parents []Context, wctx Context) (dive bool, err error)

// Walk wals the passed scope in depth-first order, calling f for each item it
// encounters.
//
// If an item has a body and f returns true, Walk will dive into it, walking it
// as well.
//
// If it dives conditionals, it walks all branches.
// Branches may be distinguished by the appropriate fields of the passed
// Context.
//
// Walk calls f with a slice of ctx.Item's parents.
// That slice is reused for each call to f, and should not be retained after f
// exits.
func Walk(s file.Scope, f Func) error {
	err := walk(make([]Context, 0, 50), s, f)
	if errors.Is(err, Stop) {
		return nil
	}

	return err
}

func walk(parents []Context, s file.Scope, f Func) error {
	var comments []file.CorgiComment

	for i, itm := range s.Items {
		i := i
		ctx := Context{
			Scope:    s,
			Index:    i,
			Item:     &s.Items[i],
			Comments: comments,
		}
		dive, err := f(parents, ctx)
		if err != nil {
			return err
		}

		if dive {
			switch itm := itm.(type) {
			case file.If:
				if then, ok := itm.Then.(file.Scope); ok {
					parents = append(parents, ctx)
					if err := walk(parents, then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}

				for iElseIf, elseIf := range itm.ElseIfs {
					if then, ok := elseIf.Then.(file.Scope); ok {
						ctx.ElseIf = &itm.ElseIfs[iElseIf]
						parents = append(parents, ctx)
						if err := walk(parents, then, f); err != nil {
							return err
						}
						parents = parents[:len(parents)-1]
					}
				}
				ctx.ElseIf = nil

				if itm.Else != nil {
					if then, ok := itm.Else.Then.(file.Scope); ok {
						ctx.Else = itm.Else
						parents = append(parents, ctx)
						if err := walk(parents, then, f); err != nil {
							return err
						}
						parents = parents[:len(parents)-1]
						ctx.Else = nil
					}
				}
			case file.Switch:
				for caseI, c := range itm.Cases {
					caseI := caseI
					ctx.Case = &itm.Cases[caseI]
					parents = append(parents, ctx)
					if err := walk(parents, c.Then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
				ctx.Case = nil
			default:
				if body, hasBody := Scope(itm); hasBody {
					parents = append(parents, ctx)
					comments = nil
					if err := walk(parents, body, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
			}
		}

		if cc, ok := itm.(file.CorgiComment); ok {
			comments = append(comments, cc)
		} else {
			comments = nil
		}
	}

	return nil
}
