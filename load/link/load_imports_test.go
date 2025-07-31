package link

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_LoadImports(t *testing.T) {
	t.Parallel()

	t.Run("load import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, nil, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, nil, "", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
		if !should.True(t, importedPkg == imp.Package) {
			t.Log(cmp.Diff(importedPkg, imp.Package))
		}
	})

	t.Run("missing import", func(t *testing.T) {
		t.Parallel()

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, "", "github.com/non/existent")

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(),
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "import: failed to load package", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("local-only mode", func(t *testing.T) {
		t.Parallel()

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, "", "github.com/some/package")

		ds := Link(context.Background(), mainPkg, Options{
			Importer: nil, // nil importer triggers local-only mode
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "local-only mode: file contains imports", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("illegal alias prefix", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, nil, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, "__corgi_illegal", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "import alias: cannot use `__corgi_` prefix", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("dot import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, nil, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}
		should.Equal(t, "", imp.Namespace)
	})

	t.Run("import with reserved package name prefix", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("__corgi_test")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, nil, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, "", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "import: import uses reserved `__corgi_` package name prefix", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("import with diagnostics", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, nil, "", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: func(_ context.Context, path string) (*file.Package, diagnostic.List, error) {
				if path == importedPkg.ImportPath {
					return importedPkg, diagnostic.List{{Message: "test diagnostic"}}, nil
				}
				return nil, nil, errors.New("unknown import")
			},
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "test diagnostic", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
		if !should.True(t, imp.LoadedWithErrors) {
			t.Error("Expected import to be marked as loaded with errors")
		}
	})

	t.Run("builtin import", func(t *testing.T) {
		t.Parallel()

		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		builtinComp := createComponent(builtinF, nil, "BuiltinComponent")
		addComponent(builtinPkg, builtinComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")

		ds := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: builtinPkg.ImportPath,
		})

		if !should.Equal(t, 0, len(ds)) {
			t.Log(ds.Short())
		}

		builtinImp := mainF.BuiltinImport()
		if !should.True(t, builtinImp != nil) {
			return
		}
		if !should.True(t, builtinPkg == builtinImp.Package) {
			t.Log(cmp.Diff(builtinPkg, builtinImp.Package))
		}
		should.Equal(t, BuiltinAlias, builtinImp.Alias)
	})

	t.Run("builtin import error", func(t *testing.T) {
		t.Parallel()

		mainPkg := createPackage("main")
		createFile(mainPkg, "main.corgi")

		ds := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(), // Empty importer will fail to find the builtin
			BuiltinPath: "github.com/non/existent",
		})

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "failed to load builtin package", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("file already has builtin import", func(t *testing.T) {
		t.Parallel()

		builtinPkg := createPackage("builtin")
		builtinF := createFile(builtinPkg, "builtin.corgi")
		builtinComp := createComponent(builtinF, nil, "BuiltinComponent")
		addComponent(builtinPkg, builtinComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")

		// Manually add a builtin import to the file
		mainF.AddBuiltinImport("__builtin", builtinPkg)

		ds := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(builtinPkg),
			BuiltinPath: "some/other/builtin/path", // Different path to trigger error
		})

		if should.Equal(t, 1, len(ds)) {
			should.Equal(t, diagnostic.InternalError, ds[0].Type)
			if !should.Equal(t, "file already has a builtin import", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}
