package link

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_LinkAttributeReferences(t *testing.T) {
	t.Parallel()

	t.Run("success", testLinker_LinkAttributeReferences_success)
	t.Run("failure", testLinker_LinkAttributeReferences_failure)
}

func testLinker_LinkAttributeReferences_success(t *testing.T) { //nolint:revive
	t.Parallel()

	for _, regexp := range []bool{false, true} {
		var name string
		if regexp {
			name = "regexp selector"
		} else {
			name = "basic selector"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, prefix := range []file.CanonicalAttributeName{"", "prefix"} {
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
						if regexp {
							createRegexpAttributeSpec(builtinF, &start, prefix, "tes.+", nil, attrtype.String)
						} else {
							createBasicAttributeSpec(builtinF, &start, prefix, "test", nil, attrtype.String)
						}

						p := createPackage("test")
						f := createFile(p, "test.corgi")

						var spec *file.AttributeSpec
						if regexp {
							spec = createRegexpAttributeSpec(f, &start, prefix, "tes.+", nil, attrtype.String)
						} else {
							spec = createBasicAttributeSpec(f, &start, prefix, "test", nil, attrtype.String)
						}

						htmlName := "test"
						if prefix != "" {
							htmlName = string(prefix) + htmlName
						}
						ref := createAttributeReference(f, &start, "", strings.ToUpper(htmlName))

						d := Link(context.Background(), p, Options{
							Importer:    ImporterFor(builtinPkg),
							BuiltinPath: builtinPkg.CorgiImportPath,
						})

						t.Log(d.Pretty(diagnostic.PrettyOptions{}))
						should.Equal(t, len(d), 0)
						if !should.True(t, spec == ref.Spec.ResultOr(nil)) {
							t.Log(cmp.Diff(spec, ref.Spec.ResultOr(nil)))
						}
					})

					t.Run("builtin", func(t *testing.T) {
						t.Parallel()

						var start ast.Position
						builtinPkg := createPackage("builtin")
						builtinF := createFile(builtinPkg, "builtin.corgi")
						var spec *file.AttributeSpec
						if regexp {
							spec = createRegexpAttributeSpec(builtinF, &start, prefix, "tes.+", nil, attrtype.String)
						} else {
							spec = createBasicAttributeSpec(builtinF, &start, prefix, "test", nil, attrtype.String)
						}

						p := createPackage("test")
						f := createFile(p, "test.corgi")

						htmlName := "test"
						if prefix != "" {
							htmlName = string(prefix) + htmlName
						}
						ref := createAttributeReference(f, &start, "", strings.ToUpper(htmlName))

						d := Link(context.Background(), p, Options{
							Importer:    ImporterFor(builtinPkg),
							BuiltinPath: builtinPkg.CorgiImportPath,
						})

						t.Log(d.Pretty(diagnostic.PrettyOptions{}))
						should.Equal(t, len(d), 0)
						if !should.True(t, spec == ref.Spec.ResultOr(nil)) {
							t.Log(cmp.Diff(spec, ref.Spec.ResultOr(nil)))
						}
					})

					importTests := []struct {
						name                      string
						packageNameDiffersFromDir bool
						alias                     file.Qualifier
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
							var importedSpec *file.AttributeSpec
							if regexp {
								importedSpec = createRegexpAttributeSpec(importedF, &start, prefix, "tes.+", nil, attrtype.String)
							} else {
								importedSpec = createBasicAttributeSpec(importedF, &start, prefix, "test", nil, attrtype.String)
							}

							htmlName := "test"
							qualifier := importedPkg.Name
							if c.alias == "." {
								qualifier = ""
								htmlName = string(prefix) + htmlName
							} else if c.alias != "" {
								qualifier = c.alias
							}

							mainPkg := createPackage("main")
							mainFile := createFile(mainPkg, "main.corgi")
							createImport(mainFile, &start, c.alias, importedPkg.CorgiImportPath)
							ref := createAttributeReference(mainFile, &start, qualifier, strings.ToUpper(htmlName))

							d := Link(context.Background(), mainPkg, Options{
								Importer: ImporterFor(importedPkg),
							})

							t.Log(d.Pretty(diagnostic.PrettyOptions{}))
							should.Equal(t, len(d), 0)
							if !should.True(t, importedSpec == ref.Spec.ResultOr(nil)) {
								t.Log(cmp.Diff(importedSpec, ref.Spec.ResultOr(nil)))
							}
						})
					}
				})
			}
		})
	}
}

func testLinker_LinkAttributeReferences_failure(t *testing.T) { //nolint:revive
	t.Parallel()

	const builtinPath = "builtin"

	tests := []struct {
		name    string
		message string
		setup   func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error)
	}{
		{
			name:    "multiple local same-specificity matching attributes",
			message: "attribute: ambiguous reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createRegexpAttributeSpec(f, &start, "", "tes.+", nil, attrtype.String)
				createRegexpAttributeSpec(f, &start, "", "tes.+", nil, attrtype.String)
				createAttributeReference(f, &start, "", "test")

				return p, nil, nil
			},
		}, {
			name:    "multiple builtin same-specificity matching attributes",
			message: "builtin: attribute: ambiguous reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				builtinPkg := createPackage("builtin")
				builtinPkg.CorgiImportPath = builtinPath
				createFile(builtinPkg, "builtin.corgi")
				createRegexpAttributeSpec(builtinPkg.Files[0], &start, "", "tes.+", nil, attrtype.String)
				createRegexpAttributeSpec(builtinPkg.Files[0], &start, "", "tes.+", nil, attrtype.String)

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createAttributeReference(f, &start, "", "test")

				return p, []*file.Package{builtinPkg}, nil
			},
		}, {
			name:    "multiple qualified same-specificity matching attributes",
			message: "attribute: ambiguous reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")
				createRegexpAttributeSpec(importedPkg.Files[0], &start, "", "tes.+", nil, attrtype.String)
				createRegexpAttributeSpec(importedPkg.Files[0], &start, "", "tes.+", nil, attrtype.String)

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, &start, "", importedPkg.CorgiImportPath)
				createAttributeReference(f, &start, importedPkg.Name, "test")

				return p, []*file.Package{importedPkg}, nil
			},
		}, {
			name:    "builtin not loaded",
			message: "failed to load builtin package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createAttributeReference(f, &start, "", "test")

				return p, nil, map[file.CorgiImportPath]error{
					builtinPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				importedPkg := createPackage("imported")
				importedPkg.PackageSymbols = nil
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, &start, "", importedPkg.CorgiImportPath)
				createAttributeReference(f, &start, "imported", "test")

				return p, []*file.Package{importedPkg}, map[file.CorgiImportPath]error{
					imp.CorgiPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "dot import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, &start, ".", "imported")
				createAttributeReference(f, &start, "", "test")

				return p, nil, map[file.CorgiImportPath]error{
					imp.CorgiPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "qualified ref to unknown package",
			message: "attribute: unresolved reference to package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createAttributeReference(f, &start, "unknown", "test")

				return p, nil, nil
			},
		}, {
			name:    "qualified ref to unknown attribute",
			message: "attribute: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[file.CorgiImportPath]error) {
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				var start ast.Position
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, &start, "", importedPkg.CorgiImportPath)
				createAttributeReference(f, &start, importedPkg.Name, "test")

				return p, []*file.Package{importedPkg}, nil
			},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, packages, errs := c.setup()
			packagesMap := make(map[file.CorgiImportPath]*file.Package, len(packages))
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
