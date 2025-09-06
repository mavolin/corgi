package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckImportCycles(t *testing.T) {
	t.Parallel()

	t.Run("no cycle", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		// A imports B, no cycle
		pkgB := createPackage("pkg/b")

		pkgA := createPackage("pkg/a")
		fileA := createFile(pkgA, "a.corgi")

		imp := createImport(fileA, &start, "", pkgB.ImportPath)
		imp.Package = pkgB

		d := Link(context.Background(), pkgA, Options{
			Importer: ImporterFor(pkgB),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})

	cycleTests := []struct {
		name  string
		setup func() (p *file.Package, packages map[importPath]*file.Package, importGraph []*file.Package)
	}{
		{
			name: "direct cycle",
			setup: func() (p *file.Package, packages map[importPath]*file.Package, importGraph []*file.Package) {
				var start ast.Position
				// A imports B, B imports A - direct cycle
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				createImport(fileA, &start, "", pkgB.ImportPath)
				createImport(fileB, &start, "", pkgA.ImportPath)

				packages = map[importPath]*file.Package{
					pkgA.ImportPath: pkgA,
					pkgB.ImportPath: pkgB,
				}
				importGraph = []*file.Package{pkgA, pkgB}

				return pkgA, packages, importGraph
			},
		}, {
			name: "indirect cycle",
			setup: func() (p *file.Package, packages map[importPath]*file.Package, importGraph []*file.Package) {
				var start ast.Position
				// A imports B, B imports C, C imports A - indirect cycle
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				pkgC := createPackage("pkg/c")
				fileC := createFile(pkgC, "c.corgi")

				createImport(fileA, &start, "", pkgB.ImportPath)
				createImport(fileB, &start, "", pkgC.ImportPath)
				createImport(fileC, &start, "", pkgA.ImportPath)

				packages = map[importPath]*file.Package{
					pkgA.ImportPath: pkgA,
					pkgB.ImportPath: pkgB,
					pkgC.ImportPath: pkgC,
				}
				importGraph = []*file.Package{pkgA, pkgB, pkgC}

				return pkgA, packages, importGraph
			},
		}, {
			name: "hook",
			setup: func() (p *file.Package, packages map[importPath]*file.Package, importGraph []*file.Package) {
				var start ast.Position
				// A imports B, B imports C, C imports B - indirect cycle with a hook
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				pkgC := createPackage("pkg/c")
				fileC := createFile(pkgC, "c.corgi")

				createImport(fileA, &start, "", pkgB.ImportPath)
				createImport(fileB, &start, "", pkgC.ImportPath)
				createImport(fileC, &start, "", pkgB.ImportPath)

				packages = map[importPath]*file.Package{
					pkgA.ImportPath: pkgA,
					pkgB.ImportPath: pkgB,
					pkgC.ImportPath: pkgC,
				}
				importGraph = []*file.Package{pkgA, pkgB, pkgC}

				return pkgA, packages, importGraph
			},
		}, {
			name: "same cycle in multiple imports", // test that we still get only a single diagnostic
			setup: func() (p *file.Package, packages map[importPath]*file.Package, importGraph []*file.Package) {
				var start ast.Position
				// A imports B, B imports A - direct cycle
				pkgA := createPackage("pkg/a")
				fileA := createFile(pkgA, "a.corgi")

				pkgB := createPackage("pkg/b")
				fileB := createFile(pkgB, "b.corgi")

				createImport(fileA, &start, "", pkgB.ImportPath)
				createImport(fileA, &start, "foo", pkgB.ImportPath)
				createImport(fileB, &start, "", pkgA.ImportPath)

				packages = map[importPath]*file.Package{
					pkgA.ImportPath: pkgA,
					pkgB.ImportPath: pkgB,
				}
				importGraph = []*file.Package{pkgA, pkgB}

				return pkgA, packages, importGraph
			},
		},
	}

	for _, c := range cycleTests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			pkg, packages, importGraph := c.setup()

			ctx := context.WithValue(context.Background(), importersGraphKey{}, importGraph)
			importer := &mockImporter{packages: packages}
			d := Link(ctx, pkg, Options{
				Importer: importer.Import,
			})

			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			if should.Equal(t, len(d), 1) {
				should.Equal(t, d[0].Message, "circular import")
			}
		})
	}
}
