package analyze

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
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

			subFields, err := a.extractFields(symbolType, field)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", symbolType, err)
			}
			fields = append(fields, subFields...)
		}
	}

	return fields, nil
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

func (a fieldAnalyzer) extractFields(parent string, field *types.Var) ([]string, error) {
	switch {
	case !field.Exported():
		return nil, nil
	case field.Name() == "AnalyzedWithErrors":
		return nil, nil
	}
	for _, dontDive := range dontDiveFields {
		if strings.HasSuffix(a.prefixedName(field.Name()), dontDive) {
			return []string{a.prefixedName(field.Name())}, nil
		}
	}

	fields := make([]string, 0, 48)

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

			isAnalyzerField, err := a.isAnalyzerField(parent, field)
			if err != nil {
				return nil, fmt.Errorf("failed to check if field %s is an analyzer field: %w", field.Name(), err)
			}
			if isAnalyzerField {
				return []string{a.prefixedName(field.Name())}, nil
			}
			return nil, nil
		case *types.Named:
			isAnalyzerField, err := a.isAnalyzerField(parent, field)
			if err != nil {
				return nil, fmt.Errorf("failed to check if field %s is an analyzer field: %w", field.Name(), err)
			}
			if isAnalyzerField {
				fields = append(fields, a.prefixedName(field.Name()))
			}

			typeName = typed.Obj().Name()

			switch {
			case typed.Obj().Pkg() != a.pkg.Types:
				return fields, nil
			case a.prefix != "" && slices.Contains(a.dontDiveTypes, typeName):
				return fields, nil
			}

			a.dontDiveTypes = append(a.dontDiveTypes, typeName)

			typ = typed.Underlying()
		case *types.Struct:
			a.prefix = a.prefixedName(field.Name())

			for field := range typed.Fields() {
				subFields, err := a.extractFields(typeName, field)
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

// isAnalyzerField checks if the field is marked with an ANALYZER comment and
// doesn't belong to another group. It analyzes the AST to find comment groups
// associated with fields.
func (a *fieldAnalyzer) isAnalyzerField(parent string, field *types.Var) (bool, error) {
	tokenFile := a.pkg.Fset.File(field.Pos())

	var file *ast.File
	for _, f := range a.pkg.Syntax {
		if a.pkg.Fset.File(f.FileStart).Name() == tokenFile.Name() {
			file = f
			break
		}
	}
	if file == nil {
		return false, fmt.Errorf("file not found: %s", tokenFile.Name())
	}

	var structSpec *ast.TypeSpec
Decls:
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)

			if !ok || typeSpec.Name.Name != parent {
				continue
			}
			structSpec = typeSpec
			break Decls
		}
	}
	if structSpec == nil {
		return false, fmt.Errorf("struct %s not found in file %s", parent, tokenFile.Name())
	}

	structStart := a.pkg.Fset.Position(structSpec.Pos())
	fieldPos := a.pkg.Fset.Position(field.Pos())

	for _, cg := range slices.Backward(file.Comments) {
		commentPos := a.pkg.Fset.Position(cg.Pos())

		// Skip comments before struct definition
		if commentPos.Line < structStart.Line || commentPos.Line >= fieldPos.Line {
			continue
		}

		// Check if comment is a group marker
		switch {
		case len(cg.List) != 2:
			continue
		case strings.TrimSpace(cg.List[0].Text) != "//":
			continue
		case strings.TrimSpace(cg.List[1].Text) == "// ANALYZER":
			return true, nil
		}
		return false, nil // different group
	}

	return false, nil
}
