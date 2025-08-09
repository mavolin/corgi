package analyze

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (z *analyzer) cannotAttributes(ctx context.Context, f *file.File, reason *file.AnalysisWithReason[ast.ContentWriter]) walk.Option {
	var numParents int
	inArrowBlock := -1 // set to numParents

	stack := make([]file.AnalysisWithReason[ast.ContentWriter], 0, 48)
	return func(w *walk.Context) walk.Action {
		diven := len(w.Parents) > numParents
		surfaced := len(w.Parents) < numParents
		numParents = len(w.Parents)

		switch {
		case surfaced:
			stack = stack[:len(stack)-1]
			if inArrowBlock == numParents {
				// We just got out of an arrow block.
				// If we cannot write attributes after the arrow block, but
				// could before, depends on whether the arrow block wrote any
				// content.
				// Simply set the reason to the reason of the last arrow block
				// item
				stack[len(stack)-1] = *reason
				inArrowBlock = -1
			}
			*reason = stack[len(stack)-1]
		case diven:
			switch parent := w.Parents[len(w.Parents)-1].Node.(type) {
			case *ast.Element:
				reason.SetFalse()
			case *ast.ComponentCall:
				reason.SetReason(parent)
			case ast.BlockSetter:
				ccAST := walk.Closest[*ast.ComponentCall](w.Parents[:len(w.Parents)-1])
				if ccAST == nil {
					reason.SetFailed()
				} else {
					cc := f.ComponentCallByNode(ccAST)
					z.AnalyzeComponentCall(ctx, cc)
					s := cc.BlockSetterByName(parent.Name())
					if s == nil || s.Block.CannotForwardAttributes.Failed() {
						reason.SetFailed()
					} else if s.Block.CannotForwardAttributes.True() {
						reason.SetReason(parent)
					}
				}
			}
			stack = append(stack, *reason)
		default:
			if inArrowBlock == numParents {
				inArrowBlock = -1
			}
			*reason = stack[len(stack)-1] // apply the last reason at the same level
		}

		// Options are executed before the walk function.
		// That means, we need to apply the reason to the stack, so that it
		// gets applied at the next iteration.
		switch n := w.Node.(type) {
		case *ast.Expression:
			return walk.Ignore | walk.NoDive
		case *ast.ComponentCall:
			cc := f.ComponentCallByNode(n)
			z.AnalyzeComponentCall(ctx, cc)

			wc := cc.WritesContent()
			if wc.Equal(true) {
				stack[len(stack)-1].SetReason(n)
			} else if wc.Failed() {
				stack[len(stack)-1].SetFailed()
			}
		case *ast.ComponentCallInterpolation:
			cc := f.ComponentCallByNode(n.ComponentCall)
			z.AnalyzeComponentCall(ctx, cc)

			wc := cc.WritesContent()
			if wc.Equal(true) {
				stack[len(stack)-1].SetReason(n.ComponentCall)
			} else if wc.Failed() {
				stack[len(stack)-1].SetFailed()
			}
		case *ast.Element:
			stack[len(stack)-1].SetReason(n)
		case *ast.Doctype:
			stack[len(stack)-1].SetReason(n)
		case *ast.Block:
			stack[len(stack)-1].SetReason(n)
		case *ast.Text:
			stack[len(stack)-1].SetReason(n)
		case *ast.CharacterReference:
			stack[len(stack)-1].SetReason(n)
		case *ast.EscapedHash:
			stack[len(stack)-1].SetReason(n)
		case *ast.EscapedRBracket:
			stack[len(stack)-1].SetReason(n)
		case *ast.ExpressionInterpolation:
			stack[len(stack)-1].SetReason(n)
		case *ast.HashSpace:
			stack[len(stack)-1].SetReason(n)
		case *ast.ArrowBlock:
			inArrowBlock = numParents
		}

		return walk.Continue
	}
}
