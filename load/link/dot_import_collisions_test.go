package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckDotImportComponentCollisions(t *testing.T) {
	t.Parallel()

	t.Run("with local import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, nil, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createComponent(mainF, nil, "Test")

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "dot import collision: multiple definitions for component of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("two dot imports", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, nil, "Test")

		imported2Pkg := createPackage("imported2")
		importedF2 := createFile(imported2Pkg, "imported2.corgi")
		createComponent(importedF2, nil, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createImport(mainF, nil, ".", imported2Pkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg, imported2Pkg),
		})
		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "dot import collision: multiple definitions for component of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("two identical dot imports", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, nil, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})
		if should.Equal(t, len(ds), 1) {
			if !should.NotEqual(t, ds[0].Message, "dot import collision: multiple definitions for component of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("collision within same dot import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createComponent(importedF, nil, "Test")
		createComponent(importedF, nil, "Test")

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if !should.Equal(t, len(ds), 0) {
			t.Log(ds.Short())
		}
	})
}

func TestLinker_CheckDotImportElementSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("with local import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, nil, "prefix", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createElementSpec(mainF, nil, "", "prefixTest", elemtype.Normal)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "dot import collision: multiple definitions for element of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("two dot imports", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, nil, "", "prefixTest", elemtype.Normal)

		imported2Pkg := createPackage("imported2")
		importedF2 := createFile(imported2Pkg, "imported2.corgi")
		createElementSpec(importedF2, nil, "prefix", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createImport(mainF, nil, ".", imported2Pkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg, imported2Pkg),
		})
		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "dot import collision: multiple definitions for element of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("two identical dot imports", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, nil, "", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})
		if should.Equal(t, len(ds), 1) {
			if !should.NotEqual(t, ds[0].Message, "dot import collision: multiple definitions for element of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("collision within same dot import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createElementSpec(importedF, nil, "", "test", elemtype.Normal)
		createElementSpec(importedF, nil, "", "test", elemtype.Normal)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if !should.Equal(t, len(ds), 0) {
			t.Log(ds.Short())
		}
	})
}

func TestLinker_CheckDotImportAttributeSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("with local import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, nil, "prefix", "test", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createBasicAttributeSpec(mainF, nil, "", "prefixTest", nil, attrtype.Innocuous)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "dot import collision: multiple definitions for attribute of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("two dot imports", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, nil, "", "prefixTest", nil, attrtype.Innocuous)

		imported2Pkg := createPackage("imported2")
		importedF2 := createFile(imported2Pkg, "imported2.corgi")
		createBasicAttributeSpec(importedF2, nil, "prefix", "test", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createImport(mainF, nil, ".", imported2Pkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg, imported2Pkg),
		})
		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "dot import collision: multiple definitions for attribute of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("two identical dot imports", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, nil, "", "test", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)
		createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})
		if should.Equal(t, len(ds), 1) {
			if !should.NotEqual(t, ds[0].Message, "dot import collision: multiple definitions for attribute of the same name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("collision within same dot import", func(t *testing.T) {
		t.Parallel()

		importedPkg := createPackage("imported")
		importedF := createFile(importedPkg, "imported.corgi")
		createBasicAttributeSpec(importedF, nil, "", "test*", nil, attrtype.Innocuous)
		createBasicAttributeSpec(importedF, nil, "", "test*", nil, attrtype.Innocuous)

		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")
		createImport(mainF, nil, ".", importedPkg.ImportPath)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(importedPkg),
		})

		if !should.Equal(t, len(ds), 0) {
			t.Log(ds.Short())
		}
	})
}
