package link

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_LinkAttributeReferences(t *testing.T) {
	t.Parallel()

	t.Run("success", testLinker_LinkAttributeReferences_success)
	t.Run("failure", testLinker_LinkAttributeReferences_failure)
}

func testLinker_LinkAttributeReferences_success(t *testing.T) {
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

						builtinPkg := createPackage("builtin")
						builtinF := createFile(builtinPkg, "builtin.corgi")
						if regexp {
							createRegexpAttributeSpec(builtinF, nil, prefix, "tes.+", nil, attrtype.Innocuous)
						} else {
							createBasicAttributeSpec(builtinF, nil, prefix, "test", nil, attrtype.Innocuous)
						}

						p := createPackage("test")
						f := createFile(p, "test.corgi")

						var spec *file.AttributeSpec
						if regexp {
							spec = createRegexpAttributeSpec(f, nil, prefix, "tes.+", nil, attrtype.Innocuous)
						} else {
							spec = createBasicAttributeSpec(f, nil, prefix, "test", nil, attrtype.Innocuous)
						}

						htmlName := "test"
						if prefix != "" {
							htmlName = prefix + htmlName
						}
						ref := createAttributeReference(f, nil, "", strings.ToUpper(htmlName))

						ds := Link(context.Background(), p, Options{
							Importer:    ImporterFor(builtinPkg),
							BuiltinPath: builtinPkg.ImportPath,
						})

						if !should.Equal(t, 0, len(ds)) {
							t.Log(ds.Short())
						}
						if !should.True(t, spec == ref.Spec) {
							t.Log(cmp.Diff(spec, ref.Spec))
						}
					})

					t.Run("builtin", func(t *testing.T) {
						t.Parallel()

						builtinPkg := createPackage("builtin")
						builtinF := createFile(builtinPkg, "builtin.corgi")
						var spec *file.AttributeSpec
						if regexp {
							spec = createRegexpAttributeSpec(builtinF, nil, prefix, "tes.+", nil, attrtype.Innocuous)
						} else {
							spec = createBasicAttributeSpec(builtinF, nil, prefix, "test", nil, attrtype.Innocuous)
						}

						p := createPackage("test")
						f := createFile(p, "test.corgi")

						htmlName := "test"
						if prefix != "" {
							htmlName = prefix + htmlName
						}
						ref := createAttributeReference(f, nil, "", strings.ToUpper(htmlName))

						ds := Link(context.Background(), p, Options{
							Importer:    ImporterFor(builtinPkg),
							BuiltinPath: builtinPkg.ImportPath,
						})

						if !should.Equal(t, 0, len(ds)) {
							t.Log(ds.Short())
						}
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

							importedPkg := createPackage("imported")
							if c.packageNameDiffersFromDir {
								importedPkg.Name += "pkg"
							}
							importedF := createFile(importedPkg, "imported.corgi")
							var importedSpec *file.AttributeSpec
							if regexp {
								importedSpec = createRegexpAttributeSpec(importedF, nil, prefix, "tes.+", nil, attrtype.Innocuous)
							} else {
								importedSpec = createBasicAttributeSpec(importedF, nil, prefix, "test", nil, attrtype.Innocuous)
							}

							htmlName := "test"
							namespace := importedPkg.Name
							if c.alias == "." {
								namespace = ""
								htmlName = prefix + htmlName
							} else if c.alias != "" {
								namespace = c.alias
							}

							mainPkg := createPackage("main")
							mainFile := createFile(mainPkg, "main.corgi")
							createImport(mainFile, nil, c.alias, importedPkg.ImportPath)
							ref := createAttributeReference(mainFile, nil, namespace, strings.ToUpper(htmlName))

							ds := Link(context.Background(), mainPkg, Options{
								Importer: ImporterFor(importedPkg),
							})

							if !should.Equal(t, 0, len(ds)) {
								t.Log(ds.Short())
							}
							if !should.True(t, importedSpec == ref.Spec) {
								t.Log(cmp.Diff(importedSpec, ref.Spec))
							}
						})
					}
				})
			}
		})
	}
}

func testLinker_LinkAttributeReferences_failure(t *testing.T) {
	t.Parallel()

	const builtinPath = "builtin"

	tests := []struct {
		name    string
		message string
		setup   func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error)
	}{
		{
			name:    "multiple local same-specificity matching attributes",
			message: "attribute: ambiguous reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createRegexpAttributeSpec(f, nil, "", "tes.+", nil, attrtype.Innocuous)
				createRegexpAttributeSpec(f, nil, "", "tes.+", nil, attrtype.Innocuous)
				createAttributeReference(f, nil, "", "test")

				return p, nil, nil
			},
		}, {
			name:    "multiple builtin same-specificity matching attributes",
			message: "builtin: attribute: ambiguous reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				builtinPkg := createPackage("builtin")
				builtinPkg.ImportPath = builtinPath
				createFile(builtinPkg, "builtin.corgi")
				createRegexpAttributeSpec(builtinPkg.Files[0], nil, "", "tes.+", nil, attrtype.Innocuous)
				createRegexpAttributeSpec(builtinPkg.Files[0], nil, "", "tes.+", nil, attrtype.Innocuous)

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createAttributeReference(f, nil, "", "test")

				return p, []*file.Package{builtinPkg}, nil
			},
		}, {
			name:    "multiple qualified same-specificity matching attributes",
			message: "attribute: ambiguous reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")
				createRegexpAttributeSpec(importedPkg.Files[0], nil, "", "tes.+", nil, attrtype.Innocuous)
				createRegexpAttributeSpec(importedPkg.Files[0], nil, "", "tes.+", nil, attrtype.Innocuous)

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, nil, "", importedPkg.ImportPath)
				createAttributeReference(f, nil, importedPkg.Name, "test")

				return p, []*file.Package{importedPkg}, nil
			},
		}, {
			name:    "builtin not loaded",
			message: "failed to load builtin package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createAttributeReference(f, nil, "", "test")

				return p, nil, map[importPath]error{
					builtinPath: errors.New("stub error"),
				}
			},
		}, {
			name:    "import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, nil, "", importedPkg.ImportPath)
				createAttributeReference(f, nil, "imported", "test")

				return p, []*file.Package{importedPkg}, map[importPath]error{
					imp.Path: errors.New("stub error"),
				}
			},
		}, {
			name:    "dot import not loaded",
			message: "import: failed to load package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				imp := createImport(f, nil, ".", "imported")
				createAttributeReference(f, nil, "", "test")

				return p, nil, map[importPath]error{
					imp.Path: errors.New("stub error"),
				}
			},
		}, {
			name:    "qualified ref to unknown package",
			message: "attribute: unresolved reference to package",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createAttributeReference(f, nil, "unknown", "test")

				return p, nil, nil
			},
		}, {
			name:    "qualified ref to unknown attribute",
			message: "attribute: unresolved reference",
			setup: func() (p *file.Package, packages []*file.Package, packageErrors map[importPath]error) {
				importedPkg := createPackage("imported")
				createFile(importedPkg, "imported.corgi")

				p = createPackage("test")
				f := createFile(p, "test.corgi")
				createImport(f, nil, "", importedPkg.ImportPath)
				createAttributeReference(f, nil, importedPkg.Name, "test")

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
				packagesMap[pkg.ImportPath] = pkg
			}

			importer := &mockImporter{packages: packagesMap, errors: errs}
			o := Options{Importer: importer.Import}
			if _, ok := packagesMap[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			} else if _, ok := errs[builtinPath]; ok {
				o.BuiltinPath = builtinPath
			}
			ds := Link(context.Background(), p, o)

			if should.Equal(t, 1, len(ds)) {
				if !should.True(t, ds[0].Message == c.message) {
					t.Log(ds[0].Short())
				}
			} else {
				t.Log(ds.Short())
			}
		})
	}
}
