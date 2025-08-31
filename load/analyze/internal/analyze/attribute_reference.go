package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AnalyzeAttributes analyzes all attribute in the package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttributes() {
	logger := z.Logger.WithGroup("attribute_references")
	logger.Debug("Analyzing attribute references")

	for _, f := range z.P.Files {
		logger := logger.With(slog.String("file", f.Name))
		walk.WalkT(f.AST, func(w *walk.ContextT[ast.Attribute]) walk.Action {
			z.AnalyzeAttribute(logger, f, w.Parents, f.AttributeByNode(w.Node))
			return walk.Continue
		})
	}
}

// AnalyzeAttribute analyzes the given attribute.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAttribute(logger *slog.Logger, f *file.File, parents []*walk.Context, attr *file.Attribute) {
	logger = logger.With(slog.String("ref_pos", attr.AST.Start().String()))

	z.AnalyzeAttributeForwarded(f, parents, attr)
	z.AnalyzeAttributeContainingElements(f, parents, attr)
	z.AnalyzeAttributeType(logger, f, parents, attr)
}

// ============================================================================
// Element
// ======================================================================================

// AnalyzeAttributeReferenceElement finds the element writer containing the
// given attribute reference.
//
// Depends on Checks: None
//
// Sets Fields:
//   - AttributeReferences.Element
//
// Depends on Fields:
//   - Components.Blocks.Instances.NotForwarded
func (z *analyzer) AnalyzeAttributeReferenceElement(f *file.File, parents []*walk.Context, ref *file.AttributeReference) {
	i := len(parents) - 1
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.ComponentCall:
			// can't directly be in a component call in a valid AST
			ref.Element.SetFailed()
			return
		case ast.BlockSetter:
			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				ref.Element.SetFailed()
				return
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
			cc := f.ComponentCallByNode(ccAST)

			s := cc.BlockSetterByName(parent.Name())
			if s == nil || s.Block == nil {
				ref.Element.SetFailed()
				return
			}
			for _, inst := range s.Block.Instances {
				if inst.NotForwarded.True() {
					ref.Element.SetResult(&ast.BlockSetterElementWriter{
						ComponentCall: ccAST,
						BlockSetter:   parent,
					})
					return
				}
			}
			i = ccI - 1 // continue with the parent of the component call
		case ast.ElementWriter:
			ref.Element.SetResult(parent)
			return
		default:
			i--
		}
	}
}
