package internal

import (
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

// Base is the base for the linker and analyzer types.
type Base struct {
	Pkg         *file.Package
	Logger      *slog.Logger
	diagnostics diagnostic.List
}

// NewBase creates a new Base.
func NewBase(p *file.Package, logger *slog.Logger) *Base {
	return &Base{
		Pkg:         p,
		Logger:      logger,
		diagnostics: make(diagnostic.List, 0, 32),
	}
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
