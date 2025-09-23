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

// AnalyzeComponentCalls analyzes the remaining component calls in the package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponentCalls() {
	logger := z.Logger.WithGroup("component_calls")
	logger.Debug("Analyzing remaining component calls")

	ctx := context.Background()
	for _, f := range z.P.Files {
		for _, cc := range f.ComponentCalls {
			z.AnalyzeComponentCall(ctx, cc)
		}
	}
}

// AnalyzeComponentCall analyzes the passed component call.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
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

	z.AnalyzeCallComponent(ctx, cc)
	z.AnalyzeReceivesAttributes(ctx, cc)

	z.AnalyzeComponentArguments(cc)

	cc.Analyzed = true
}

func (z *analyzer) checkNoInfiniteRecursion(logger *slog.Logger, cc *file.ComponentCall, callerChain []*file.Component) {
	if len(callerChain) < 2048 {
		return
	}

	cc.Circular = true

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

// AnalyzeReceivesAttributes finds the first attribute writer and the first
// &-placeholder writer that fills the &-placeholder of the called component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.ReceivesAttributes
//   - ComponentCalls.ReceivesAndPlaceholder
//
// Depends on Fields: None
func (z *analyzer) AnalyzeReceivesAttributes(ctx context.Context, cc *file.ComponentCall) {
	cc.ReceivesAttributes.SetFalse()
	cc.ReceivesAndPlaceholder.SetFalse()

	if cc.AST.Header.Arguments != nil {
		z.analyzeReceivedAttributesInArgs(cc)
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

					fa := subCC.ForwardsAttributes()
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

					fap := subCC.ForwardsAndPlaceholder()
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

func (z *analyzer) analyzeReceivedAttributesInArgs(cc *file.ComponentCall) {
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

// AnalyzeElementWithAndPlaceholder sets the ElementWithAndPlaceholder field of the passed
// component call.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.ElementWithAndPlaceholder
//
// Depends on Fields: None
func (z *analyzer) AnalyzeElementWithAndPlaceholder(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.ElementsWithAndPlaceholder.SetFailed()
		return
	} else if cc.Component.PermanentElementsWithAndPlaceholder.Failed() {
		cc.ElementsWithAndPlaceholder.SetFailed()
		return
	}

	var res []ast.AttributeReceiver

	if cc.Component.PermanentElementsWithAndPlaceholder.Result().Len() > 0 {
		res = append(res, cc.Component.PermanentElementsWithAndPlaceholder.Result().Get()...)
	}

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil || instance.DefaultOverwritten(cc) {
				continue
			} else if instance.Default.ElementsWithAndPlaceholder.Failed() {
				cc.ElementsWithAndPlaceholder.SetFailed()
				return
			}

			if instance.Default.ElementSpecsWithAndPlaceholder.Result().Len() > 0 {
				res = append(res, instance.Default.ElementsWithAndPlaceholder.Result().Get()...)
			}
		}
	}

	cc.ElementsWithAndPlaceholder.SetResult(file.SliceRefFrom(slices.Clip(res)))
}

// ============================================================================
// Element Specs With &-Placeholder
// ======================================================================================

// AnalyzeCallComponentElementSpecsWithAndPlaceholder sets the
// ElementSpecsWithAndPlaceholder field of the passed component call.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.ElementSpecsWithAndPlaceholder
//
// Depends on Fields: None
func (z *analyzer) AnalyzeCallComponentElementSpecsWithAndPlaceholder(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.ElementSpecsWithAndPlaceholder.SetFailed()
		return
	}

	// fast path
	if cc.ElementsWithAndPlaceholder.Failed() {
		cc.ElementSpecsWithAndPlaceholder.SetFailed()
		return
	} else if cc.ElementsWithAndPlaceholder.Result().Len() == 0 {
		cc.ElementSpecsWithAndPlaceholder.SetResult(file.NilSliceRef[*file.ElementSpec]())
		return
	}

	specSet := make(map[*file.ElementSpec]struct{})

	for _, spec := range cc.Component.PermanentElementSpecsWithAndPlaceholder.Result().Get() {
		specSet[spec] = struct{}{}
	}

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil || instance.DefaultOverwritten(cc) {
				continue
			} else if instance.Default.ElementSpecsWithAndPlaceholder.Failed() {
				cc.ElementSpecsWithAndPlaceholder.SetFailed()
				return
			}

			for _, spec := range instance.Default.ElementSpecsWithAndPlaceholder.Result().Get() {
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
