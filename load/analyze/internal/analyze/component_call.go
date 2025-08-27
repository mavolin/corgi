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
	z.AnalyzeComponentForwardsAttributes(cc)
	z.AnalyzeReceivesAttributes(ctx, cc)
	z.AnalyzeAcceptsAttributes(cc)
	z.AnalyzeForwardsReceivedAttributes(cc)

	cc.Analyzed = true
}

func (z *analyzer) checkNoInfiniteRecursion(logger *slog.Logger, cc *file.ComponentCall, callerChain []*file.Component) {
	if len(callerChain) < 2048 {
		return
	}

	var sb strings.Builder
	sb.Grow(48 * len("github.com/mavolin/corgi/v2/mycomponents/foo/bar.Baz\n"))
	for _, c := range slices.Backward(callerChain[2000:]) {
		sb.WriteByte('\n')
		sb.WriteString(c.File.Package.ImportPath)
		sb.WriteByte('.')
		sb.WriteString(c.AST.Header.Name.Name)
	}

	logger.Error("AnalyzeComponentCall: Recursion depth exceeded")
	z.Report(&diagnostic.Diagnostic{
		Type:    diagnostic.InternalError,
		Message: "AnalyzeComponent: recursion depth exceeded",
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
// Forwards Received Attributes
// ======================================================================================

// AnalyzeForwardsReceivedAttributes analyzes whether the component call
// forwards received attributes.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.ForwardsReceivedAttributes
//
// Depends on Fields: None
func (z *analyzer) AnalyzeForwardsReceivedAttributes(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.ForwardsReceivedAttributes.SetFailed()
		return
	}

	if cc.Component.CouldForwardReceivedAttributes.False() {
		cc.ForwardsReceivedAttributes.SetFalse()
		return
	}

	if cc.Component.AlwaysForwardsAndPlaceholder.True() {
		cc.ForwardsReceivedAttributes.SetReason(cc.Component.AlwaysForwardsAndPlaceholder.Reason())
		return
	}

	cc.ForwardsReceivedAttributes.SetFalse()
	if cc.Component.AlwaysForwardsAndPlaceholder.Failed() {
		cc.ForwardsReceivedAttributes.SetFailed()
	}

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil {
				continue
			} else if instance.DefaultOverwritten(cc) {
				continue
			}

			forwardsAndPlaceholder := file.ConditionalAnalysis(instance.Forwarded, instance.Default.ForwardsAndPlaceholder)
			if forwardsAndPlaceholder.True() {
				cc.ForwardsReceivedAttributes.SetReason(instance.Default.ForwardsAndPlaceholder.Reason())
				return
			} else if forwardsAndPlaceholder.Failed() {
				cc.ForwardsReceivedAttributes.SetFailed()
			}
		}
	}
}

// ============================================================================
// Accepts Attributes
// ======================================================================================

// AnalyzeAcceptsAttributes analyzes whether the call's component accepts
// attributes.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.AcceptsAttributes
//
// Depends on Fields:
//   - ComponentCalls.ForwardsReceivedAttributes
func (z *analyzer) AnalyzeAcceptsAttributes(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.AcceptsAttributes.SetFailed()
		return
	}

	if cc.Component.CouldAcceptAttributes.False() {
		cc.AcceptsAttributes.SetFalse()
		return
	}

	if cc.ForwardsReceivedAttributes.True() {
		cc.AcceptsAttributes.SetReason(cc.ForwardsReceivedAttributes.Reason())
		return
	} else if cc.Component.AlwaysWritesAndPlaceholder.True() {
		cc.AcceptsAttributes.SetReason(cc.Component.AlwaysWritesAndPlaceholder.Reason())
		return
	}

	cc.AcceptsAttributes.SetFalse()
	if cc.Component.AlwaysWritesAndPlaceholder.Failed() {
		cc.AcceptsAttributes.SetFailed()
	}

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil {
				continue
			} else if instance.DefaultOverwritten(cc) {
				continue
			}

			if instance.Default.WritesAndPlaceholder.True() {
				cc.AcceptsAttributes.SetReason(instance.Default.WritesAndPlaceholder.Reason())
				return
			} else if instance.Default.WritesAndPlaceholder.Failed() {
				cc.AcceptsAttributes.SetFailed()
			}
		}
	}
}

// ============================================================================
// Component Forwards Attributes
// ======================================================================================

// AnalyzeComponentForwardsAttributes finds the first top-level
// attribute writer in the component call.
// It prefers attribute writers inside the component call's component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.ComponentForwardsAttributes
//
// Depends on Fields:
//   - ComponentCalls.ForwardsReceivedAttributes
//   - ComponentCalls.ReceivesAttributes
//   - ComponentCalls.BlockSetters.Instances.ForwardsAttributes
func (z *analyzer) AnalyzeComponentForwardsAttributes(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.ComponentForwardsAttributes.SetFailed()
		return
	}

	if cc.Component.AlwaysForwardsAttributes.True() {
		cc.ComponentForwardsAttributes.SetReason(cc.Component.AlwaysForwardsAttributes.Reason())
		return
	}

	cc.ComponentForwardsAttributes.SetFalse()
	if cc.Component.AlwaysForwardsAttributes.Failed() {
		cc.ComponentForwardsAttributes.SetFailed()
	}

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil || instance.DefaultOverwritten(cc) {
				continue
			}

			forwardedAttr := file.ConditionalAnalysis(instance.Forwarded, instance.Default.ForwardsAttributes)
			if forwardedAttr.True() {
				cc.ComponentForwardsAttributes.SetReason(forwardedAttr.Reason())
				return
			} else if forwardedAttr.Failed() {
				cc.ComponentForwardsAttributes.SetFailed()
			}
		}
	}
}

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
	scope, _ := cc.AST.Body.(*ast.Scope)
	if scope == nil {
		return
	}

	walk.Walk(scope, func(w *walk.Context) walk.Action {
		switch n := w.Node.(type) {
		case *ast.AndPlaceholder:
			if cc.ReceivesAndPlaceholder.True() {
				return walk.Continue
			}

			cc.ReceivesAndPlaceholder.SetReason(n)
			if cc.ReceivesAttributes.True() {
				return walk.Break
			}
		case *ast.ComponentCall:
			subCC := cc.File.ComponentCallByNode(n)
			z.AnalyzeComponentCall(ctx, subCC)

			if !cc.ReceivesAttributes.True() {
				fa := subCC.ForwardsAttributes()
				if fa.Equal(true) {
					cc.ReceivesAttributes.SetReason(subCC.AST)
				} else if fa.Failed() {
					cc.ReceivesAttributes.SetFailed()
				}
			}

			if !cc.ReceivesAndPlaceholder.True() {
				fap := subCC.ForwardsAndPlaceholder()
				if fap.Equal(true) {
					cc.ReceivesAndPlaceholder.SetReason(subCC.AST)
				} else if fap.Failed() {
					cc.ReceivesAndPlaceholder.SetFailed()
				}
			}

			if cc.ReceivesAndPlaceholder.True() && subCC.ReceivesAndPlaceholder.True() {
				return walk.Break
			}
			return walk.NoDive
		case ast.AttributeWriter:
			if cc.ReceivesAttributes.True() {
				return walk.Continue
			}

			cc.ReceivesAttributes.SetReason(n)
			if cc.ReceivesAndPlaceholder.True() {
				return walk.Break
			}
		}
		return walk.Continue
	}, walk.DontDive[ast.BlockSetter]())
}

func (z *analyzer) analyzeReceivedAttributesInArgs(cc *file.ComponentCall) {
	for _, arg := range cc.AST.Header.Arguments.List {
		switch n := arg.(type) {
		case *ast.AndPlaceholder:
			if cc.ReceivesAndPlaceholder.True() {
				continue
			}

			cc.ReceivesAndPlaceholder.SetReason(n)
			if cc.ReceivesAttributes.True() {
				return
			}
		case ast.AttributeWriter:
			if cc.ReceivesAttributes.True() {
				continue
			}

			cc.ReceivesAttributes.SetReason(n)
			if cc.ReceivesAndPlaceholder.True() {
				return
			}
		}
	}
}
