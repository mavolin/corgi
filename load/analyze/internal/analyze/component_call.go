package analyze

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

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
		z.secondaryCircularComponentCallCheck(logger, cc, callerChain)
		if len(callerChain) == 0 {
			callerChain = make([]*file.Component, 0, 1024)
		}
		callerChain = append(callerChain, cc.Component)
		ctx = context.WithValue(ctx, callerChainKey{}, callerChain)
	}

	z.ComponentCallFindFirstDelegatedAttributes(ctx, logger, cc)
}

func (z *analyzer) secondaryCircularComponentCallCheck(logger *slog.Logger, cc *file.ComponentCall, callerChain []*file.Component) {
	if cc.Circular {
		return
	}

	for i, c := range slices.Backward(callerChain) {
		if c != cc.Component {
			continue
		}
		cc.Circular = true

		logger.Error("Undetected circular component call")

		secondaries := make([]diagnostic.Annotation, len(callerChain)-i)
		for i, c := range callerChain[i:] {
			secondaries[i] = anno.Node(c.File, c.AST.Header.Name, fmt.Sprint(i+1, ": `", c.AST.Header.Name.Name, "`"))
		}
		z.Report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "AnalyzeComponentCall: undetected circular component call",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST, "circular component call"),
			},
			Secondary: secondaries,
			Explanation: "Although this does not affect the correctness of the program, " +
				"this should've been caught earlier.\n" +
				"\n" +
				"This is a bug in the analyzer, please open an issue and report it.",
		})
	}
}

type callerChainKey struct{}

// ComponentCallFindFirstDelegatedAttributes finds the first attribute
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
func (z *analyzer) ComponentCallFindFirstDelegatedAttributes(ctx context.Context, _ *slog.Logger, cc *file.ComponentCall) {
	cc.FirstDelegatedAttributeWriter.SetZero()
	cc.FirstDelegatedAndPlaceholderWriter.SetZero()

	if cc.AST.Header.Arguments != nil {
		z.componentCallFindFirstDelegatedAttributesInArgs(cc)
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

			if cc.FirstDelegatedAttributeWriter.Failed || cc.FirstDelegatedAndPlaceholderWriter.Result == nil {
				writesTopLevelAttributes := subCC.WritesTopLevelAttributes()
				if writesTopLevelAttributes.Equal(true) {
					cc.FirstDelegatedAttributeWriter.Set(n)
				} else if writesTopLevelAttributes.Failed {
					cc.FirstDelegatedAttributeWriter.SetFailed()
				}
			}

			if cc.FirstDelegatedAndPlaceholderWriter.Failed || cc.FirstDelegatedAndPlaceholderWriter.Result == nil {
				forwardsTopLevelAndPlaceholder := subCC.ForwardsTopLevelAndPlaceholder()
				if forwardsTopLevelAndPlaceholder.Equal(true) {
					cc.FirstDelegatedAndPlaceholderWriter.Set(n)
				} else if forwardsTopLevelAndPlaceholder.Failed {
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

func (z *analyzer) componentCallFindFirstDelegatedAttributesInArgs(cc *file.ComponentCall) {
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
