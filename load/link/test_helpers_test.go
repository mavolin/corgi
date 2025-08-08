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
	packages map[importPath]*file.Package
	errors   map[importPath]error
}

var _ Importer = (*mockImporter)(nil).Import

func (m *mockImporter) Import(_ context.Context, path importPath) (*file.Package, diagnostic.List, error) {
	if p, ok := m.packages[path]; ok {
		var err error
		if m.errors != nil {
			err = m.errors[path]
		}
		return p, nil, err
	}
	return nil, nil, fmt.Errorf("package %q not found", path)
}

func ImporterFor(ps ...*file.Package) Importer {
	packages := make(map[string]*file.Package)
	for _, p := range ps {
		packages[p.ImportPath] = p
	}
	return (&mockImporter{packages: packages}).Import
}

// createPackage is a helper to create a package for testing.
func createPackage(pathInModule string) *file.Package {
	p := &file.Package{
		Module:         "github.com/mavolin/linktest",
		PathInModule:   pathInModule,
		ImportPath:     "linktest/" + pathInModule,
		Name:           path.Base(pathInModule),
		PackageSymbols: &file.PackageSymbols{},
	}
	return p
}

// createFile is a helper to create a file for testing.
func createFile(p *file.Package, name string) *file.File {
	f := &file.File{
		Package: p,
		Name:    name,
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

func createComponent(f *file.File, start *ast.Position, name string) *file.Component {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	compAST := &ast.Component{Comp: start}
	compAST.Header = &ast.ComponentHeader{
		Name: &ast.Identifier{
			Name:     name,
			Position: spaceAfter(compAST),
		},
	}

	c := &file.Component{
		File: f,
		AST:  compAST,
	}
	addComponent(f.Package, c)
	return c
}

// createElementSpec creates an element spec for testing
func createElementSpec(f *file.File, start *ast.Position, prefix, name string, typ elemtype.Type) *file.ElementSpec {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	definitionAST := &ast.ElementDefinition{Elem: start}
	if prefix != "" {
		definitionAST.Prefix = &ast.ElementName{
			Name:     prefix,
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
		File:       f,
		Definition: definitionAST,
		AST:        specAST,
	}
	spec.Type.SetResult(typ)

	addElementSpec(f.Package, spec)
	return spec
}

// createBasicAttributeSpec creates an attribute spec for testing.
//
// if elemSpec is nil, the attribute spec will match all elements
func createBasicAttributeSpec(f *file.File, start *ast.Position, prefix, name string, elemSpec *file.ElementSpec, typ attrtype.Type) *file.AttributeSpec {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	definitionAST := &ast.AttributeDefinition{Attr: start}
	if prefix != "" {
		definitionAST.Prefix = &ast.AttributeName{
			Name:     prefix,
			Position: spaceAfter(definitionAST),
		}
	}

	specAST := &ast.AttributeSpec{}
	definitionAST.Specs = []*ast.AttributeSpec{specAST}
	selAST := &ast.BasicAttributeSelector{
		Name:     name,
		Position: spaceAfter(definitionAST),
	}
	specAST.Selector = selAST
	if strings.HasSuffix(selAST.Name, "*") {
		selAST.Name = strings.TrimSuffix(selAST.Name, "*")
		selAST.Wildcard = true
	}

	specAST.Ruleset = &ast.AttributeRuleset{LBrace: spaceAfter(definitionAST)}

	var ruleSelAST ast.ElementSelector
	if elemSpec == nil {
		ruleSelAST = &ast.WildcardElementSelector{Asterisk: spaceAfter(definitionAST)}
	} else {
		var namespace string
		for _, imp := range f.Imports {
			if imp.Package == elemSpec.File.Package {
				namespace = imp.Namespace
			}
		}

		ruleSelAST = &ast.ListElementSelector{
			List: []*ast.ElementReference{
				createElementReference(f, spaceAfter(definitionAST), namespace, elemSpec.HTMLName()).AST,
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
	}

	addAttributeSpec(f.Package, spec)
	return spec
}

// createRegexpAttributeSpec creates an attribute spec for testing.
//
// if elemSpec is nil, the attribute spec will match all elements
func createRegexpAttributeSpec(f *file.File, start *ast.Position, prefix, regex string, elemSpec *file.ElementSpec, typ attrtype.Type) *file.AttributeSpec {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	definitionAST := &ast.AttributeDefinition{Attr: start}
	if prefix != "" {
		definitionAST.Prefix = &ast.AttributeName{
			Name:     prefix,
			Position: spaceAfter(definitionAST),
		}
	}

	selAST := &ast.RegexpAttributeSelector{Regexp: spaceAfter(definitionAST)}
	specAST := &ast.AttributeSpec{Selector: selAST}
	definitionAST.Specs = []*ast.AttributeSpec{specAST}

	selAST.LParen = directlyAfter(definitionAST)
	selAST.Raw = &ast.StaticString{
		Open:     directlyAfter(selAST),
		Quote:    '"',
		Contents: strconv.Quote(regex),
	}
	selAST.Compiled = regexp.MustCompile(regex)
	selAST.Raw.Close = deltaPos(*selAST.LParen, 0, len(`"`)+len(selAST.Raw.Contents))
	selAST.RParen = directlyAfter(definitionAST)

	specAST.Ruleset = &ast.AttributeRuleset{LBrace: spaceAfter(definitionAST)}

	var ruleSelAST ast.ElementSelector
	if elemSpec == nil {
		ruleSelAST = &ast.WildcardElementSelector{Asterisk: spaceAfter(definitionAST)}
	} else {
		var namespace string
		for _, imp := range f.Imports {
			if imp.Package == elemSpec.File.Package {
				namespace = imp.Namespace
			}
		}

		ruleSelAST = &ast.ListElementSelector{
			List: []*ast.ElementReference{
				createElementReference(f, spaceAfter(definitionAST), namespace, elemSpec.HTMLName()).AST,
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
	}

	addAttributeSpec(f.Package, spec)
	return spec
}

// createImport is a helper to create an import for testing.
func createImport(f *file.File, start *ast.Position, alias, impPath string) *file.Import {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	impAST := &ast.Import{Import: start}

	impSpecAST := &ast.ImportSpec{}
	if alias != "" {
		impSpecAST.Alias = &ast.Identifier{
			Name:     alias,
			Position: spaceAfter(impAST),
		}
	}

	impSpecAST.Path = &ast.StaticString{
		Open:     spaceAfter(impAST),
		Quote:    '"',
		Contents: strconv.Quote(impPath),
	}
	impSpecAST.Path.Close = deltaPos(*impSpecAST.Path.Open, 0, len(`"`)+len(impSpecAST.Path.Contents))
	impAST.Specs = []*ast.ImportSpec{impSpecAST}

	imp := &file.Import{
		AST:   impSpecAST,
		Path:  impPath,
		Alias: alias,
	}
	addImport(f, imp)
	return imp
}

// createComponentCall is a helper to create a component call for testing.
func createComponentCall(f *file.File, start *ast.Position, namespace, name string) *file.ComponentCall {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	ccAST := &ast.ComponentCall{Colon: start}

	if namespace != "" {
		nameAST := &ast.QualifiedIdentifier{
			Package: &ast.Identifier{
				Name:     namespace,
				Position: directlyAfter(ccAST),
			},
		}
		ccAST.Header = &ast.ComponentCallHeader{Name: nameAST}
		nameAST.Dot = directlyAfter(ccAST)
		nameAST.Name = &ast.Identifier{
			Name:     name,
			Position: directlyAfter(ccAST),
		}
	} else {
		ccAST.Header = &ast.ComponentCallHeader{
			Name: &ast.Identifier{
				Name:     name,
				Position: directlyAfter(ccAST),
			},
		}
	}

	cc := &file.ComponentCall{
		AST:  ccAST,
		File: f,
	}

	addComponentCall(f, cc)
	return cc
}

// createElementReference creates an element reference for testing
func createElementReference(f *file.File, start *ast.Position, namespace, name string) *file.ElementReference {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	refAST := &ast.ElementReference{}
	if namespace != "" {
		refAST.Package = &ast.Identifier{
			Name:     namespace,
			Position: start,
		}
		refAST.Dot = directlyAfter(refAST)
		refAST.Name = &ast.ElementName{
			Name:     name,
			Position: directlyAfter(refAST),
		}
	} else {
		refAST.Name = &ast.ElementName{
			Name:     name,
			Position: start,
		}
	}

	ref := &file.ElementReference{AST: refAST}
	addElementReference(f, ref)
	return ref
}

// createAttributeReference creates an attribute reference for testing
func createAttributeReference(f *file.File, start *ast.Position, namespace, name string) *file.AttributeReference {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	refAST := &ast.AttributeReference{}
	if namespace != "" {
		refAST.Package = &ast.Identifier{
			Name:     namespace,
			Position: start,
		}
		refAST.Dot = directlyAfter(refAST)
		refAST.Name = &ast.AttributeName{
			Name:     name,
			Position: directlyAfter(refAST),
		}
	} else {
		refAST.Name = &ast.AttributeName{
			Name:     name,
			Position: start,
		}
	}

	ref := &file.AttributeReference{AST: refAST}
	addAttributeReference(f, ref)
	return ref
}

func createBlock(comp *file.Component, name string) *file.Block {
	block := &file.Block{
		Name: name,
	}
	comp.Blocks = append(comp.Blocks, block)
	return block
}

func createBlockSetter(cc *file.ComponentCall, name string) *file.BlockSetter {
	blockSetter := &file.BlockSetter{
		Name: name,
	}
	cc.BlockSetters = append(cc.BlockSetters, blockSetter)
	return blockSetter
}

func createWith(group *file.BlockSetter, start *ast.Position) *file.BlockSetterInstance {
	if start == nil {
		start = &ast.Position{Line: 1, Col: 1}
	}

	withAST := &ast.With{With: start}
	withAST.Identifier = &ast.Identifier{
		Name:     group.Name,
		Position: spaceAfter(withAST),
	}
	bodyAST := &ast.Scope{LBrace: spaceAfter(withAST)}
	withAST.Body = bodyAST
	bodyAST.RBrace = directlyAfter(withAST)

	return &file.BlockSetterInstance{
		Group: group,
		AST:   withAST,
	}
}

func deltaPos(p ast.Position, dLine, dCol int) *ast.Position {
	if dLine == 0 {
		return &ast.Position{
			Line: p.Line,
			Col:  p.Col + dCol,
		}
	}
	return &ast.Position{
		Line: p.Line + dLine,
		Col:  1,
	}
}

func spaceAfter(n ast.Node) *ast.Position {
	return deltaPos(n.End(), 0, len(" "))
}

func directlyAfter(n ast.Node) *ast.Position {
	p := n.End()
	return &p
}
