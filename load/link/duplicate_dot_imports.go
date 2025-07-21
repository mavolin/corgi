package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckDuplicateDotImports(_ context.Context) {
	logger := l.logger.WithGroup("check.duplicate_dot_imports")
	logger.Info("Checking for duplicate dot imports")

	for _, f := range l.p.Files {
		(&duplicateDotImportChecker{
			duplDotImports: make([]*file.Import, 0, 8),
			checked:        l.takeStringSet(),
		}).checkFile(l, l.logger, f)
	}
}

type duplicateDotImportChecker struct { // file level
	duplDotImports []*file.Import
	checked        set.Set[importPath]
}

func (c *duplicateDotImportChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.With(slog.String("file", f.Name))
	logger.Debug("Checking file")

	if len(f.Imports) <= 1 {
		logger.Debug("One or no imports, skipping")
		return
	}

	for ai, a := range f.Imports[:len(f.Imports)-1] {
		if c.shouldCheck(a) {
			continue
		}

		logger := logger.With(
			slog.String("pos", a.AST.Start().String()),
			slog.String("import", a.Path))
		logger.Debug("Checking import")

		if c.checked.Contains(a.Path) {
			logger.Debug("Already checked imports with that path")
			continue
		}
		c.checked.Add(a.Path)
		c.resetDuplicates()

		for _, b := range f.Imports[ai+1:] {
			if c.shouldCheck(b) {
				continue
			}
			if a.Path == b.Path {
				logger.Debug("Found duplicate", slog.String("duplicate_position", b.AST.Start().String()))
				c.recordDuplicate(b)
			}
		}

		if len(c.duplDotImports) > 0 {
			c.reportDuplicate(l, logger, f, a, c.duplDotImports)
		}
	}
}

func (c *duplicateDotImportChecker) reportDuplicate(l *linker, logger *slog.Logger, f *file.File, first *file.Import, dupls []*file.Import) {
	logger.Error("Duplicate dot imports", slog.Int("n_duplicates", len(dupls)))

	primaries := make([]diagnostic.Annotation, 1, 1+len(dupls))
	primaries[0] = anno.Node(f, first.AST, "package dot-imported for the first time here")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.AST, "then again here"))
	}

	l.report(&diagnostic.Diagnostic{
		Message:     "duplicated dot imports",
		Primary:     primaries,
		Explanation: "You can only dot-import a package once per file.",
		Hints: []diagnostic.Hint{
			{Hint: "Remove the duplicate dot imports."},
		},
	})
}

func (c *duplicateDotImportChecker) resetDuplicates() {
	c.duplDotImports = c.duplDotImports[:0]
}

func (c *duplicateDotImportChecker) recordDuplicate(imp *file.Import) {
	c.duplDotImports = append(c.duplDotImports, imp)
}

func (c duplicateDotImportChecker) shouldCheck(imp *file.Import) bool {
	return imp.Path != "" && imp.AST != nil && imp.AST.Alias != nil && imp.AST.Alias.Name == "."
}
