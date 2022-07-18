// Package imports provides a resolver for imports.
package imports

import (
	"path/filepath"

	"github.com/mavolin/corgi/corgi/file"
)

// Resolver resolves all imports in a file.
//
// For that, it prevents duplicate imports of the same package and namespace.
//
// Additionally, it checks that there are no two imports of different packages
// that have the same namespace.
// If there are, it reports them.
type Resolver struct {
	f *file.File
}

// NewResolver creates a new Resolver for the passed file.
func NewResolver(f *file.File) *Resolver {
	return &Resolver{f: f}
}

// Resolve attempts to resolve all imports.
// It stores the resolved imports in f.Imports.
func (r *Resolver) Resolve() error {
	if err := r.resolveExtend(); err != nil {
		return err
	}

	if err := r.resolveUses(); err != nil {
		return err
	}

	if err := r.resolveIncludes(); err != nil {
		return err
	}

	return nil
}

// addImport checks if this file's imports allow an import of the passed
// alias and path.
//
// If there is already an import for the same package under the same namespace,
// addImport does not add the import.
//
// If the alias is '.' or '_', addImport adds the import if the path is not
// already imported.
//
// If another package uses the same alias and that alias is not '.' or '_',
// addImport returns an *CollisionError.
//
// In all other cases it adds the import.
func (r *Resolver) addImport(imp file.Import) error {
	impNamespace := resolveNamespace(imp)

	if imp.Alias == "." || imp.Alias == "_" {
		for _, cmp := range r.f.Imports {
			if cmp.Path == imp.Path {
				return nil
			}
		}

		r.f.Imports = append(r.f.Imports, imp)

		return nil
	}

	for i, cmp := range r.f.Imports {
		cmpNamespace := resolveNamespace(cmp)
		if cmpNamespace == "_" {
			r.f.Imports[i].Alias = imp.Alias
			return nil
		}

		if impNamespace != cmpNamespace {
			continue
		}

		// both have no alias, or both have the same alias and both have the
		// same path -> don't add this import, it already exists
		if imp.Path == cmp.Path {
			return nil
		}

		// namespace collision 😕
		return &CollisionError{
			Source:      r.f.Source,
			File:        r.f.Name,
			Line:        imp.Line,
			Col:         imp.Col,
			OtherSource: cmp.Source,
			OtherFile:   cmp.File,
			OtherLine:   cmp.Line,
			OtherCol:    cmp.Col,
			Namespace:   impNamespace,
		}
	}

	r.f.Imports = append(r.f.Imports, imp)
	return nil
}

func resolveNamespace(imp file.Import) string {
	namespace := string(imp.Alias)

	if namespace == "" {
		namespace = filepath.Base(imp.Path)
	}

	return namespace
}
