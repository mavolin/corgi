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

	z.AnalyzeElementSpecs_Circular(logger)

	for _, spec := range z.Pkg.ElementSpecs {
		if spec.Analyzed {
			continue
		}

		logger := logger.With(
			slog.String("file", string(spec.File.Name)),
			slog.String("name", spec.StylizedHTMLName),
			slog.String("pos", spec.AST.Start().String()))

		z.AnalyzeElementSpec(logger, spec)

		spec.Analyzed = true
	}
}

func (z *analyzer) AnalyzeElementSpec(logger *slog.Logger, spec *file.ElementSpec) {
	z.AnalyzeElementSpec_Type(logger, spec)
}

// ============================================================================
// Check Cycles
// ======================================================================================

type elementSpec_Circular struct{}

func (z *analyzer) AnalyzeElementSpecs_Circular(logger *slog.Logger) {
	logger = logger.WithGroup("cycles")

	for _, spec := range z.Pkg.ElementSpecs {
		if !spec.Analyzed {
			spec.Circular = false // reset
		}
	}

	chain := make([]*file.ElementSpec, 1, 24)
	for _, spec := range z.Pkg.ElementSpecs {
		if spec.Circular || spec.Analyzed {
			continue
		}

		chain[0] = spec
		chain = chain[:1] // reset the chain
		z.analyzeElementSpec_Circular(logger, chain)
		z.Ran(spec, elementSpec_Circular{})
	}
}

func (z *analyzer) analyzeElementSpec_Circular(logger *slog.Logger, chain []*file.ElementSpec) {
	spec := chain[len(chain)-1]
	if spec.Circular {
		for _, spec := range chain[:len(chain)-1] {
			spec.Circular = true
		}
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
		z.analyzeElementSpec_Circular(logger, chain)
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

	logger.Error("Found element definition cycle")
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

type elementSpec_Type struct{}

func (z *analyzer) AnalyzeElementSpec_Type(logger *slog.Logger, spec *file.ElementSpec) {
	z.Ran(spec, elementSpec_Type{})

	if z.elementSpec_Circular(spec) {
		return
	}

	switches.ElementType(spec.AST.Type,
		func(typ *ast.AliasElementType) {
			ref := spec.File.ElementReferenceByNode(typ.Name)
			if ref == nil {
				spec.Type.SetFailed()
				return
			}
			if z.elementSpec_Circular(spec) {
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
