package analyze

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (z *analyzer) AnalyzeComponentCalls() {
	logger := z.Logger.WithGroup("component_calls")
	logger.Debug("Analyzing remaining component calls")

	ctx := context.Background()
	for _, f := range z.Pkg.Files {
		for _, cc := range f.ComponentCalls {
			z.AnalyzeComponentCall(ctx, cc)
		}
	}
}

func (z *analyzer) AnalyzeComponentCall(ctx context.Context, cc *file.ComponentCall) {
	if cc.Analyzed {
		return
	}

	logger := z.Logger.
		WithGroup("component_calls").
		With(slog.String("call_name", cc.AST.Header.Name.Full()),
			slog.String("call_pos", cc.AST.Start().String()))

	if cc.Component != nil {
		callerChain, _ := ctx.Value(callerChainKey{}).([]*file.Component)
		if len(callerChain) == 0 {
			callerChain = make([]*file.Component, 0, 100)
		} else {
			z.checkNoInfiniteRecursion(logger, cc, callerChain)
		}
		callerChain = append(callerChain, cc.Component)
		ctx = context.WithValue(ctx, callerChainKey{}, callerChain)
	}

	z.AnalyzeComponentCall_Component(ctx, cc)
	z.AnalyzeComponentCall_ReceivesAttributes_ReceivesAndPlaceholder(ctx, cc)
	z.AnalyzeComponentCall_ElementsWithAndPlaceholder(cc)
	z.AnalyzeComponentCall_ElementSpecsWithAndPlaceholder(cc)

	z.AnalyzeComponentCall_ComponentArguments(cc)

	cc.Analyzed = true
}

func (z *analyzer) checkNoInfiniteRecursion(logger *slog.Logger, cc *file.ComponentCall, callerChain []*file.Component) {
	if get.Component.Circular(z, cc.Component) {
		return
	} else if len(callerChain) < 2048 {
		return
	}

	cc.Component.Circular = true

	var sb strings.Builder
	sb.Grow(48 * len("github.com/mavolin/corgi/v2/mycomponents/foo/bar.Baz\n"))
	for _, c := range slices.Backward(callerChain[2000:]) {
		sb.WriteByte('\n')
		sb.WriteString(string(c.File.Package.CorgiImportPath))
		sb.WriteByte('.')
		sb.WriteString(c.AST.Header.Name.Name)
	}

	logger.Error("AnalyzeComponentCall: Recursion depth exceeded")
	z.Report(&diagnostic.Diagnostic{
		Type:    diagnostic.InternalError,
		Message: "AnalyzeComponent: maximum recursion depth exceeded",
		Primary: []diagnostic.Annotation{
			anno.Node(cc.File, callerChain[0].AST, "while analyzing this component"),
		},
		Explanation: "Components are analyzed recursively, with a maximum recursion depth of 2048. " +
			"The component being analyzed either calls 2048 different components, exceeding this " +
			"maximum, or a component call cycle was undetected. " +
			"If there is a component call cycle that led to this, otherwise infinite, recursion " +
			"in the analyzer, it should've been caught elsewhere. " +
			"This is a bug, please report it.\n" +
			"And if you actually called 2048 different components, you ought to rethink what you are doing.\n" +
			sb.String(),
	})
}

type callerChainKey struct{}

// ============================================================================
// Receives Attributes
// ======================================================================================

func (z *analyzer) AnalyzeComponentCall_ReceivesAttributes_ReceivesAndPlaceholder(ctx context.Context, cc *file.ComponentCall) {
	defer analyzed.ComponentCall.ReceivesAndPlaceholder(z, cc)
	defer analyzed.ComponentCall.ReceivesAttributes(z, cc)

	cc.ReceivesAttributes.SetFalse()
	cc.ReceivesAndPlaceholder.SetFalse()

	if cc.AST.Header.Arguments != nil {
		z.analyzeComponentCall_ReceivesAttributes_ReceivesAndPlaceholder_throughArgs(cc)
		if cc.ReceivesAttributes.True() && cc.ReceivesAndPlaceholder.True() {
			return // we found both, no need to walk the body
		}
	}

	if cc.AST.Body == nil {
		return
	}

	var scope *ast.Scope
	switches.ComponentCallBody(cc.AST.Body,
		func(*ast.DefaultBlockShorthand) {},
		func(s *ast.Scope) { scope = s })
	if scope == nil {
		return
	}

	walk.Walk(scope, func(w *walk.Context) walk.Action {
		if aw, _ := w.Node.(ast.AttributeWriter); aw != nil {
			switches.AttributeWriter(aw,
				func(s *ast.ClassShorthand) {
					if !cc.ReceivesAttributes.True() {
						cc.ReceivesAttributes.SetReason(s)
					}
				},
				func(subCCAST *ast.ComponentCall) {
					subCC := cc.File.ComponentCallByNode(subCCAST)
					z.AnalyzeComponentCall(ctx, subCC)

					if !cc.ReceivesAttributes.True() {
						return
					}

					fa := get.ComponentCall.ForwardsAttributes(z, subCC)
					if fa.Equal(true) {
						cc.ReceivesAttributes.SetReason(subCC.AST)
					} else if fa.Failed() {
						cc.ReceivesAttributes.SetFailed()
					}
				},
				func(s *ast.IDShorthand) {
					if !cc.ReceivesAttributes.True() {
						cc.ReceivesAttributes.SetReason(s)
					}
				},
				func(na *ast.NamedAttribute) {
					if !cc.ReceivesAttributes.True() {
						cc.ReceivesAttributes.SetReason(na)
					}
				})
		}
		if apw, _ := w.Node.(ast.AndPlaceholderWriter); apw != nil {
			switches.AndPlaceholderWriter(apw,
				func(n *ast.AndPlaceholder) {
					if !cc.ReceivesAndPlaceholder.True() {
						cc.ReceivesAndPlaceholder.SetReason(n)
					}
				},
				func(subCCAST *ast.ComponentCall) {
					subCC := cc.File.ComponentCallByNode(subCCAST)
					z.AnalyzeComponentCall(ctx, subCC)

					if !cc.ReceivesAndPlaceholder.True() {
						return
					}

					fap := get.ComponentCall.ForwardsAndPlaceholder(z, subCC)
					if fap.Equal(true) {
						cc.ReceivesAndPlaceholder.SetReason(subCC.AST)
					} else if fap.Failed() {
						cc.ReceivesAndPlaceholder.SetFailed()
					}
				})
		}

		if cc.ReceivesAttributes.True() && cc.ReceivesAndPlaceholder.True() {
			return walk.Break
		}
		return walk.Continue
	}, walk.DontDive[*ast.ComponentCall](), walk.DontDive[ast.BlockSetter]())
}

func (z *analyzer) analyzeComponentCall_ReceivesAttributes_ReceivesAndPlaceholder_throughArgs(cc *file.ComponentCall) {
	for _, arg := range cc.AST.Header.Arguments.List {
		if aw, _ := arg.(ast.AttributeWriter); aw != nil {
			switches.AttributeWriter(aw,
				func(s *ast.ClassShorthand) {
					if !cc.ReceivesAttributes.True() {
						cc.ReceivesAttributes.SetReason(s)
					}
				},
				func(*ast.ComponentCall) {},
				func(s *ast.IDShorthand) {
					if !cc.ReceivesAttributes.True() {
						cc.ReceivesAttributes.SetReason(s)
					}
				},
				func(n *ast.NamedAttribute) {
					if !cc.ReceivesAttributes.True() {
						cc.ReceivesAttributes.SetReason(n)
					}
				})
		}

		if apw, _ := arg.(ast.AndPlaceholderWriter); apw != nil {
			switches.AndPlaceholderWriter(apw,
				func(n *ast.AndPlaceholder) {
					if !cc.ReceivesAndPlaceholder.True() {
						cc.ReceivesAndPlaceholder.SetReason(n)
					}
				},
				func(*ast.ComponentCall) {})
		}

		if cc.ReceivesAttributes.True() && cc.ReceivesAndPlaceholder.True() {
			return
		}
	}
}

// ============================================================================
// Elements With &-Placeholder
// ======================================================================================

func (z *analyzer) AnalyzeComponentCall_ElementsWithAndPlaceholder(cc *file.ComponentCall) {
	defer analyzed.ComponentCall.ElementsWithAndPlaceholder(z, cc)

	if cc.Component == nil {
		cc.ElementsWithAndPlaceholder.SetFailed()
		return
	}

	permanentElementsWithAndPlaceholder := get.Component.PermanentElementsWithAndPlaceholder(z, cc.Component)
	if permanentElementsWithAndPlaceholder.Failed() {
		cc.ElementsWithAndPlaceholder.SetFailed()
		return
	}

	var res []ast.AttributeReceiver

	if permanentElementsWithAndPlaceholder.Result().Len() > 0 {
		res = append(res, permanentElementsWithAndPlaceholder.Result().Get()...)
	}

	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil || get.BlockInstance.DefaultOverwritten(z, bi, cc) {
				continue
			}
			elementsWithAndPlaceholder := get.BlockInstance.Default.ElementsWithAndPlaceholder(z, bi)
			if elementsWithAndPlaceholder.Failed() {
				cc.ElementsWithAndPlaceholder.SetFailed()
				return
			}

			if get.BlockInstance.Default.ElementSpecsWithAndPlaceholder(z, bi).Result().Len() > 0 {
				res = append(res, elementsWithAndPlaceholder.Result().Get()...)
			}
		}
	}

	cc.ElementsWithAndPlaceholder.SetResult(file.SliceRefFrom(slices.Clip(res)))
}

// ============================================================================
// Element Specs With &-Placeholder
// ======================================================================================

func (z *analyzer) AnalyzeComponentCall_ElementSpecsWithAndPlaceholder(cc *file.ComponentCall) {
	defer analyzed.ComponentCall.ElementSpecsWithAndPlaceholder(z, cc)

	if cc.Component == nil {
		cc.ElementSpecsWithAndPlaceholder.SetFailed()
		return
	}

	// fast path
	elementsWithAndPlaceholder := get.ComponentCall.ElementsWithAndPlaceholder(z, cc)
	if elementsWithAndPlaceholder.Failed() {
		cc.ElementSpecsWithAndPlaceholder.SetFailed()
		return
	} else if elementsWithAndPlaceholder.Result().Len() == 0 {
		cc.ElementSpecsWithAndPlaceholder.SetResult(file.NilSliceRef[*file.ElementSpec]())
		return
	}

	specSet := make(map[*file.ElementSpec]struct{})

	permanentElementSpecsWithAndPlaceholder := get.Component.PermanentElementSpecsWithAndPlaceholder(z, cc.Component)
	for _, spec := range permanentElementSpecsWithAndPlaceholder.Result().Get() {
		specSet[spec] = struct{}{}
	}

	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil || get.BlockInstance.DefaultOverwritten(z, bi, cc) {
				continue
			}
			elementSpecsWithAndPlaceholder := get.BlockInstance.Default.ElementSpecsWithAndPlaceholder(z, bi)
			if elementSpecsWithAndPlaceholder.Failed() {
				cc.ElementSpecsWithAndPlaceholder.SetFailed()
				return
			}

			for _, spec := range elementSpecsWithAndPlaceholder.Result().Get() {
				specSet[spec] = struct{}{}
			}
		}
	}

	res := make([]*file.ElementSpec, 0, len(specSet))
	for spec := range specSet {
		res = append(res, spec)
	}

	cc.ElementSpecsWithAndPlaceholder.SetResult(file.SliceRefFrom(res))
}
