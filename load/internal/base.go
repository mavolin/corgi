package internal

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
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
func (b *Base) Ran(for_ any, dep any) {
	if assert.DebugEnabled {
		ranDeps := b.ranDependencies[for_]
		if ranDeps == nil {
			ranDeps = make(map[any]bool)
			b.ranDependencies[for_] = ranDeps
		}
		if ranDeps[dep] {
			panic(fmt.Sprintf("%T ran multiple times for %#v", dep, for_))
		}
		ranDeps[dep] = true
	}
}

// Require requires that the given dependency has been run for the given
// target.
func (b *Base) Require(for_ any, dep any) {
	if assert.DebugEnabled {
		ranDeps := b.ranDependencies[for_]
		if ranDeps == nil || !ranDeps[dep] {
			panic(fmt.Sprintf("%T didn't run for %#v", dep, for_))
		}
	}
}
