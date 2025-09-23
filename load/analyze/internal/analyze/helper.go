package analyze

import (
	"context"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
	"github.com/mavolin/corgi/v2/load/analyze/internal/candidate"
)

type stackItem struct {
	// A nil stack item indicates to inherit the reason from the node before.
	// A non-nil stack item indicates the next reason at that level should be
	// that item.
	reason *file.AnalysisWithReason[ast.AttributeInhibitor]
	// before is only set for component calls and indicates the reason
	// before entering the component call.
	before *file.AnalysisWithReason[ast.AttributeInhibitor]
	// branch indicates whether this stack item belongs to a
	// conditional/switch branch.
	branch bool
}

func (z *analyzer) cannotAttributes(
	ctx context.Context, f *file.File, reason *file.AnalysisWithReason[ast.AttributeInhibitor],
) walk.Option {
	var numParents int

	// Stack for each level of the AST, storing the reason for the next node
	// at that level.
	// Most of the time, we want to keep the same reason after surfacing or
	// diving: Consider a for loop: if we dive, we want to keep the reason
	// from before the for loop, and if we surface, we also want to keep that
	// reason.
	// However, if we dive an element, we want to allow elements again, and if
	// we surface (or process the node behind), we want to forbid them.
	stack := make([]stackItem, 1, 48)
	return func(w *walk.Context) walk.Action {
		diven := len(w.Parents) > numParents
		surfaced := len(w.Parents) < numParents
		numParents = len(w.Parents)

		switch {
		case surfaced:
			// We surfaced.
			// Pop the stack one by one and correctly adjust the reason until
			// we have processed the node at the same level right before this
			// node.
			for _, child := range slices.Backward(stack[numParents:]) {
				if child.branch {
					// child is a branch of a conditional/switch.
					// That means its parent is the conditional/switch.
					// The value of reason thus far is the reason to use for
					// that branch.
					// Combine that reason with the reason from the previous
					// branches, if any, and store it as the reason to use
					// after the conditional/switch.
					// The previous branches' is the reason stored in the
					// conditional/switch's stack item.
					if !stack[len(stack)-2].reason.True() && !reason.False() {
						stack[len(stack)-2].reason = stack[len(stack)-1].reason
					}
					// A branch also always has a reason, used to restore the
					// reason before the branch started, so that when we enter
					// the next branch, it starts with that same reason.
					// Note that the reason of the conditional/switch is one
					// up and represents the cumulative reason of all branches,
					// so this logic is still sound.
					*reason = *child.reason
				} else if child.reason != nil {
					*reason = *child.reason
				}
			}

			// Finally, adjust the length of the stack.
			stack = stack[:numParents+1]
		case diven:
			// Decide if we want to inherit the reason from our parent, or if
			// we want to set a new reason.
			// Default to inheriting.
			candidate.SwitchAttributeInhibitor(w.Parents[len(w.Parents)-1].Node,
				func(*ast.Block) {},
				func(parent ast.BlockSetter) {
					ccI := walk.ClosestIndex[*ast.ComponentCall](w.Parents[:len(w.Parents)-1])
					if ccI < 0 {
						reason.SetFailed()
						return
					}

					ccAST := w.Parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck

					cc := f.ComponentCallByNode(ccAST)
					z.AnalyzeComponentCall(ctx, cc)
					s := cc.BlockSetterByNode(parent)
					if s == nil {
						reason.SetFailed()
						return
					}

					blockCannotForwardAttrs := s.Block.CannotForwardAttributes()
					switch {
					case blockCannotForwardAttrs.Failed():
						reason.SetFailed()
					case blockCannotForwardAttrs.True():
						reason.SetReason(&ast.BlockSetterAttributeInhibitor{
							ComponentCall: ccAST,
							BlockSetter:   parent,
						})
					default:
						ccStackItem := stack[ccI+1]
						*reason = *ccStackItem.before
					}
				},
				func(*ast.CharacterEscape) {},
				func(*ast.CharacterReference) {},
				func(parent *ast.ComponentCall) {
					cc := f.ComponentCallByNode(parent)
					z.AnalyzeComponentCall(ctx, cc)

					acceptsAttributes := cc.AcceptsAttributes()
					switch {
					case acceptsAttributes.True():
						reason.SetFalse()
					case acceptsAttributes.False():
						reason.SetReason(parent)
					case acceptsAttributes.Failed():
						reason.SetFailed()
					}
				},
				func(*ast.Doctype) {},
				func(*ast.Element) { reason.SetFalse() },
				func(*ast.ExpressionInterpolation) {},
				func(*ast.RawElement) {},
				func(*ast.Text) {})

			// By default, assume that we want to inherit the reason from the
			// child again, after surfacing.
			stack = append(stack, stackItem{})
		default:
			if stack[len(stack)-1].reason != nil {
				*reason = *stack[len(stack)-1].reason
			}
		}

		// Options are executed before the walk function.
		// That means, we need to apply the reason to the stack, so that it
		// gets applied at the next iteration.

		// Reset the stack item at this level, so that by default the next
		// node would inherit the reason from the last-walked node at this or a
		// deeper level.
		stack[len(stack)-1].reason = nil
		stack[len(stack)-1].before = nil
		stack[len(stack)-1].branch = false

		clone := *reason

		switch w.Node.(type) {
		case *ast.Expression:
			// Expressions don't affect attribute forwarding.
			stack[len(stack)-1].reason = &clone
			return walk.Continue

		// Note that we are entering a conditional/switch.
		// The reason after the conditional/switch defaults to the reason
		// before the conditional/switch, to be updated with the reason of
		// each branch as we complete them.
		case *ast.Conditional:
			stack[len(stack)-1].reason = &clone
			return walk.Continue
		case *ast.Switch:
			stack[len(stack)-1].reason = &clone
			return walk.Continue

		// We entered a branch of a conditional/switch.
		// We note to restore the current reason after the branch is completed,
		// that way each branch starts with the same reason.
		// The conditional/switch handles combining the reasons of all branches.
		case *ast.If:
			stack[len(stack)-1].reason = &clone
			stack[len(stack)-1].branch = true
			return walk.Continue
		case *ast.ElseIf:
			stack[len(stack)-1].reason = &clone
			stack[len(stack)-1].branch = true
			return walk.Continue
		case *ast.Else:
			stack[len(stack)-1].reason = &clone
			stack[len(stack)-1].branch = true
			return walk.Continue
		case *ast.Case:
			stack[len(stack)-1].reason = &clone
			stack[len(stack)-1].branch = true
			return walk.Continue
		}

		// Determine the reason for the next node at this level.
		// Don't touch this variable to inherit the reason from the last-walked
		// node this or a deeper level.
		var next file.AnalysisWithReason[ast.AttributeInhibitor]

		candidate.SwitchAttributeInhibitor(w.Parents[len(w.Parents)-1].Node,
			func(n *ast.Block) { next.SetReason(n) },
			func(ast.BlockSetter) {},
			func(n *ast.CharacterEscape) { next.SetReason(n) },
			func(n *ast.CharacterReference) { next.SetReason(n) },
			func(ccAST *ast.ComponentCall) {
				stack[len(stack)-1].before = &clone
				cc := f.ComponentCallByNode(ccAST)
				z.AnalyzeComponentCall(ctx, cc)

				wc := cc.WritesContent()
				if wc.Equal(true) {
					next.SetReason(ccAST)
				} else if wc.Failed() {
					next.SetFailed()
					stack[len(stack)-1].reason = &next
				}
			},
			func(n *ast.Doctype) { next.SetReason(n) },
			func(n *ast.Element) { next.SetReason(n) },
			func(n *ast.ExpressionInterpolation) { next.SetReason(n) },
			func(n *ast.RawElement) { next.SetReason(n) },
			func(n *ast.Text) { next.SetReason(n) })
		if !next.Failed() {
			stack[len(stack)-1].reason = &next
		}

		return walk.Continue
	}
}
