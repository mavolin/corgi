// Package analyze implements the analysis of a corgi package.
//
// It establishes that certain preconditions are met, that are necessary to
// run the analysis.
// After that it runs the actual analysis.
// A positive outcome is required to run the postchecks.
package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/internal/meta"
	"github.com/mavolin/corgi/v2/load/internal"
)

type analyzer struct {
	*internal.Base
	Logger                 *slog.Logger
	analyzedComponentCalls map[*file.ComponentCall]bool
}

func Analyze(b *internal.Base) {
	var numCCs int
	for _, f := range b.Pkg.Files {
		numCCs += len(f.ComponentCalls)
	}

	z := &analyzer{
		Base:                   b,
		Logger:                 b.Logger.WithGroup("analysis"),
		analyzedComponentCalls: make(map[*file.ComponentCall]bool, numCCs),
	}
	z.Logger.Info("Running analysis")

	if z.CheckPackageNamesMatch() {
		z.SetPackageName()
	}

	z.AnalyzeElementSpecs()
	z.AnalyzeComponents()
	z.AnalyzeComponentCalls()
	z.AnalyzeAttributes()

	z.Pkg.Analyzed = true
	for _, f := range z.Pkg.Files {
		f.Analyzed = true
	}
}

// SafeImport returns the import for the safe package for the given file, or
// adds it if it does not exist yet.
func (z *analyzer) SafeImport(f *file.File) *file.Import {
	if imp := f.ImportByGoPath(file.SafeImport); imp != nil {
		return imp
	}

	imp := &file.Import{
		Alias:     "__corgi_safe",
		GoPath:    file.SafeImport,
		Qualifier: "__corgi_safe",
		Forward:   true,
	}
	imp.EnsureUniqueQualifier(f)
	f.AddImport(imp)
	return imp
}

const htmlImportPath file.GoImportPath = meta.Module + "/std/html"

// HTMLImport returns the import for the std/html package for the given file,
// or adds it if it does not exist yet.
func (z *analyzer) HTMLImport(f *file.File) *file.Import {
	if imp := f.ImportByGoPath(htmlImportPath); imp != nil {
		return imp
	}

	imp := &file.Import{
		Alias:     "__corgi_html",
		GoPath:    htmlImportPath,
		Qualifier: "__corgi_html",
		Forward:   true,
	}
	imp.EnsureUniqueQualifier(f)
	f.AddImport(imp)
	return imp
}
