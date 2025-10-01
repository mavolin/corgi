package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
)

func (ch *checker) CheckBlockInstances(logger *slog.Logger, block *file.Block) {
	for _, instance := range block.Instances {
		logger := logger.With(slog.String("instance_pos", instance.AST.Start().String()))

		ch.CheckBlockInstance_InAllowedElement(logger, instance)
	}
}

// ============================================================================
// In Allowed Elements
// ======================================================================================

func (ch *checker) CheckBlockInstance_InAllowedElement(logger *slog.Logger, bi *file.BlockInstance) {
	if bi.ContainingElements.Failed() {
		return
	}

	f := bi.Group.Component.File
	for _, e := range bi.ContainingElements.Result().Get() {
		switches.ContainingElement(e,
			// Handled when CheckBlockInstance_InAllowedElement is called on the
			// component for that call.
			func(*ast.BlockSetterContainingElement) {},
			func(e *ast.Element) {
				ref := f.ElementReferenceByNode(e.Header.Name)
				if ref.Spec.Type.Failed() {
					return
				}

				switch ref.Spec.Type.Result() {
				case elemtype.JS:
					logger.
						WithGroup("checks.not_in_script").
						Error("Block instance in JS-typed element", slog.String("block", string(bi.Group.Name)))
					ch.Report(&diagnostic.Diagnostic{
						Message: "block placed in `js`-typed element",
						Primary: []diagnostic.Annotation{
							anno.Node(f, bi.AST, "cannot place `block` here"),
						},
						Explanation: "You cannot place blocks inside `js`-typed elements.",
					})
				case elemtype.CSS:
					logger.
						WithGroup("checks.not_in_script").
						Error("Block instance in CSS-typed element", slog.String("block", string(bi.Group.Name)))
					ch.Report(&diagnostic.Diagnostic{
						Message: "block placed in `css`-typed element",
						Primary: []diagnostic.Annotation{
							anno.Node(f, bi.AST, "cannot place `block` here"),
						},
						Explanation: "You cannot place blocks inside `css`-typed elements.",
					})
				case elemtype.Unknown, elemtype.Void, elemtype.Nothing, elemtype.Text, elemtype.Normal: // for linting
					// do nothing
				}
			})
	}
}
