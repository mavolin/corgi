package analyze

import (
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
	logger := z.Logger.WithGroup("check.package_names_match")
	logger.Info("Checking if package names match")

	if len(z.P.Files) <= 1 {
		logger.Debug("One or no files, skipping")
		return true
	}

	expect := z.P.Files[0].Name
	primaries := make([]diagnostic.Annotation, 1, len(z.P.Files))
	primaries[0] = anno.Node(z.P.Files[0], z.P.Files[0].AST.Package.Name, "found this name here")
	for _, f := range z.P.Files[1:] {
		if f.Name != expect {
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
	logger.Info("Setting package name")

	if len(z.P.Files) == 0 {
		logger.Warn("No files in package")
		return
	}

	z.P.Name = z.P.Files[0].Name
	logger.Debug("Set package name", "name", z.P.Name)
}
