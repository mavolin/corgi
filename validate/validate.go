// Package validate provides contextual validation for [file.File].
package validate

import (
	"sort"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/list"
)

type errList = list.List[*corgierr.Error]

// PreLink validates that there are no namespace collisions in the file's uses.
//
// It should be run before linking.
//
// It expects the file's metadata to be set.
//
// If it returns an error, that error will be of type [corgierr.List].
func PreLink(f *file.File) error {
	var errs errList

	errs.PushBackList(useNamespaces(f))
	errs.PushBackList(usePathBaseIsValidIdent(f))

	if errs.Len() == 0 {
		return nil
	}

	return corgierr.List(errs.ToSlice())
}

// File runs all contextual validation for the file, and all the other files
// it uses.
//
// If you have a library that you want to validate, you should call [Library]
// instead, which will run all the validation File does, plus extra
// library-specific validation.
//
// It expects the file to be linked and its metadata to be set and
// [PreLink] to be run.
//
// Since File recursively validates all files it encounters, it is neither
// necessary nor performant to validate the files f depends on individually.
// Instead, you should fully link f and all its dependencies and then run File
// on f.
//
// If File returns an error, that error will be of type [corgierr.List].
func File(f *file.File) error {
	valedFiles := make(map[string]struct{})
	impNamespaces := make(map[string]importNamespace)

	errs := _file(f, valedFiles, impNamespaces)
	if errs.Len() == 0 {
		return nil
	}

	errSlice := corgierr.List(errs.ToSlice())
	sort.Stable(errSlice)
	return errSlice
}

func _file(f *file.File, valedFiles map[string]struct{}, impNamespaces map[string]importNamespace) *errList {
	if _, ok := valedFiles[f.Module+f.PathInModule]; ok {
		return &errList{}
	}

	valedFiles[f.Module+f.PathInModule] = struct{}{}

	var errs errList

	errs.PushBackList(importNamespaces(impNamespaces, f))

	errs.PushBackList(duplicateImports(f))
	errs.PushBackList(unusedUses(f))

	errs.PushBackList(mainFile(f))
	errs.PushBackList(templateFile(f))
	errs.PushBackList(extendingFile(f))
	errs.PushBackList(libraryFile(f))

	errs.PushBackList(onlyTemplateFilesContainBlockPlaceholders(f))
	errs.PushBackList(duplicateTemplateBlocks(f))
	errs.PushBackList(nonExistentTemplateBlocks(f))

	errs.PushBackList(mixinChecks(f))
	errs.PushBackList(mixinsInMixins(f))
	errs.PushBackList(duplicateMixinNames(f))

	errs.PushBackList(mixinCallChecks(f))
	errs.PushBackList(andPlaceholderPlacement(f))

	errs.PushBackList(interpolatedMixinCallChecks(f))

	errs.PushBackList(mixinCallAttributeChecks(f))

	errs.PushBackList(attributePlacement(f))
	errs.PushBackList(topLevelAttribute(f))
	errs.PushBackList(topLevelTemplateBlockAnds(f))

	if f.Extend != nil {
		errs.PushBackList(_file(f.Extend.File, valedFiles, impNamespaces))
	}

	for _, use := range f.Uses {
		for _, spec := range use.Uses {
			errs.PushBackList(libraryMixinNameConflicts(spec.Library.Files))
			for _, libFile := range spec.Library.Files {
				errs.PushBackList(_file(libFile, valedFiles, impNamespaces))
			}
		}
	}

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		incl, ok := (*ctx.Item).(file.Include)
		if !ok {
			return false, nil
		}

		cincl, ok := incl.Include.(file.CorgiInclude)
		if !ok {
			return false, nil
		}

		errs.PushBackList(_file(cincl.File, valedFiles, impNamespaces))
		return false, err
	})

	return &errs
}

// Library runs all the rules [File] runs and some additional library-specific
// rules.
//
// Just like [File], Library recursively validates all files and should
// therefore only be called if compiling a library.
//
// Read the doc of [File] for more information about requirements and Library's
// return value.
func Library(l *file.Library) error {
	var errs errList

	impNamespaces := make(map[string]importNamespace)

	if l.Precompiled {
		for _, f := range l.Files {
			errs.PushBackList(importNamespaces(impNamespaces, f))
		}

		if errs.Len() == 0 {
			return nil
		}

		errSlice := corgierr.List(errs.ToSlice())
		sort.Stable(errSlice)
		return errSlice
	}

	errs.PushBackList(libraryMixinNameConflicts(l.Files))

	valedFiles := make(map[string]struct{})

	for _, f := range l.Files {
		errs.PushBackList(_file(f, valedFiles, impNamespaces))
	}

	if errs.Len() == 0 {
		return nil
	}

	errSlice := corgierr.List(errs.ToSlice())
	sort.Stable(errSlice)
	return errSlice
}
