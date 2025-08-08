package analyze

import (
	"errors"
	"fmt"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Helpers

// dontDiveFields is a list of fields that should not be traversed further
// to prevent including fields that are otherwise included in the output.
var dontDiveFields = []string{
	"ComponentCalls.BlockSetters.Block",
}

// ignoredFields is a list of fields that are not included in the output
// because they are not relevant for the analyzer.
var ignoredFields = []string{
	"AttributeReferences.Spec",
}

func getAllFields() ([]string, error) {
	symbolTypes := []string{"Symbols", "PackageSymbols"}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}, "../../../../file")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package, got %d", len(pkgs))
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		errs := make([]error, len(pkg.Errors))
		for i, err := range pkg.Errors {
			errs[i] = err
		}
		return nil, errors.Join(errs...)
	}

	rootTypes := make([]string, 0, 16)
	for _, symbolType := range symbolTypes {
		structType, err := symbolStruct(pkg, symbolType)
		if err != nil {
			return nil, err
		}

		for field := range structType.Fields() {
			if !field.Exported() {
				continue
			}

			name := field.Name()
			if name == "Linked" || name == "Analyzed" {
				continue
			}

			sliceType, _ := field.Type().(*types.Slice)
			if sliceType == nil {
				return nil, fmt.Errorf("field %s is not a slice type", field.Name())
			}

			pointerType, _ := sliceType.Elem().(*types.Pointer)
			if pointerType == nil {
				return nil, fmt.Errorf("field %s is not a pointer type", field.Name())
			}

			namedType, _ := pointerType.Elem().(*types.Named)
			if namedType == nil {
				return nil, fmt.Errorf("field %s is not a named type", field.Name())
			}

			rootTypes = append(rootTypes, namedType.Obj().Name())
		}
	}

	fields := make([]string, 0, 100)
	for _, symbolType := range symbolTypes {
		structType, err := symbolStruct(pkg, symbolType)
		if err != nil {
			return nil, err
		}

		for field := range structType.Fields() {
			a := &fieldAnalyzer{
				pkg:           pkg,
				dontDiveTypes: rootTypes,
			}

			name := field.Name()
			if name == "Linked" || name == "Analyzed" {
				continue
			}

			subFields, err := a.extractFields(field)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", symbolType, err)
			}
			fields = append(fields, subFields...)
		}
	}

	return slices.DeleteFunc(fields, func(s string) bool {
		return slices.Contains(ignoredFields, s)
	}), nil
}

func symbolStruct(pkg *packages.Package, symbolType string) (*types.Struct, error) {
	typ := pkg.Types.Scope().Lookup(symbolType)
	if typ == nil {
		return nil, fmt.Errorf("symbol type %s not found", symbolType)
	}

	namedType, _ := typ.Type().(*types.Named)
	if namedType == nil {
		return nil, fmt.Errorf("symbol type %s is not a named type", symbolType)
	}

	structType, _ := namedType.Underlying().(*types.Struct)
	if structType == nil {
		return nil, fmt.Errorf("symbol type %s is not a struct type", symbolType)
	}

	return structType, nil
}

type fieldAnalyzer struct {
	pkg           *packages.Package
	dontDiveTypes []string
	prefix        string
}

func (a *fieldAnalyzer) prefixedName(n string) string {
	if a.prefix == "" {
		return n
	}
	return a.prefix + "." + n
}

func (a fieldAnalyzer) extractFields(field *types.Var) ([]string, error) {
	if !field.Exported() {
		return nil, nil
	}
	for _, dontDive := range dontDiveFields {
		if strings.HasSuffix(a.prefixedName(field.Name()), dontDive) {
			if isAnalyzerField(field) {
				return []string{a.prefixedName(field.Name())}, nil
			}
			return nil, nil
		}
	}

	typ := field.Type()
	var typeName string
	for {
		switch typed := typ.(type) {
		case *types.Pointer:
			typ = typed.Elem()
		case *types.Slice:
			typ = typed.Elem()
		case *types.Basic:
			if a.prefix == "" {
				return nil, fmt.Errorf("root field %s is not a struct", field.Name())
			}

			if isAnalyzerField(field) {
				return []string{a.prefixedName(field.Name())}, nil
			}

			return nil, nil
		case *types.Named:
			typeName = typed.Obj().Name()

			switch {
			case typed.Obj().Pkg() != a.pkg.Types:
				return nil, nil
			case a.prefix != "" && slices.Contains(a.dontDiveTypes, typeName):
				if isAnalyzerField(field) {
					return []string{a.prefixedName(field.Name())}, nil
				}
				return nil, nil
			}

			a.dontDiveTypes = append(a.dontDiveTypes, typeName)
			if isAnalyzerField(field) {
				return []string{a.prefixedName(field.Name())}, nil
			}

			typ = typed.Underlying()
		case *types.Struct:
			a.prefix = a.prefixedName(field.Name())

			var fields []string
			for field := range typed.Fields() {
				subFields, err := a.extractFields(field)
				if err != nil {
					return nil, err
				}
				fields = append(fields, subFields...)
			}

			return fields, nil
		default:
			return nil, fmt.Errorf("unsupported field type %T for field %s", typ, field.Name())
		}
	}
}

// isAnalyzerField checks if the field is an Analysis[T] type.
func isAnalyzerField(field *types.Var) bool {
	named, ok := field.Type().(*types.Named)
	if !ok {
		return false
	}

	// Check if it's from the file package and named "Analysis"
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != "github.com/mavolin/corgi/v2/file" {
		return false
	}

	return obj.Name() == "Analysis" || obj.Name() == "AnalysisWithReason"
}
