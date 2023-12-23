package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"

	"github.com/antchfx/htmlquery"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var fail bool

	specAttrs, err := loadSpecAttrs()
	if err != nil {
		return err
	}

	corgiAttrs, err := loadCorgiAttrs(os.Args[1])
	if err != nil {
		return err
	}

	for _, attr := range specAttrs.attrs {
		if !slices.Contains(corgiAttrs.attrs, attr) {
			fmt.Fprintf(os.Stderr, "missing attribute %q\n", attr)
			fail = true
		}
	}
	for _, attr := range corgiAttrs.attrs {
		if !slices.Contains(specAttrs.attrs, attr) {
			fmt.Fprintf(os.Stderr, "unknown attribute %q\n", attr)
		}
	}
	for _, eventHandler := range specAttrs.eventHandlers {
		if !slices.Contains(corgiAttrs.eventHandlers, eventHandler) {
			fmt.Fprintf(os.Stderr, "missing event handler %q\n", eventHandler)
			fail = true
		}
	}
	for _, eventHandler := range corgiAttrs.eventHandlers {
		if !slices.Contains(specAttrs.eventHandlers, eventHandler) {
			fmt.Fprintf(os.Stderr, "unknown event handler %q\n", eventHandler)
		}
	}

	if fail {
		os.Exit(2)
	}

	return nil
}

type attrs struct {
	attrs         []string
	eventHandlers []string
}

const (
	specURL             = "https://html.spec.whatwg.org/multipage/indices.html"
	attrTableID         = "attributes-1"
	eventHandlerTableID = "ix-event-handlers"
)

func loadSpecAttrs() (*attrs, error) {
	var specAttrs attrs

	spec, err := htmlquery.LoadURL(specURL)
	if err != nil {
		return nil, fmt.Errorf("could not load spec: %w", err)
	}

	attrTable := htmlquery.FindOne(spec, fmt.Sprintf("//table[@id=%q]", attrTableID))
	if attrTable == nil {
		return nil, fmt.Errorf("could not find attribute table with id %q", attrTableID)
	}

	attrs := htmlquery.Find(attrTable, "//tbody/tr/th[1]/code/text()")
	if len(attrs) == 0 {
		return nil, fmt.Errorf("could not find any attributes")
	}

	specAttrs.attrs = make([]string, len(attrs))
	for i, attr := range attrs {
		specAttrs.attrs[i] = htmlquery.InnerText(attr)
	}

	eventHandlerTable := htmlquery.FindOne(spec, fmt.Sprintf("//table[@id=%q]", eventHandlerTableID))
	if eventHandlerTable == nil {
		return nil, fmt.Errorf("could not find event handler table with id %q", eventHandlerTableID)
	}

	eventHandlers := htmlquery.Find(eventHandlerTable, "//tbody/tr/th[1]/code/text()")
	if len(eventHandlers) == 0 {
		return nil, fmt.Errorf("could not find any event handlers")
	}

	specAttrs.eventHandlers = make([]string, len(eventHandlers))
	for i, eventHandler := range eventHandlers {
		specAttrs.eventHandlers[i], _ = strings.CutPrefix(htmlquery.InnerText(eventHandler), "on")
	}

	return &specAttrs, nil
}

func loadCorgiAttrs(filename string) (*attrs, error) {
	var corgiAttrs attrs

	f, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return nil, err
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, specI := range genDecl.Specs {
			spec, ok := specI.(*ast.ValueSpec)
			if !ok {
				continue
			}

			if spec.Names[0].Name == "html5Types" {
				lit, ok := spec.Values[0].(*ast.CompositeLit)
				if !ok {
					return nil, fmt.Errorf("html5Types: expected composite literal, got %T", spec.Values[0])
				}

				corgiAttrs.attrs, err = mapKeys(lit)
			} else if spec.Names[0].Name == "html5EventHandlers" {
				lit, ok := spec.Values[0].(*ast.CompositeLit)
				if !ok {
					return nil, fmt.Errorf("html5EventHandlers: expected composite literal, got %T", spec.Values[0])
				}

				corgiAttrs.eventHandlers, err = mapKeys(lit)
				for i, h := range corgiAttrs.eventHandlers {
					corgiAttrs.eventHandlers[i] = h
				}
			}
		}
	}

	if len(corgiAttrs.attrs) == 0 {
		return nil, fmt.Errorf("could not find attribute map")
	} else if len(corgiAttrs.eventHandlers) == 0 {
		return nil, fmt.Errorf("could not find any event handler map")
	}

	return &corgiAttrs, nil
}

func mapKeys(lit *ast.CompositeLit) ([]string, error) {
	var keys []string

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("expected key-value expression, got %T", elt)
		}

		key, ok := kv.Key.(*ast.BasicLit)
		if !ok {
			return nil, fmt.Errorf("expected basic literal, got %T", kv.Key)
		}

		keys = append(keys, strings.Trim(key.Value, `"`))
	}

	return keys, nil
}
