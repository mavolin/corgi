package link

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_LoadImports(t *testing.T) {
	t.Parallel()

	t.Run("load import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, &start, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, &start, "", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
		if !should.True(t, importedPkg == imp.Package) {
			t.Log(cmp.Diff(importedPkg, imp.Package))
		}
	})

	t.Run("missing import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, "", "github.com/non/existent")

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "import: failed to load package")
		}
	})

	t.Run("dot import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, &start, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
		should.Equal(t, imp.Qualifier, "")
	})

	t.Run("import with reserved package name prefix", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("__corgi_test")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, &start, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, "", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "import: import uses reserved `__corgi_` package name prefix")
		}
	})

	t.Run("import with diagnostics", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, &start, "", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: func(_ context.Context, path file.CorgiImportPath) (*file.Package, diagnostic.List, error) {
				if path == importedPkg.CorgiImportPath {
					return importedPkg, diagnostic.List{{Message: "test diagnostic"}}, nil
				}
				return nil, nil, errors.New("unknown import")
			},
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "test diagnostic")
		}
		should.True(t, imp.Loaded)
	})

	t.Run("import with errors", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, &start, "", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: (&mockImporter{
				errors: map[file.CorgiImportPath]error{
					importedPkg.CorgiImportPath: errors.New("test error"),
				},
			}).Import,
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "import: failed to load package")
		}
		should.True(t, imp.Loaded)
		should.Equal(t, imp.Package, nil)
	})

	t.Run("builtin import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		builtinComp := createComponent(builtinF, &start, "BuiltinComponent")
		addComponent(builtinPkg, builtinComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")

		d := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.CorgiImportPath,
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)

		builtinImp := mainF.BuiltinImport()
		if !should.True(t, builtinImp != nil) {
			return
		}
		if !should.True(t, builtinPkg == builtinImp.Package) {
			t.Log(cmp.Diff(builtinPkg, builtinImp.Package))
		}
		should.Equal(t, builtinImp.Alias, BuiltinAlias)
	})

	t.Run("builtin import error", func(t *testing.T) {
		t.Parallel()

		mainPkg := createPackage("main")
		createFile(mainPkg, "main.corgi")

		d := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(), // Empty importer will fail to find the builtin
			BuiltinPath: "github.com/non/existent",
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "failed to load builtin package")
		}
	})

	t.Run("file already has builtin import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		builtinComp := createComponent(builtinF, &start, "BuiltinComponent")
		addComponent(builtinPkg, builtinComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")

		// Manually add a builtin import to the file
		mainF.AddBuiltinImport("__builtin", builtinPkg)

		d := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: "some/other/builtin/path", // Different path to trigger error
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Type, diagnostic.InternalError)
			should.Equal(t, d[0].Message, "file already has a builtin import")
		}
	})
}

func TestLinker_ImportCycles(t *testing.T) {
	t.Parallel()

	t.Run("no cycle", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		// A imports B, no cycle
		pkgB := createPackage("pkg/b")

		pkgA := createPackage("pkg/a")
		fileA := createFile(pkgA, "a.corgi")

		imp := createImport(fileA, &start, "", pkgB.CorgiImportPath)
		imp.Package = pkgB

		d := Link(context.Background(), pkgA, Options{
			Importer: ImporterFor(pkgB),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})

	cycleTests := []struct {
		name  string
		setup func() []*file.Package
	}{
		{
			name: "direct cycle",
			setup: func() []*file.Package {
				var start ast.Position
				// A imports B, B imports A - direct cycle
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				createImport(fileA, &start, "", pkgB.CorgiImportPath)
				createImport(fileB, &start, "", pkgA.CorgiImportPath)

				return []*file.Package{pkgA, pkgB}
			},
		}, {
			name: "indirect cycle",
			setup: func() []*file.Package {
				var start ast.Position
				// A imports B, B imports C, C imports A - indirect cycle
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				pkgC := createPackage("pkg/c")
				fileC := createFile(pkgC, "c.corgi")

				createImport(fileA, &start, "", pkgB.CorgiImportPath)
				createImport(fileB, &start, "", pkgC.CorgiImportPath)
				createImport(fileC, &start, "", pkgA.CorgiImportPath)

				return []*file.Package{pkgA, pkgB, pkgC}
			},
		}, {
			name: "hook",
			setup: func() []*file.Package {
				var start ast.Position
				// A imports B, B imports C, C imports B - indirect cycle with a hook
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				pkgC := createPackage("pkg/c")
				fileC := createFile(pkgC, "c.corgi")

				createImport(fileA, &start, "", pkgB.CorgiImportPath)
				createImport(fileB, &start, "", pkgC.CorgiImportPath)
				createImport(fileC, &start, "", pkgB.CorgiImportPath)

				return []*file.Package{pkgA, pkgB, pkgC}
			},
		},
	}

	for _, c := range cycleTests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			pkgs := c.setup()

			d := Link(t.Context(), pkgs[0], Options{
				Importer: LinkingImporterFor(pkgs...),
			})

			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			if should.Equal(t, len(d), 1) {
				should.Equal(t, d[0].Message, "circular import")
			}
		})
	}
}
