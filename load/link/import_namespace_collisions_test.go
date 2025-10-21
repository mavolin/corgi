package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_CheckImportNamespaceCollisions(t *testing.T) {
	t.Parallel()

	t.Run("no collisions", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkgA := createPackage("pkg/a")
		pkgB := createPackage("pkg/b")

		mainPkg := createPackage("main")
		mainFile := createFile(mainPkg, "main.corgi")

		createImport(mainFile, &start, "", pkgA.CorgiImportPath)
		createImport(mainFile, &start, "", pkgB.CorgiImportPath)

		d := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(pkgA, pkgB),
		})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		should.Equal(t, len(d), 0)
	})

	t.Run("collision", func(t *testing.T) {
		const wantMessage = "import collision"

		tests := []struct {
			name                 string
			aliasA, packageNameA file.Qualifier
			aliasB, packageNameB file.Qualifier
		}{
			{
				name:   "same alias",
				aliasA: "same", packageNameA: "a",
				aliasB: "same", packageNameB: "b",
			}, {
				name:         "same package name",
				packageNameA: "same",
				packageNameB: "same",
			}, {
				name:   "alias and package name",
				aliasA: "alias", packageNameA: "a",
				packageNameB: "alias",
			},
		}

		for _, c := range tests {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				var start ast.Position
				pkgA := createPackage(file.PackagePath(c.packageNameA))
				pkgB := createPackage(file.PackagePath(c.packageNameB))

				mainPkg := createPackage("main")
				mainFile := createFile(mainPkg, "main.corgi")

				createImport(mainFile, &start, c.aliasA, pkgA.CorgiImportPath)
				createImport(mainFile, &start, c.aliasB, pkgB.CorgiImportPath)

				d := Link(context.Background(), mainPkg, Options{
					Importer: ImporterFor(pkgA, pkgB),
				})

				t.Log(d.Pretty(diagnostic.PrettyOptions{}))
				if should.Equal(t, len(d), 1) {
					should.Equal(t, d[0].Message, wantMessage)
				}
			})
		}
	})
}
