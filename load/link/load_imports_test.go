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

		if !should.Equal(t, len(ds), 0) {
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

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "import: failed to load package") {
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

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "local-only mode: file contains imports") {
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

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "import alias: cannot use `__corgi_` prefix") {
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

		if !should.Equal(t, len(ds), 0) {
			t.Log(ds.Short())
		}
		should.Equal(t, imp.Namespace, "")
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

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "import: import uses reserved `__corgi_` package name prefix") {
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

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "test diagnostic") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
		should.True(t, imp.Loaded)
	})

	t.Run("import with errors", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		imp := createImport(mainF, nil, "", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: (&mockImporter{
				errors: map[importPath]error{
					importedPkg.ImportPath: errors.New("test error"),
				},
			}).Import,
		})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "import: failed to load package") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
		should.True(t, imp.Loaded)
		should.Equal(t, imp.Package, nil)
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

		if !should.Equal(t, len(ds), 0) {
			t.Log(ds.Short())
		}

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

		ds := Link(context.Background(), mainPkg, Options{
			Importer:    ImporterFor(), // Empty importer will fail to find the builtin
			BuiltinPath: "github.com/non/existent",
		})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "failed to load builtin package") {
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

		if should.Equal(t, len(ds), 1) {
			should.Equal(t, ds[0].Type, diagnostic.InternalError)
			if !should.Equal(t, ds[0].Message, "file already has a builtin import") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}
