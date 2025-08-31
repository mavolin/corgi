package file

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
)

// BuildSymbols builds the symbols of the package.
// That means it populates the PackageSymbols field of the package
// and the Symbols field on the files of the package.
//
// This is usually done prior to linking, after all files have been parsed and
// added to the package.
// BuildSymbols fills the all fields on the package's PackageSymbols struct,
// and all fields on the Symbols struct of each file.
func BuildSymbols(p *Package) {
	p.PackageSymbols = &PackageSymbols{
		Components:     make([]*Component, 0, 64),
		State:          make([]*State, 0, 64),
		ElementSpecs:   make([]*ElementSpec, 0, 64),
		AttributeSpecs: make([]*AttributeSpec, 0, 256),
	}
	defer func() {
		p.Components = slices.Clip(p.Components)
		p.State = slices.Clip(p.State)
		p.ElementSpecs = slices.Clip(p.ElementSpecs)
		p.AttributeSpecs = slices.Clip(p.AttributeSpecs)
	}()

	for _, f := range p.Files {
		for _, n := range f.AST.TopLevel {
			switch n := n.(type) {
			case *ast.Component:
				c := &Component{
					AST:  n,
					File: f,
				}
				if n.Header != nil && n.Header.Parameters != nil && len(n.Header.Parameters.List) > 0 {
					c.Parameters = make([]*ComponentParameter, len(n.Header.Parameters.List))
					for i, param := range n.Header.Parameters.List {
						c.Parameters[i] = &ComponentParameter{AST: param}
					}
				}
				p.Components = append(p.Components, c)
			case *ast.StateDeclaration:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					for i := range spec.Names {
						s := &State{AST: spec, File: f, Index: i}
						p.State = append(p.State, s)
					}
				}
			case *ast.ElementDefinition:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					p.ElementSpecs = append(p.ElementSpecs, &ElementSpec{Definition: n, AST: spec, File: f})
				}
			case *ast.AttributeDefinition:
				for _, spec := range n.Specs {
					if spec == nil {
						continue
					}
					p.AttributeSpecs = append(p.AttributeSpecs, &AttributeSpec{Definition: n, AST: spec, File: f})
				}
			}
		}
	}

	p.RebuildLookupTables()

	for _, f := range p.Files {
		buildSymbols(f)
	}
}

func buildSymbols(f *File) {
	f.Symbols = &Symbols{
		Imports:             make([]*Import, 0, 64),
		ComponentCalls:      make([]*ComponentCall, 0, 256),
		ElementReferences:   make([]*ElementReference, 0, 256),
		Attributes:          make([]*Attribute, 0, 512),
		AttributeReferences: make([]*AttributeReference, 0, 512),
	}
	defer func() {
		f.Imports = slices.Clip(f.Imports)
		f.ComponentCalls = slices.Clip(f.ComponentCalls)
		f.ElementReferences = slices.Clip(f.ElementReferences)
		f.Attributes = slices.Clip(f.Attributes)
		f.AttributeReferences = slices.Clip(f.AttributeReferences)
	}()

	for _, impStmt := range f.AST.Imports {
		for _, spec := range impStmt.Specs {
			imp := &Import{AST: spec}
			if spec.Alias != nil {
				imp.Alias = spec.Alias.Name
			}
			if spec.Path != nil {
				imp.Path = spec.Path.Unquote()
			}
			f.Imports = append(f.Imports, imp)
		}
	}

	var (
		cc          *ComponentCall
		attr        *Attribute
		comp        *Component
		parentBlock *BlockInstance
	)
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch n := n.(type) {
		case *ast.ComponentCall:
			ccw := &ComponentCall{AST: n, File: f, BlockSetters: make([]*BlockSetter, 0, 24)}
			f.ComponentCalls = append(f.ComponentCalls, ccw)
			f.componentCallsByNode[n] = ccw

			oldCC := cc
			cc = ccw
			n.Walk(walk)
			cc = oldCC

			ccw.BlockSetters = slices.Clip(ccw.BlockSetters)
			for _, with := range ccw.BlockSetters {
				with.Instances = slices.Clip(with.Instances)
			}
		case ast.BlockSetter:
			if cc == nil {
				n.Walk(walk)
				break
			}
			instance := &BlockSetterInstance{AST: n}
			group := cc.BlockSetterByName(n.Name())
			if group == nil {
				group = &BlockSetter{Name: n.Name(), Instances: make([]*BlockSetterInstance, 0, 16)}
				cc.BlockSetters = append(cc.BlockSetters, group)
			}
			instance.Group = group
			group.Instances = append(group.Instances, instance)

			n.Walk(walk)
		case *ast.Block:
			instance := &BlockInstance{AST: n, Parent: parentBlock}
			if n.Default != nil {
				instance.Default = &BlockInstanceDefault{AST: n.Default}
			}
			group := comp.BlockByName(n.Name())
			if group == nil {
				group = &Block{Name: n.Name(), Instances: make([]*BlockInstance, 0, 16)}
				comp.Blocks = append(comp.Blocks, group)
			}
			instance.Group = group
			group.Instances = append(group.Instances, instance)

			oldParent := parentBlock
			parentBlock = instance
			n.Walk(walk)
			parentBlock = oldParent
		case *ast.ElementReference:
			f.ElementReferences = append(f.ElementReferences, &ElementReference{AST: n})
			n.Walk(walk)
		case ast.Attribute:
			attr = &Attribute{AST: n}
			f.Attributes = append(f.Attributes, attr)
			n.Walk(walk)
		case *ast.AttributeReference:
			ref := &AttributeReference{AST: n}
			if attr != nil {
				attr.Reference = ref
				attr = nil
			}
			f.AttributeReferences = append(f.AttributeReferences, ref)
			n.Walk(walk)
		}
	}
	for _, n := range f.AST.TopLevel {
		astC, _ := n.(*ast.Component)
		if astC == nil {
			continue
		}
		c := f.Package.ComponentByNode(astC)
		c.Blocks = make([]*Block, 0, 24)

		ccsStart := len(f.ComponentCalls)
		n.Walk(walk)
		ccEnd := len(f.ComponentCalls)
		if ccEnd > ccsStart {
			c.ComponentCalls = f.ComponentCalls[ccsStart:ccEnd:ccEnd]
		}

		c.Blocks = slices.Clip(c.Blocks)
		for _, block := range c.Blocks {
			block.Instances = slices.Clip(block.Instances)
		}
	}

	f.RebuildLookupTables()
}
