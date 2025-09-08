package link

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_LinkElementReferences(t *testing.T) {
	t.Parallel()

	t.Run("success", testLinker_LinkElementReferences_success)
	t.Run("failure", testLinker_LinkElementReferences_failure)
}

func testLinker_LinkElementReferences_success(t *testing.T) { //nolint:revive
	t.Parallel()

	for _, prefix := range []string{"", "prefix"} {
		var name string
		if prefix == "" {
			name = "without prefix"
		} else {
			name = "with prefix"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Run("local", func(t *testing.T) {
				t.Parallel()

				var start ast.Position
				builtinPkg := createPackage("builtin")
				builtinF := createFile(builtinPkg, "builtin.corgi")
				createElementSpec(builtinF, &start, prefix, "test", elemtype.Normal)

				p := createPackage("test")
				f := createFile(p, "test.corgi")
				spec := createElementSpec(f, &start, prefix, "test", elemtype.Normal)

				htmlName := spec.AST.Name.Name
				if prefix != "" {
					htmlName = prefix + htmlName
				}
				ref := createElementReference(f, &start, "", strings.ToUpper(htmlName))

				d := Link(context.Background(), p, Options{
					Importer:    ImporterFor(builtinPkg),
					BuiltinPath: builtinPkg.CorgiImportPath,
				})

				t.Log(d.Pretty(diagnostic.PrettyOptions{}))
				should.Equal(t, len(d), 0)
				if !should.True(t, spec == ref.Spec) {
					t.Log(cmp.Diff(spec, ref.Spec))
				}
			})

			t.Run("builtin", func(t *testing.T) {
				t.Parallel()

				var start ast.Position
				builtinPkg := createPackage("builtin")
				builtinF := createFile(builtinPkg, "builtin.corgi")
				spec := createElementSpec(builtinF, &start, prefix, "test", elemtype.Normal)

				p := createPackage("test")
				f := createFile(p, "test.corgi")

				htmlName := spec.AST.Name.Name
				if prefix != "" {
					htmlName = prefix + htmlName
				}
				ref := createElementReference(f, &start, "", strings.ToUpper(htmlName))

				d := Link(context.Background(), p, Options{
					Importer:    ImporterFor(builtinPkg),
					BuiltinPath: builtinPkg.CorgiImportPath,
				})

				t.Log(d.Pretty(diagnostic.PrettyOptions{}))
				should.Equal(t, len(d), 0)
				if !should.True(t, spec == ref.Spec) {
					t.Log(cmp.Diff(spec, ref.Spec))
				}
			})

			importTests := []struct {
				name                      string
				packageNameDiffersFromDir bool
				alias                     string
			}{
				{
					name: "qualified/package name/matches directory",
				}, {
					name: "qualified/package name/matches directory",
				}, {
					name:                      "qualified/package name/doesn't match directory",
					packageNameDiffersFromDir: true,
				}, {
					name:  "qualified/alias",
					alias: "myalias",
				}, {
					name:  "dot import",
					alias: ".",
				},
			}
			for _, c := range importTests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					var start ast.Position
					importedPkg := createPackage("imported")
					if c.packageNameDiffersFromDir {
						importedPkg.Name += "pkg"
					}
					importedF := createFile(importedPkg, "imported.corgi")
					importedSpec := createElementSpec(importedF, &start, prefix, "test", elemtype.Normal)

					htmlName := importedSpec.AST.Name.Name
					namespace := importedPkg.Name
					if c.alias == "." {
						namespace = ""
						htmlName = prefix + htmlName
					} else if c.alias != "" {
						namespace = c.alias
					}

					mainPkg := createPackage("main")
					mainFile := createFile(mainPkg, "main.corgi")
					createImport(mainFile, &start, c.alias, importedPkg.CorgiImportPath)
					ref := createElementReference(mainFile, &start, namespace, strings.ToUpper(htmlName))

					d := Link(context.Background(), mainPkg, Options{
						Importer: ImporterFor(importedPkg),
					})

					t.Log(d.Pretty(diagnostic.PrettyOptions{}))
					should.Equal(t, len(d), 0)
					if !should.True(t, importedSpec == ref.Spec) {
						t.Log(cmp.Diff(importedSpec, ref.Spec))
					}
				})
			}
		})
	}
}

func testLinker_LinkElementReferences_failure(t *testing.T) { //nolint:revive
	t.Parallel()

	const builtinPath = "builtin"

	tests := []struct {
		name    string
		message string
		setup   func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error)
	}{
		{
			name:    "unresolved unqualified ref",
			message: "element: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createElementReference(f, &start, "", "test")

				return p, nil, nil
			},
		}, {
			name:    "builtin not loaded",
			message: "failed to load builtin package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createElementReference(f, &start, "", "test")

				return p, nil, map[importPath]error{
					builtinPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				importedPkg.PackageSymbols = nil
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, &start, "", importedPkg.CorgiImportPath)
				createElementReference(f, &start, "imported", "test")

				return p, []*file.Package{importedPkg},
					map[importPath]error{
						imp.CorgiPath: errors.New("stub error"),
					}
			},
		}, {
			name:    "dot import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, &start, ".", "imported")
				createElementReference(f, &start, "", "test")

				return p, nil, map[importPath]error{
					imp.CorgiPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "qualified ref to unknown package",
			message: "element: unresolved reference to package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createElementReference(f, &start, "unknown", "test")

				return p, nil, nil
			},
		}, {
			name:    "qualified ref to unknown element",
			message: "element: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, &start, "", importedPkg.CorgiImportPath)
				createElementReference(f, &start, importedPkg.Name, "test")

				return p, []*file.Package{importedPkg}, nil
			},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, packages, errs := c.setup()
			packagesMap := make(map[importPath]*file.Package, len(packages))
			for _, pkg := range packages {
				packagesMap[pkg.CorgiImportPath] = pkg
			}

			importer := &mockImporter{packages: packagesMap, errors: errs}
			o := Options{Importer: importer.Import}
			if _, ok := packagesMap[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			} else if _, ok := errs[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			}
			d := Link(context.Background(), p, o)

			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			if should.Equal(t, len(d), 1) {
				should.True(t, d[0].Message == c.message)
			}
		})
	}
}
