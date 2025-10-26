package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
)

func (z *analyzer) AnalyzeElementSpecs() {
	logger := z.Logger.WithGroup("element_specs")
	logger.Debug("Analyzing element specs")

	z.CheckElementSpecs_Cycles(logger)

	for _, spec := range z.Pkg.ElementSpecs {
		logger := logger.With(
			slog.String("file", string(spec.File.Name)),
			slog.String("name", spec.StylizedHTMLName),
			slog.String("pos", spec.AST.Start().String()))

		z.AnalyzeElementSpec(logger, spec)

		spec.Analyzed = true
	}
}

// AnalyzeElementSpec analyzes the given element spec.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeElementSpec(logger *slog.Logger, spec *file.ElementSpec) {
	z.AnalyzeElementSpec_Type(logger, spec)
}

// ============================================================================
// Check Cycles
// ======================================================================================

// CheckElementSpecs_Cycles checks for element spec cycles in the given element
// spec.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckElementSpecs_Cycles(logger *slog.Logger) {
	logger = logger.WithGroup("cycles")

	chain := make([]*file.ElementSpec, 1, 24)
	for _, spec := range z.Pkg.ElementSpecs {
		chain[0] = spec
		chain = chain[:1] // reset the chain
		z.checkElementSpecs_Cycles(logger, chain)
	}
}

func (z *analyzer) checkElementSpecs_Cycles(logger *slog.Logger, chain []*file.ElementSpec) {
	spec := chain[len(chain)-1]
	if spec.Circular {
		return
	}

	t, _ := spec.AST.Type.(*ast.AliasElementType)
	if t == nil {
		return
	}

	ref := spec.File.ElementReferenceByNode(t.Name)
	if ref == nil {
		return
	}

	if len(chain) == 1 || chain[0] != ref.Spec {
		chain = append(chain, ref.Spec)
		z.checkElementSpecs_Cycles(logger, chain)
		return
	}

	// We have a cycle

	for _, spec := range chain {
		spec.Circular = true
	}

	primaries := make([]diagnostic.Annotation, len(chain))
	for i, spec := range chain {
		primaries[i] = anno.Node(spec.File, spec.AST.Type, "references itself")
	}

	logger.Error("Found element definition cycle",
		slog.String("name", spec.StylizedHTMLName),
		slog.String("file", string(spec.File.Name)),
		slog.String("pos", spec.AST.Start().String()))
	z.Report(&diagnostic.Diagnostic{
		Message: "element definition cycle",
		Primary: primaries,
		Explanation: "This element defines itself through an alias that references itself.\n" +
			"To fix this error, change the definition of at least one of the elements in the cycle.",
	})
}

// ============================================================================
// Type
// ======================================================================================

// AnalyzeElementSpec_Type analyzes the type of the given element spec.
//
// Depends on Checks:
//   - CheckElementSpecs_Cycles
//
// Sets Fields:
//   - ElementSpec.Type
//
// Depends on Fields: None
func (z *analyzer) AnalyzeElementSpec_Type(logger *slog.Logger, spec *file.ElementSpec) {
	if spec.Circular {
		spec.Type.SetFailed()
		return
	} else if spec.Type.Successful() {
		return // already analyzed
	}

	switches.ElementType(spec.AST.Type,
		func(typ *ast.AliasElementType) {
			ref := spec.File.ElementReferenceByNode(typ.Name)
			if ref == nil {
				spec.Type.SetFailed()
				return
			}
			if ref.Spec.Circular {
				spec.Type.SetFailed()
				return
			}
			z.AnalyzeElementSpec_Type(logger, ref.Spec)
			spec.Type = ref.Spec.Type
		},
		func(typ *ast.BasicElementType) {
			spec.Type.SetResult(typ.Type.Type)
		})
}
