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
	"Components.Blocks.Instances.Group",
	"Components.Blocks.Instances.ChildOf",
	"ComponentCalls.Withs.Block",
	"ComponentCalls.Withs.Instances.Group",
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
				pkg:       pkg,
				rootTypes: rootTypes,
			}

			subFields, err := a.extractFields(field)
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
	pkg       *packages.Package
	rootTypes []string
	prefix    string
}

func (a *fieldAnalyzer) prefixedName(n string) string {
	if a.prefix == "" {
		return n
	}
	return a.prefix + "." + n
}

func (a fieldAnalyzer) extractFields(field *types.Var) ([]string, error) {
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

	typ := field.Type()
	var typeName string
Deref:
	for {
		switch typed := typ.(type) {
		case *types.Pointer:
			typ = typed.Elem()
		case *types.Slice:
			typ = typed.Elem()
		case *types.Named:
			switch {
			case typed.Obj().Pkg() != a.pkg.Types:
				break Deref
			case a.prefix != "" && slices.Contains(a.rootTypes, typed.Obj().Name()):
				break Deref
			}

			typeName = typed.Obj().Name()
			typ = typed.Underlying()
		default:
			break Deref
		}
	}

	switch typ := typ.(type) {
	case *types.Basic:
		if a.prefix == "" {
			return nil, fmt.Errorf("root field %s is not a struct", field.Name())
		}
		return []string{a.prefixedName(field.Name())}, nil
	case *types.Named:
		if a.prefix == "" {
			return nil, fmt.Errorf("root field %s is not a struct", field.Name())
		}
		return []string{a.prefixedName(field.Name())}, nil
	case *types.Struct:
		a.prefix = a.prefixedName(field.Name())
		fields, err := a.extractStructFields(typeName, typ)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field.Name(), err)
		}

		// don't add root fields to the list
		if a.prefix == field.Name() {
			return fields, nil
		}
		return append([]string{a.prefix}, fields...), nil
	default:
		return nil, fmt.Errorf("unsupported field type %T for field %s", typ, field.Name())
	}
}

func (a fieldAnalyzer) extractStructFields(typeName string, s *types.Struct) ([]string, error) {
	fields := make([]string, 0, 50)
	for field := range s.Fields() {
		analyzerField, err := a.isAnalyzerField(typeName, field)
		if err != nil {
			return nil, fmt.Errorf("failed to check if field %s is an analyzer field: %w", field.Name(), err)
		}
		if !analyzerField {
			continue // skip fields not marked with ANALYZER
		}

		subFields, err := a.extractFields(field)
		if err != nil {
			return nil, err
		}
		fields = append(fields, subFields...)
	}

	return fields, nil
}

// isAnalyzerField checks if the field is marked with an ANALYZER comment and
// doesn't belong to another group. It analyzes the AST to find comment groups
// associated with fields.
func (a *fieldAnalyzer) isAnalyzerField(parent string, field *types.Var) (bool, error) {
	// Only root fields need to marked with an ANALYZER comment.
	// All other fields are considered analyzer fields by default.
	if strings.Contains(a.prefix, ".") {
		return true, nil
	}

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
		case strings.TrimSpace(cg.List[1].Text) != "//":
			continue
		case strings.TrimSpace(cg.List[0].Text) == "// ANALYZER":
			return true, nil
		}
		return false, nil // different group
	}

	return false, nil
}
