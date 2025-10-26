package analyze

import (
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// ============================================================================
// Package Name
// ======================================================================================

// CheckPackageNamesMatch checks if all files in the package have the same
// package name.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckPackageNamesMatch() (ok bool) {
	logger := z.Logger.WithGroup("checks.package_names_match")

	if len(z.Pkg.Files) <= 1 {
		return true
	}

	expect := file.Qualifier(z.Pkg.Files[0].AST.Package.Name.Name)
	primaries := make([]diagnostic.Annotation, 1, len(z.Pkg.Files))
	primaries[0] = anno.Node(z.Pkg.Files[0], z.Pkg.Files[0].AST.Package.Name, "found this name here")
	for _, f := range z.Pkg.Files[1:] {
		if file.Qualifier(f.AST.Package.Name.Name) != expect {
			primaries = append(primaries, anno.Node(f, f.AST.Package.Name, "but found other name here"))
		}
	}

	if len(primaries) > 1 {
		logger.Error("Package names do not match")
		z.Report(&diagnostic.Diagnostic{
			Message: "package names do not match",
			Primary: primaries,
		})
	}

	return false
}

// SetPackageName sets the package name to the name of the first file in the
// package.
//
// Depends on Checks:
//   - CheckPackageNamesMatch - So that we don't assign an incorrect package
//     name.
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) SetPackageName() {
	logger := z.Logger.WithGroup("set_package_name")

	if len(z.Pkg.Files) == 0 {
		logger.Warn("No files in package")
		return
	}

	z.Pkg.Name = file.Qualifier(z.Pkg.Files[0].AST.Package.Name.Name)
	logger.Debug("Set package name", "name", z.Pkg.Name)
}
