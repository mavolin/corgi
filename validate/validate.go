// Package validate provides contextual validation for [file.File].
package validate

import (
	"sort"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/internal/list"
)

type errList = list.List[*corgierr.Error]

// UseNamespaces validates that there are no namespace collisions in the file's
// uses.
//
// It should be run before linking.
//
// It expects the file's metadata to be set.
//
// If it returns an error, that error will be of type [errList].
func UseNamespaces(f *file.File) error {
	errs := useNamespaces(f)
	if errs.Len() == 0 {
		return nil
	}

	return corgierr.List(errs.ToSlice())
}

// Validate runs all contextual validation for the file, except for
// [UseNamespaces] and [Package], which should be called separately beforehand.
//
// It expects the file to be linked and its metadata to be set.
//
// If it returns an error, that error will be of type [errList].
func Validate(f *file.File) error {
	var errs errList

	errs.PushBackList(ptr(importNamespaces(f)))
	errs.PushBackList(ptr(unusedUses(f)))

	errs.PushBackList(ptr(mainFile(f)))
	errs.PushBackList(ptr(extendFile(f)))
	errs.PushBackList(ptr(extendingFile(f)))
	errs.PushBackList(ptr(libraryFile(f)))

	errs.PushBackList(ptr(duplicateTemplateBlocks(f)))
	errs.PushBackList(ptr(nonExistentTemplateBlocks(f)))

	errs.PushBackList(ptr(mixinsInMixins(f)))
	errs.PushBackList(ptr(duplicateMixinNames(f)))

	errs.PushBackList(ptr(mixinCallChecks(f)))
	errs.PushBackList(ptr(andPlaceholderPlacement(f)))

	errs.PushBackList(ptr(interpolatedMixinCallChecks(f)))

	errs.PushBackList(ptr(mixinCallAttributeChecks(f)))

	errs.PushBackList(ptr(attributePlacement(f)))
	errs.PushBackList(ptr(topLevelAttribute(f)))
	errs.PushBackList(ptr(topLevelTemplateBlockAnds(f)))

	if errs.Len() == 0 {
		return nil
	}

	errSlice := corgierr.List(errs.ToSlice())

	sort.SliceStable(errSlice, func(i, j int) bool {
		return errSlice[i].ErrorAnnotation.Line < errSlice[j].ErrorAnnotation.Line ||
			(errSlice[i].ErrorAnnotation.Line == errSlice[j].ErrorAnnotation.Line &&
				errSlice[i].ErrorAnnotation.Start < errSlice[j].ErrorAnnotation.Start)
	})

	return errSlice
}

// Package runs package-specific validation rules on the files in a package.
//
// It expects the files to be linked and have passed [Validate], and have their
// metadata filled.
func Package(fs []file.File) error {
	var errs errList

	errs.PushBackList(ptr(packageMixinNameConflicts(fs)))

	if errs.Len() == 0 {
		return nil
	}
	return corgierr.List(errs.ToSlice())
}

func ptr[T any](v T) *T {
	return &v
}
