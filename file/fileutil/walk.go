package fileutil

import (
	"errors"

	"github.com/mavolin/corgi/file"
)

// StopWalk can be returned from a WalkFunc to indicate that Walk should return
// immediately, but without actually returning an error.
//
//nolint:revive,errname
var StopWalk = errors.New("stop walk")

type WalkContext struct {
	// Scope is the scope in which the item was found.
	Scope file.Scope
	// Index is the index of the item.
	Index int
	// Item is a pointer to the item, allowing it to be edited.
	Item *file.ScopeItem

	// Case is the case of Item.(file.Switch) that we are walking.
	//
	// Note that if this is set, the parent context's item will be the switch's
	// parent, not the switch itself.
	Case *file.Case
	// ElseIf is the else if of Item.(file.If) or Item.(file.IfBlock) that we
	// are walking.
	//
	// Note that if this is set, the parent context's item will be if's
	// parent, not the if itself.
	ElseIf *file.ElseIf
	// ElseIfBlock is the else if of Item.(file.IfBlock) that we are walking.
	//
	// Note that if this is set, the parent context's item will be if's
	// parent, not the if itself.
	ElseIfBlock *file.ElseIfBlock
	// Else is the else of Item.(file.If) that we are walking.
	//
	// Note that if this is set, the parent context's item will be if's
	// parent, not the if itself.
	Else *file.Else

	// Comments are the corgi comments preceding the item.
	Comments []file.CorgiComment
}

type WalkFunc func(parents []WalkContext, ctx WalkContext) (dive bool, err error)

// Walk wals the passed scope in depth-first order, calling f for each item it
// encounters.
//
// If an item has a body and f returns true, Walk will dive into it, walking it
// as well.
//
// If it dives conditionals, it walks all branches.
// Branches may be distinguished by the appropriate fields of the passed
// WalkContext.
//
// Walk calls f with a slice of ctx.Item's parents.
// That slice is reused for each call to f, and should not be retained after f
// exits.
func Walk(s file.Scope, f WalkFunc) error {
	err := walk(make([]WalkContext, 0, 50), s, f)
	if errors.Is(err, StopWalk) {
		return nil
	}

	return err
}

func walk(parents []WalkContext, s file.Scope, f WalkFunc) error {
	var cs []file.CorgiComment

	for i, itm := range s {
		i := i
		ctx := WalkContext{
			Scope:    s,
			Index:    i,
			Item:     &s[i],
			Comments: cs,
		}
		dive, err := f(parents, ctx)
		if err != nil {
			return err
		}

		if dive {
			switch itm := itm.(type) {
			case file.If:
				parents = append(parents, ctx)
				if err := walk(parents, itm.Then, f); err != nil {
					return err
				}
				parents = parents[:len(parents)-1]

				for elseIfI, elseIf := range itm.ElseIfs {
					elseIfI := elseIfI
					ctx.ElseIf = &itm.ElseIfs[elseIfI]
					parents = append(parents, ctx)
					if err := walk(parents, elseIf.Then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
				ctx.ElseIf = nil

				if itm.Else != nil {
					ctx.Else = itm.Else
					parents = append(parents, ctx)
					if err := walk(parents, itm.Else.Then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
			case file.IfBlock:
				parents = append(parents, ctx)
				if err := walk(parents, itm.Then, f); err != nil {
					return err
				}
				parents = parents[:len(parents)-1]

				for elseIfI, elseIf := range itm.ElseIfs {
					elseIfI := elseIfI
					ctx.ElseIfBlock = &itm.ElseIfs[elseIfI]
					parents = append(parents, ctx)
					if err := walk(parents, elseIf.Then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
				ctx.ElseIf = nil

				if itm.Else != nil {
					ctx.Else = itm.Else
					parents = append(parents, ctx)
					if err := walk(parents, itm.Else.Then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
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

				if itm.Default != nil {
					ctx.Case = itm.Default
					parents = append(parents, ctx)
					if err := walk(parents, itm.Default.Then, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
			default:
				if body, hasBody := Body(itm); hasBody {
					parents = append(parents, ctx)
					cs = nil
					if err := walk(parents, body, f); err != nil {
						return err
					}
					parents = parents[:len(parents)-1]
				}
			}
		}

		if cc, ok := itm.(file.CorgiComment); ok {
			cs = append(cs, cc)
		} else {
			cs = nil
		}
	}

	return nil
}
