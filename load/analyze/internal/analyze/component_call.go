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

	z.AnalyzeCallComponent(cc)
	z.FindFirstTopLevelAttributeWriter(cc)
	z.FindFirstTopLevelAndPlaceholderWriter(cc)
	z.FindFirstDelegatedAttributes(ctx, cc)
	z.AnalyzeAcceptsAttributes(cc)
	z.AnalyzeForwardsDelegatedAttributes(cc)
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
			"This is a bug, please report it. " +
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

	ap := cc.Component.FirstPermanentTopLevelAndPlaceholder
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

			firstTopLevelAndPlaceholder := file.ConditionalAnalysis(instance.TopLevel, instance.Default.FirstTopLevelAndPlaceholder)
			if firstTopLevelAndPlaceholder.NotZero() {
				cc.ForwardsDelegatedAttributes.Set(true)
				return
			}
			failed = failed || firstTopLevelAndPlaceholder.Failed
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

	ap := cc.Component.FirstPermanentAndPlaceholder
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

			if instance.Default.FirstAndPlaceholder.NotZero() {
				cc.AcceptsAttributes.Set(true)
				return
			}
			failed = failed || instance.Default.FirstAndPlaceholder.Failed
		}
	}

	cc.AcceptsAttributes.SetIf(false, !failed)
}

// ============================================================================
// First Top-Level Attribute Writer
// ======================================================================================

// FindFirstTopLevelAttributeWriter finds the first top-level
// attribute writer in the component call.
// It prefers attribute writers inside the component call's component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.FirstTopLevelAttributeWriter
//
// Depends on Fields:
//   - ComponentCalls.Blocks.Instances.FirstTopLevelAttributeWriter
//   - ComponentCalls.Blocks.Instances.TopLevel
//   - ComponentCalls.FirstDelegatedAttributeWriter
//   - ComponentCalls.ForwardsDelegatedAttributes
func (z *analyzer) FindFirstTopLevelAttributeWriter(cc *file.ComponentCall) {
	if cc.Component == nil {
		cc.FirstTopLevelAttributeWriter.SetFailed()
		return
	}

	aw := cc.Component.FirstPermanentTopLevelAttributeWriter
	if aw.NotZero() {
		cc.FirstTopLevelAttributeWriter.Set(cc.AST)
		return
	}
	failed := cc.Component.FirstPermanentTopLevelAttributeWriter.Failed

	for _, block := range cc.Component.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil || instance.DefaultOverwritten(cc) {
				continue
			}

			topLevelAttr := file.ConditionalAnalysis(instance.TopLevel, instance.Default.FirstTopLevelAttributeWriter)
			if topLevelAttr.NotZero() {
				cc.FirstTopLevelAttributeWriter.Set(cc.AST)
				return
			}
			failed = failed || topLevelAttr.Failed
		}
	}

	topLevelAndPlaceholderFiller := file.ConditionalAnalysis(cc.ForwardsDelegatedAttributes, cc.FirstDelegatedAttributeWriter)
	if topLevelAndPlaceholderFiller.NotZero() {
		cc.FirstTopLevelAttributeWriter.Set(cc.AST)
		return
	}
	failed = failed || topLevelAndPlaceholderFiller.Failed

	for _, s := range cc.BlockSetters {
		topLevelBlockAttr := s.FirstTopLevelAttributeWriter()
		if s.Block == nil {
			failed = failed || topLevelBlockAttr.Failed || topLevelBlockAttr.Result != nil
			continue
		}

		topLevelAttr := file.ConditionalAnalysis(s.Block.TopLevel(file.AtLeastOne), topLevelBlockAttr)
		if topLevelAttr.NotZero() {
			cc.FirstTopLevelAttributeWriter.Set(cc.AST)
			return
		}
		failed = failed || topLevelAttr.Failed
	}

	cc.FirstTopLevelAttributeWriter.SetIf(nil, !failed)
}

// ============================================================================
// First Top-Level &-Placeholder
// ======================================================================================

// FindFirstTopLevelAndPlaceholderWriter finds the first &-placeholder that fills the
// &-placeholder of the called component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.FirstTopLevelAndPlaceholderWriter
//
// Depends on Fields:
//   - ComponentCalls.Blocks.TopLevel
//   - ComponentCalls.Blocks.Instances.FirstTopLevelAndPlaceholderWriter
//   - ComponentCalls.FirstDelegatedAndPlaceholderWriter
//   - ComponentCalls.ForwardsDelegatedAttributes
func (z *analyzer) FindFirstTopLevelAndPlaceholderWriter(cc *file.ComponentCall) {
	firstTopLevelAndPlaceholder := file.ConditionalAnalysis(cc.ForwardsDelegatedAttributes, cc.FirstDelegatedAndPlaceholderWriter)
	if firstTopLevelAndPlaceholder.NotZero() {
		cc.FirstTopLevelAndPlaceholderWriter.Set(cc.AST)
		return
	}

	failed := firstTopLevelAndPlaceholder.Failed
	for _, s := range cc.BlockSetters {
		topLevelBlockAndPlaceholder := s.FirstTopLevelAndPlaceholderWriter()
		if s.Block == nil {
			failed = failed || topLevelBlockAndPlaceholder.Failed || topLevelBlockAndPlaceholder.Result != nil
			continue
		}

		topLevelAndPlaceholder := file.ConditionalAnalysis(s.Block.TopLevel(file.AtLeastOne), topLevelBlockAndPlaceholder)
		if topLevelAndPlaceholder.NotZero() {
			cc.FirstTopLevelAttributeWriter.Set(cc.AST)
			return
		}
		failed = failed || topLevelAndPlaceholder.Failed
	}

	cc.FirstTopLevelAndPlaceholderWriter.SetIf(nil, !failed)
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

	walk.Walk(scope, func(wctx *walk.Context) error {
		switch n := wctx.Node.(type) {
		case *ast.AndPlaceholder:
			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return nil
			}

			cc.FirstDelegatedAndPlaceholderWriter.Set(n)
			if cc.FirstDelegatedAttributeWriter.NotZero() {
				return walk.Stop
			}
		case *ast.ComponentCall:
			subCC := cc.File.ComponentCallByNode(n)
			z.AnalyzeComponentCall(ctx, subCC)

			if !cc.FirstDelegatedAttributeWriter.NotZero() {
				if subCC.FirstTopLevelAttributeWriter.NotZero() {
					cc.FirstDelegatedAttributeWriter.Set(subCC.AST)
				} else if subCC.FirstTopLevelAttributeWriter.Failed {
					cc.FirstDelegatedAttributeWriter.SetFailed()
				}
			}

			if !cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				if subCC.FirstTopLevelAndPlaceholderWriter.NotZero() {
					cc.FirstDelegatedAndPlaceholderWriter.Set(subCC.AST)
				} else if subCC.FirstTopLevelAndPlaceholderWriter.Failed {
					cc.FirstDelegatedAndPlaceholderWriter.SetFailed()
				}
			}

			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() && subCC.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return walk.Stop
			}
			return walk.NoDive
		case ast.AttributeWriter:
			if cc.FirstDelegatedAttributeWriter.NotZero() {
				return nil
			}

			cc.FirstDelegatedAttributeWriter.Set(n)
			if cc.FirstDelegatedAndPlaceholderWriter.NotZero() {
				return walk.Stop
			}
		}
		return nil
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
