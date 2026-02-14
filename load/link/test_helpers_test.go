package link

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

type mockImporter struct {
	packages map[file.CorgiImportPath]*file.Package
	errors   map[file.CorgiImportPath]error
}

var (
	_ Importer = (*mockImporter)(nil).Import
	_ Importer = (*mockImporter)(nil).LinkedImport
)

func (m *mockImporter) Import(_ context.Context, path file.CorgiImportPath) (*file.Package, diagnostic.List, error) {
	if p, ok := m.packages[path]; ok {
		var err error
		if m.errors != nil {
			err = m.errors[path]
		}
		return p, nil, err
	}
	return nil, nil, fmt.Errorf("package %q not found", path)
}

func (m *mockImporter) LinkedImport(ctx context.Context, path file.CorgiImportPath) (*file.Package, diagnostic.List, error) {
	if p, ok := m.packages[path]; ok {
		var err error
		if m.errors != nil {
			err = m.errors[path]
			return p, nil, err
		}
		return p, Link(ctx, p, Options{Importer: m.LinkedImport}), nil
	}
	return nil, nil, fmt.Errorf("package %q not found", path)
}

func ImporterFor(ps ...*file.Package) Importer {
	packages := make(map[file.CorgiImportPath]*file.Package)
	for _, p := range ps {
		packages[p.CorgiImportPath] = p
	}
	return (&mockImporter{packages: packages}).Import
}

func LinkingImporterFor(ps ...*file.Package) Importer {
	packages := make(map[file.CorgiImportPath]*file.Package)
	for _, p := range ps {
		packages[p.CorgiImportPath] = p
	}
	return (&mockImporter{packages: packages}).LinkedImport
}

// createPackage is a helper to create a package for testing.
func createPackage(pathInModule file.PackagePath) *file.Package {
	p := &file.Package{
		Module:          "github.com/mavolin/linktest",
		PathInModule:    pathInModule,
		CorgiImportPath: file.CorgiImportPath("linktest/" + pathInModule),
		Name:            file.Qualifier(path.Base(string(pathInModule))),
		PackageSymbols:  &file.PackageSymbols{},
	}
	return p
}

// createFile is a helper to create a file for testing.
func createFile(p *file.Package, name file.Name) *file.File {
	lines := make([]string, 256)
	for i := range lines {
		lines[i] = strings.Repeat(" ", 180) // so diagnostics work
	}

	f := &file.File{
		Package: p,
		Name:    name,
		Lines:   lines,
		Symbols: &file.Symbols{},
	}
	p.Files = append(p.Files, f)
	return f
}

func addComponent(p *file.Package, component *file.Component) {
	p.Components = append(p.Components, component)
	p.RebuildLookupTables()
}

func addElementSpec(p *file.Package, spec *file.ElementSpec) {
	p.ElementSpecs = append(p.ElementSpecs, spec)
	p.RebuildLookupTables()
}

func addAttributeSpec(p *file.Package, spec *file.AttributeSpec) {
	p.AttributeSpecs = append(p.AttributeSpecs, spec)
	p.RebuildLookupTables()
}

func addImport(f *file.File, imp *file.Import) {
	f.Imports = append(f.Imports, imp)
	f.RebuildLookupTables()
}

func addComponentCall(f *file.File, call *file.ComponentCall) {
	f.ComponentCalls = append(f.ComponentCalls, call)
	f.RebuildLookupTables()
}

func addElementReference(f *file.File, ref *file.ElementReference) {
	f.ElementReferences = append(f.ElementReferences, ref)
	f.RebuildLookupTables()
}

func addAttributeReference(f *file.File, ref *file.AttributeReference) {
	f.AttributeReferences = append(f.AttributeReferences, ref)
	f.RebuildLookupTables()
}

func createComponent(f *file.File, start *ast.Position, name file.Identifier) *file.Component {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	compAST := &ast.Component{Comp: clonePos(start)}
	compAST.Header = &ast.ComponentHeader{
		Name: &ast.Identifier{
			Name:     string(name),
			Position: spaceAfter(compAST),
		},
	}

	c := &file.Component{
		File: f,
		AST:  compAST,
		Name: name,
	}
	addComponent(f.Package, c)
	start.Line++
	start.Col = 1
	return c
}

// createElementSpec creates an element spec for testing
func createElementSpec(
	f *file.File, start *ast.Position, prefix file.CanonicalElementName, name string, typ elemtype.Type,
) *file.ElementSpec {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	definitionAST := &ast.ElementDefinition{Elem: clonePos(start)}
	if prefix != "" {
		definitionAST.Prefix = &ast.ElementName{
			Name:     string(prefix),
			Position: spaceAfter(definitionAST),
		}
	}
	specAST := &ast.ElementSpec{
		Name: &ast.ElementName{
			Name:     name,
			Position: spaceAfter(definitionAST),
		},
	}
	definitionAST.Specs = []*ast.ElementSpec{specAST}

	spec := &file.ElementSpec{
		File:             f,
		Definition:       definitionAST,
		AST:              specAST,
		StylizedHTMLName: string(prefix) + name,
		HTMLName:         prefix + file.CanonicalElementName(strings.ToLower(name)),
		QualifiableName:  file.CanonicalQualifiableElementName(strings.ToLower(name)),
	}
	spec.Type.SetResult(typ)

	addElementSpec(f.Package, spec)
	start.Line++
	start.Col = 1
	return spec
}

// createBasicAttributeSpec creates an attribute spec for testing.
//
// if elemSpec is nil, the attribute spec will match all elements
func createBasicAttributeSpec(
	f *file.File, start *ast.Position, prefix file.CanonicalAttributeName, name string, elemSpec *file.ElementSpec,
	typ attrtype.Type,
) *file.AttributeSpec {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	definitionAST := &ast.AttributeDefinition{Attr: clonePos(start)}
	if prefix != "" {
		definitionAST.Prefix = &ast.AttributeName{
			Name:     string(prefix),
			Position: spaceAfter(definitionAST),
		}
	}

	specAST := &ast.AttributeSpec{}
	definitionAST.Specs = []*ast.AttributeSpec{specAST}
	selAST := &ast.BasicAttributeSelector{
		Name:          name,
		CanonicalName: strings.ToLower(name),
		Position:      spaceAfter(definitionAST),
	}
	specAST.Selector = selAST
	if strings.HasSuffix(selAST.Name, "*") {
		selAST.Name = strings.TrimSuffix(selAST.Name, "*")
		selAST.CanonicalName = strings.TrimSuffix(selAST.CanonicalName, "*")
		selAST.Wildcard = true
	}

	specAST.Ruleset = &ast.AttributeRuleset{LBrace: spaceAfter(definitionAST)}

	var ruleSelAST ast.ElementSelector
	if elemSpec == nil {
		ruleSelAST = &ast.WildcardElementSelector{Asterisk: spaceAfter(definitionAST)}
	} else {
		var qualifier file.Qualifier
		for _, imp := range f.Imports {
			if imp.Package == elemSpec.File.Package {
				qualifier = imp.Qualifier
			}
		}

		ruleSelAST = &ast.ListElementSelector{
			List: []*ast.ElementReference{
				createElementReference(f, spaceAfter(definitionAST), qualifier, elemSpec.StylizedHTMLName).AST,
			},
		}
	}

	ruleAST := &ast.AttributeRule{Selector: ruleSelAST}
	specAST.Ruleset.List = []*ast.AttributeRule{ruleAST}

	ruleAST.Type = &ast.AttributeTypeName{
		Type:     typ,
		Position: spaceAfter(definitionAST),
	}
	if typ != nil {
		ruleAST.Type.Name = typ.String()
	}
	specAST.Ruleset.RBrace = spaceAfter(definitionAST)

	spec := &file.AttributeSpec{
		File:       f,
		Definition: definitionAST,
		AST:        specAST,
		Prefix:     prefix,
	}

	addAttributeSpec(f.Package, spec)
	start.Line++
	start.Col = 1
	return spec
}

// createRegexpAttributeSpec creates an attribute spec for testing.
//
// if elemSpec is nil, the attribute spec will match all elements
func createRegexpAttributeSpec(
	f *file.File, start *ast.Position, prefix file.CanonicalAttributeName, regex string,
	elemSpec *file.ElementSpec, typ attrtype.Type,
) *file.AttributeSpec {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	definitionAST := &ast.AttributeDefinition{Attr: clonePos(start)}
	if prefix != "" {
		definitionAST.Prefix = &ast.AttributeName{
			Name:     string(prefix),
			Position: spaceAfter(definitionAST),
		}
	}

	selAST := &ast.RegexpAttributeSelector{Regexp: spaceAfter(definitionAST)}
	specAST := &ast.AttributeSpec{Selector: selAST}
	definitionAST.Specs = []*ast.AttributeSpec{specAST}

	selAST.LParen = directlyAfter(definitionAST)
	raw := &ast.String{
		Open:  directlyAfter(definitionAST),
		Quote: '"',
		Contents: []ast.StringNode{
			&ast.StringText{
				Text:     interpretedStringContents(regex),
				Position: directlyAfter(definitionAST),
			},
		},
	}
	selAST.Raw = raw
	raw.Close = directlyAfter(raw)
	selAST.Compiled = regexp.MustCompile(regex)
	selAST.RParen = directlyAfter(definitionAST)

	specAST.Ruleset = &ast.AttributeRuleset{LBrace: spaceAfter(definitionAST)}

	var ruleSelAST ast.ElementSelector
	if elemSpec == nil {
		ruleSelAST = &ast.WildcardElementSelector{Asterisk: spaceAfter(definitionAST)}
	} else {
		var qualifier file.Qualifier
		for _, imp := range f.Imports {
			if imp.Package == elemSpec.File.Package {
				qualifier = imp.Qualifier
			}
		}

		ruleSelAST = &ast.ListElementSelector{
			List: []*ast.ElementReference{
				createElementReference(f, spaceAfter(definitionAST), qualifier, elemSpec.StylizedHTMLName).AST,
			},
		}
	}

	ruleAST := &ast.AttributeRule{Selector: ruleSelAST}
	specAST.Ruleset.List = []*ast.AttributeRule{ruleAST}

	ruleAST.Type = &ast.AttributeTypeName{
		Name:     typ.String(),
		Type:     typ,
		Position: spaceAfter(definitionAST),
	}
	specAST.Ruleset.RBrace = spaceAfter(definitionAST)

	spec := &file.AttributeSpec{
		File:       f,
		Definition: definitionAST,
		AST:        specAST,
		Prefix:     prefix,
	}

	addAttributeSpec(f.Package, spec)
	start.Line++
	start.Col = 1
	return spec
}

// createImport is a helper to create an import for testing.
func createImport(f *file.File, start *ast.Position, alias file.Qualifier, impPath file.CorgiImportPath) *file.Import {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	impAST := &ast.Import{Import: clonePos(start)}

	impSpecAST := &ast.ImportSpec{}
	if alias != "" {
		impSpecAST.Alias = &ast.Identifier{
			Name:     string(alias),
			Position: spaceAfter(impAST),
		}
	}

	impSpecAST.Path = &ast.String{
		Open:  spaceAfter(impAST),
		Quote: '"',
		Contents: []ast.StringNode{
			&ast.StringText{
				Text:     interpretedStringContents(string(impPath)),
				Position: directlyAfter(impAST),
			},
		},
	}
	impSpecAST.Path.Close = directlyAfter(impAST)
	impAST.Specs = []*ast.ImportSpec{impSpecAST}

	imp := &file.Import{
		AST:       impSpecAST,
		CorgiPath: impPath,
		Alias:     alias,
	}
	addImport(f, imp)
	start.Line++
	start.Col = 1
	return imp
}

// createComponentCall is a helper to create a component call for testing.
func createComponentCall(f *file.File, start *ast.Position, qualifier file.Qualifier, name file.Identifier) *file.ComponentCall {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	cc := &file.ComponentCall{File: f}

	ccAST := &ast.ComponentCall{Colon: clonePos(start)}
	cc.AST = ccAST

	if qualifier != "" {
		nameAST := &ast.QualifiedIdentifier{
			Package: &ast.Identifier{
				Name:     string(qualifier),
				Position: directlyAfter(ccAST),
			},
		}
		ccAST.Header = &ast.ComponentCallHeader{Name: nameAST}
		nameAST.Dot = directlyAfter(ccAST)
		nameAST.Name = &ast.Identifier{
			Name:     string(name),
			Position: directlyAfter(ccAST),
		}
		cc.Qualifier = qualifier
		cc.Name = name
	} else {
		ccAST.Header = &ast.ComponentCallHeader{
			Name: &ast.Identifier{
				Name:     string(name),
				Position: directlyAfter(ccAST),
			},
		}
		cc.Name = name
	}

	addComponentCall(f, cc)
	start.Line++
	start.Col = 1
	return cc
}

// createElementReference creates an element reference for testing
func createElementReference(f *file.File, start *ast.Position, qualifier file.Qualifier, name string) *file.ElementReference {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	ref := &file.ElementReference{}

	refAST := &ast.ElementReference{}
	ref.AST = refAST

	if qualifier != "" {
		refAST.Package = &ast.Identifier{
			Name:     string(qualifier),
			Position: clonePos(start),
		}
		refAST.Dot = directlyAfter(refAST)
		refAST.Name = &ast.ElementName{
			Name:     name,
			Position: directlyAfter(refAST),
		}
		ref.Qualifier = qualifier
		ref.QualifiableName = file.CanonicalQualifiableElementName(strings.ToLower(name))
	} else {
		refAST.Name = &ast.ElementName{
			Name:     name,
			Position: clonePos(start),
		}
		ref.UnqualifiedName = file.CanonicalElementName(strings.ToLower(name))
	}

	addElementReference(f, ref)
	start.Line++
	start.Col = 1
	return ref
}

// createAttributeReference creates an attribute reference for testing
func createAttributeReference(f *file.File, start *ast.Position, qualifier file.Qualifier, name string) *file.AttributeReference {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	ref := &file.AttributeReference{}

	refAST := &ast.AttributeReference{}
	ref.AST = refAST

	if qualifier != "" {
		refAST.Package = &ast.Identifier{
			Name:     string(qualifier),
			Position: clonePos(start),
		}
		refAST.Dot = directlyAfter(refAST)
		refAST.Name = &ast.AttributeName{
			Name:     name,
			Position: directlyAfter(refAST),
		}
		ref.Qualifier = qualifier
		ref.QualifiableName = file.CanonicalQualifiableAttributeName(strings.ToLower(name))
	} else {
		refAST.Name = &ast.AttributeName{
			Name:     name,
			Position: clonePos(start),
		}
		ref.UnqualifiedName = file.CanonicalAttributeName(strings.ToLower(name))
	}

	addAttributeReference(f, ref)
	start.Line++
	start.Col = 1
	return ref
}

func createBlock(comp *file.Component, name file.Identifier) *file.Block {
	block := &file.Block{
		Name: name,
	}
	comp.Blocks = append(comp.Blocks, block)
	return block
}

func createBlockSetter(cc *file.ComponentCall, name file.Identifier) *file.BlockSetter {
	blockSetter := &file.BlockSetter{
		Name: name,
	}
	cc.BlockSetters = append(cc.BlockSetters, blockSetter)
	return blockSetter
}

func createWith(group *file.BlockSetter, start *ast.Position) *file.BlockSetterInstance {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}

	withAST := &ast.With{With: clonePos(start)}
	withAST.Identifier = &ast.Identifier{
		Name:     string(group.Name),
		Position: spaceAfter(withAST),
	}
	bodyAST := &ast.Scope{LBrace: spaceAfter(withAST)}
	withAST.Body = bodyAST
	bodyAST.RBrace = directlyAfter(withAST)

	instance := &file.BlockSetterInstance{
		Group: group,
		AST:   withAST,
	}
	group.Instances = append(group.Instances, instance)
	start.Line++
	start.Col = 1
	return instance
}

// createParameter adds a parameter with the given name to the component.
func createParameter(comp *file.Component, start *ast.Position, name file.Identifier) *file.ComponentParameter {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}
	param := &file.ComponentParameter{
		Name:      name,
		Component: comp,
		AST: &ast.ComponentParameter{
			Name: &ast.Identifier{
				Name:     string(name),
				Position: clonePos(start),
			},
		},
	}
	comp.Parameters = append(comp.Parameters, param)
	start.Line++
	start.Col = 1
	return param
}

// createArgument adds an argument with the given name to the component call.
func createArgument(call *file.ComponentCall, start *ast.Position, name file.Identifier) *file.ComponentArgument {
	if *start == ast.NoPosition {
		*start = ast.Position{Line: 1, Col: 1}
	}
	arg := &file.ComponentArgument{
		Name:          name,
		ComponentCall: call,
		AST: &ast.ComponentArgument{
			Name: &ast.Identifier{
				Name:     string(name),
				Position: clonePos(start),
			},
		},
	}
	call.ComponentArguments = append(call.ComponentArguments, arg)
	start.Line++
	start.Col = 1
	return arg
}

func interpretedStringContents(s string) string {
	q := strconv.Quote(s)
	return q[len(`"`) : len(q)-len(`"`)]
}

func clonePos(pos *ast.Position) *ast.Position {
	pos2 := *pos
	return &pos2
}

func deltaPos(p ast.Position, dCol int) *ast.Position {
	return &ast.Position{
		Line: p.Line,
		Col:  ast.Col(int(p.Col) + dCol),
	}
}

func spaceAfter(n ast.Node) *ast.Position {
	return deltaPos(n.End(), len(" "))
}

func directlyAfter(n ast.Node) *ast.Position {
	p := n.End()
	return &p
}
