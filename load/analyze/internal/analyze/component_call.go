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
	z.FindFirstForwardedAttributeWriter(cc)
	z.FindFirstForwardedAndPlaceholderWriter(cc)
	z.FindFirstDelegatedAttributes(ctx, cc)
	z.AnalyzeAcceptsAttributes(cc)
	z.AnalyzeForwardsDelegatedAttributes(cc)

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
// Forwards Delegated Attributes
// ======================================================================================

// AnalyzeForwardsDelegatedAttributes analyzes whether the component call
// forwards delegated attributes.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.ForwardsDelegatedAttributes
//
// Depends on Fields: None
func (z *analyzer) AnalyzeForwardsDelegatedAttributes(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.ForwardsDelegatedAttributes.SetFailed()
		return
	}

	if cc.Component.CouldForwardAttributes.Equal(false) {
		cc.ForwardsDelegatedAttributes.Set(false)
		return
	}

	ap := cc.Component.FirstPermanentForwardedAndPlaceholderWriter
	if ap.NotZero() {
		cc.ForwardsDelegatedAttributes.Set(true)
		return
	}

	failed := ap.Failed
	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil {
				continue
			} else if instance.DefaultOverwritten(cc) {
				continue
			}

			firstForwardedAndPlaceholder := file.ConditionalAnalysis(instance.Forwarded, instance.Default.FirstForwardedAndPlaceholderWriter)
			if firstForwardedAndPlaceholder.NotZero() {
				cc.ForwardsDelegatedAttributes.Set(true)
				return
			}
			failed = failed || firstForwardedAndPlaceholder.Failed
		}
	}

	cc.ForwardsDelegatedAttributes.SetIf(false, !failed)
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
//   - ComponentCalls.ForwardsDelegatedAttributes
func (z *analyzer) AnalyzeAcceptsAttributes(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.AcceptsAttributes.SetFailed()
		return
	}

	if cc.ForwardsDelegatedAttributes.NotZero() {
		cc.AcceptsAttributes.Set(true)
		return
	} else if cc.Component.CouldAcceptAttributes.Equal(false) {
		cc.AcceptsAttributes.Set(false)
		return
	}

	ap := cc.Component.FirstPermanentAndPlaceholderWriter
	if ap.NotZero() {
		cc.AcceptsAttributes.Set(true)
		return
	}

	failed := ap.Failed
	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil {
				continue
			} else if instance.DefaultOverwritten(cc) {
				continue
			}

			if instance.Default.FirstAndPlaceholderWriter.NotZero() {
				cc.AcceptsAttributes.Set(true)
				return
			}
			failed = failed || instance.Default.FirstAndPlaceholderWriter.Failed
		}
	}

	cc.AcceptsAttributes.SetIf(false, !failed)
}

// ============================================================================
// First Top-Level Attribute Writer
// ======================================================================================

// FindFirstForwardedAttributeWriter finds the first top-level
// attribute writer in the component call.
// It prefers attribute writers inside the component call's component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.FirstForwardedAttributeWriter
//
// Depends on Fields:
//   - ComponentCalls.ForwardsDelegatedAttributes
//   - ComponentCalls.FirstDelegatedAttributeWriter
//   - ComponentCalls.BlockSetters.Instances.FirstForwardedAttributeWriter
func (z *analyzer) FindFirstForwardedAttributeWriter(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.FirstForwardedAttributeWriter.SetFailed()
		return
	}

	aw := cc.Component.FirstPermanentForwardedAttributeWriter
	if aw.NotZero() {
		cc.FirstForwardedAttributeWriter.Set(cc.AST)
		return
	}
	failed := cc.Component.FirstPermanentForwardedAttributeWriter.Failed

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil || instance.DefaultOverwritten(cc) {
				continue
			}

			forwardedAttr := file.ConditionalAnalysis(instance.Forwarded, instance.Default.FirstForwardedAttributeWriter)
			if forwardedAttr.NotZero() {
				cc.FirstForwardedAttributeWriter.Set(cc.AST)
				return
			}
			failed = failed || forwardedAttr.Failed
		}
	}

	forwardedDelegatedAttributeWriter := file.ConditionalAnalysis(cc.ForwardsDelegatedAttributes, cc.FirstDelegatedAttributeWriter)
	if forwardedDelegatedAttributeWriter.NotZero() {
		cc.FirstForwardedAttributeWriter.Set(cc.AST)
		return
	}
	failed = failed || forwardedDelegatedAttributeWriter.Failed

	for _, s := range cc.BlockSetters {
		forwardedBlockAttr := s.FirstForwardedAttributeWriter()
		if s.Block == nil {
			failed = failed || forwardedBlockAttr.Failed || forwardedBlockAttr.Result != nil
			continue
		}

		forwardedAttr := file.ConditionalAnalysis(s.Block.Forwarded, forwardedBlockAttr)
		if forwardedAttr.NotZero() {
			cc.FirstForwardedAttributeWriter.Set(cc.AST)
			return
		}
		failed = failed || forwardedAttr.Failed
	}

	cc.FirstForwardedAttributeWriter.SetIf(nil, !failed)
}

// ============================================================================
// First Top-Level &-Placeholder
// ======================================================================================

// FindFirstForwardedAndPlaceholderWriter finds the first &-placeholder that fills the
// &-placeholder of the called component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.FirstForwardedAndPlaceholderWriter
//
// Depends on Fields:
//   - ComponentCalls.ForwardsDelegatedAttributes
//   - ComponentCalls.FirstDelegatedAndPlaceholderWriter
//   - ComponentCalls.BlockSetters.Instances.FirstForwardedAndPlaceholderWriter
func (z *analyzer) FindFirstForwardedAndPlaceholderWriter(cc *file.ComponentCall) {
	firstForwardedAndPlaceholder := file.ConditionalAnalysis(cc.ForwardsDelegatedAttributes, cc.FirstDelegatedAndPlaceholderWriter)
	if firstForwardedAndPlaceholder.NotZero() {
		cc.FirstForwardedAndPlaceholderWriter.Set(cc.AST)
		return
	}

	failed := firstForwardedAndPlaceholder.Failed
	for _, s := range cc.BlockSetters {
		forwardedBlockAndPlaceholder := s.FirstForwardedAndPlaceholderWriter()
		if s.Block == nil {
			failed = failed || forwardedBlockAndPlaceholder.Failed || forwardedBlockAndPlaceholder.Result != nil
			continue
		}

		forwardedAndPlaceholder := file.ConditionalAnalysis(s.Block.Forwarded, forwardedBlockAndPlaceholder)
		if forwardedAndPlaceholder.NotZero() {
			cc.FirstForwardedAttributeWriter.Set(cc.AST)
			return
		}
		failed = failed || forwardedAndPlaceholder.Failed
	}

	cc.FirstForwardedAndPlaceholderWriter.SetIf(nil, !failed)
}

// ============================================================================
// First Delegated Attributes
// ======================================================================================

// FindFirstDelegatedAttributes finds the first attribute
// writer and the first &-placeholder that fills the &-placeholder of the
// called component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.FirstDelegatedAttributeWriter
//   - ComponentCalls.FirstDelegatedAndPlaceholderWriter
//
// Depends on Fields: None
func (z *analyzer) FindFirstDelegatedAttributes(ctx context.Context, cc *file.ComponentCall) {
	cc.FirstDelegatedAttributeWriter.SetZero()
	cc.FirstDelegatedAndPlaceholderWriter.SetZero()

	if cc.AST.Header.Arguments != nil {
		z.findFirstDelegatedAttributesInArgs(cc)
		if cc.FirstDelegatedAttributeWriter.NotZero() && cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
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
			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return walk.Continue
			}

			cc.FirstDelegatedAndPlaceholderWriter.Set(n)
			if cc.FirstDelegatedAttributeWriter.NotZero() {
				return walk.Break
			}
		case *ast.ComponentCall:
			subCC := cc.File.ComponentCallByNode(n)
			z.AnalyzeComponentCall(ctx, subCC)

			if !cc.FirstDelegatedAttributeWriter.NotZero() {
				if subCC.FirstForwardedAttributeWriter.NotZero() {
					cc.FirstDelegatedAttributeWriter.Set(subCC.AST)
				} else if subCC.FirstForwardedAttributeWriter.Failed {
					cc.FirstDelegatedAttributeWriter.SetFailed()
				}
			}

			if !cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				if subCC.FirstForwardedAndPlaceholderWriter.NotZero() {
					cc.FirstDelegatedAndPlaceholderWriter.Set(subCC.AST)
				} else if subCC.FirstForwardedAndPlaceholderWriter.Failed {
					cc.FirstDelegatedAndPlaceholderWriter.SetFailed()
				}
			}

			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() && subCC.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return walk.Break
			}
			return walk.NoDive
		case ast.AttributeWriter:
			if cc.FirstDelegatedAttributeWriter.NotZero() {
				return walk.Continue
			}

			cc.FirstDelegatedAttributeWriter.Set(n)
			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return walk.Break
			}
		}
		return walk.Continue
	}, walk.DontDive[ast.BlockSetter]())
}

func (z *analyzer) findFirstDelegatedAttributesInArgs(cc *file.ComponentCall) {
	for _, arg := range cc.AST.Header.Arguments.List {
		switch n := arg.(type) {
		case *ast.AndPlaceholder:
			if cc.FirstDelegatedAndPlaceholderWriter.Result != nil {
				continue
			}

			cc.FirstDelegatedAndPlaceholderWriter.Set(n)
			if cc.FirstDelegatedAttributeWriter.NotZero() {
				return
			}
		case ast.AttributeWriter:
			if cc.FirstDelegatedAttributeWriter.Result != nil {
				continue
			}

			cc.FirstDelegatedAttributeWriter.Set(n)
			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return
			}
		}
	}
}
