package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_CheckDotImportComponentCollisions(t *testing.T) {
	t.Parallel()

	t.Run("with local import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")

		createComponent(importedF, &start, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createComponent(mainF, &start, "Test")

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "dot import collision: multiple definitions for component of the same name")
		}
	})

	t.Run("two dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, &start, "Test")

		imported2Pkg := createPackage("imported2")
		importedF2 := createFile(imported2Pkg, "imported2.corgi")
		createComponent(importedF2, &start, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createImport(mainF, &start, ".", imported2Pkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg, imported2Pkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "dot import collision: multiple definitions for component of the same name")
		}
	})

	t.Run("two identical dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, &start, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.NotEqual(t, d[0].Message, "dot import collision: multiple definitions for component of the same name")
		}
	})

	t.Run("collision within same dot import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, &start, "Test")
		createComponent(importedF, &start, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})
}

func TestLinker_CheckDotImportElementSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("with local import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, &start, "prefix", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createElementSpec(mainF, &start, "", "prefixTest", elemtype.Normal)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "dot import collision: multiple definitions for element of the same name")
		}
	})

	t.Run("two dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, &start, "", "prefixTest", elemtype.Normal)

		imported2Pkg := createPackage("imported2")
		importedF2 := createFile(imported2Pkg, "imported2.corgi")
		createElementSpec(importedF2, &start, "prefix", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createImport(mainF, &start, ".", imported2Pkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg, imported2Pkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "dot import collision: multiple definitions for element of the same name")
		}
	})

	t.Run("two identical dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, &start, "", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.NotEqual(t, d[0].Message, "dot import collision: multiple definitions for element of the same name")
		}
	})

	t.Run("collision within same dot import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, &start, "", "test", elemtype.Normal)
		createElementSpec(importedF, &start, "", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})
}

func TestLinker_CheckDotImportAttributeSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("with local import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, &start, "prefix", "test", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createBasicAttributeSpec(mainF, &start, "", "prefixTest", nil, attrtype.Innocuous)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "dot import collision: multiple definitions for attribute of the same name")
		}
	})

	t.Run("two dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, &start, "", "prefixTest", nil, attrtype.Innocuous)

		imported2Pkg := createPackage("imported2")
		importedF2 := createFile(imported2Pkg, "imported2.corgi")
		createBasicAttributeSpec(importedF2, &start, "prefix", "test", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createImport(mainF, &start, ".", imported2Pkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg, imported2Pkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "dot import collision: multiple definitions for attribute of the same name")
		}
	})

	t.Run("two identical dot imports", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, &start, "", "test", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.NotEqual(t, d[0].Message, "dot import collision: multiple definitions for attribute of the same name")
		}
	})

	t.Run("collision within same dot import", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, &start, "", "test*", nil, attrtype.Innocuous)
		createBasicAttributeSpec(importedF, &start, "", "test*", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, &start, ".", importedPkg.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})
}
