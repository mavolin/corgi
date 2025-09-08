package link

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
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

	t.Run("local-only mode", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, "", "github.com/some/package")

		d := Link(context.Background(), mainPkg, Options{
			Importer: nil, // nil importer triggers local-only mode
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "local-only mode: file contains imports")
		}
	})

	t.Run("illegal alias prefix", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		importedComp := createComponent(importedF, &start, "Test")
		addComponent(importedPkg, importedComp)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, "__corgi_illegal", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "import alias: cannot use `__corgi_` prefix")
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
		should.Equal(t, imp.Namespace, "")
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
			Importer: func(_ context.Context, path string) (*file.Package, diagnostic.List, error) {
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
				errors: map[importPath]error{
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
