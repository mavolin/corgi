package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// ============================================================================
// And
// ======================================================================================

func (ch *checker) CheckAnd(logger *slog.Logger, f *file.File, _ []*walk.Context, a *ast.And) {
	logger = logger.WithGroup("and").
		With(slog.String("and_pos", a.Start().String()))
	logger.Debug("Checking and")

	ch.CheckAndNoEmptyAttributeList(logger, f, a)
	ch.CheckAndContainsOnlyAttributes(logger, f, a)
}

func (ch *checker) CheckAndNoEmptyAttributeList(logger *slog.Logger, f *file.File, a *ast.And) {
	logger = logger.WithGroup("no_empty_attribute_list")
	logger.Debug("Checking for empty attribute list")

	if a.Attributes == nil {
		logger.Debug("No attribute list, skipping")
		return
	} else if len(a.Attributes.Args) > 0 {
		logger.Debug("At least one attribute, skipping")
		return
	}

	logger.Error("Empty attribute list")
	ch.Report(&diagnostic.Diagnostic{
		Message: "and: empty attribute list",
		Primary: []diagnostic.Annotation{
			anno.Node(f, a.Attributes, "remove this empty attribute list"),
		},
		Hints: []diagnostic.Hint{
			{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
		},
	})
}

func (ch *checker) CheckAndContainsOnlyAttributes(logger *slog.Logger, f *file.File, a *ast.And) {
	logger = logger.WithGroup("contains_only_attributes")
	logger.Debug("Checking that the arguments contain only attributes")

	if a.Attributes == nil {
		logger.Debug("No attributes, skipping")
		return
	}

	for _, arg := range a.Attributes.Args {
		logger := logger.With(slog.String("pos", arg.Start().String()))
		logger.Debug("Checking argument")

		if _, ok := arg.(ast.Attribute); ok {
			continue
		}

		logger.Error("Attribute list contains non-attribute")
		ch.Report(&diagnostic.Diagnostic{
			Message: "and: attribute list contains non-attribute",
			Primary: []diagnostic.Annotation{
				anno.Node(f, arg, "this is not an attribute"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
			},
		})
	}
}

// ============================================================================
// Attribute Type
// ======================================================================================

func (ch *checker) CheckAttributeType(logger *slog.Logger, f *file.File, parents []*walk.Context, t *ast.AttributeType) {
	logger = logger.WithGroup("attribute_type").
		With(slog.String("attribute_type", t.Name.Name),
			slog.String("type_pos", t.Start().String()))
	logger.Debug("Checking attribute type")

	ch.CheckAttributeTypeSuperfluousAttributeName(logger, parents, f, t)
}

func (ch *checker) CheckAttributeTypeSuperfluousAttributeName(logger *slog.Logger, parents []*walk.Context, f *file.File, t *ast.AttributeType) {
	logger = logger.WithGroup("superfluous_attribute_name")
	logger.Debug("Checking for superfluous attribute name")

	if t.Name == nil {
		logger.Debug("No attribute name, skipping")
		return
	}

	if t.Name.Type != attrtype.Unsafe && t.Name.Type != attrtype.UnsafeBool {
		logger.Error("Attribute name on non-unsafe type")
		ch.Report(&diagnostic.Diagnostic{
			Message: "attribute type: attribute name on non-unsafe type",
			Primary: []diagnostic.Annotation{
				anno.Node(f, t.Name, "remove this attribute name"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
			},
			Explanation: "Attribute types other than `unsafe` and `unsafeBool` " +
				"do not need to be tied to a specific attribute.",
		})
	} else if walk.IsChildOf[*ast.NamedAttribute](parents) {
		logger.Error("Attribute name on non-unsafe type")
		ch.Report(&diagnostic.Diagnostic{
			Message: "attribute type: superfluous attribute name",
			Primary: []diagnostic.Annotation{
				anno.Node(f, t.Name, "remove this attribute name"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
			},
			Explanation: "Attribute names need not be specified a second time " +
				"in brackets when writing a typed named attribute.",
		})
	}
}

// ============================================================================
// Block Function
// ======================================================================================

func (ch *checker) CheckBlockFunction(logger *slog.Logger, f *file.File, parents []*walk.Context, bf *ast.BlockFunction) {
	logger = logger.WithGroup("block_function").
		With(slog.String("block_name", bf.BlockName.Ident),
			slog.String("block_function_pos", bf.Start().String()))
	logger.Debug("Checking block function")

	ch.CheckBlockFunctionDefined(logger, f, parents, bf)
}

func (ch *checker) CheckBlockFunctionDefined(logger *slog.Logger, f *file.File, parents []*walk.Context, bf *ast.BlockFunction) {
	logger = logger.WithGroup("block_defined")
	logger.Debug("Checking block function only uses blocks defined by the component")

	astComp, _ := parents[0].Node.(*ast.Component) // todo: aliaz
	if astComp == nil {
		logger.Debug("Not a component, skipping")
		return
	}

	c := f.Package.ComponentByNode(astComp)
	if c.BlockByName(bf.BlockName.Ident) != nil {
		return
	}

	logger.Error("Block function uses block not defined by the component")
	ch.Report(&diagnostic.Diagnostic{
		Message: "block function: reference to undefined block",
		Primary: []diagnostic.Annotation{
			anno.Node(f, bf.BlockName, "this block is never used in this component"),
		},
	})
}

// ============================================================================
// Break
// ======================================================================================

func (ch *checker) CheckBreak(logger *slog.Logger, f *file.File, parents []*walk.Context, b *ast.Break) {
	logger = logger.WithGroup("break").
		With(slog.String("break_pos", b.Start().String()))
	logger.Debug("Checking break")

	ch.CheckBreakAllowed(logger, parents, f, b)
}

func (ch *checker) CheckBreakAllowed(logger *slog.Logger, parents []*walk.Context, f *file.File, b *ast.Break) {
	logger = logger.WithGroup("break_allowed")
	logger.Debug("Checking that break is in a loop or switch")

	if !walk.IsChildOf[*ast.For](parents) && !walk.IsChildOf[*ast.Switch](parents) {
		logger.Error("Break not in loop or switch")
		ch.Report(&diagnostic.Diagnostic{
			Message: "break used outside of loop and switch",
			Primary: []diagnostic.Annotation{
				anno.Node(f, b, "`break`s can only be used inside loops or switch statements"),
			},
		})
	}
}

// ============================================================================
// Continue
// ======================================================================================

func (ch *checker) CheckContinue(logger *slog.Logger, f *file.File, parents []*walk.Context, c *ast.Continue) {
	logger = logger.WithGroup("continue").
		With(slog.String("continue_pos", c.Start().String()))
	logger.Debug("Checking continue")

	ch.CheckContinueInLoop(logger, parents, f, c)
}

func (ch *checker) CheckContinueInLoop(logger *slog.Logger, parents []*walk.Context, f *file.File, c *ast.Continue) {
	logger = logger.WithGroup("continue_in_loop")
	logger.Debug("Checking that continue is in a loop")

	if !walk.IsChildOf[*ast.For](parents) {
		logger.Error("Continue not in loop")
		ch.Report(&diagnostic.Diagnostic{
			Message: "continue used outside of loop",
			Primary: []diagnostic.Annotation{
				anno.Node(f, c, "`continue`s can only be used inside loops"),
			},
		})
	}
}

// ============================================================================
// Fallthrough
// ======================================================================================

func (ch *checker) CheckFallthrough(logger *slog.Logger, f *file.File, parents []*walk.Context, ft *ast.Fallthrough) {
	logger = logger.WithGroup("fallthrough").
		With(slog.String("fallthrough_pos", ft.Start().String()))
	logger.Debug("Checking fallthrough")

	ch.CheckFallthroughInSwitch(logger, parents, f, ft)
}

func (ch *checker) CheckFallthroughInSwitch(logger *slog.Logger, parents []*walk.Context, f *file.File, ft *ast.Fallthrough) {
	logger = logger.WithGroup("fallthrough_in_switch")
	logger.Debug("Checking that fallthrough is in a switch")

	if !walk.IsChildOf[*ast.Switch](parents) {
		logger.Error("Fallthrough not in switch")
		ch.Report(&diagnostic.Diagnostic{
			Message: "fallthrough used outside of switch",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ft, "`fallthrough`s can only be used inside switch statements"),
			},
		})
	}
}
