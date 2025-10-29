package internal

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/assert"
)

// Base is the base for the linker and analyzer types.
type Base struct {
	Pkg         *file.Package
	Logger      *slog.Logger
	diagnostics diagnostic.List

	ranDependencies map[any]map[any]bool
}

// NewBase creates a new Base.
func NewBase(p *file.Package, logger *slog.Logger) *Base {
	b := Base{
		Pkg:         p,
		Logger:      logger,
		diagnostics: make(diagnostic.List, 0, 32),
	}
	if assert.DebugEnabled {
		b.ranDependencies = make(map[any]map[any]bool)
	}
	return &b
}

// Report adds the given diagnostic to the Base's diagnostics.
func (b *Base) Report(ds ...*diagnostic.Diagnostic) {
	b.diagnostics = append(b.diagnostics, ds...)
}

// Diagnostics returns the diagnostics reported to the Base.
func (b *Base) Diagnostics() diagnostic.List {
	if len(b.diagnostics) == 0 {
		return nil
	}
	return slices.Clip(b.diagnostics)
}

// Ran marks the given dependency as ran for the given target.
//
// Try to use the most broad type possible for target, best case the package
// itself.
func (b *Base) Ran(target any, dep any) {
	if assert.DebugEnabled {
		ranDeps := b.ranDependencies[target]
		if ranDeps == nil {
			ranDeps = make(map[any]bool)
			b.ranDependencies[target] = ranDeps
		}
		if ranDeps[dep] {
			panic(fmt.Sprintf("%T ran multiple times for %#v", dep, target))
		}
		ranDeps[dep] = true
	}
}

// Require requires that the given dependency has been run for the given
// target.
func (b *Base) Require(target any, dep any) {
	if assert.DebugEnabled {
		ranDeps := b.ranDependencies[target]
		if ranDeps == nil || !ranDeps[dep] {
			panic(fmt.Sprintf("%T didn't run for %#v", dep, target))
		}
	}
}

func (b *Base) AssertNoError(err error, msg string, f *file.File, n ast.Node, annotation string) bool {
	if err == nil || !assert.DebugEnabled {
		return true
	}

	var primaries []diagnostic.Annotation
	if f != nil {
		if n != nil {
			primaries = []diagnostic.Annotation{anno.Node(f, n, annotation)}
		} else {
			primaries = []diagnostic.Annotation{anno.Position(f, ast.Position{Line: 1, Col: 1}, annotation)}
		}
	}

	b.Report(&diagnostic.Diagnostic{
		Type:    diagnostic.InternalError,
		Message: msg,
		Primary: primaries,
		Cause:   err,
	})
	return false
}
